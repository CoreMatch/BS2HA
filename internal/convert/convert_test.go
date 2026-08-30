package convert

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lnb/bs2ha/internal/bsdump"
)

func TestRunCopiesBcryptAndKeepsMatchingPlayer(t *testing.T) {
	tmpDir := t.TempDir()
	bsTextures := filepath.Join(tmpDir, "bs-textures")
	if err := os.MkdirAll(bsTextures, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bsTextures, "skinhash"), []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}

	sql := `
CREATE TABLE ` + "`options`" + ` (
  ` + "`option_name`" + ` varchar(255) NOT NULL,
  ` + "`option_value`" + ` varchar(255) NOT NULL
);
CREATE TABLE ` + "`users`" + ` (
  ` + "`uid`" + ` int NOT NULL,
  ` + "`email`" + ` varchar(255) DEFAULT NULL,
  ` + "`nickname`" + ` varchar(255) DEFAULT NULL,
  ` + "`password`" + ` varchar(255) NOT NULL,
  ` + "`ip`" + ` varchar(255) DEFAULT NULL,
  ` + "`permission`" + ` int DEFAULT 0,
  ` + "`last_sign_at`" + ` datetime DEFAULT NULL,
  ` + "`register_at`" + ` datetime DEFAULT NULL,
  ` + "`verified`" + ` tinyint(1) DEFAULT 0,
  ` + "`remember_token`" + ` varchar(100) DEFAULT NULL
);
CREATE TABLE ` + "`players`" + ` (
  ` + "`pid`" + ` int NOT NULL,
  ` + "`uid`" + ` int NOT NULL,
  ` + "`name`" + ` varchar(255) NOT NULL,
  ` + "`tid_skin`" + ` int DEFAULT 0,
  ` + "`tid_cape`" + ` int DEFAULT 0,
  ` + "`last_modified`" + ` datetime DEFAULT NULL
);
CREATE TABLE ` + "`textures`" + ` (
  ` + "`tid`" + ` int NOT NULL,
  ` + "`type`" + ` varchar(32) NOT NULL,
  ` + "`hash`" + ` varchar(64) NOT NULL
);
INSERT INTO ` + "`options`" + ` (` + "`option_name`" + `,` + "`option_value`" + `) VALUES ('pwd_method','BCRYPT');
INSERT INTO ` + "`users`" + ` (` + "`uid`" + `,` + "`email`" + `,` + "`nickname`" + `,` + "`password`" + `,` + "`ip`" + `,` + "`permission`" + `,` + "`last_sign_at`" + `,` + "`register_at`" + `,` + "`verified`" + `,` + "`remember_token`" + `) VALUES
(1,'alice@example.com','Alice','$2y$10$keepme','127.0.0.1',1,'2026-08-01 00:00:00','2026-07-01 00:00:00',1,'remember');
INSERT INTO ` + "`players`" + ` (` + "`pid`" + `,` + "`uid`" + `,` + "`name`" + `,` + "`tid_skin`" + `,` + "`tid_cape`" + `,` + "`last_modified`" + `) VALUES
(1,1,'NotAlice',1,0,'2026-08-02 00:00:00'),
(2,1,'Alice',1,0,'2026-08-03 00:00:00');
INSERT INTO ` + "`textures`" + ` (` + "`tid`" + `,` + "`type`" + `,` + "`hash`" + `) VALUES (1,'alex','skinhash');
`

	dump, err := bsdump.Parse(sql)
	if err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		BSTexturesDir:         bsTextures,
		HACallbackURL:         "https://auth.example.com",
		HATexturesOutDir:      filepath.Join(tmpDir, "ha-textures"),
		BSUsernameColumn:      "nickname",
		AllowUnsignedTextures: true,
	}
	result, err := Run(cfg, dump)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(result.SQL, "'$2y$10$keepme'") {
		t.Fatalf("expected bcrypt hash to be copied into SQL:\n%s", result.SQL)
	}
	if !strings.Contains(result.SQL, "'Alice','slim'") {
		t.Fatalf("expected matching player profile to be kept with slim model:\n%s", result.SQL)
	}
	if _, err := os.Stat(filepath.Join(cfg.HATexturesOutDir, "skinhash")); err != nil {
		t.Fatalf("expected copied texture file: %v", err)
	}

	var report Report
	if err := json.Unmarshal(result.ReportJSON, &report); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if report.BcryptPasswordCopies != 1 {
		t.Fatalf("BcryptPasswordCopies = %d, want 1", report.BcryptPasswordCopies)
	}
	if report.ResetRequiredUsers != 0 {
		t.Fatalf("ResetRequiredUsers = %d, want 0", report.ResetRequiredUsers)
	}
	if report.DroppedPlayers != 1 {
		t.Fatalf("DroppedPlayers = %d, want 1", report.DroppedPlayers)
	}
}

func TestRunMarksNonBcryptPasswordsForReset(t *testing.T) {
	sql := `
CREATE TABLE ` + "`options`" + ` (
  ` + "`option_name`" + ` varchar(255) NOT NULL,
  ` + "`option_value`" + ` varchar(255) NOT NULL
);
CREATE TABLE ` + "`users`" + ` (
  ` + "`uid`" + ` int NOT NULL,
  ` + "`email`" + ` varchar(255) DEFAULT NULL,
  ` + "`nickname`" + ` varchar(255) DEFAULT NULL,
  ` + "`password`" + ` varchar(255) NOT NULL,
  ` + "`ip`" + ` varchar(255) DEFAULT NULL,
  ` + "`permission`" + ` int DEFAULT 0,
  ` + "`last_sign_at`" + ` datetime DEFAULT NULL,
  ` + "`register_at`" + ` datetime DEFAULT NULL,
  ` + "`verified`" + ` tinyint(1) DEFAULT 0,
  ` + "`remember_token`" + ` varchar(100) DEFAULT NULL
);
INSERT INTO ` + "`options`" + ` (` + "`option_name`" + `,` + "`option_value`" + `) VALUES ('pwd_method','SHA256');
INSERT INTO ` + "`users`" + ` (` + "`uid`" + `,` + "`email`" + `,` + "`nickname`" + `,` + "`password`" + `,` + "`ip`" + `,` + "`permission`" + `,` + "`last_sign_at`" + `,` + "`register_at`" + `,` + "`verified`" + `,` + "`remember_token`" + `) VALUES
(1,'alice@example.com','Alice','oldhash','127.0.0.1',0,'2026-08-01 00:00:00','2026-07-01 00:00:00',0,NULL);
`

	dump, err := bsdump.Parse(sql)
	if err != nil {
		t.Fatal(err)
	}

	cfg := Config{
		BSTexturesDir:         t.TempDir(),
		HACallbackURL:         "https://auth.example.com",
		HATexturesOutDir:      filepath.Join(t.TempDir(), "ha-textures"),
		BSUsernameColumn:      "nickname",
		AllowUnsignedTextures: true,
	}
	result, err := Run(cfg, dump)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(result.SQL, "'"+resetRequiredMarker+"'") {
		t.Fatalf("expected reset marker in SQL:\n%s\nreport:\n%s", result.SQL, string(result.ReportJSON))
	}
}
