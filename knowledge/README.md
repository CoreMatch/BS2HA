# Knowledge Base

## What This Is

This knowledge base is **BS2HA's memo to itself**. It serves self-hosting and
maintenance of this repository: what the tool consumes and produces, how it maps
BlessingSkin data onto the HA schema, and how textures and signing are handled.
It is maintained alongside the code and is not meant to make external projects
accommodate this one.

## Order of Authority

This knowledge base **does not define outward behavior**. BS2HA is a migration
tool whose *target* schema is defined elsewhere: the HRPAuth core database
baseline and the HRPAuth-Yggdrasil-API database baseline (and, for ecosystem
outward behavior, [HA-Contract](https://github.com/CoreMatch/HA-Contract)). If
an entry here ever appears to contradict the target schema or a contract,
**the target/contract wins and the entry is stale — fix the entry, never the
target**, and surface the conflict to the owner.

## Format

Each entry is one page in ADR style:

- **Status** (`accepted` / `superseded` / ...)
- **Context** — the problem and constraints
- **Decision** — what was chosen
- **Consequences** — including costs and limitations
- **Evidence** — `file:line`, protocol entries. Every piece of
  evidence must have been actually checked; anything that could not be verified
  is marked "not verified". Do not fabricate.

## Entries

| # | Title | Status |
|---|-------|--------|
| [0001](0001-one-shot-migration-cli-not-a-service.md) | One-Shot Migration CLI, Not a Service | accepted |
| [0002](0002-dual-baseline-aligned-sql-output.md) | Dual Baseline-Aligned SQL Output (ha + ygg) | accepted |
| [0003](0003-texture-copy-and-sha1-rsa-signing.md) | Texture Copy and SHA1-RSA Signing with Auto-Generated Key | accepted |
