# @mountos-app/admin-sdk

TypeScript SDK for the mountOS vendor API. ESM-only, Node 18+/Bun/Deno.

## Install

This package is not published to npm. Use one of the following methods:

**GitHub dependency (package.json):**

```json
{
  "dependencies": {
    "@mountos-app/admin-sdk": "github:mountos-app/mountos-admin-sdk#main"
  }
}
```

Note: since the TS package lives under `ts/`, you may need to reference the subdirectory. With npm/pnpm you can use:

```json
{
  "dependencies": {
    "@mountos-app/admin-sdk": "github:mountos-app/mountos-admin-sdk#main&path:ts"
  }
}
```

**Git submodule:**

```bash
git submodule add https://github.com/mountos-app/mountos-admin-sdk.git vendor/mountos-admin-sdk
```

Then reference in your tsconfig paths or use a workspace/symlink pointing to `vendor/mountos-admin-sdk/ts`.

**Private registry:**

If your organization runs a private npm registry (Verdaccio, GitHub Packages, etc.), publish the `ts/` package there and install normally:

```bash
npm install @mountos-app/admin-sdk --registry=https://npm.your-org.com
```

## Usage

```typescript
import { MountOSAdmin } from "@mountos-app/admin-sdk";

const client = new MountOSAdmin({
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
await client.accounts.activate(id);
await client.accounts.deactivate(id);

// Users
const user = await client.users.add({ accountId: 1, email: "a@b.com", username: "alice" });
const users = await client.users.list({ accountId: 1 });
const u = await client.users.get(user.id);
await client.users.edit(user.id, { username: "bob", email: "b@c.com" });
await client.users.activate(user.id);
await client.users.deactivate(user.id);

// Regions
const region = await client.regions.create({ accountId: 1, name: "us-east" });
const regions = await client.regions.list();
const r = await client.regions.get(region.id);
await client.regions.edit(region.id, { accountId: 1, name: "us-west" });
await client.regions.activate(region.id);
await client.regions.deactivate(region.id);

// Storages
const storage = await client.storages.create({
  accountId: 1, regionId: 1, name: "prod-s3",
  storageType: "object", providerType: "s3", endpoint: "https://s3.example.com",
});
const storages = await client.storages.list({ accountId: 1 });
const s = await client.storages.get(storage.id);
await client.storages.edit(storage.id, { name: "new-name" });
await client.storages.activate(storage.id);
await client.storages.deactivate(storage.id);

// Volumes
const vol = await client.volumes.create({ accountId: 1, storageId: 1, name: "data", volumeType: "standard" });
const vols = await client.volumes.list({ accountId: 1 });
const v = await client.volumes.get(vol.id);
await client.volumes.edit(vol.id, { name: "data-v2" });
await client.volumes.lock(vol.id);
await client.volumes.unlock(vol.id);
await client.volumes.activate(vol.id);
await client.volumes.deactivate(vol.id);
await client.volumes.updateQuota(vol.id, { quotaLimit: 1073741824 });
const keys = await client.volumes.generateAPIKeys(vol.id, { userId: 1 });
await client.volumes.revokeAPIKey(vol.id, { apiKey: keys.apiKey });
const stats = await client.volumes.stats(vol.id);

// Audit logs (cursor-based)
const logs = await client.auditLogs.list({ accountId: 1, cursor: 0, limit: 20 });

// Service nodes
const nodes = await client.serviceNodes.list(region.id);
await client.serviceNodes.drain(region.id, "node-1");
await client.serviceNodes.activate(region.id, "node-1");
await client.serviceNodes.remove(region.id, "node-1");

// Discover
const meta = await client.discover.meta("AKID...");

// Cache
await client.cache.refresh();
```

## Error Handling

```typescript
import { MountOSError } from "@mountos-app/admin-sdk";

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

## License

MIT
