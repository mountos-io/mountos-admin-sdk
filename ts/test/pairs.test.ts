// Fixture/mock-server contract test for the block HA-pair placement admin
// surface (admin-sdk.md §5 step 2): exercises the generated client against
// a hand-written fake RequestFn, no live appserv. Covers the "accepted, not
// completed" response shapes (drainPair/cancelDrain/updateConfig) and
// regression-guards the reactivateMember GET-vs-POST generator bug (a
// no-request endpoint with a named responseType on a mutating method was
// silently generated as GET in all three SDK languages until fixed
// alongside this test).
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

test('getConfig: read-back half of updateConfig (D40)', async () => {
  const { client, calls } = fakeClient({
    'GET /api/v1/storages/7/config': { id: 's1', k: 3, algorithmVersion: 1, epochPolicyVersion: 1 },
  })
  const cfg = await client.storages.getConfig(7)
  assert.equal(cfg.k, 3)
  assert.equal(calls[0].method, 'GET')
})

test('listPairs: full state per pair, no per-pair follow-up call needed', async () => {
  const { client, calls } = fakeClient({
    'GET /api/v1/storages/7/pairs': [
      { id: 'p1', storageId: 's1', state: 'active', memberA: 'bv1', memberB: 'bv2', placementGroupA: 1, placementGroupB: 2 },
      { id: 'p2', storageId: 's1', state: 'draining', memberA: 'bv3', memberB: 'bv4', pendingSyncJobsA: 3, pendingSyncJobsB: 0 },
    ],
  })
  const pairs = await client.storages.listPairs(7)
  assert.equal(pairs.length, 2)
  assert.equal(pairs[1].state, 'draining')
  assert.equal(pairs[1].pendingSyncJobsA, 3)
  assert.deepEqual(calls, [{ method: 'GET', path: '/api/v1/storages/7/pairs', body: undefined }])
})

test('getPairStatus: nullable drain-only fields absent on an active pair', async () => {
  const { client } = fakeClient({
    'GET /api/v1/storages/7/pairs/p1': { id: 'p1', storageId: 's1', state: 'active', memberA: 'bv1', memberB: 'bv2' },
  })
  const pair = await client.storages.getPairStatus(7, 'p1')
  assert.equal(pair.state, 'active')
  assert.equal(pair.pendingSyncJobsA, undefined)
})

test('drainPair: idempotent-ack shape (D9) - state reads "draining", not "drained"', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/pairs/p1/drain': { id: 'p1', state: 'draining' },
  })
  const res = await client.storages.drainPair(7, 'p1')
  assert.equal(res.state, 'draining')
  assert.equal(calls[0].method, 'POST')
})

test('cancelDrain (D27): active-again ack', async () => {
  const { client } = fakeClient({
    'POST /api/v1/storages/7/pairs/p1/cancel-drain': { id: 'p1', state: 'active' },
  })
  const res = await client.storages.cancelDrain(7, 'p1')
  assert.equal(res.state, 'active')
})

test('updateConfig: partial-success surfaces reason, not swallowed', async () => {
  const { client, calls } = fakeClient({
    'PUT /api/v1/storages/7/config': {
      id: 's1', pairsFormed: 2, pairsRequested: 3, partial: true,
      reason: 'placement cluster B has no unused members',
    },
  })
  const res = await client.storages.updateConfig(7, { k: 3 })
  assert.equal(res.partial, true)
  assert.equal(res.reason, 'placement cluster B has no unused members')
  assert.deepEqual(calls[0].body, { k: 3 })
})

test('reactivateMember: sends POST (regression guard for the responseType/no-request generator bug)', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/members/bv1/reactivate': {
      id: 'bv1', name: 'originator', regionId: 1, regionClusterId: 2, memberState: 'active',
    },
  })
  const res = await client.storages.reactivateMember(7, 'bv1')
  assert.equal(res.memberState, 'active')
  assert.equal(calls[0].method, 'POST')
})

test('registerMember: required name, immediately poolable (memberState active)', async () => {
  const { client, calls } = fakeClient({
    'POST /api/v1/storages/7/members': {
      id: 'bv5', name: 'new-member', regionId: 1, regionClusterId: 3, memberState: 'active',
    },
  })
  const res = await client.storages.registerMember(7, { regionClusterId: 3, name: 'new-member' })
  assert.equal(res.memberState, 'active')
  assert.deepEqual(calls[0].body, { regionClusterId: 3, name: 'new-member' })
})
