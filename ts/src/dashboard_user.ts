import type { DashboardUser } from './types_gen.js'

const DEFAULT_TTL_MS = 10 * 60 * 1000

function encodeBase64(bytes: Uint8Array): string {
  let bin = ''
  for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i])
  return btoa(bin)
}

function toBase64Url(bytes: Uint8Array): string {
  return encodeBase64(bytes).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

async function deriveDashboardHMACKey(secret: string): Promise<CryptoKey> {
  const raw = new TextEncoder().encode(secret)
  const baseKey = await crypto.subtle.importKey('raw', raw, { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
  const derived = new Uint8Array(await crypto.subtle.sign('HMAC', baseKey, new TextEncoder().encode('dashboard-user')))
  return crypto.subtle.importKey('raw', derived, { name: 'HMAC', hash: 'SHA-256' }, false, ['sign'])
}

/**
 * Request target as fetch sends it: the path and query of the URL. Pass the same
 * value to signDashboardUser for the request it signs.
 */
export function requestTarget(url: string): string {
  const u = new URL(url)
  return `${u.pathname}${u.search}`
}

/**
 * Produces the signed header value for X-MountOS-Dashboard-User for one request.
 * method is the HTTP method and target is the request path and query (see
 * requestTarget). Sets exp to now + 10 minutes and a fresh nonce. hmacSecret is the
 * dedicated dashboard-user HMAC secret shared with appserv (DASHBOARD_USER_HMAC_KEY);
 * it is a separate secret from the admin signing key, never the public verification key.
 * Format: base64url(json).base64url(hmac-sha256(base64url(json), derived_key))
 */
export async function signDashboardUser(
  user: DashboardUser,
  hmacSecret: string,
  method: string,
  target: string,
): Promise<string> {
  if (!hmacSecret) throw new Error('mountos: dashboard HMAC secret is required to sign a dashboard user header')

  const nonce = new Uint8Array(16)
  crypto.getRandomValues(nonce)
  const withExp = {
    ...user,
    exp: Math.floor((Date.now() + DEFAULT_TTL_MS) / 1000),
    method,
    requestUri: target,
    nonce: toBase64Url(nonce),
  }
  const payload = new TextEncoder().encode(JSON.stringify(withExp))
  const encoded = toBase64Url(payload)

  const key = await deriveDashboardHMACKey(hmacSecret)
  const sig = new Uint8Array(await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(encoded)))
  return `${encoded}.${toBase64Url(sig)}`
}
