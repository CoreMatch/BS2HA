package convert

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lnb/bs2ha/internal/bsdump"
)

const resetRequiredMarker = "BS2HA$RESET_REQUIRED"

type Config struct {
	BSSQLPath             string
	BSTexturesDir         string
	HACallbackURL         string
	HAPrivateKeyPath      string
	HAPrivateKeyContent   string
	GeneratedKeyPath      string // Where to save the generated key
	OutSQLPath            string
	OutReportPath         string
	HATexturesOutDir      string
	BSUsernameColumn      string
	AllowUnsignedTextures bool
}

type Result struct {
	SQL        string
	ReportJSON []byte
}

type Report struct {
	DetectedPasswordMethod string        `json:"detected_password_method"`
	ImportedUsers          int           `json:"imported_users"`
	ImportedProfiles       int           `json:"imported_profiles"`
	PlaceholderProfiles    int           `json:"placeholder_profiles"`
	DroppedPlayers         int           `json:"dropped_players"`
	ImportedTextureProps   int           `json:"imported_texture_properties"`
	UnsignedTextureProps   int           `json:"unsigned_texture_properties"`
	BcryptPasswordCopies   int           `json:"bcrypt_password_copies"`
	ResetRequiredUsers     int           `json:"reset_required_users"`
	Skipped                []ReportIssue `json:"skipped"`
	Warnings               []ReportIssue `json:"warnings"`
}

type ReportIssue struct {
	Kind    string `json:"kind"`
	Table   string `json:"table,omitempty"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message"`
}

type userRecord struct {
	UID           uint64
	UUID          string
	Email         string
	Avatar        string
	Username      string
	Password      string
	IP            string
	Permission    int64
	LastSignAt    string
	RegisterAt    string
	Verified      bool
	RememberToken string
	RegIP         string
}

type profileRecord struct {
	ID        string
	UserID    string
	Name      string
	Model     string
	CreatedAt string
	UpdatedAt string
}

type propertyRecord struct {
	ProfileID  string
	Name       string
	Value      string
	Signature  string
	Unsigned   bool
	SourceHash []string
}

type texturePayload struct {
	Timestamp   int64        `json:"timestamp"`
	ProfileID   string       `json:"profileId"`
	ProfileName string       `json:"profileName"`
	Textures    textureSlots `json:"textures"`
}

type textureSlots struct {
	Skin *textureInfo `json:"SKIN,omitempty"`
	Cape *textureInfo `json:"CAPE,omitempty"`
}

type textureInfo struct {
	URL      string                 `json:"url"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type signer struct {
	key *rsa.PrivateKey
}

func Run(cfg Config, dump *bsdump.Dump) (*Result, error) {
	if dump == nil {
		return nil, fmt.Errorf("nil SQL dump")
	}

	callbackURL, err := url.Parse(cfg.HACallbackURL)
	if err != nil || callbackURL.Scheme == "" || callbackURL.Host == "" {
		return nil, fmt.Errorf("invalid HA callback URL %q", cfg.HACallbackURL)
	}

	method, err := detectPasswordMethod(dump)
	if err != nil {
		return nil, err
	}

	sign, err := loadSigner(cfg)
	if err != nil {
		return nil, err
	}

	usersTable := dump.Tables["users"]
	if len(usersTable) == 0 {
		return nil, fmt.Errorf("BlessingSkin users table is empty or missing")
	}

	playersByUID := map[uint64][]bsdump.Row{}
	for _, row := range dump.Tables["players"] {
		uid, ok := parseUint(row, "uid")
		if !ok {
			continue
		}
		playersByUID[uid] = append(playersByUID[uid], row)
	}

	texturesByID := map[int64]bsdump.Row{}
	for _, row := range dump.Tables["textures"] {
		tid, ok := parseInt(row, "tid")
		if !ok {
			continue
		}
		texturesByID[tid] = row
	}

	if cfg.BSTexturesDir != "" {
		if err := os.MkdirAll(cfg.HATexturesOutDir, 0o755); err != nil {
			return nil, fmt.Errorf("create texture output directory: %w", err)
		}
	}

	report := Report{
		DetectedPasswordMethod: method,
	}

	var users []userRecord
	var profiles []profileRecord
	var props []propertyRecord
	usedProfileNames := map[string]string{}

	for _, row := range usersTable {
		user, ok := mapUser(row, cfg.BSUsernameColumn, method, &report)
		if !ok {
			continue
		}

		userPlayers := playersByUID[user.UID]
		selectedPlayer, dropped, placeholder, ok := selectPlayer(user.Username, userPlayers)
		report.DroppedPlayers += dropped
		if !ok {
			report.Skipped = append(report.Skipped, ReportIssue{
				Kind:    "user",
				Table:   "users",
				ID:      strconv.FormatUint(user.UID, 10),
				Message: "failed to select or create profile for user",
			})
			continue
		}

		profile := profileFromSelection(user, selectedPlayer, placeholder)
		if !placeholder && selectedPlayer != nil {
			if tid, ok := parseInt(selectedPlayer, "tid_skin"); ok && tid > 0 {
				if tex, exists := texturesByID[tid]; exists {
					profile.Model = modelFromTexture(tex)
				}
			}
		}
		if owner, exists := usedProfileNames[profile.Name]; exists {
			report.Skipped = append(report.Skipped, ReportIssue{
				Kind:    "profile_conflict",
				Table:   "profiles",
				ID:      profile.Name,
				Message: fmt.Sprintf("profile name already used by user %s", owner),
			})
			continue
		}
		usedProfileNames[profile.Name] = strconv.FormatUint(user.UID, 10)

		if placeholder {
			report.PlaceholderProfiles++
		}

		prop, warningIssues, err := buildTextureProperty(cfg, sign, user, profile, selectedPlayer, texturesByID)
		if err != nil {
			report.Skipped = append(report.Skipped, ReportIssue{
				Kind:    "profile_property",
				Table:   "profile_properties",
				ID:      profile.ID,
				Message: err.Error(),
			})
			continue
		}
		report.Warnings = append(report.Warnings, warningIssues...)

		users = append(users, user)
		profiles = append(profiles, profile)
		if prop != nil {
			props = append(props, *prop)
			report.ImportedTextureProps++
			if prop.Unsigned {
				report.UnsignedTextureProps++
			}
		}
	}

	report.ImportedUsers = len(users)
	report.ImportedProfiles = len(profiles)

	sqlText, err := renderSQL(users, profiles, props)
	if err != nil {
		return nil, err
	}
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}

	return &Result{SQL: sqlText, ReportJSON: reportJSON}, nil
}

func mapUser(row bsdump.Row, usernameColumn string, passwordMethod string, report *Report) (userRecord, bool) {
	uid, ok := parseUint(row, "uid")
	if !ok {
		report.Skipped = append(report.Skipped, ReportIssue{
			Kind:    "user",
			Table:   "users",
			Message: "missing or invalid uid",
		})
		return userRecord{}, false
	}

	username, ok := row.Get(usernameColumn)
	if !ok || strings.TrimSpace(username) == "" {
		report.Skipped = append(report.Skipped, ReportIssue{
			Kind:    "user",
			Table:   "users",
			ID:      strconv.FormatUint(uid, 10),
			Message: fmt.Sprintf("missing BlessingSkin username column %q", usernameColumn),
		})
		return userRecord{}, false
	}

	password, _ := row.Get("password")
	if strings.EqualFold(passwordMethod, "BCRYPT") {
		report.BcryptPasswordCopies++
	} else {
		password = resetRequiredMarker
		report.ResetRequiredUsers++
	}

	ip, _ := row.Get("ip")
	email, _ := row.Get("email")
	lastSignAt, _ := row.Get("last_sign_at")
	registerAt, _ := row.Get("register_at")
	rememberToken, _ := row.Get("remember_token")
	permission, _ := parseInt(row, "permission")
	verified := parseBool(row, "verified")

	return userRecord{
		UID:           uid,
		UUID:          deriveID("user", strconv.FormatUint(uid, 10)),
		Email:         email,
		Avatar:        "",
		Username:      username,
		Password:      password,
		IP:            ip,
		Permission:    permission,
		LastSignAt:    normalizeDateTime(lastSignAt),
		RegisterAt:    normalizeDateTime(registerAt),
		Verified:      verified,
		RememberToken: rememberToken,
		RegIP:         ip,
	}, true
}

func selectPlayer(username string, players []bsdump.Row) (bsdump.Row, int, bool, bool) {
	switch len(players) {
	case 0:
		return nil, 0, true, true
	case 1:
		return players[0], 0, false, true
	default:
		dropped := len(players)
		for _, player := range players {
			name, _ := player.Get("name")
			if name == "" {
				name, _ = player.Get("player_name")
			}
			if name == username {
				return player, dropped - 1, false, true
			}
		}
		return nil, dropped, true, true
	}
}

func profileFromSelection(user userRecord, player bsdump.Row, placeholder bool) profileRecord {
	name := user.Username
	model := "default"
	timestamp := user.RegisterAt
	if timestamp == "" {
		timestamp = "1970-01-01 00:00:00"
	}

	if !placeholder && player != nil {
		if playerName, ok := player.Get("name"); ok && playerName != "" {
			name = playerName
		} else if playerName, ok := player.Get("player_name"); ok && playerName != "" {
			name = playerName
		}
		if lastModified, ok := player.Get("last_modified"); ok && lastModified != "" {
			timestamp = normalizeDateTime(lastModified)
		}
	}

	return profileRecord{
		ID:        deriveID("profile", strconv.FormatUint(user.UID, 10), name),
		UserID:    user.UUID,
		Name:      name,
		Model:     model,
		CreatedAt: timestamp,
		UpdatedAt: timestamp,
	}
}

func buildTextureProperty(cfg Config, sign signer, user userRecord, profile profileRecord, player bsdump.Row, texturesByID map[int64]bsdump.Row) (*propertyRecord, []ReportIssue, error) {
	if player == nil || cfg.BSTexturesDir == "" {
		return nil, nil, nil
	}

	var warnings []ReportIssue
	slots := textureSlots{}
	var copiedHashes []string

	if tid, ok := parseInt(player, "tid_skin"); ok && tid > 0 {
		if tex, exists := texturesByID[tid]; exists {
			hash, info, warning, err := materializeTexture(cfg, tex, true)
			if err != nil {
				return nil, warnings, err
			}
			if warning != nil {
				warnings = append(warnings, *warning)
			}
			if info != nil {
				slots.Skin = info
				copiedHashes = append(copiedHashes, hash)
			}
		}
	}

	if tid, ok := parseInt(player, "tid_cape"); ok && tid > 0 {
		if tex, exists := texturesByID[tid]; exists {
			hash, info, warning, err := materializeTexture(cfg, tex, false)
			if err != nil {
				return nil, warnings, err
			}
			if warning != nil {
				warnings = append(warnings, *warning)
			}
			if info != nil {
				slots.Cape = info
				copiedHashes = append(copiedHashes, hash)
			}
		}
	}

	if slots.Skin == nil && slots.Cape == nil {
		return nil, warnings, nil
	}

	payload := texturePayload{
		Timestamp:   timestampMillis(profile.UpdatedAt),
		ProfileID:   profile.ID,
		ProfileName: profile.Name,
		Textures:    slots,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, warnings, fmt.Errorf("marshal textures payload for profile %s: %w", profile.ID, err)
	}

	value := base64.StdEncoding.EncodeToString(body)
	signature := ""
	unsigned := false
	if sign.key == nil {
		unsigned = true
	} else {
		signature, err = sign.sign(value)
		if err != nil {
			return nil, warnings, err
		}
	}

	return &propertyRecord{
		ProfileID:  profile.ID,
		Name:       "textures",
		Value:      value,
		Signature:  signature,
		Unsigned:   unsigned,
		SourceHash: copiedHashes,
	}, warnings, nil
}

func materializeTexture(cfg Config, tex bsdump.Row, isSkin bool) (string, *textureInfo, *ReportIssue, error) {
	hash, ok := tex.Get("hash")
	if !ok || hash == "" {
		return "", nil, nil, nil
	}

	source := filepath.Join(cfg.BSTexturesDir, hash)
	if _, err := os.Stat(source); err != nil {
		return hash, nil, &ReportIssue{
			Kind:    "texture_missing",
			Table:   "textures",
			ID:      hash,
			Message: "texture file not found in BlessingSkin textures directory",
		}, nil
	}

	target := filepath.Join(cfg.HATexturesOutDir, hash)
	if err := copyFile(source, target); err != nil {
		return hash, nil, nil, fmt.Errorf("copy texture %s: %w", hash, err)
	}

	textureURL := strings.TrimRight(cfg.HACallbackURL, "/") + "/textures/" + hash
	info := &textureInfo{URL: textureURL}
	if isSkin {
		if model := modelFromTexture(tex); model == "slim" {
			info.Metadata = map[string]interface{}{"model": "slim"}
		}
	}
	return hash, info, nil, nil
}

func loadSigner(cfg Config) (signer, error) {
	if cfg.HAPrivateKeyPath == "" && cfg.HAPrivateKeyContent == "" {
		if cfg.AllowUnsignedTextures {
			return signer{}, nil
		}
		// Auto-generate if not provided and not explicitly unsigned
		return generateAndSaveKey(cfg.GeneratedKeyPath)
	}

	var data []byte
	var err error
	if cfg.HAPrivateKeyContent != "" {
		data = []byte(cfg.HAPrivateKeyContent)
	} else {
		data, err = os.ReadFile(cfg.HAPrivateKeyPath)
		if err != nil {
			return signer{}, fmt.Errorf("read HA private key: %w", err)
		}
	}
	return parseKey(data)
}

func parseKey(data []byte) (signer, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return signer{}, fmt.Errorf("failed to parse PEM block")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return signer{}, fmt.Errorf("parse RSA private key: %w", err)
	}
	return signer{key: key}, nil
}

func generateAndSaveKey(path string) (signer, error) {
	if path == "" {
		path = "hrpauth-generated.key"
	}
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return signer{}, fmt.Errorf("generate RSA key: %w", err)
	}

	data := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	if err := os.WriteFile(path, data, 0600); err != nil {
		return signer{}, fmt.Errorf("save generated key: %w", err)
	}

	fmt.Printf("\n[INFO] Generated new HRPAuth private key and saved to: %s\n", path)
	fmt.Println("[IMPORTANT] Please keep this key safe and configure it in your HRPAuth server.")

	return signer{key: key}, nil
}

func (s signer) sign(value string) (string, error) {
	if s.key == nil {
		return "", nil
	}
	hashed := sha1.Sum([]byte(value))
	sig, err := rsa.SignPKCS1v15(nil, s.key, crypto.SHA1, hashed[:])
	if err != nil {
		return "", fmt.Errorf("sign texture payload: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func detectPasswordMethod(dump *bsdump.Dump) (string, error) {
	known := map[string]struct{}{
		"BCRYPT":            {},
		"MD5":               {},
		"SHA256":            {},
		"SHA512":            {},
		"SALTED2MD5":        {},
		"SALTED2SHA256":     {},
		"SALTED2SHA512":     {},
		"PHP_PASSWORD_HASH": {},
		"ARGON2I":           {},
	}
	keyCandidates := map[string]struct{}{
		"PWD_METHOD":      {},
		"PASSWORD_METHOD": {},
		"SECURE.CIPHER":   {},
		"CIPHER":          {},
	}
	keyColumns := []string{"option_name", "key", "name", "config_key", "setting_key"}
	valueColumns := []string{"option_value", "value", "config_value", "setting_value"}

	for _, rows := range dump.Tables {
		for _, row := range rows {
			for _, valueCol := range []string{"pwd_method", "password_method", "cipher"} {
				if value, ok := row.Get(valueCol); ok {
					upper := strings.ToUpper(strings.TrimSpace(value))
					if _, exists := known[upper]; exists {
						return upper, nil
					}
				}
			}
			for _, keyCol := range keyColumns {
				key, ok := row.Get(keyCol)
				if !ok {
					continue
				}
				if _, exists := keyCandidates[strings.ToUpper(strings.TrimSpace(key))]; !exists {
					continue
				}
				for _, valueCol := range valueColumns {
					value, ok := row.Get(valueCol)
					if !ok {
						continue
					}
					upper := strings.ToUpper(strings.TrimSpace(value))
					if _, exists := known[upper]; exists {
						return upper, nil
					}
				}
			}
		}
	}

	users := dump.Tables["users"]
	if len(users) == 0 {
		return "", fmt.Errorf("could not determine BlessingSkin password method from SQL input")
	}

	seenPassword := false
	allBcrypt := true
	for _, row := range users {
		password, ok := row.Get("password")
		if !ok || strings.TrimSpace(password) == "" {
			continue
		}
		seenPassword = true
		if !looksLikeBcrypt(password) {
			allBcrypt = false
			break
		}
	}

	if seenPassword && allBcrypt {
		return "BCRYPT", nil
	}

	return "NON_BCRYPT_OR_MIXED", nil
}

func looksLikeBcrypt(password string) bool {
	password = strings.TrimSpace(password)
	return strings.HasPrefix(password, "$2y$") || strings.HasPrefix(password, "$2b$") || strings.HasPrefix(password, "$2a$")
}

func renderSQL(users []userRecord, profiles []profileRecord, props []propertyRecord) (string, error) {
	sort.Slice(users, func(i, j int) bool { return users[i].UID < users[j].UID })
	sort.Slice(profiles, func(i, j int) bool { return profiles[i].ID < profiles[j].ID })
	sort.Slice(props, func(i, j int) bool { return props[i].ProfileID < props[j].ProfileID })

	var b strings.Builder
	b.WriteString("BEGIN;\n\n")
	b.WriteString("INSERT INTO `users` (`uid`,`uuid`,`email`,`cbh`,`mojang_uuid`,`avatar`,`password`,`ip`,`permission`,`last_sign_at`,`register_at`,`verified`,`remember_token`,`username`,`mbe`,`regip`,`totp`,`2FA`) VALUES\n")
	for i, user := range users {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString("(")
		b.WriteString(strings.Join([]string{
			strconv.FormatUint(user.UID, 10),
			quoteSQL(user.UUID),
			nullableSQL(user.Email),
			"1",
			"NULL",
			nullableSQL(user.Avatar),
			quoteSQL(user.Password),
			nullableSQL(user.IP),
			strconv.FormatInt(user.Permission, 10),
			nullableSQL(user.LastSignAt),
			nullableSQL(user.RegisterAt),
			boolSQL(user.Verified),
			nullableSQL(user.RememberToken),
			nullableSQL(user.Username),
			"0",
			nullableSQL(user.RegIP),
			"NULL",
			"0",
		}, ","))
		b.WriteString(")")
	}
	b.WriteString(";\n\n")

	b.WriteString("INSERT INTO `profiles` (`id`,`user_id`,`name`,`model`,`created_at`,`updated_at`) VALUES\n")
	for i, profile := range profiles {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString("(")
		b.WriteString(strings.Join([]string{
			quoteSQL(profile.ID),
			quoteSQL(profile.UserID),
			quoteSQL(profile.Name),
			quoteSQL(profile.Model),
			nullableSQL(profile.CreatedAt),
			nullableSQL(profile.UpdatedAt),
		}, ","))
		b.WriteString(")")
	}
	b.WriteString(";\n")

	if len(props) > 0 {
		b.WriteString("\nINSERT INTO `profile_properties` (`profile_id`,`name`,`value`,`signature`) VALUES\n")
		for i, prop := range props {
			if i > 0 {
				b.WriteString(",\n")
			}
			b.WriteString("(")
			b.WriteString(strings.Join([]string{
				quoteSQL(prop.ProfileID),
				quoteSQL(prop.Name),
				quoteSQL(prop.Value),
				nullableSQL(prop.Signature),
			}, ","))
			b.WriteString(")")
		}
		b.WriteString(";\n")
	}

	b.WriteString("\nCOMMIT;\n")
	return b.String(), nil
}

func quoteSQL(value string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "'", "''") + "'"
}

func nullableSQL(value string) string {
	if strings.TrimSpace(value) == "" {
		return "NULL"
	}
	return quoteSQL(value)
}

func boolSQL(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

func parseUint(row bsdump.Row, key string) (uint64, bool) {
	raw, ok := row.Get(key)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	return n, err == nil
}

func parseInt(row bsdump.Row, key string) (int64, bool) {
	raw, ok := row.Get(key)
	if !ok {
		return 0, false
	}
	n, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return n, err == nil
}

func parseBool(row bsdump.Row, key string) bool {
	raw, ok := row.Get(key)
	if !ok {
		return false
	}
	raw = strings.TrimSpace(strings.ToLower(raw))
	return raw == "1" || raw == "true"
}

func normalizeDateTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "0000-00-00 00:00:00") {
		return ""
	}
	return value
}

func timestampMillis(dt string) int64 {
	if dt == "" {
		return 0
	}
	t, err := time.Parse("2006-01-02 15:04:05", dt)
	if err != nil {
		return 0
	}
	return t.UnixMilli()
}

func deriveID(kind string, parts ...string) string {
	sum := sha256.Sum256([]byte(kind + ":" + strings.Join(parts, ":")))
	return hex.EncodeToString(sum[:16])
}

func modelFromTexture(row bsdump.Row) string {
	textureType, _ := row.Get("type")
	if strings.EqualFold(textureType, "alex") {
		return "slim"
	}
	return "default"
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
