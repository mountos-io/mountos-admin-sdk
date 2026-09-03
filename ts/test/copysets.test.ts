// Fixture/mock-server contract test for the block copyset placement admin
// surface: exercises the generated client against a hand-written fake
// RequestFn, no live appserv. Covers the "accepted, not completed" response
// shapes (drainCopyset/cancelDrain) and regression-guards the GET-vs-POST
// generator bug (a no-request endpoint with a named responseType on a
// mutating method was silently generated as GET in all three SDK languages
// until fixed) via addCopysetMember, the surviving action with that same
// shape.
import { test } from 'node:test'
import assert from 'node:assert/strict'
// Imports the tsc-compiled output (dist/), not src/ directly: this repo's
// src/*.ts files cross-import each other with .js specifiers per NodeNext
// module resolution, which only resolve once tsc has produced dist/*.js -
// Node's own ESM loader (even with native TS type-stripping) does not
// implement TypeScript's module resolution, so it can't follow a .js
// specifier to a sibling .ts source file directly. `npm test` runs `tsc`
// before this file, so dist/ is guaranteed fresh.
import { createClient, type RequestFn } from '../dist/index.js'

// Records every call for assertion, and answers from a path->response map
// keyed by "METHOD path".
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

test('listCopysets: full state per copyset, no per-copyset follow-up call needed', async () => {
  const { client, calls } = fakeClient({
    'GET /api/v1/storages/7/copysets': [
      { id: 'p1', storageId: 's1', state: 'active', memberA: 'bv1', memberB: 'bv2' },
      { id: 'p2', storageId: 's1', state: 'draining', memberA: 'bv3', memberB: 'bv4', pendingSyncJobsA: 3, pendingSyncJobsB: 0 },
    ],
  })
  const copysets = await client.storages.listCopysets(7)
  assert.equal(copysets.length, 2)
  assert.equal(copysets[1].state, 'draining')
  assert.equal(copysets[1].pendingSyncJobsA, 3)
  assert.deepEqual(calls, [{ method: 'GET', path: '/api/v1/storages/7/copysets', body: undefined }])
})

test('getCopysetStatus: nullable drain-only fields absent on an active copyset', async () => {
  const { client } = fakeClient({
    'GET /api/v1/storages/7/copysets/p1': { id: 'p1', storageId: 's1', state: 'active', memberA: 'bv1', memberB: 'bv2' },
  })
  const copyset = await client.storages.getCopysetStatus(7, 'p1')
  assert.equal(copyset.state, 'active')
  assert.equal(copyset.pendingSyncJobsA, undefined)
})

test('drainCopyset: idempotent-ack shape - state reads "draining", not "drained"', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/copysets/p1/drain': { id: 'p1', state: 'draining' },
  })
  const res = await client.storages.drainCopyset(7, 'p1', {})
  assert.equal(res.state, 'draining')
  assert.equal(calls[0].method, 'POST')
})

test('cancelDrain: active-again ack', async () => {
  const { client } = fakeClient({
    'POST /api/v1/storages/7/copysets/p1/cancel-drain': { id: 'p1', state: 'active' },
  })
  const res = await client.storages.cancelDrain(7, 'p1', {})
  assert.equal(res.state, 'active')
})

test('registerCopyset: explicit name, both members immediately poolable', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/copysets': {
      id: 'p5', storageId: 's1', name: 'mos-block-a', state: 'active', memberA: 'bv5', memberB: 'bv6',
    },
  })
  const res = await client.storages.registerCopyset(7, { name: 'mos-block-a' })
  assert.equal(res.name, 'mos-block-a')
  assert.equal(res.memberA, 'bv5')
  assert.equal(res.memberB, 'bv6')
  assert.deepEqual(calls[0].body, { name: 'mos-block-a' })
})

test('registerCopyset: omitted name auto-fills server-side, both members derive from it', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/copysets': {
      id: 'p6', storageId: 's1', name: 'riveted-truss-4f2a', state: 'active', memberA: 'bv7', memberB: 'bv8',
    },
  })
  const res = await client.storages.registerCopyset(7, {})
  assert.equal(res.state, 'active')
  assert.equal(res.name, 'riveted-truss-4f2a')
  assert.deepEqual(calls[0].body, {})
})

test('registerCopysetsBulk: count-only, every copyset auto-named', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/copysets/bulk': {
      copysets: [
        { id: 'p10', storageId: 's1', name: 'riveted-truss-1a2b', state: 'active', memberA: 'bv10', memberB: 'bv11' },
        { id: 'p11', storageId: 's1', name: 'coupled-beam-3c4d', state: 'active', memberA: 'bv12', memberB: 'bv13' },
      ],
    },
  })
  const res = await client.storages.registerCopysetsBulk(7, { count: 2 })
  assert.equal(res.copysets.length, 2)
  assert.equal(res.copysets[0].name, 'riveted-truss-1a2b')
  assert.equal(res.copysets[1].name, 'coupled-beam-3c4d')
  assert.deepEqual(calls[0].body, { count: 2 })
})

// failureDomainA/failureDomainB are optional, one per member: omitted
// entirely, they must never reach the wire, and a given pair must reach it
// under their exact field names.
test('registerCopyset: failureDomainA/failureDomainB reach the wire under their own names', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/copysets': {
      id: 'p12', storageId: 's1', name: 'mos-block-fd', state: 'active', memberA: 'bv14', memberB: 'bv15',
    },
  })
  await client.storages.registerCopyset(7, { name: 'mos-block-fd', failureDomainA: 'rack-1', failureDomainB: 'rack-2' })
  assert.deepEqual(calls[0].body, { name: 'mos-block-fd', failureDomainA: 'rack-1', failureDomainB: 'rack-2' })
})

test('addCopysetMember: replaces a vacant slot on an existing copyset, no request body, sends POST (regression guard for the responseType/no-request generator bug)', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/copysets/p1/members': {
      id: 'bv9', name: 'mos-block-a-b', regionId: 1, memberState: 'active',
    },
  })
  const res = await client.storages.addCopysetMember(7, 'p1')
  assert.equal(res.memberState, 'active')
  assert.equal(res.name, 'mos-block-a-b')
  assert.equal(calls[0].method, 'POST')
  assert.equal(calls[0].body, undefined)
})
