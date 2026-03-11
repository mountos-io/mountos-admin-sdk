import { SignJWT, importPKCS8 } from 'jose'

const TOKEN_TTL = 3600
const REFRESH_MARGIN = 300

function ed25519PemFromRaw(raw64: Uint8Array): string {
  const pkcs8Prefix = new Uint8Array([
    0x30, 0x2e, 0x02, 0x01, 0x00, 0x30, 0x05, 0x06,
    0x03, 0x2b, 0x65, 0x70, 0x04, 0x22, 0x04, 0x20,
  ])
  const seed = raw64.subarray(0, 32)
  const der = new Uint8Array(pkcs8Prefix.length + seed.length)
  der.set(pkcs8Prefix)
  der.set(seed, pkcs8Prefix.length)
  const b64 = btoa(String.fromCharCode(...der))
  return `-----BEGIN PRIVATE KEY-----\n${b64}\n-----END PRIVATE KEY-----`
}

// keyFingerprint computes hex(sha256(pubkey)[:16]) matching the server-side KeyFingerprint().
async function keyFingerprint(raw: Uint8Array): Promise<string> {
  const pub = raw.subarray(32, 64)
  const hash = new Uint8Array(await crypto.subtle.digest('SHA-256', pub as ArrayBufferView<ArrayBuffer>))
  return Array.from(hash.subarray(0, 16), b => b.toString(16).padStart(2, '0')).join('')
}

export class TokenSigner {
  private token?: string
  private expiry = 0
  private key?: CryptoKey
  private kfp?: string
  private readonly privateKeyBase64: string

  constructor(privateKeyBase64: string) {
    this.privateKeyBase64 = privateKeyBase64
  }

  async getToken(): Promise<string> {
    const now = Math.floor(Date.now() / 1000)
    if (this.token && now < this.expiry - REFRESH_MARGIN) {
      return this.token
    }

    if (!this.key) {
      const raw = Uint8Array.from(atob(this.privateKeyBase64), c => c.charCodeAt(0))
      const pem = ed25519PemFromRaw(raw)
      this.key = await importPKCS8(pem, 'EdDSA')
      this.kfp = await keyFingerprint(raw)
    }

    const exp = now + TOKEN_TTL
    const token = await new SignJWT({ scope: 'service', kfp: this.kfp })
      .setProtectedHeader({ alg: 'EdDSA', typ: 'JWT' })
      .setSubject('vendor')
      .setAudience('mountos/appserv')
      .setIssuedAt(now)
      .setNotBefore(now)
      .setExpirationTime(exp)
      .setJti(crypto.randomUUID())
      .sign(this.key)

    this.token = token
    this.expiry = exp
    return token
  }
}
