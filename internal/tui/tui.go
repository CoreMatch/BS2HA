package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lnb/bs2ha/internal/convert"
)

func detectBSTextures() string {
	paths := []string{
		"storage/app/public/textures",
		"textures",
		"blessing-skin-server/storage/app/public/textures",
	}
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if info, err := os.Stat(abs); err == nil && info.IsDir() {
			return abs
		}
	}
	return ""
}

func detectHAPrivateKey() string {
	paths := []string{
		"yggdrasil.key",
		"private.key",
		"blessing-skin-server/yggdrasil.key",
	}
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			return abs
		}
	}
	return ""
}

// FillConfig fills the missing configuration values using a TUI.
func FillConfig(cfg *convert.Config) error {
	var (
		bsSQLPath      = cfg.BSSQLPath
		bsTexturesDir  = cfg.BSTexturesDir
		haCallbackURL  = cfg.HACallbackURL
		haPrivateKey   = cfg.HAPrivateKeyPath
		haKeyContent   = cfg.HAPrivateKeyContent
		outSQLPath     = cfg.OutSQLPath
		outReportPath  = cfg.OutReportPath
		haTexturesOut  = cfg.HATexturesOutDir
		usernameColumn = cfg.BSUsernameColumn
		unsigned       = cfg.AllowUnsignedTextures
	)

	// Set defaults if empty
	if usernameColumn == "" {
		usernameColumn = "nickname"
	}

	// Define a high-contrast theme
	theme := huh.ThemeBase()
	white := lipgloss.Color("15")
	cyan := lipgloss.Color("6")
	red := lipgloss.Color("9")
	lightGray := lipgloss.Color("7")

	theme.Focused.Title = theme.Focused.Title.Foreground(cyan).Bold(true)
	theme.Focused.Description = theme.Focused.Description.Foreground(lightGray)
	theme.Focused.TextInput.Text = theme.Focused.TextInput.Text.Foreground(white)
	theme.Focused.TextInput.Prompt = theme.Focused.TextInput.Prompt.Foreground(cyan)
	theme.Focused.TextInput.Cursor = theme.Focused.TextInput.Cursor.Foreground(cyan)
	theme.Focused.TextInput.Placeholder = theme.Focused.TextInput.Placeholder.Foreground(lightGray)
	theme.Blurred.Title = theme.Blurred.Title.Foreground(lightGray)
	theme.Blurred.Description = theme.Blurred.Description.Foreground(lightGray)
	theme.Blurred.TextInput.Text = theme.Blurred.TextInput.Text.Foreground(lightGray)
	theme.Blurred.TextInput.Prompt = theme.Blurred.TextInput.Prompt.Foreground(lightGray)
	theme.Focused.ErrorMessage = theme.Focused.ErrorMessage.Foreground(red)

	// 1. SQL Dump
	if bsSQLPath == "" {
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewNote().
					Title("BlessingSkin to HRPAuth Migration").
					Description("Welcome to the migration tool. Please fill in the missing parameters."),
				huh.NewInput().
					Title("BlessingSkin SQL Dump Path").
					Description("Path to the .sql file. Leave empty for default (bs.sql).").
					Placeholder("bs.sql").
					Value(&bsSQLPath).
					Validate(func(s string) error {
						if s == "" {
							return nil // Fallback handled below
						}
						if _, err := os.Stat(s); os.IsNotExist(err) {
							return fmt.Errorf("file %s not found", s)
						}
						return nil
					}),
			),
		).WithTheme(theme).Run()
		if err != nil {
			return err
		}
		if bsSQLPath == "" {
			bsSQLPath = "bs.sql"
			if _, err := os.Stat(bsSQLPath); os.IsNotExist(err) {
				return fmt.Errorf("default SQL dump %s not found", bsSQLPath)
			}
		}
	}

	// 2. Textures
	if bsTexturesDir == "" {
		detected := detectBSTextures()
		useDetected := true
		if detected != "" {
			err := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Detected textures at: %s. Use it?", detected)).
						Value(&useDetected),
				),
			).WithTheme(theme).Run()
			if err != nil {
				return err
			}
			if useDetected {
				bsTexturesDir = detected
			}
		}

		if !useDetected || detected == "" {
			err := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("BlessingSkin Textures Directory").
						Description("Leave empty to skip texture migration.").
						Placeholder("Leave empty to skip").
						Value(&bsTexturesDir).
						Validate(func(s string) error {
							if s != "" {
								if info, err := os.Stat(s); err != nil || !info.IsDir() {
									return fmt.Errorf("invalid directory")
								}
							}
							return nil
						}),
				),
			).WithTheme(theme).Run()
			if err != nil {
				return err
			}
		}
	}

	// 3. HA URL
	if haCallbackURL == "" {
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("HRPAuth Callback URL").
					Description("Base URL for HRPAuth (e.g., https://auth.example.com).").
					Value(&haCallbackURL).
					Validate(func(s string) error {
						if s == "" {
							return errors.New("required")
						}
						return nil
					}),
			),
		).WithTheme(theme).Run()
		if err != nil {
			return err
		}
	}

	// 4. Private Key
	if !unsigned && haPrivateKey == "" && haKeyContent == "" {
		detected := detectHAPrivateKey()
		useDetected := true
		if detected != "" {
			err := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Detected private key at: %s. Use it?", detected)).
						Value(&useDetected),
				),
			).WithTheme(theme).Run()
			if err != nil {
				return err
			}
			if useDetected {
				haPrivateKey = detected
			}
		}

		if !useDetected || detected == "" {
			err := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().Title("Allow Unsigned Textures?").Value(&unsigned),
				),
			).WithTheme(theme).Run()
			if err != nil {
				return err
			}

			if !unsigned {
				mode := "path"
				err := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title("Private Key Input Method").
							Options(
								huh.NewOption("File Path", "path"),
								huh.NewOption("Paste Content", "content"),
							).
							Value(&mode),
					),
				).WithTheme(theme).Run()
				if err != nil {
					return err
				}

				if mode == "path" {
					err := huh.NewForm(
						huh.NewGroup(
							huh.NewInput().
								Title("Private Key File Path").
								Description("Leave empty for default (yggdrasil.key).").
								Placeholder("yggdrasil.key").
								Value(&haPrivateKey).
								Validate(func(s string) error {
									if s == "" {
										return nil // Fallback handled below
									}
									if _, err := os.Stat(s); os.IsNotExist(err) {
										return fmt.Errorf("file %s not found", s)
									}
									return nil
								}),
						),
					).WithTheme(theme).Run()
					if err != nil {
						return err
					}
					if haPrivateKey == "" {
						haPrivateKey = "yggdrasil.key"
						// Last check for default
						if _, err := os.Stat(haPrivateKey); os.IsNotExist(err) {
							return fmt.Errorf("default private key %s not found", haPrivateKey)
						}
					}
				} else {
					err := huh.NewForm(
						huh.NewGroup(
							huh.NewText().
								Title("Private Key Content").
								Description("Paste your PEM private key content here (Ctrl+D to finish).").
								Value(&haKeyContent).
								Validate(func(s string) error {
									if s == "" {
										return errors.New("required")
									}
									if !strings.Contains(s, "BEGIN") {
										return errors.New("invalid PEM")
									}
									return nil
								}),
						),
					).WithTheme(theme).Run()
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// 5. Outputs
	if outSQLPath == "" || outReportPath == "" {
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Output SQL Path").
					Placeholder("ha-import.sql").
					Value(&outSQLPath),
				huh.NewInput().
					Title("Output Report Path").
					Placeholder("ha-import-report.json").
					Value(&outReportPath),
				huh.NewInput().
					Title("Output Textures Directory").
					Placeholder("Default: textures/").
					Value(&haTexturesOut),
				huh.NewInput().
					Title("BS Username Column").
					Placeholder("nickname").
					Value(&usernameColumn).
					Validate(func(s string) error {
						if s == "" {
							return nil
						} // Use default nickname
						return nil
					}),
			),
		).WithTheme(theme).Run()
		if err != nil {
			return err
		}
	}

	// Set defaults if still empty
	if outSQLPath == "" {
		outSQLPath = "ha-import.sql"
	}
	if outReportPath == "" {
		outReportPath = "ha-import-report.json"
	}
	if usernameColumn == "" {
		usernameColumn = "nickname"
	}

	// Update config
	cfg.BSSQLPath = bsSQLPath
	cfg.BSTexturesDir = bsTexturesDir
	cfg.HACallbackURL = haCallbackURL
	cfg.HAPrivateKeyPath = haPrivateKey
	cfg.HAPrivateKeyContent = haKeyContent
	cfg.OutSQLPath = outSQLPath
	cfg.OutReportPath = outReportPath
	cfg.HATexturesOutDir = haTexturesOut
	cfg.BSUsernameColumn = usernameColumn
	cfg.AllowUnsignedTextures = unsigned

	if cfg.HATexturesOutDir == "" {
		cfg.HATexturesOutDir = filepath.Join(filepath.Dir(cfg.OutSQLPath), "textures")
	}

	return nil
}
