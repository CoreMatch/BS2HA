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
- Release the HRPAuth baseline SQL templates (`ha-baseline.sql` and `ygg-baseline.sql`).
- Detect your `bs.sql` and textures directory.
- Guide you through the configuration via a TUI if any parameters are missing.
- Auto-generate a new signing key if you don't provide one.

For advanced CLI usage:

```bash
./bs2ha convert \
  --bs-sql ./bs.sql \
  --bs-textures ./storage/textures \
  --ha-callback-url https://auth.example.com \
  --ha-out-sql ./ha-import.sql \
  --ygg-out-sql ./ygg-import.sql
```

## Docker image

The image is published on Docker Hub: [`limeight/ha-bs2ha`](https://hub.docker.com/r/limeight/ha-bs2ha)
(`linux/amd64` + `linux/arm64` multi-arch; tag `6232db3` alongside `latest`). BS2HA is a
one-shot migration tool and is not part of the platform container orchestration; run it
interactively against your Blessing-Skin-Server data.

Notes:

- copied HA texture files default to a `textures/` directory next to the
  `--ha-out-sql` output when `--ha-textures-out` is omitted
- pass `--unsigned` only for non-production imports without a signing key
- the current BlessingSkin dump format uses `users.nickname` as the canonical source for `HA.users.username`
- password method is detected from SQL; `BCRYPT` hashes are copied directly, other methods become `BS2HA$RESET_REQUIRED`
- generated output is split into `ha-import.sql` for the core database and `ygg-import.sql` for the Yggdrasil database
