# ADR-0001: One-Shot Migration CLI, Not a Service

- **Status**: accepted
- **Date**: 2026-11-04

## Context

Operators migrating from BlessingSkin Server to the HA ecosystem need to move
users, players, textures, and skins into HRPAuth-format storage. The choice was
between a long-running sync service and a one-time offline conversion.

## Decision

BS2HA is a **one-shot, offline migration CLI**, not a resident service:

- It reads a BlessingSkin **SQL dump** plus a **textures directory**, converts
  them, and writes output files. It never connects to a live database and never
  writes to a database directly.
- It has **no HTTP server / daemon code**: there is no web framework or
  database driver in its dependencies, and no listener anywhere.
- Delivery forms are a CLI (`bs2ha convert ...`), an interactive TUI when
  required parameters are missing, and a container image for running the same
  one-shot command.

## Consequences

- After a migration the tool is done; the design goal is that the migrated HA
  can run independently without the old BlessingSkin staying online.
- Inputs are plain-text `CREATE TABLE`/`INSERT INTO` dumps (tables
  `users`/`players`/`textures`), so the source must be dumped first.
- Cost/limitation: there is no incremental or live sync; a re-run means
  regenerating the output files. The tool's correctness depends on the dump
  format being one it can parse.

## Evidence

- Positioning: `README.md:L2` ("Transform your Blessing-Skin-Server database
  to HRPAuth format."); `docs/design.md:L5` (offline migration tool),
  `docs/design.md:L12` (migrated HA runs independently, old BS need not stay
  online); `Dockerfile:L12-L13` ("bs2ha is a one-shot migration CLI").
- Inputs: `internal/app/app.go:L64-L65` (default `--bs-textures
  storage/textures`; `--bs-sql` has no flag default); parser accepts only
  `CREATE TABLE`/`INSERT INTO` text
  `internal/bsdump/parse.go::createTablePattern`,
  `internal/bsdump/parse.go::insertPattern`; tables read
  `internal/convert/convert.go:L157-L172`; username from
  `users.nickname` `internal/app/app.go:L76`, `README.md:L36`.
- Outputs are files only: `internal/app/app.go::HAOutSQLPath`,
  `internal/app/app.go::YGGOutSQLPath`, `internal/app/app.go::OutReportPath`
  (write `HAOutSQLPath`, `YGGOutSQLPath`, `OutReportPath` JSON); report
  struct `internal/convert/convert.go::Report`.
- No server code: repo-wide grep for `http.ListenAndServe|gin|echo|net/http`
  returns zero matches; `go.mod:L5-L8` dependencies are only TUI libraries
  (`charmbracelet/huh`, `lipgloss`).
- Delivery forms: CLI flags `internal/app/app.go::runConvert`,
  `README.md:L24-L30`; TUI auto-entry `internal/app/app.go::FillConfig`,
  `internal/tui/tui.go::FillConfig`; Docker multi-stage
  `Dockerfile:L4-L10`, `Dockerfile:L16`, `Dockerfile:L20-L21`.
- Commits: `32c70fb` (initial tool), `436acfc` (modern dump formats),
  `c8bc6a1` (interactive TUI), `0d4eb47` (container image).
- Not verified: actual runtime behavior was not executed (read-only research);
  correctness is asserted from code reading.
