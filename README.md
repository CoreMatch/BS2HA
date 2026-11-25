<div align="center">
  <h1>Armorial BS2HA Migration</h1>
  <p><strong>Move from BlessingSkin with the details that matter.</strong></p>
  <p><a href="./README.zh-CN.md">简体中文</a></p>
</div>

Armorial BS2HA Migration is an offline migration tool that converts BlessingSkin SQL and texture data into Armorial core and Yggdrasil outputs through a guided TUI.

It produces import files and a report for administrators planning a controlled migration.

## Why Armorial BS2HA Migration

- **Guide the conversion.** Walk administrators through missing choices with a TUI.
- **Sign textures securely.** Generate RSA-4096 keys and signed texture output.
- **Preserve usable passwords.** Copy bcrypt hashes without rehashing.
- **Produce both imports.** Emit separate core and Yggdrasil SQL outputs plus a migration report.
- **Migrate offline.** Work directly from SQL dumps and texture files.

## Learn more

- [Migration design](docs/design.md) — local design reference
- [Docker image](https://hub.docker.com/r/limeight/ha-bs2ha) — published container image
- [Armorial Wiki](https://wiki.mcnb.dev/) — ecosystem documentation
- [Armorial API Contracts](https://github.com/CoreMatch/HA-Contract) — ecosystem API contracts

## Migration limits

Sessions, tokens, and OAuth clients are excluded; non-bcrypt passwords require reset. Back up and test first—migration is never lossless.

## License & provenance

[AGPL-3.0](LICENSE). Source: [CoreMatch/BS2HA](https://github.com/CoreMatch/BS2HA).
