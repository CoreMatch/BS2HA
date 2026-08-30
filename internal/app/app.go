package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lnb/bs2ha/internal/assets"
	"github.com/lnb/bs2ha/internal/bsdump"
	"github.com/lnb/bs2ha/internal/convert"
	"github.com/lnb/bs2ha/internal/tui"
)

func Run(args []string) error {
	// Default to "convert" command if no arguments are provided
	if len(args) == 0 {
		if err := releaseBaselineSQL(); err != nil {
			fmt.Printf("Warning: failed to release baseline SQL: %v\n", err)
		}
		return runConvert([]string{})
	}

	switch args[0] {
	case "convert":
		// Release baseline SQL early
		if err := releaseBaselineSQL(); err != nil {
			fmt.Printf("Warning: failed to release baseline SQL: %v\n", err)
		}
		return runConvert(args[1:])
	case "-h", "--help", "help":
		return usageError()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usageError() error {
	return errors.New("usage: bs2ha convert --bs-sql <path> --ha-callback-url <url> --out-sql <path> --out-report <path> [--bs-textures <path>] [--ha-private-key <path>] [--ha-private-key-content <pem>] [--ha-textures-out <path>] [--bs-username-column <name>] [--unsigned]")
}

func releaseBaselineSQL() error {
	const filename = "ha-baseline.sql"
	if _, err := os.Stat(filename); err == nil {
		return nil // Already exists
	}
	return os.WriteFile(filename, assets.HABaselineSQL, 0644)
}

func runConvert(args []string) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := convert.Config{}
	fs.StringVar(&cfg.BSSQLPath, "bs-sql", "", "path to BlessingSkin SQL dump")
	fs.StringVar(&cfg.BSTexturesDir, "bs-textures", "storage/textures", "path to BlessingSkin textures directory")
	fs.StringVar(&cfg.HACallbackURL, "ha-callback-url", "", "HRPAuth callback base URL")
	fs.StringVar(&cfg.HAPrivateKeyPath, "ha-private-key", "", "path to HRPAuth Yggdrasil private key")
	fs.StringVar(&cfg.HAPrivateKeyContent, "ha-private-key-content", "", "raw HRPAuth Yggdrasil private key PEM content")
	fs.StringVar(&cfg.HABaselineSQLPath, "ha-baseline-sql", "", "path to HRPAuth baseline SQL template")
	fs.StringVar(&cfg.GeneratedKeyPath, "ha-gen-key-path", "hrpauth-generated.key", "where to save the auto-generated key")
	fs.StringVar(&cfg.OutSQLPath, "out-sql", "", "path to generated HRPAuth SQL file")
	fs.StringVar(&cfg.OutReportPath, "out-report", "", "path to generated migration report JSON")
	fs.StringVar(&cfg.HATexturesOutDir, "ha-textures-out", "", "path to copied HRPAuth textures output directory")
	fs.StringVar(&cfg.BSUsernameColumn, "bs-username-column", "nickname", "BlessingSkin canonical username column name in users table")
	fs.BoolVar(&cfg.AllowUnsignedTextures, "unsigned", false, "allow unsigned texture properties when no private key is provided")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Check if any required parameters are missing
	isMissingRequired := cfg.BSSQLPath == "" ||
		cfg.HACallbackURL == "" ||
		cfg.OutSQLPath == "" ||
		cfg.OutReportPath == ""

	if isMissingRequired {
		fmt.Println("Missing required parameters. Launching TUI...")
		if err := tui.FillConfig(&cfg); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
	}

	if cfg.HATexturesOutDir == "" {
		cfg.HATexturesOutDir = filepath.Join(filepath.Dir(cfg.OutSQLPath), "textures")
	}
	if cfg.BSUsernameColumn == "" {
		cfg.BSUsernameColumn = "nickname"
	}

	// If using default textures dir, check if it exists
	if cfg.BSTexturesDir == "storage/textures" {
		if _, err := os.Stat(cfg.BSTexturesDir); os.IsNotExist(err) {
			cfg.BSTexturesDir = ""
		}
	}

	dump, err := bsdump.ParseFile(cfg.BSSQLPath)
	if err != nil {
		return fmt.Errorf("parse BlessingSkin SQL dump: %w", err)
	}

	result, err := convert.Run(cfg, dump)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cfg.OutSQLPath), 0o755); err != nil {
		return fmt.Errorf("create SQL output directory: %w", err)
	}
	if err := os.WriteFile(cfg.OutSQLPath, []byte(result.SQL), 0o644); err != nil {
		return fmt.Errorf("write SQL output: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.OutReportPath), 0o755); err != nil {
		return fmt.Errorf("create report output directory: %w", err)
	}
	if err := os.WriteFile(cfg.OutReportPath, result.ReportJSON, 0o644); err != nil {
		return fmt.Errorf("write report output: %w", err)
	}

	return nil
}
