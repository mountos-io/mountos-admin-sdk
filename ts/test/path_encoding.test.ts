// Regression coverage for the generator emitting encodeURIComponent on
// every string :path param (not just the one endpoint that originally
// needed it). Exercises the real fetch-based transport (createServerClient),
// not the fake RequestFn used elsewhere in this directory: fetch's own URL
// parser treats a raw '/' as a structural separator (splitting the path
// into extra segments) while auto-escaping other unsafe characters, so only
// the literal bytes actually sent on the wire (a plain http.Server's
// req.url, which Node never auto-decodes) can tell an escaped "%2F" apart
// from a real segment boundary.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import http from 'node:http'
import type { AddressInfo } from 'node:net'
import { createServerClient } from '../dist/index.js'

test('string path params are URL-encoded on the wire', async () => {
  const rawCopysetId = 'copyset/id with spaces/café'

  let requestURL = ''
  const server = http.createServer((req, res) => {
    requestURL = req.url ?? ''
    res.setHeader('content-type', 'application/json')
    res.end(JSON.stringify({
      status: 'success', message: 'ok',
      data: { id: rawCopysetId, storageId: 's1', state: 'active' },
    }))
  })
  await new Promise<void>(resolve => server.listen(0, resolve))
  const { port } = server.address() as AddressInfo

  const privateKey = Buffer.from(new Uint8Array(32).fill(7)).toString('base64')
  const client = createServerClient({ baseUrl: `http://127.0.0.1:${port}`, privateKey })

  try {
    const copyset = await client.storages.getCopysetStatus(7, rawCopysetId)
    assert.equal(copyset.id, rawCopysetId)
    assert.match(requestURL, /copyset%2Fid%20with%20spaces%2Fcaf%C3%A9/)
    assert.doesNotMatch(requestURL, /\/copysets\/copyset\//)
  } finally {
    await new Promise<void>(resolve => server.close(() => resolve()))
  }
})
