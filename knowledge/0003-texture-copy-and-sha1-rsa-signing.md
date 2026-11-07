# ADR-0003: Texture Copy and SHA1-RSA Signing with Auto-Generated Key

- **Status**: accepted

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
  `docs/design.md:L146-L153 fp=feb21e51 ~ BS 的 `textures.hash` 与材质文件名天然对应，一期直接复用该哈希。` (reuse hash, URL `/textures/{hash}`).
- Default directory handling: `internal/app/app.go::runConvert` (default
  `storage/textures`), `internal/app/app.go::runConvert` (skip if missing);
  TUI candidates `internal/tui/tui.go::detectBSTextures`.
- Key auto-generation: `internal/convert/convert.go::loadSigner`,
  `internal/convert/convert.go::generateAndSaveKey` (RSA-4096
  → `hrpauth-generated.key`), `internal/app/app.go::runConvert`; TUI option
  "Auto-generate New Key (Recommended)" `internal/tui/tui.go::FillConfig`.
- Signing scheme: `internal/convert/convert.go::sign` (`sign()` SHA1+PKCS1v15
  over base64 payload), written to `profile_properties.signature`
  `internal/convert/convert.go::renderYGGSQL`.
- Consumer match: HRPAuth-Yggdrasil-API
  `/home/lnb/Desktop/HA/HRPAuth-Yggdrasil-API/services/texture_service.go::SignTextureValue`
  (`SignTextureValue`, SHA1+RSA, reads `Yggdrasil.Server.SignaturePrivateKey`);
  `docs/design.md:L182-L184 fp=cbbb705e ~ 为了让迁移后的 Yggdrasil 资料立即可用，迁移工具应该使用 HA 私钥对 `textur` ("使用 HA 私钥对 textures 属性签名").
- Not verified: how the generated key is loaded on the HRPAuth/Yggdrasil side in
  a real deployment (`SignaturePrivateKey` is confirmed in the Yggdrasil service
  config, but the hand-off is an operational step outside this repo).
