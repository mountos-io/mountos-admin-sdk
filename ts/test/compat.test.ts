// Regression guard for the Pair -> Copyset rename (1.14.0, shipped with no
// migration aliases). Exercises the compat.ts shims: old method names,
// old-named request fields, and old-named response fields populated from
// the new wire fields.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { createClient, type RequestFn } from '../dist/index.js'

function fakeClient(responses: Record<string, unknown>) {
  const calls: Array<{ method: string; path: string; body?: unknown }> = []
  const request: RequestFn = async (method, path, body) => {
    calls.push({ method, path, body })
    const key = `${method} ${path}`
    if (!(key in responses)) throw new Error(`fixture has no response for ${key}`)
    return responses[key] as never
  }
  return { client: createClient(request), calls }
}

test('listPairs is an alias for listCopysets', async () => {
  const copyset = { id: 'c1', storageId: 's1', state: 'active', tags: [] }
  const { client, calls } = fakeClient({
    'GET /api/v1/storages/7/copysets': [copyset],
  })
  const pairs = await client.storages.listPairs(7)
  assert.deepEqual(pairs, [copyset])
  assert.equal(calls[0].path, '/api/v1/storages/7/copysets')
})

test('getPairStatus/drainPair are aliases for getCopysetStatus/drainCopyset', async () => {
  const { client } = fakeClient({
    'GET /api/v1/storages/7/copysets/c1': { id: 'c1', storageId: 's1', state: 'active', tags: [] },
    'POST /api/v1/storages/7/copysets/c1/drain': { id: 'c1', state: 'draining' },
  })
  const status = await client.storages.getPairStatus(7, 'c1')
  assert.equal(status.state, 'active')
  const drained = await client.storages.drainPair(7, 'c1')
  assert.equal(drained.state, 'draining')
})

test('getPairConfig/updatePairConfig alias getCopysetConfig/updateCopysetConfig and map request/response fields', async () => {
  const { client, calls } = fakeClient({
    'GET /api/v1/volumes/5/copyset-config': { id: 5, targetCopysetCount: 3, currentEpoch: 1, copysetIds: ['c1', 'c2'] },
    'PUT /api/v1/volumes/5/copyset-config': {
      id: 5, targetCopysetCount: 4, copysetCountBefore: 3, copysetsAdded: 1, copysetsRemoved: 0,
      copysetCountAfter: 4, epoch: 2, partial: false,
    },
  })
  const cfg = await client.volumes.getPairConfig(5)
  assert.equal(cfg.targetPairCount, 3)
  assert.deepEqual(cfg.pairIds, ['c1', 'c2'])

  const resize = await client.volumes.updatePairConfig(5, { targetPairCount: 4 })
  assert.equal(calls[1].body && (calls[1].body as { targetCopysetCount: number }).targetCopysetCount, 4)
  assert.equal(resize.targetPairCount, 4)
  assert.equal(resize.pairCountBefore, 3)
  assert.equal(resize.pairsAdded, 1)
  assert.equal(resize.pairsRemoved, 0)
  assert.equal(resize.pairCountAfter, 4)
})

test('getCopysetConfig/updateCopysetConfig (new names) still carry the deprecated fields too', async () => {
  const { client } = fakeClient({
    'GET /api/v1/volumes/5/copyset-config': { id: 5, targetCopysetCount: 3, currentEpoch: 1, copysetIds: ['c1'] },
  })
  const cfg = await client.volumes.getCopysetConfig(5)
  assert.equal(cfg.targetPairCount, 3)
  assert.deepEqual(cfg.pairIds, ['c1'])
})

test('listBlockVolumes response carries pairId shimmed from copysetId', async () => {
  const { client } = fakeClient({
    'GET /api/v1/storages/7/block-volumes': [
      { id: 'bv1', shardId: 1, isActive: true, createdAt: '', updatedAt: '', copysetId: 'c1' },
      { id: 'bv2', shardId: 2, isActive: true, createdAt: '', updatedAt: '' },
    ],
  })
  const vols = await client.storages.listBlockVolumes(7)
  assert.equal(vols[0].pairId, 'c1')
  assert.equal(vols[1].pairId, undefined)
})
