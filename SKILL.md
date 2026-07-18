---
name: mountos-admin-sdk
description: Integrate the mountOS Admin (provider) API into TypeScript, Go, or Rust applications using @mountos-io/admin-sdk, github.com/mountos-io/mountos-admin-sdk/go, or the mountos-admin-sdk crate. Use when a task involves calling the mountOS provider API (managing accounts, users, regions, storages, volumes, quotas, volume API keys, audit logs, service nodes, client sessions, licenses, alerts, or vault) or when wiring up ED25519/JWT auth against appserv. Also use when porting the API to another language from api.yaml.
---

# mountOS Admin SDK

Provider-facing SDK for the mountOS Admin API (`/api/v1/*` on appserv), with ED25519
JWT authentication. Three first-party SDKs share one spec: **TypeScript** (`ts/`,
`@mountos-io/admin-sdk`), **Go** (`go/`, `github.com/mountos-io/mountos-admin-sdk/go`),
and **Rust** (`rust/`, `mountos-admin-sdk`).

Repo: https://github.com/mountos-io/mountos-admin-sdk

## When to use which entry point

- **Server-side TS/Go/Rust service** holding a provider private key → use the SDK's JWT
  client. Tokens are signed locally with ED25519, cached ~55 min (1h TTL, 5 min refresh margin).
- **Browser / cookie-or-session app** (no private key in the client) → use the TS
  `createClient(request)` form with your own transport; auth is handled by your backend.
- **Another language** → generate from [api.yaml](api.yaml); see [api.md](api.md).

## Reference map

- [api.md](api.md) - language-neutral REST reference: every endpoint, request/response
  shape, query params, error codes, and the JWT claim contract. Start here for the wire
  protocol or non-TS/Go/Rust callers.
- [docs/ts.md](docs/ts.md) / [docs/go.md](docs/go.md) / [docs/rust.md](docs/rust.md) - full
  per-language method/type reference.
- [ts/README.md](ts/README.md) / [go/README.md](go/README.md) / [rust/README.md](rust/README.md) - install + quickstart.
- Auth internals (how the JWT is built) - read the source for exact claims/encoding:
  - TS: https://github.com/mountos-io/mountos-admin-sdk/blob/main/ts/src/auth.ts
  - Go: https://github.com/mountos-io/mountos-admin-sdk/blob/main/go/auth.go
  - Rust: https://github.com/mountos-io/mountos-admin-sdk/blob/main/rust/src/auth.rs

## TypeScript usage

ESM-only, Node 18+/Bun/Deno. `npm install @mountos-io/admin-sdk`.

Server-side with JWT:

```typescript
import { createServerClient } from "@mountos-io/admin-sdk";

const client = createServerClient({
  baseUrl: "https://appserv.example.com",
  privateKey: "base64-ed25519-key", // 32-byte seed or 64-byte seed+pubkey
});

const { id } = await client.accounts.create({ name: "Acme" });
const { items, pagination } = await client.accounts.list({ page: 1, limit: 10 });
await client.volumes.updateQuota(volId, { quotaLimit: 1073741824 });
const keys = await client.volumes.generateAPIKeys(volId, { userId: 1 });
const logs = await client.auditLogs.list({ accountId: 1, cursor: 0, limit: 20 });
```

Browser / custom transport (auth handled by your backend):

```typescript
import { createClient, type RequestFn } from "@mountos-io/admin-sdk";

const request: RequestFn = async (method, path, body, signal) => {
  const res = await fetch(`https://api.example.com${path}`, {
    method, credentials: "include",
    headers: body !== undefined ? { "Content-Type": "application/json" } : {},
    body: body !== undefined ? JSON.stringify(body) : undefined, signal,
  });
  const json = await res.json();
  if (json.status !== "success") throw new Error(json.message);
  return json.data;
};
const client = createClient(request);
```

Errors throw `MountOSError` with `.message`, `.status`, `.errorCode`.

## Go usage

Pure stdlib, zero external deps. `go get github.com/mountos-io/mountos-admin-sdk/go`.

```go
import sdk "github.com/mountos-io/mountos-admin-sdk/go"

client, err := sdk.NewClient(sdk.Config{
  BaseURL:    "https://appserv.example.com",
  PrivateKey: "base64-ed25519-key", // 64-byte seed+pubkey (ed25519.PrivateKey)
})
// every call takes context.Context first:
acct, err := client.Accounts.Create(ctx, &sdk.CreateAccountRequest{Name: "Acme"})
list, err := client.Accounts.List(ctx, &sdk.ListOptions{Page: 1, Limit: 10})
_, err = client.Volumes.UpdateQuota(ctx, volID, &sdk.UpdateVolumeQuotaRequest{QuotaLimit: 1 << 30})
```

Errors: `errors.As(err, &sdkErr)` where `sdkErr` is `*sdk.Error` with `Message`,
`Status`, `ErrorCode`. Token cache is `sync.Mutex`-guarded (thread-safe).

> Note: TS accepts a 32-byte seed **or** 64-byte key; Go requires the full 64-byte
> `ed25519.PrivateKey`.

## Rust usage

Async (`tokio`/`reqwest`), edition 2024 (Rust 1.85+). `cargo add mountos-admin-sdk`.
TLS uses `rustls` with the OS trust store (self-hosted appserv behind a private CA works).

```rust
use mountos_admin_sdk::{Client, Config, CreateAccountRequest, UpdateVolumeQuotaRequest};

let client = Client::new(Config {
    base_url: "https://appserv.example.com".into(),
    private_key: "base64-ed25519-key".into(), // 32-byte seed or 64-byte seed+pubkey
    ..Default::default()
})?;

let created = client.accounts.create(&CreateAccountRequest {
    name: "Acme".into(), description: None, icon_url: None, provider_info: None,
}).await?;
let page = client.accounts.list(None).await?;                     // Option<&...ListOptions>
client.volumes.update_quota(created.id, &UpdateVolumeQuotaRequest { quota_limit: 1 << 30 }).await?;
```

Methods are `async` on snake_case accessors (`client.volume_fork_trees`, `client.region_audit_logs`).
Errors are `mountos_admin_sdk::Error`; match `Error::Api { message, status, error_code }` vs
`Error::Http(..)`/`Error::Serde(..)`. Build `serde_json::Value` fields via the re-exported
`mountos_admin_sdk::serde_json`.

## Resources

All three SDKs expose the same resource groups (TS `client.<group>.<action>`, Go
`client.<Group>.<Action>`, Rust `client.<group>.<action>` snake_case): accounts, users,
regions, regionClusters, clusters, storages, volumes (+ quota, API keys, stats),
volumeForkTrees/Entries/Searches, auditLogs, regionAuditLogs, serviceNodes, nodes,
clientSessions, discover, dashboard, metrics, license, alerts, regionAlerts,
gcWorkerEvents, vault. See [api.md](api.md) for the authoritative list and shapes.

## Auth contract (quick reference)

EdDSA (ED25519) JWT, subject `mountos:provider`, audience `mountos/appserv`,
scope `service`, claim `kfp = hex(sha256(ed25519_pubkey)[:16])` (also the JWT `kid`).
The SDK signs and caches tokens for you; read [api.md](api.md) `jwt:` section or the
auth source files above to reproduce token construction in another language.

## Maintaining the SDK (contributors only)

`ts/src/*_gen.*`, `go/*_gen.go`, and `rust/src/*_gen.rs` carry `Code generated by gen; DO
NOT EDIT.` - never hand-edit them. Change [api.yaml](api.yaml) (and the generators under
`gen/` if behavior changes), then `make gen` to regenerate Go + TS + Rust + docs. `make
check` type-checks TS, vets Go, and clippy-lints Rust. Keep all three in lockstep; they
come from the same spec.
