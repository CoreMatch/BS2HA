# BS2HA
Transform your Blessing-Skin-Server database to HRPAuth format.

See [docs/design.md](./docs/design.md) for the current migration design.

## Usage

Simply run the tool in your Blessing-Skin-Server root directory:

```bash
go build -o bs2ha ./cmd/bs2ha
./bs2ha
```

The tool will automatically:
- Release the HRPAuth baseline SQL template (`ha-baseline.sql`).
- Detect your `bs.sql` and textures directory.
- Guide you through the configuration via a TUI if any parameters are missing.
- Auto-generate a new signing key if you don't provide one.

For advanced CLI usage:

```bash
./bs2ha convert \
  --bs-sql ./bs.sql \
  --bs-textures ./storage/textures \
  --ha-callback-url https://auth.example.com \
  --out-sql ./ha-import.sql
```

Notes:

- copied HA texture files default to `./out/textures` when `--ha-textures-out` is omitted
- pass `--unsigned` only for non-production imports without a signing key
- the current BlessingSkin dump format uses `users.nickname` as the canonical source for `HA.users.username`
- password method is detected from SQL; `BCRYPT` hashes are copied directly, other methods become `BS2HA$RESET_REQUIRED`
