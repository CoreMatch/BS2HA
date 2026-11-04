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

- Positioning: `README.md:2` ("Transform your Blessing-Skin-Server database to
  HRPAuth format."); `docs/design.md:5` (offline migration tool), `:12`
  (migrated HA runs independently, old BS need not stay online);
  `Dockerfile:12-13` ("bs2ha is a one-shot migration CLI").
- Inputs: `internal/app/app.go:64-65` (default `--bs-textures
  storage/textures`; `--bs-sql` has no flag default); parser accepts only
  `CREATE TABLE`/`INSERT INTO` text `internal/bsdump/parse.go:137-139`; tables
  read `internal/convert/convert.go:157,163,172`; username from
  `users.nickname` `internal/app/app.go:76`, `README.md:36`.
- Outputs are files only: `internal/app/app.go:124,131,138` (write
  `HAOutSQLPath`, `YGGOutSQLPath`, `OutReportPath` JSON); report struct
  `internal/convert/convert.go:51-63`.
- No server code: repo-wide grep for `http.ListenAndServe|gin|echo|net/http`
  returns zero matches; `go.mod:5-8` dependencies are only TUI libraries
  (`charmbracelet/huh`, `lipgloss`).
- Delivery forms: CLI flags `internal/app/app.go:59-77`, `README.md:24-30`;
  TUI auto-entry `internal/app/app.go:84-95`, `internal/tui/tui.go:53-54`;
  Docker multi-stage `Dockerfile:4-10,16,20-21`.
- Commits: `32c70fb` (initial tool), `436acfc` (modern dump formats),
  `c8bc6a1` (interactive TUI), `0d4eb47` (container image).
- Not verified: actual runtime behavior was not executed (read-only research);
  correctness is asserted from code reading.
