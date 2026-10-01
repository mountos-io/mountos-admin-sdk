// Contract test for the fork tree listing cursor: it is an opaque string.
// The client must send it back unchanged and read nextCursor as a string,
// with JSON null (final page) read as null. Uses the real fetch transport so
// the literal query string on the wire is what gets asserted.
import { test } from 'node:test'
import assert from 'node:assert/strict'
import http from 'node:http'
import type { AddressInfo } from 'node:net'
import { createServerClient } from '../dist/index.js'

test('tree list cursor is an opaque string round trip', async () => {
  const opaque = 'n1.dGVzdC1fbmFtZQ'

  const urls: string[] = []
  const server = http.createServer((req, res) => {
    urls.push(req.url ?? '')
    const hasCursor = new URL(req.url ?? '', 'http://x').searchParams.has('cursor')
    res.setHeader('content-type', 'application/json')
    res.end(JSON.stringify({
      status: 'success', message: 'ok',
      data: {
        items: [{ name: 'a', kind: 'file', inode: 5, size: 1, mtime: 2, ctime: 3 }],
        nextCursor: hasCursor ? null : opaque,
      },
    }))
  })
  await new Promise<void>(resolve => server.listen(0, resolve))
  const { port } = server.address() as AddressInfo

  const privateKey = Buffer.from(new Uint8Array(32).fill(7)).toString('base64')
  const client = createServerClient({ baseUrl: `http://127.0.0.1:${port}`, privateKey })

  try {
    const first = await client.volumeForkTrees.list(7, 'main', { path: '/' })
    assert.equal(first.nextCursor, opaque)
    assert.doesNotMatch(urls[0], /cursor=/)

    const last = await client.volumeForkTrees.list(7, 'main', { path: '/', cursor: first.nextCursor as string })
    assert.match(urls[1], new RegExp(`cursor=${opaque.replace('.', '\\.')}`))
    assert.equal(last.nextCursor, null)
  } finally {
    await new Promise<void>(resolve => server.close(() => resolve()))
  }
})
