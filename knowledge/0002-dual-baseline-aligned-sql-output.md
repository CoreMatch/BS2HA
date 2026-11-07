# ADR-0002: Dual Baseline-Aligned SQL Output (ha + ygg)

- **Status**: accepted

## Context

The HA ecosystem splits account data across two databases: the HRPAuth
**core** database holds `users`, while the HRPAuth-Yggdrasil-API database
holds Minecraft game accounts/profiles. A BlessingSkin migration must populate
both, and the generated SQL must match the real target schemas or the import
will fail.

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

- Two output files and their destinations: `README.md:L38-L38 ~ - password method is detected from SQL; `BCRYPT`` ("split into
  `ha-import.sql` for the core database and `ygg-import.sql` for the Yggdrasil
  database"); `docs/design.md:L54-L66 ~ 1. HA Core 导入 SQL` (ha = Core `users`, ygg = Yggdrasil
  `accounts/profiles/profile_properties`).
- Embedded baselines: `internal/assets/assets.go::HABaselineSQL`
  (`//go:embed ha-baseline.sql`), `internal/assets/assets.go::YGGBaselineSQL`
  (`//go:embed ygg-baseline.sql`); emitted at runtime
  `internal/app/app.go::releaseBaselineSQLs`, `README.md:L16-L16 ~ - Release the HRPAuth baseline SQL templates (`h`.
- Baseline identity (verified by diff): `internal/assets/ha-baseline.sql` is
  byte-identical to HRPAuth
  `/home/lnb/Desktop/HA/HRPAuth/database/migrations/000001_baseline.up.sql`
  (tables `users`, `oauth2_clients`, `oauth2_authorization_codes`,
  `oauth2_access_tokens`, `oauth2_refresh_tokens`, `deleted_accounts`,
  `webauthn_credentials`);
  `internal/assets/ygg-baseline.sql` is byte-identical to HRPAuth-Yggdrasil-API
  `/home/lnb/Desktop/HA/HRPAuth-Yggdrasil-API/database/migrations/000001_init.up.sql`
  (`accounts`, `profiles`, `profile_properties`, `sessions`, `tokens`,
  `profile_keys`).
- Generated INSERTs match columns: `internal/convert/convert.go::renderHASQL`
  (`INSERT INTO users`, 19 columns),
  `internal/convert/convert.go::renderYGGSQL` (`accounts`/`profiles`/
  `profile_properties`).
- Password mapping: bcrypt passthrough vs reset marker
  `internal/convert/convert.go::resetRequiredMarker`,
  `internal/convert/convert.go::mapUser` (`BS2HA$RESET_REQUIRED`).
- Not verified: none material; baseline identity was confirmed by direct diff.
