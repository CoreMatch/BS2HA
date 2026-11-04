# ADR-0002: Dual Baseline-Aligned SQL Output (ha + ygg)

- **Status**: accepted
- **Date**: 2026-11-04

## Context

The HA ecosystem splits account data across two databases after the Yggdrasil
split: the HRPAuth **core** database holds `users`, while the
HRPAuth-Yggdrasil-API database holds Minecraft game accounts/profiles. A
BlessingSkin migration must populate both, and the generated SQL must match the
real target schemas or the import will fail.

## Decision

BS2HA emits **two separate SQL files**, each aligned byte-for-byte with the
target service's current baseline schema:

- `ha-import.sql` → the **HRPAuth core** database (`users`).
- `ygg-import.sql` → the **Yggdrasil API** database
  (`accounts`/`profiles`/`profile_properties`).

To guarantee alignment, the tool **embeds the baseline SQL templates** of both
targets and emits them, then appends generated `INSERT` rows whose columns
match those baselines. Password hashes that are bcrypt are passed through;
non-bcrypt hashes are replaced with a `BS2HA$RESET_REQUIRED` marker forcing a
password reset.

## Consequences

- The output is guaranteed to be loadable against the exact baseline schema of
  each target, because the embedded baselines are identical to the targets'
  migration files.
- The split reflects the ecosystem's separation of core auth from
  Minecraft-specific game accounts.
- Cost/limitation: if a target's baseline schema changes, the embedded template
  here becomes stale and must be re-synced manually. Users with non-bcrypt
  passwords cannot keep them and must reset.

## Evidence

- Two output files and their destinations: `README.md:38` ("split into
  `ha-import.sql` for the core database and `ygg-import.sql` for the Yggdrasil
  database"); `docs/design.md:54-66` (ha = Core `users`, ygg = Yggdrasil
  `accounts/profiles/profile_properties`).
- Embedded baselines: `internal/assets/assets.go:7`
  (`//go:embed ha-baseline.sql`), `:12` (`//go:embed ygg-baseline.sql`);
  emitted at runtime `internal/app/app.go:43-57`, `README.md:16`; introduced by
  `50c4ae7`.
- Baseline identity (verified by diff): `internal/assets/ha-baseline.sql` is
  byte-identical to HRPAuth
  `database/migrations/000001_baseline.up.sql` (tables `users`,
  `oauth2_clients`, `oauth2_authorization_codes`, `oauth2_access_tokens`,
  `oauth2_refresh_tokens`, `deleted_accounts`, `webauthn_credentials`);
  `internal/assets/ygg-baseline.sql` is byte-identical to HRPAuth-Yggdrasil-API
  `database/migrations/000001_init.up.sql` (`accounts`, `profiles`,
  `profile_properties`, `sessions`, `tokens`, `profile_keys`).
- Generated INSERTs match columns: `internal/convert/convert.go:690`
  (`INSERT INTO users`, 19 columns), `:749,772,792` (`accounts`/`profiles`/
  `profile_properties`).
- Password mapping: bcrypt passthrough vs reset marker
  `internal/convert/convert.go:26,304-309` (`BS2HA$RESET_REQUIRED`).
- Commits: `50c4ae7` (baseline templates), `507997f` (separate ha/ygg files).
- Not verified: none material; baseline identity was confirmed by direct diff.
