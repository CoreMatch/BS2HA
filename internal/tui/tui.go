package tui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lnb/bs2ha/internal/convert"
)

// FillConfig fills the missing configuration values using a TUI.
func FillConfig(cfg *convert.Config) error {
	var (
		bsSQLPath      = cfg.BSSQLPath
		bsTexturesDir  = cfg.BSTexturesDir
		haCallbackURL  = cfg.HACallbackURL
		haPrivateKey   = cfg.HAPrivateKeyPath
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

	var groups []*huh.Group

	// Group 1: Source Info
	var group1Fields []huh.Field
	group1Fields = append(group1Fields, huh.NewNote().
		Title("BlessingSkin to HRPAuth Migration").
		Description("Welcome to the migration tool. Please fill in the missing parameters."))

	if bsSQLPath == "" {
		group1Fields = append(group1Fields, huh.NewInput().
			Title("BlessingSkin SQL Dump Path").
			Description("Path to the .sql file exported from BlessingSkin.").
			Value(&bsSQLPath).
			Validate(func(s string) error {
				if s == "" {
					return errors.New("SQL dump path is required")
				}
				if _, err := os.Stat(s); os.IsNotExist(err) {
					return fmt.Errorf("file %s does not exist", s)
				}
				return nil
			}))
	}

	if bsTexturesDir == "" {
		group1Fields = append(group1Fields, huh.NewInput().
			Title("BlessingSkin Textures Directory").
			Description("Path to the directory containing texture files. Leave empty to skip texture migration.").
			Placeholder("Leave empty to skip").
			Value(&bsTexturesDir).
			Validate(func(s string) error {
				if s != "" {
					if info, err := os.Stat(s); err != nil || !info.IsDir() {
						return fmt.Errorf("directory %s does not exist or is not a directory", s)
					}
				}
				return nil
			}))
	}

	if haCallbackURL == "" {
		group1Fields = append(group1Fields, huh.NewInput().
			Title("HRPAuth Callback URL").
			Description("The base URL for HRPAuth callback (e.g., https://auth.example.com).").
			Value(&haCallbackURL).
			Validate(func(s string) error {
				if s == "" {
					return errors.New("callback URL is required")
				}
				return nil
			}))
	}

	if len(group1Fields) > 1 {
		groups = append(groups, huh.NewGroup(group1Fields...))
	}

	// Group 2: Signing
	var group2Fields []huh.Field
	if !unsigned && haPrivateKey == "" {
		group2Fields = append(group2Fields, huh.NewConfirm().
			Title("Allow Unsigned Textures?").
			Description("If you don't have the Yggdrasil private key yet, you can skip signing.").
			Value(&unsigned))

		group2Fields = append(group2Fields, huh.NewInput().
			Title("HRPAuth Yggdrasil Private Key Path").
			Description("Required for signing texture properties unless 'Unsigned' is enabled.").
			Value(&haPrivateKey).
			Validate(func(s string) error {
				if !unsigned && s == "" {
					return errors.New("private key is required unless unsigned is enabled")
				}
				if s != "" {
					if _, err := os.Stat(s); os.IsNotExist(err) {
						return fmt.Errorf("file %s does not exist", s)
					}
				}
				return nil
			}))
	}

	if len(group2Fields) > 0 {
		groups = append(groups, huh.NewGroup(group2Fields...))
	}

	// Group 3: Output Info
	var group3Fields []huh.Field
	if outSQLPath == "" {
		group3Fields = append(group3Fields, huh.NewInput().
			Title("Output SQL Path").
			Description("Where to save the generated HRPAuth SQL.").
			Value(&outSQLPath).
			Validate(func(s string) error {
				if s == "" {
					return errors.New("output SQL path is required")
				}
				return nil
			}))
	}

	if outReportPath == "" {
		group3Fields = append(group3Fields, huh.NewInput().
			Title("Output Report Path").
			Description("Where to save the migration report JSON.").
			Value(&outReportPath).
			Validate(func(s string) error {
				if s == "" {
					return errors.New("output report path is required")
				}
				return nil
			}))
	}

	if haTexturesOut == "" {
		group3Fields = append(group3Fields, huh.NewInput().
			Title("Output Textures Directory").
			Description("Where to copy the textures. Defaults to 'textures' in the same directory as output SQL.").
			Placeholder("Leave empty for default").
			Value(&haTexturesOut))
	}

	if cfg.BSUsernameColumn == "" {
		group3Fields = append(group3Fields, huh.NewInput().
			Title("BS Username Column").
			Description("The column name in BS users table that stores the canonical username.").
			Value(&usernameColumn).
			Validate(func(s string) error {
				if s == "" {
					return errors.New("username column is required")
				}
				return nil
			}))
	}

	if len(group3Fields) > 0 {
		groups = append(groups, huh.NewGroup(group3Fields...))
	}

	if len(groups) > 0 {
		// Define a high-contrast theme
		theme := huh.ThemeBase()

		// Use standard ANSI colors for maximum compatibility
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

		form := huh.NewForm(groups...).WithTheme(theme)
		err := form.Run()
		if err != nil {
			return err
		}
	}

	// Update config with values from TUI
	cfg.BSSQLPath = bsSQLPath
	cfg.BSTexturesDir = bsTexturesDir
	cfg.HACallbackURL = haCallbackURL
	cfg.HAPrivateKeyPath = haPrivateKey
	cfg.OutSQLPath = outSQLPath
	cfg.OutReportPath = outReportPath
	cfg.HATexturesOutDir = haTexturesOut
	cfg.BSUsernameColumn = usernameColumn
	cfg.AllowUnsignedTextures = unsigned

	// Apply defaults if still empty
	if cfg.HATexturesOutDir == "" {
		cfg.HATexturesOutDir = filepath.Join(filepath.Dir(cfg.OutSQLPath), "textures")
	}

	return nil
}
