import { createServerClient, MountOSError } from '@mountos-io/admin-sdk'

const BASE_URL = process.env.MOUNTOS_BASE_URL || 'https://appserv.example.com'
const PRIVATE_KEY = process.env.MOUNTOS_PRIVATE_KEY || ''

async function main() {
  const client = createServerClient({
    baseUrl: BASE_URL,
    privateKey: PRIVATE_KEY,
  })

  // --- Accounts ---
  const { id: accountId } = await client.accounts.create({
    name: 'Acme Corp',
    description: 'Demo account',
    providerInfo: { tier: 'enterprise' },
  })
  console.log('Created account ID:', accountId)

  const account = await client.accounts.get(accountId)
  console.log(`Account: ${account.name} (active=${account.isActive}, locked=${account.locked})`)

  const { items: accounts, pagination } = await client.accounts.list({ page: 1, limit: 10 })
  console.log(`Accounts: ${pagination.total} total, page ${pagination.page} of ${pagination.totalPages}`)

  await client.accounts.edit(accountId, {
    name: 'Acme Corp Updated',
    description: 'Updated description',
  })

  // --- Users ---
  const { id: userId } = await client.users.add({
    accountId,
    username: 'alice',
    email: 'alice@acme.com',
    name: 'Alice Smith',
  })
  console.log('Added user ID:', userId)

  const { items: users } = await client.users.list({ accountId, page: 1, limit: 20 })
  for (const u of users) {
    console.log(`  User: ${u.username} <${u.email}> (id=${u.id})`)
  }

  await client.users.edit(userId, {
    username: 'alice-updated',
    email: 'alice-new@acme.com',
  })

  // --- Regions ---
  const { id: regionId } = await client.regions.create({
    accountId,
    name: 'us-east-1',
  })
  console.log('Created region ID:', regionId)

  const region = await client.regions.get(regionId)
  console.log(`Region: ${region.name} (exportId=${region.exportId})`)

  // --- Storages ---
  const { id: storageId } = await client.storages.create({
    accountId,
    regionId,
    name: 'prod-s3-bucket',
    storageType: 'object',
    providerType: 's3',
    endpoint: 'https://s3.us-east-1.amazonaws.com',
    bucket: 'my-mountos-bucket',
    region: 'us-east-1',
  })
  console.log('Created storage ID:', storageId)

  const { items: storages } = await client.storages.list({ accountId })
  for (const s of storages) {
    console.log(`  Storage: ${s.name} (type=${s.storageType}, active=${s.isActive})`)
  }

  // --- Volumes ---
  try {
    await client.volumes.updateQuota(1, { quotaLimit: 10 * 1024 * 1024 * 1024 })
  } catch (err) {
    if (err instanceof MountOSError) {
      console.log(`UpdateQuota error: ${err.message} (status=${err.status})`)
    }
  }

  // --- Audit Logs ---
  const { items: logs, nextCursor } = await client.auditLogs.list({
    accountId,
    limit: 5,
  })
  console.log(`Audit logs: ${logs.length} entries (nextCursor=${nextCursor})`)
  for (const entry of logs) {
    console.log(`  [${entry.id}] ${entry.title} (success=${entry.success})`)
  }

  // --- Service Nodes ---
  try {
    const nodes = await client.serviceNodes.list(regionId)
    console.log(`Service nodes: ${nodes.length}`)
    for (const n of nodes) {
      console.log(`  Node: ${n.nodeId} (type=${n.serviceType}, status=${n.status})`)
    }
  } catch (err) {
    if (err instanceof MountOSError) {
      console.log(`List nodes: ${err.message}`)
    }
  }

  // --- License ---
  const license = await client.license.get()
  console.log(`License: ${license.licensee} (${license.status})`)

  // Upload signed payloads; the HUB verifies each and rejects the batch if any is invalid.
  try {
    const { loaded, ignored } = await client.license.load({
      payloads: [process.env.MOUNTOS_LICENSE_PAYLOAD || ''],
    })
    console.log(`License loaded: ${loaded} new, ${ignored} ignored`)
  } catch (err) {
    if (err instanceof MountOSError) {
      console.log(`License load rejected: ${err.message} (status=${err.status})`)
    }
  }

  const { items: licenses } = await client.license.list()
  console.log(`Stored license payloads: ${licenses.length}`)
  for (const r of licenses) {
    // active is the newest non-expired payload for its license id; expiry retires it.
    console.log(`  ${r.key}: ${r.licensee} (${r.status}, active=${r.active})`)
  }

  // --- Vault ---
  try {
    await client.vault.resync()
    console.log('Vault resynced')
  } catch (err) {
    if (err instanceof MountOSError) {
      console.log(`Vault resync: ${err.message}`)
    }
  }

  console.log('Done.')
}

main().catch(console.error)
