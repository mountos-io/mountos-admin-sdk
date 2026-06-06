# @mountos-io/admin-sdk

TypeScript SDK for the mountOS provider API. ESM-only, Node 18+/Bun/Deno.

## Install

```bash
npm install @mountos-io/admin-sdk
```

Bun (`bun add @mountos-io/admin-sdk`) and Deno (`deno add npm:@mountos-io/admin-sdk`) work too.

## Usage

### Server-side (Node/Bun/Deno) with JWT auth

```typescript
import { createServerClient } from "@mountos-io/admin-sdk";

const client = createServerClient({
  baseUrl: "https://appserv.example.com",
  privateKey: "base64-encoded-ed25519-key", // 32-byte seed or 64-byte seed+pubkey
});

// Accounts
const { id } = await client.accounts.create({ name: "Acme" });
const { items, pagination } = await client.accounts.list({ page: 1, limit: 10 });
const account = await client.accounts.get(id);
await client.accounts.edit(id, { name: "Acme Corp" });
await client.accounts.lock(id);
await client.accounts.unlock(id);
await client.accounts.deactivate(id);

// Users
const user = await client.users.add({ accountId: 1, email: "a@b.com", username: "alice" });
const users = await client.users.list({ accountId: 1 });
const u = await client.users.get(user.id);
await client.users.edit(user.id, { username: "bob", email: "b@c.com" });
await client.users.deactivate(user.id);

// Regions
const region = await client.regions.create({ accountId: 1, name: "us-east" });
const regions = await client.regions.list();
const r = await client.regions.get(region.id);
await client.regions.edit(region.id, { accountId: 1, name: "us-west" });
await client.regions.deactivate(region.id);

// Storages
const storage = await client.storages.create({
  accountId: 1, regionId: 1, name: "prod-s3",
  storageType: "object", providerType: "s3", endpoint: "https://s3.example.com",
});
const storages = await client.storages.list({ accountId: 1 });
const s = await client.storages.get(storage.id);
await client.storages.edit(storage.id, { name: "new-name" });
await client.storages.deactivate(storage.id);

// Volumes
const vol = await client.volumes.create({ accountId: 1, storageId: 1, name: "data", volumeType: "standard" });
const vols = await client.volumes.list({ accountId: 1 });
const v = await client.volumes.get(vol.id);
await client.volumes.edit(vol.id, { name: "data-v2" });
await client.volumes.lock(vol.id);
await client.volumes.unlock(vol.id);
await client.volumes.deactivate(vol.id);
await client.volumes.updateQuota(vol.id, { quotaLimit: 1073741824 });
const keys = await client.volumes.generateAPIKeys(vol.id, { userId: 1 });
await client.volumes.revokeAPIKey(vol.id, { apiKey: keys.apiKey });
const stats = await client.volumes.stats(vol.id);

// Audit logs (cursor-based)
const logs = await client.auditLogs.list({ accountId: 1, cursor: 0, limit: 20 });

// Service nodes
const nodes = await client.serviceNodes.list(region.id);
const nodeStats = await client.serviceNodes.stats(region.id, "node-1");

// Discover
const meta = await client.discover.meta("AKID...");

// Cache
await client.cache.refresh();
```

### Browser / custom transport

For browser apps (cookie/session auth, token refresh, etc.) supply your own `RequestFn`:

```typescript
import { createClient, type RequestFn } from "@mountos-io/admin-sdk";

const request: RequestFn = async (method, path, body, signal) => {
  const res = await fetch(`https://api.example.com${path}`, {
    method,
    credentials: "include",
    headers: body !== undefined ? { "Content-Type": "application/json" } : {},
    body: body !== undefined ? JSON.stringify(body) : undefined,
    signal,
  });
  const json = await res.json();
  if (json.status !== "success") {
    throw new Error(json.message);
  }
  return json.data;
};

const client = createClient(request);
await client.accounts.list({ page: 1, limit: 10 });
```

## Error Handling

```typescript
import { MountOSError } from "@mountos-io/admin-sdk";

try {
  await client.accounts.get(999);
} catch (err) {
  if (err instanceof MountOSError) {
    console.error(err.message); // "account not found"
    console.error(err.status); // 404
    console.error(err.errorCode); // 10900
  }
}
```

## Auth

The `privateKey` accepts a base64-encoded Ed25519 key in either format:
- **32 bytes** (seed only) — standard format, e.g. from `openssl genpkey`
- **64 bytes** (seed + public key concatenated) — Go's `ed25519.PrivateKey` format

JWT tokens are auto-generated and cached for ~55 minutes (1h TTL with 5min refresh margin).

## Reference

The published package bundles, alongside `dist/`:

- **`api.yaml`** — the source spec the SDK is generated from; use it to port the API to another language.
- **`api.md`** — language-neutral REST reference: every endpoint, request/response shape, query params, error codes, and the JWT claim contract.
- **`SKILL.md`** — an agent skill describing how to use this SDK (TS/Go usage, auth, resources). AI coding agents can load it for accurate, up-to-date guidance.

Resolve their on-disk paths from the package root:

```typescript
import { createRequire } from "node:module";
const pkgDir = createRequire(import.meta.url).resolve("@mountos-io/admin-sdk/package.json");
// api.md and SKILL.md sit next to package.json
```

## License

Apache-2.0
