# mountOS Admin SDK

Provider-facing SDK for the mountOS Admin API. Wraps the appserv provider API (`/api/v1/*`) with ED25519 JWT authentication.

## SDKs

| Language                   | Path  | Package                                             | Reference                        |
| -------------------------- | ----- | --------------------------------------------------- | -------------------------------- |
| [TypeScript](ts/README.md) | `ts/` | `@mountos-io/admin-sdk` (npm)                        | [docs/ts.md](docs/ts.md)         |
| [Go](go/README.md)         | `go/` | `github.com/mountos-io/mountos-admin-sdk/go`        | [docs/go.md](docs/go.md)         |

## Examples

See [examples/](examples/) for working Go and TypeScript sample code.

## API Reference

- Language-neutral: [api.md](api.md)
- TypeScript SDK: [docs/ts.md](docs/ts.md)
- Go SDK: [docs/go.md](docs/go.md)

All references are regenerated from [api.yaml](api.yaml) via `make gen` (or `make docs` for the language-specific docs only).

## For AI Agents

[SKILL.md](SKILL.md) is an agent skill covering TS/Go usage, ED25519/JWT auth, and the full resource set. It ships in the npm package (next to `api.md` and `api.yaml`) so coding agents can load accurate, version-matched guidance.

## Porting to Other Languages

Use [api.yaml](api.yaml) for generating code for other languages or contact support@mountos.app.

## License

Apache-2.0
