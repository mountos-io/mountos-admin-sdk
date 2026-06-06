---
name: mountos-admin-sdk
description: Integrate the mountOS Admin (provider) API into TypeScript or Go applications using @mountos-io/admin-sdk or github.com/mountos-io/mountos-admin-sdk/go. Use when a task involves calling the mountOS provider API — managing accounts, users, regions, storages, volumes, quotas, volume API keys, audit logs, service nodes, client sessions, licenses, alerts, or vault — or when wiring up ED25519/JWT auth against appserv. Also use when porting the API to another language from api.yaml.
---

# mountOS Admin SDK

Provider-facing SDK for the mountOS Admin API (`/api/v1/*` on appserv), with ED25519
JWT authentication. Two first-party SDKs share one spec: **TypeScript** (`ts/`,
`@mountos-io/admin-sdk`) and **Go** (`go/`, `github.com/mountos-io/mountos-admin-sdk/go`).

Repo: https://github.com/mountos-io/mountos-admin-sdk

## When to use which entry point

- **Server-side TS/Go service** holding a provider private key → use the SDK's JWT client.
  Tokens are signed locally with ED25519, cached ~55 min (1h TTL, 5 min refresh margin).
- **Browser / cookie-or-session app** (no private key in the client) → use the TS
  `createClient(request)` form with your own transport; auth is handled by your backend.
- **Another language** → generate from [api.yaml](api.yaml); see [api.md](api.md).

## Reference map

- [api.md](api.md) — language-neutral REST reference: every endpoint, request/response
  shape, query params, error codes, and the JWT claim contract. Start here for the wire
  protocol or non-TS/Go callers.
- [docs/ts.md](docs/ts.md) — full TypeScript method/type reference.
- [docs/go.md](docs/go.md) — full Go method/type reference.
- [ts/README.md](ts/README.md) / [go/README.md](go/README.md) — install + quickstart.
- Auth internals (how the JWT is built) — read the source for exact claims/encoding:
  - TS: https://github.com/mountos-io/mountos-admin-sdk/blob/main/ts/src/auth.ts
  - Go: https://github.com/mountos-io/mountos-admin-sdk/blob/main/go/auth.go

## TypeScript usage

ESM-only, Node 18+/Bun/Deno. Install per [ts/README.md](ts/README.md) (GitHub dep,
submodule, or private registry — not on public npm).

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

## Resources

Both SDKs expose the same resource groups (TS `client.<group>.<action>`, Go
`client.<Group>.<Action>`): accounts, users, regions, regionClusters, storages,
volumes (+ quota, API keys, stats), volumeForkTrees/Entries/Searches, auditLogs,
regionAuditLogs, serviceNodes, nodes, clientSessions, discover, dashboard, license,
alerts, regionAlerts, vault. See [api.md](api.md) for the authoritative list and shapes.

## Auth contract (quick reference)

EdDSA (ED25519) JWT, subject `mountos:provider`, audience `mountos/appserv`,
scope `service`, claim `kfp = hex(sha256(ed25519_pubkey)[:16])` (also the JWT `kid`).
The SDK signs and caches tokens for you; read [api.md](api.md) `jwt:` section or the
auth source files above to reproduce token construction in another language.

## Maintaining the SDK (contributors only)

`ts/src/*_gen.*` and `go/*_gen.go` carry `Code generated by gen; DO NOT EDIT.` — never
hand-edit them. Change [api.yaml](api.yaml) (and the generators under `gen/` if behavior
changes), then `make gen` to regenerate Go + TS + docs. `make check` type-checks TS and
vets Go. Keep TS and Go in lockstep — both come from the same spec.
