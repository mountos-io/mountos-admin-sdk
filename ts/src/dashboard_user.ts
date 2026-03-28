import type { DashboardUser } from './types_gen.js'

const DEFAULT_TTL_MS = 10 * 60 * 1000

const ED25519_PKCS8_PREFIX = new Uint8Array([
  0x30, 0x2e, 0x02, 0x01, 0x00, 0x30, 0x05, 0x06,
  0x03, 0x2b, 0x65, 0x70, 0x04, 0x22, 0x04, 0x20,
])

function decodeBase64(b64: string): Uint8Array {
  const bin = atob(b64)
  const out = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i)
  return out
}

function encodeBase64(bytes: Uint8Array): string {
  let bin = ''
  for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i])
  return btoa(bin)
}

function toBase64Url(bytes: Uint8Array): string {
  return encodeBase64(bytes).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

const verificationKeyCache = new Map<string, string>()

async function deriveVerificationKey(signingKeyBase64: string): Promise<string> {
  const cached = verificationKeyCache.get(signingKeyBase64)
  if (cached) return cached

  const seed = decodeBase64(signingKeyBase64)
  if (seed.length !== 32) throw new Error(`signing key must be 32 bytes, got ${seed.length}`)

  const pkcs8 = new Uint8Array(ED25519_PKCS8_PREFIX.length + seed.length)
  pkcs8.set(ED25519_PKCS8_PREFIX)
  pkcs8.set(seed, ED25519_PKCS8_PREFIX.length)

  const key = await crypto.subtle.importKey('pkcs8', pkcs8, { name: 'Ed25519' }, true, ['sign'])
  const jwk = await crypto.subtle.exportKey('jwk', key)
  const b64url = jwk.x as string
  const verificationKey = b64url.replace(/-/g, '+').replace(/_/g, '/') + '='.repeat((4 - b64url.length % 4) % 4)

  verificationKeyCache.set(signingKeyBase64, verificationKey)
  return verificationKey
}

async function deriveDashboardHMACKey(verificationKey: string): Promise<CryptoKey> {
  const raw = new TextEncoder().encode(verificationKey)
  const baseKey = await crypto.subtle.importKey('raw', raw, { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
  const derived = new Uint8Array(await crypto.subtle.sign('HMAC', baseKey, new TextEncoder().encode('dashboard-user')))
  return crypto.subtle.importKey('raw', derived, { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
}

/**
 * Produces the signed header value for X-MountOS-Dashboard-User.
 * Sets exp to now + 10 minutes. HMAC key derived from the Ed25519 verification
 * key (derived from privateKey) with domain separation.
 * Format: base64url(json).base64url(hmac-sha256(base64url(json), derived_key))
 */
export async function signDashboardUser(user: DashboardUser, privateKey: string): Promise<string> {
  const verificationKey = await deriveVerificationKey(privateKey)

  const withExp = { ...user, exp: Math.floor((Date.now() + DEFAULT_TTL_MS) / 1000) }
  const payload = new TextEncoder().encode(JSON.stringify(withExp))
  const encoded = toBase64Url(payload)

  const key = await deriveDashboardHMACKey(verificationKey)
  const sig = new Uint8Array(await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(encoded)))
  return `${encoded}.${toBase64Url(sig)}`
}
