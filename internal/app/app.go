package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lnb/bs2ha/internal/bsdump"
	"github.com/lnb/bs2ha/internal/convert"
)

func Run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "convert":
		return runConvert(args[1:])
	case "-h", "--help", "help":
		return usageError()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usageError() error {
	return errors.New("usage: bs2ha convert --bs-sql <path> --bs-textures <path> --ha-callback-url <url> --out-sql <path> --out-report <path> [--ha-private-key <path>] [--ha-textures-out <path>] [--bs-username-column <name>] [--unsigned]")
}

func runConvert(args []string) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := convert.Config{}
	fs.StringVar(&cfg.BSSQLPath, "bs-sql", "", "path to BlessingSkin SQL dump")
	fs.StringVar(&cfg.BSTexturesDir, "bs-textures", "", "path to BlessingSkin textures directory")
	fs.StringVar(&cfg.HACallbackURL, "ha-callback-url", "", "HRPAuth callback base URL")
	fs.StringVar(&cfg.HAPrivateKeyPath, "ha-private-key", "", "path to HRPAuth Yggdrasil private key")
	fs.StringVar(&cfg.OutSQLPath, "out-sql", "", "path to generated HRPAuth SQL file")
	fs.StringVar(&cfg.OutReportPath, "out-report", "", "path to generated migration report JSON")
	fs.StringVar(&cfg.HATexturesOutDir, "ha-textures-out", "", "path to copied HRPAuth textures output directory")
	fs.StringVar(&cfg.BSUsernameColumn, "bs-username-column", "nickname", "BlessingSkin canonical username column name in users table")
	fs.BoolVar(&cfg.AllowUnsignedTextures, "unsigned", false, "allow unsigned texture properties when no private key is provided")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cfg.BSSQLPath == "" || cfg.BSTexturesDir == "" || cfg.HACallbackURL == "" || cfg.OutSQLPath == "" || cfg.OutReportPath == "" {
		return usageError()
	}
	if cfg.HATexturesOutDir == "" {
		cfg.HATexturesOutDir = filepath.Join(filepath.Dir(cfg.OutSQLPath), "textures")
	}
	if cfg.BSUsernameColumn == "" {
		cfg.BSUsernameColumn = "nickname"
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
