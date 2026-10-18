# mountOS Admin SDK

SDKs for [mountOS](https://mountos.io), a cross-platform, forkable, time-traveling distributed filesystem over object storage. Automate accounts, users, volumes, forks, quotas, and audit logs from TypeScript, Go, or Rust. The SDKs wrap the appserv provider API (`/api/v1/*`) with ED25519 JWT authentication.

## SDKs

| Language                   | Path    | Package                                      | Reference                    |
| -------------------------- | ------- | -------------------------------------------- | ---------------------------- |
| [TypeScript](ts/README.md) | `ts/`   | `@mountos-io/admin-sdk` (npm)                | [docs/ts.md](docs/ts.md)     |
| [Go](go/README.md)         | `go/`   | `github.com/mountos-io/mountos-admin-sdk/go` | [docs/go.md](docs/go.md)     |
| [Rust](rust/README.md)     | `rust/` | `mountos-admin-sdk` (crates.io)              | [docs/rust.md](docs/rust.md) |

## Examples

See [examples/](examples/) for working Go, TypeScript, and Rust sample code.

## API Reference

- Language-neutral: [api.md](api.md)
- TypeScript SDK: [docs/ts.md](docs/ts.md)
- Go SDK: [docs/go.md](docs/go.md)
- Rust SDK: [docs/rust.md](docs/rust.md)

All references are regenerated from [api.yaml](api.yaml) via `make gen` (or `make docs` for the language-specific docs only).

## Dashboard user (opt-in access scoping)

`SignDashboardUser` signs an operator identity into the `X-MountOS-Dashboard-User` header (HMAC, 10-min TTL). This header is **opt-in** and defines the split between what appserv enforces and what you own:

- **Absent header** → appserv treats the caller as an unrestricted admin.
- **`role: "user"`** → the only role the hub understands. appserv confines the caller to their own `accountId` / `volumeId` and uses `userId` for audit + volume file-operation attribution. Nothing else.
- **Any other `role`** → treated as admin (unrestricted).

appserv does **not** enforce which operations or fields a role may use (the role→capability matrix, forbidden fields, per-endpoint rules). **That access-level policy is yours to own.** The reference admin dashboard implements it in its open-source backend proxy — fork it, replace it, or write your own. appserv only guarantees account/volume scoping + attribution for `role: "user"`, plus a system-admin floor on provider-infrastructure endpoints (vault/license/raft/regions/clusters) that cannot be account-scoped.

**If you send the header, it must be well-formed** — appserv rejects a signed-but-malformed identity (403) instead of falling back to admin. `role` must be non-empty, and `role: "user"` must set both `accountId` and `userId`. An absent header is the only way to be treated as an unrestricted admin; a corrupt one is treated as an attempted bypass. Any other non-empty `role` is accepted as an admin (role names are extensible — you may define your own in the backend policy).

## For AI Agents

[SKILL.md](SKILL.md) is an agent skill covering TS/Go usage, ED25519/JWT auth, and the full resource set. It ships in the npm package (next to `api.md` and `api.yaml`) so coding agents can load accurate, version-matched guidance.

## Versioning and compatibility

Public structs are built by field name. In Go, every struct starts with a blank `_ struct{}` field, so a positional literal does not compile. In Rust, every public struct and enum is `#[non_exhaustive]`: build a struct with `new(required...)`, `Default` and the `with_<field>(value)` setters, and keep a wildcard arm when matching an enum. TypeScript types are structural and need no guard.

A minor version can add fields and enum variants. A major version can remove or rename them.

## 0.x versions

0.x versions are pre-release versions which may have breaking changes intentionally.

## Porting to Other Languages

Use [api.yaml](api.yaml) for generating code for other languages or contact support@mountos.io.

## License

Apache-2.0
