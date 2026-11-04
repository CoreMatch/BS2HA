# ADR-0003: Texture Copy and SHA1-RSA Signing with Auto-Generated Key

- **Status**: accepted
- **Date**: 2026-11-04

## Context

BlessingSkin stores skins/capes as texture files addressed by hash, referenced
by player records. The HA ecosystem serves textures over a callback URL and, for
Yggdrasil profile properties, signs the texture payload with the server's RSA
private key. A migration must move the files and produce valid signatures, or
clients will reject the textures.

## Decision

BS2HA migrates textures by **content hash** and signs the Yggdrasil texture
payload with an RSA key:

- Each BlessingSkin texture is located at `<BSTexturesDir>/<hash>` and copied to
  `<HATexturesOutDir>/<hash>`; the payload URL becomes
  `<HACallbackURL>/textures/<hash>`. Reusing the hash keeps deduplication.
- If no private key is supplied (and not `--unsigned`), the tool
  **auto-generates an RSA-4096 key** and saves it as
  `hrpauth-generated.key`.
- Signing uses **SHA1 + PKCS1v15 over the base64 texture payload**, written to
  `profile_properties.signature`. This matches the consumer: HRPAuth-Yggdrasil-API's
  `SignTextureValue` uses the same SHA1+RSA scheme reading
  `Yggdrasil.Server.SignaturePrivateKey`.

## Consequences

- Migrated profiles carry signatures that the live Yggdrasil service can
  reproduce only if it is configured with the same private key generated here.
- The default textures directory is `storage/textures`; if it is missing the
  step is skipped rather than failing.
- Cost/limitation: the generated private key must be handed to the Yggdrasil
  deployment out-of-band; the tool does not configure the target service. The
  signing key is for **texture property signatures, not JWTs**.

## Evidence

- Texture copy and URL: `internal/convert/convert.go::materializeTexture`
  (copy by hash; `payload URL = HACallbackURL + "/textures/" + hash`);
  `docs/design.md:L146-L153` (reuse hash, URL `/textures/{hash}`).
- Default directory handling: `internal/app/app.go:L65` (default
  `storage/textures`), `internal/app/app.go:L105-L109` (skip if missing);
  TUI candidates `internal/tui/tui.go::detectBSTextures`; commit `2320cb8`.
- Key auto-generation: `internal/convert/convert.go::loadSigner`,
  `internal/convert/convert.go::generateAndSaveKey` (RSA-4096
  → `hrpauth-generated.key`), `internal/app/app.go:L71`; TUI option
  "Auto-generate New Key (Recommended)" `internal/tui/tui.go:L225`; commit
  `7bb6779`.
- Signing scheme: `internal/convert/convert.go::sign` (`sign()` SHA1+PKCS1v15
  over base64 payload), written to `profile_properties.signature`
  `internal/convert/convert.go::renderYGGSQL`.
- Consumer match: HRPAuth-Yggdrasil-API
  `/home/lnb/Desktop/HA/HRPAuth-Yggdrasil-API/services/texture_service.go::SignTextureValue`
  (`SignTextureValue`, SHA1+RSA, reads `Yggdrasil.Server.SignaturePrivateKey`);
  `docs/design.md:L182-L184` ("使用 HA 私钥对 textures 属性签名").
- Commits: `2320cb8` (default texture dir), `7bb6779` (auto key generation).
- Not verified: how the generated key is loaded on the HRPAuth/Yggdrasil side in
  a real deployment (`SignaturePrivateKey` is confirmed in the Yggdrasil service
  config, but the hand-off is an operational step outside this repo).
