# @mountos/admin-sdk

TypeScript SDK for the mountOS vendor API. ESM-only, Node 18+/Bun/Deno.

## Install

```bash
npm install @mountos/admin-sdk
```

## Usage

```typescript
import { MountOSAdmin } from '@mountos/admin-sdk'

const client = new MountOSAdmin({
  baseUrl: 'https://appserv.example.com',
  privateKey: 'base64-encoded-ed25519-private-key',
})

// Accounts
const { id } = await client.accounts.create({ name: 'Acme' })
const { items, pagination } = await client.accounts.list({ page: 1, limit: 10 })
const account = await client.accounts.get(id)
await client.accounts.edit(id, { name: 'Acme Corp' })
await client.accounts.lock(id)
await client.accounts.unlock(id)
await client.accounts.activate(id)
await client.accounts.deactivate(id)

// Users
const user = await client.users.add({ accountId: 1, email: 'a@b.com', username: 'alice' })
const users = await client.users.list({ accountId: 1 })
await client.users.edit(user.id, { username: 'bob', email: 'b@c.com' })
await client.users.activate(user.id)
await client.users.deactivate(user.id)

// Regions
const region = await client.regions.create({ accountId: 1, name: 'us-east', dns: 'us.example.com' })
const regions = await client.regions.list()
await client.regions.edit(region.id, { accountId: 1, name: 'us-west', dns: 'us-w.example.com' })

// Storages
const storage = await client.storages.create({
  accountId: 1, regionId: 1, name: 'prod-s3',
  storageType: 'object', providerType: 's3', endpoint: 'https://s3.example.com',
})
await client.storages.edit(storage.id, { name: 'new-name' })

// Volumes
await client.volumes.updateQuota('volume-uuid', { quotaLimit: 1073741824 })

// Audit logs (cursor-based)
const logs = await client.auditLogs.list({ accountId: 1, cursor: 0, limit: 20 })
```

## Error Handling

```typescript
import { MountOSError } from '@mountos/admin-sdk'

try {
  await client.accounts.get(999)
} catch (err) {
  if (err instanceof MountOSError) {
    console.error(err.message)    // "account not found"
    console.error(err.status)     // 404
    console.error(err.errorCode)  // 10900
  }
}
```

## Auth

JWT tokens are auto-generated using your ED25519 private key and cached for ~55 minutes (1h TTL with 5min refresh margin).

## License

MIT
