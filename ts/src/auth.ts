import { SignJWT, importPKCS8, exportJWK } from 'jose'

const TOKEN_TTL = 3600
const REFRESH_MARGIN = 300
const CLOCK_SKEW_LEEWAY = 5

function decodeBase64(base64: string): Uint8Array {
  const bin = atob(base64)
  const out = new Uint8Array(bin.length)
  for (let i = 0; i < bin.length; i++) out[i] = bin.charCodeAt(i)
  return out
}

function encodeBase64(bytes: Uint8Array): string {
  let bin = ''
  for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i])
  return btoa(bin)
}

function decodeBase64Url(b64url: string): Uint8Array {
  const b64 = b64url.replace(/-/g, '+').replace(/_/g, '/') + '=='.slice(0, (4 - b64url.length % 4) % 4)
  return decodeBase64(b64)
}

function ed25519SeedFrom(raw: Uint8Array): Uint8Array {
  if (raw.length === 32) return raw
  if (raw.length === 64) return raw.subarray(0, 32)
  throw new Error(`invalid Ed25519 private key length: expected 32 or 64 bytes, got ${raw.length}`)
}

function ed25519Pkcs8PemFromSeed(seed: Uint8Array): string {
  const pkcs8Prefix = new Uint8Array([
    0x30, 0x2e, 0x02, 0x01, 0x00, 0x30, 0x05, 0x06,
    0x03, 0x2b, 0x65, 0x70, 0x04, 0x22, 0x04, 0x20,
  ])

  const der = new Uint8Array(pkcs8Prefix.length + seed.length)
  der.set(pkcs8Prefix)
  der.set(seed, pkcs8Prefix.length)

  const b64 = encodeBase64(der)
  const wrapped = b64.match(/.{1,64}/g)?.join('\n') ?? b64

  return `-----BEGIN PRIVATE KEY-----\n${wrapped}\n-----END PRIVATE KEY-----`
}

async function keyFingerprint(key: CryptoKey): Promise<string> {
  const jwk = await exportJWK(key)
  const pub = decodeBase64Url(jwk.x!)
  const digest = new Uint8Array(await crypto.subtle.digest('SHA-256', pub as ArrayBufferView<ArrayBuffer>))
  return Array.from(digest.subarray(0, 16), b => b.toString(16).padStart(2, '0')).join('')
}

export class TokenSigner {
  private token?: string
  private expiry = 0
  private key?: CryptoKey
  private kfp?: string
  private refreshPromise?: Promise<string>

  constructor(private readonly privateKeyBase64: string) {}

  async getToken(): Promise<string> {
    const now = Math.floor(Date.now() / 1000)

    if (this.token && now < this.expiry - REFRESH_MARGIN) {
      return this.token
    }

    if (this.refreshPromise) {
      return this.refreshPromise
    }

    this.refreshPromise = this.refreshToken(now)

    try {
      return await this.refreshPromise
    } finally {
      this.refreshPromise = undefined
    }
  }

  private async refreshToken(now: number): Promise<string> {
    if (!this.key || !this.kfp) {
      const seed = ed25519SeedFrom(decodeBase64(this.privateKeyBase64))
      const pem = ed25519Pkcs8PemFromSeed(seed)

      this.key = await importPKCS8(pem, 'EdDSA', { extractable: true })
      this.kfp = await keyFingerprint(this.key)
    }

    const exp = now + TOKEN_TTL

    // hub understands only one role: user and rest of them are admin
    // since we dont pass in role info, token level opt-in auth will be ignored
    // DashboardUser (opt-in) may provide user level auth through http header
    const token = await new SignJWT({
      scope: 'service',
      kfp: this.kfp,
    })
      .setProtectedHeader({ alg: 'EdDSA', typ: 'JWT', kid: this.kfp })
      .setSubject('mountos:provider')
      .setAudience('mountos/appserv')
      .setIssuedAt(now)
      .setNotBefore(now - CLOCK_SKEW_LEEWAY)
      .setExpirationTime(exp)
      .setJti(crypto.randomUUID())
      .sign(this.key)

    this.token = token
    this.expiry = exp
    return token
  }
}
