# Skill: Port mountOS Admin SDK to a New Language

Use this skill to generate an SDK for the mountOS Admin API in any language.

## Steps

1. Read `api.md` at repo root for full API surface (endpoints, types, envelopes)
2. Read existing implementations as reference:
   - **TypeScript**: `ts/src/` — client, auth, types, errors
   - **Go**: `go/` — client, auth, types, errors, per-resource files
3. Implement the SDK following the patterns below

## Auth: JWT Construction (ED25519)

Sign-only. No verification needed.

```
Header:  {"alg":"EdDSA","typ":"JWT"}
Payload: {"sub":"vendor","aud":["mountos/appserv"],"iat":<now>,"nbf":<now>,"exp":<now+3600>,"jti":"<nanos>","scope":"service"}
Signature: ed25519_sign(base64url(header) + "." + base64url(payload), privateKey)
Token:     base64url(header) + "." + base64url(payload) + "." + base64url(signature)
```

- Key input: base64-encoded 64-byte ED25519 private key (standard encoding, not URL-safe)
- Cache token, re-sign when <5min until expiry

## Response Envelope

All API responses: `{ "status": "success"|"failure", "message": string, "data"?: T, "errorCode"?: int }`

SDK must:
- On `status=success`: return `data`
- On `status=failure`: throw/return typed error with `message`, `errorCode`, HTTP status

## Error Type

Fields: `message` (string), `status` (HTTP int), `errorCode` (optional int from response)

## Client Config

```
baseUrl:    string   — appserv base URL (no trailing slash)
privateKey: string   — base64-encoded ED25519 private key
```

## Resource Namespaces

Group methods under resource namespaces: `client.accounts.*`, `client.users.*`, etc.

## Naming Conventions

| Concept | TypeScript | Go | Python | Rust | Java |
|---------|------------|-----|--------|------|------|
| Client class | `MountOSAdmin` | `Client` | `MountOSAdmin` | `Client` | `MountOSAdmin` |
| Create | `accounts.create()` | `Accounts.Create()` | `accounts.create()` | `accounts().create()` | `accounts().create()` |
| List | `accounts.list()` | `Accounts.List()` | `accounts.list()` | `accounts().list()` | `accounts().list()` |
| Error class | `MountOSError` | `Error` | `MountOSError` | `Error` | `MountOSException` |

## Module/Package Name

- Python: `mountos-admin-sdk` (pip), `mountos_admin`
- Rust: `mountos-admin-sdk` (crates.io), `mountos_admin`
- Java: `com.mountos.admin`
- Ruby: `mountos-admin-sdk` (gem), `MountOS::Admin`

## Checklist

- [ ] JWT construction (ED25519 sign-only, cached)
- [ ] HTTP client with Authorization header
- [ ] Response envelope unwrapping
- [ ] Typed errors with errorCode
- [ ] All resource methods from api.md
- [ ] Pagination support (page-based + cursor-based)
- [ ] Request/response types matching api.md exactly
