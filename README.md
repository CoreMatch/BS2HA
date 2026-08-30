# BS2HA
Transform your Blessing-Skin-Server database to HRPAuth format.

See [docs/design.md](./docs/design.md) for the current migration design.

## Usage

```bash
go run ./cmd/bs2ha convert \
  --bs-sql ./bs.sql \
  --bs-textures ./textures \
  --ha-callback-url https://auth.example.com \
  --ha-private-key ./private_key.pem \
  --out-sql ./out/ha-import.sql \
  --out-report ./out/ha-import-report.json
```

Notes:

- copied HA texture files default to `./out/textures` when `--ha-textures-out` is omitted
- pass `--unsigned` only for non-production imports without a signing key
- the current BlessingSkin dump format uses `users.nickname` as the canonical source for `HA.users.username`
- password method is detected from SQL; `BCRYPT` hashes are copied directly, other methods become `BS2HA$RESET_REQUIRED`
