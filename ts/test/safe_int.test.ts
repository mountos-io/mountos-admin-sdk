// BHA-063: server int64 fields (ids, byte counts, ...) can exceed
// Number.MAX_SAFE_INTEGER; JSON.parse silently rounds such a value to the
// nearest representable double. assertSafeIntegers must catch that instead
// of letting a corrupted id/count flow through unnoticed.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { assertSafeIntegers, UnsafeIntegerError, createServerClient } from '../dist/index.js'

test('assertSafeIntegers passes safe nested values', () => {
  assert.doesNotThrow(() => assertSafeIntegers({
    id: 42,
    nested: { arr: [1, 2, 3], name: 'ok', ratio: 3.14159 },
  }))
})

test('assertSafeIntegers throws on an out-of-range integer, with a path', () => {
  assert.throws(
    () => assertSafeIntegers({ copysets: [{ placementGroupA: 9007199254740993 }] }),
    (err: unknown) => err instanceof UnsafeIntegerError && err.message.includes('$.copysets[0].placementGroupA'),
  )
})

test('assertSafeIntegers does not flag ordinary floats', () => {
  assert.doesNotThrow(() => assertSafeIntegers({ ratio: 0.1, negRatio: -3.5 }))
})

function fakePrivateKey(): string {
  return Buffer.from(new Uint8Array(32).fill(7)).toString('base64')
}

test('a real response carrying an unsafe int64 rejects loudly instead of returning a corrupted value', async () => {
  const originalFetch = globalThis.fetch
  globalThis.fetch = (async () => {
    // Raw literal, not JSON.stringify'd: preserves the full digit sequence
    // so JSON.parse (inside res.json()) is what performs the rounding this
    // test is guarding against, exactly as a real fetch response would.
    const raw = '{"status":"success","message":"ok","data":'
      + '{"id":"p1","storageId":"s1","state":"active","placementGroupA":9223372036854775807}}'
    return new Response(raw, { status: 200, headers: { 'content-type': 'application/json' } })
  }) as typeof fetch

  try {
    const client = createServerClient({ baseUrl: 'http://example.invalid', privateKey: fakePrivateKey() })
    await assert.rejects(() => client.storages.getCopysetStatus(7, 'p1'), UnsafeIntegerError)
  } finally {
    globalThis.fetch = originalFetch
  }
})

test('a response with only safe integers resolves normally', async () => {
  const originalFetch = globalThis.fetch
  globalThis.fetch = (async () => {
    const raw = '{"status":"success","message":"ok","data":'
      + '{"id":"p1","storageId":"s1","state":"active","placementGroupA":2}}'
    return new Response(raw, { status: 200, headers: { 'content-type': 'application/json' } })
  }) as typeof fetch

  try {
    const client = createServerClient({ baseUrl: 'http://example.invalid', privateKey: fakePrivateKey() })
    const copyset = await client.storages.getCopysetStatus(7, 'p1')
    assert.equal(copyset.placementGroupA, 2)
  } finally {
    globalThis.fetch = originalFetch
  }
})
