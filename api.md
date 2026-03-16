# mountOS Admin API Reference

Base path: `/api/v1`
Auth: `Authorization: Bearer <JWT>` (ED25519/EdDSA, sub=vendor, aud=mountos/appserv)

## Response Envelope

All responses use `StandardResponse`:
```
{ "status": "success"|"failure", "message": string, "data"?: object, "errorCode"?: int }
```

Paginated responses nest in `data`:
```
{ "items": T[], "pagination": { "page": int, "limit": int, "total": int64, "totalPages": int64 } }
```

Cursor-paginated responses nest in `data`:
```
{ "items": T[], "nextCursor": int64|null }
```

## Error Codes (AppServ 1XXXX)

| Code  | Name                   |
|-------|------------------------|
| 10001 | AUTHENTICATION_REQUIRED |
| 10002 | INVALID_SESSION        |
| 10003 | INVALID_CREDENTIALS    |
| 10004 | SESSION_EXPIRED        |
| 10200 | INVALID_REQUEST_FORMAT |
| 10201 | VALIDATION_FAILED      |
| 10202 | MISSING_PARAMETER      |
| 10900 | INTERNAL_ERROR         |
| 10901 | SERVICE_UNAVAILABLE    |
| 10902 | DATABASE_ERROR         |

---

## Accounts

### POST /api/v1/accounts/create
Request:
```
{
  "name": string(required),
  "description"?: string,
  "iconUrl"?: string,
  "vendorInfo"?: object
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/accounts/list
Query: `page=int(default 1)`, `limit=int(default 10)`
Response data: `{ "items": Account[], "pagination": PaginationMeta }`

### GET /api/v1/accounts/:accountId
Param: `accountId`
Response data: `Account`

### PUT /api/v1/accounts/:accountId/edit
Param: `accountId`
Request:
```
{
  "name": string(required),
  "description"?: string,
  "iconUrl"?: string,
  "vendorInfo"?: object
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/accounts/:accountId/lock
Param: `accountId`
Response data: `{ "id": int64 }`

### POST /api/v1/accounts/:accountId/unlock
Param: `accountId`
Response data: `{ "id": int64 }`

### POST /api/v1/accounts/:accountId/activate
Param: `accountId`
Response data: `{ "id": int64 }`

### POST /api/v1/accounts/:accountId/deactivate
Param: `accountId`
Response data: `{ "id": int64 }`

### Account Type
```
{
  "id": int64,
  "name": string,
  "description": string,
  "iconUrl"?: string,
  "vendorInfo"?: object,
  "isActive": bool,
  "locked": bool,
  "createdAt": RFC3339,
  "updatedAt": RFC3339
}
```

---

## Users

### POST /api/v1/users/add
Request:
```
{
  "accountId": int64(required),
  "username": string(required),
  "email": string(required),
  "name"?: string,
  "vendorInfo"?: object
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/users/list
Query: `accountId=int64(required)`, `page=int(default 1)`, `limit=int(default 10)`
Response data: `{ "items": User[], "pagination": PaginationMeta }`

### GET /api/v1/users/:userId
Param: `userId`
Response data: `User`

### PUT /api/v1/users/:userId/edit
Param: `userId`
Request:
```
{
  "username": string(required),
  "email": string(required),
  "name"?: string,
  "vendorInfo"?: object
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/users/:userId/activate
Param: `userId`
Response data: `{ "id": int64 }`

### POST /api/v1/users/:userId/deactivate
Param: `userId`
Response data: `{ "id": int64 }`

### User Type
```
{
  "id": int64,
  "accountId": int64,
  "username": string,
  "email": string,
  "name": string,
  "isActive": bool
}
```

---

## Regions

### POST /api/v1/regions/create
Request:
```
{
  "accountId": int64(required),
  "name": string(required),
  "dns": string(required)
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/regions/list
Query: `page=int(default 1)`, `limit=int(default 10)`
Response data: `{ "items": Region[], "pagination": PaginationMeta }`

### GET /api/v1/regions/:regionId
Param: `regionId`
Response data: `Region`

### PUT /api/v1/regions/:regionId/edit
Param: `regionId`
Request:
```
{
  "accountId": int64(required),
  "name": string(required),
  "dns": string(required)
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/regions/:regionId/activate
Param: `regionId`
Response data: `{ "id": int64 }`

### POST /api/v1/regions/:regionId/deactivate
Param: `regionId`
Response data: `{ "id": int64 }`

### Region Type
```
{
  "id": int64,
  "exportId": string,
  "accountId": int64,
  "name": string,
  "dns": string,
  "isActive": bool,
  "createdAt": RFC3339,
  "updatedAt": RFC3339
}
```

---

## Storages

### POST /api/v1/storages/create
Request:
```
{
  "accountId": int64(required),
  "regionId": int64(required),
  "name": string(required),
  "description"?: string,
  "storageType": string(required),
  "providerType": string(required),
  "endpoint": string(required),
  "region"?: string,
  "bucket"?: string,
  "base"?: string,
  "blockRegion"?: string,
  "blockType"?: string,
  "blockSize"?: int32,
  "accessKey"?: string,
  "secretKey"?: string
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/storages/list
Query: `accountId=int64(required)`, `page=int(default 1)`, `limit=int(default 10)`
Response data: `{ "items": Storage[], "pagination": PaginationMeta }`

### GET /api/v1/storages/:storageId
Param: `storageId`
Response data: `Storage`

### PUT /api/v1/storages/:storageId/edit
Param: `storageId`
Request:
```
{
  "name": string(required),
  "description"?: string,
  "endpoint"?: string,
  "accessKey"?: string,
  "secretKey"?: string
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/storages/:storageId/activate
Param: `storageId`
Response data: `{ "id": int64 }`

### POST /api/v1/storages/:storageId/deactivate
Param: `storageId`
Response data: `{ "id": int64 }`

### POST /api/v1/storages/test-bucket
Request:
```
{
  "endpoint": string(required),
  "region"?: string,
  "bucket": string(required),
  "accessKey": string(required),
  "secretKey": string(required)
}
```
Response data: `{ "bucketExists": bool, "list": bool, "write": bool, "read": bool, "delete": bool, "multipart": bool }`

### Storage Type
```
{
  "id": int64,
  "accountId": int64,
  "regionId": int64,
  "name": string,
  "description"?: string,
  "storageType": string,
  "providerType": string,
  "blockType"?: string,
  "endpoint": string,
  "region"?: string,
  "bucket"?: string,
  "base"?: string,
  "blockRegion"?: string,
  "blockSize"?: int32,
  "isActive": bool,
  "createdAt": RFC3339,
  "updatedAt": RFC3339
}
```

---

## Volumes

### POST /api/v1/volumes/create
Request:
```
{
  "accountId": int64(required),
  "storageId": int64(required),
  "name": string(required),
  "description"?: string,
  "volumeType": string(required),
  "encryption"?: bool,
  "encryptionKey"?: string,
  "retentionPeriod"?: int32,
  "gracePeriod"?: int32,
  "gcOnDeactivation"?: bool,
  "quotaLimit"?: int64
}
```
Response data: `{ "id": int64, "encryptionKey": string }`

### GET /api/v1/volumes/list
Query: `accountId=int64(required)`, `page=int(default 1)`, `limit=int(default 10)`
Response data: `{ "items": Volume[], "pagination": PaginationMeta }`

### GET /api/v1/volumes/:volumeId
Param: `volumeId`
Response data: `Volume`

### PUT /api/v1/volumes/:volumeId/edit
Param: `volumeId`
Request:
```
{
  "name": string(required),
  "description"?: string,
  "encryption"?: bool,
  "retentionPeriod"?: int32,
  "gracePeriod"?: int32,
  "gcOnDeactivation"?: bool
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/lock
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/unlock
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/activate
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/deactivate
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/api-keys/generate
Param: `volumeId`
Request:
```
{
  "userId": int64(required)
}
```
Response data: `{ "apiKey": string, "apiSecret": string }`

### POST /api/v1/volumes/:volumeId/api-keys/revoke
Param: `volumeId`
Request:
```
{
  "apiKey": string(required)
}
```

### PUT /api/v1/volumes/:volumeId/quota
Param: `volumeId`
Request:
```
{
  "quotaLimit": int64(required)
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/volumes/:volumeId/stats
Param: `volumeId`
Response data: `{ "volumeId": string, "diskSize": int64, "activeSize": int64, "size": int64 }`

### Volume Type
```
{
  "id": int64,
  "accountId": int64,
  "storageId": int64,
  "regionId": int64,
  "name": string,
  "description"?: string,
  "encryption": bool,
  "quotaLimit": int64,
  "quotaUsed": int64,
  "locked": bool,
  "isActive": bool,
  "createdAt": RFC3339,
  "updatedAt": RFC3339
}
```

---

## AuditLogs

### GET /api/v1/audit-logs/list
Query: `accountId=int64`, `cursor=int64`, `limit=int(default 20)`, `subject=string`
Response data: `{ "items": AuditLog[], "nextCursor": int64|null }`

### AuditLog Type
```
{
  "id": int64,
  "title": string,
  "description"?: string,
  "subject"?: string,
  "success": bool,
  "data"?: object,
  "createdBy"?: string,
  "accountId"?: int64,
  "createdAt"?: RFC3339,
  "updatedAt"?: RFC3339
}
```

---

## ServiceNodes

### GET /api/v1/regions/:regionId/nodes
Param: `regionId`
Response data: `ServiceNode[]`

### POST /api/v1/regions/:regionId/nodes/:nodeId/drain
Param: `regionId`
Param: `nodeId`

### POST /api/v1/regions/:regionId/nodes/:nodeId/activate
Param: `regionId`
Param: `nodeId`

### DELETE /api/v1/regions/:regionId/nodes/:nodeId
Param: `regionId`
Param: `nodeId`

### ServiceNode Type
```
{
  "id": int64,
  "regionId": int64,
  "serviceType": string,
  "nodeId": string,
  "advertiseAddr": string,
  "httpAddr"?: string,
  "metadata"?: object,
  "status": string,
  "lastHeartbeat"?: string,
  "isActive": bool
}
```

---

## ClientSessions

### GET /api/v1/client-sessions/list
Query: `accountId=int64`, `regionId=int64`, `clientType=string`, `status=int`, `page=int(default 1)`, `limit=int(default 20)`
Response data: `{ "items": ClientSession[], "pagination": PaginationMeta }`

### GET /api/v1/client-sessions/:sessionId
Param: `sessionId`
Response data: `ClientSession`

### GET /api/v1/client-sessions/summary
Response data: `SessionSummary[]`

### ClientSession Type
```
{
  "id": int64,
  "accountId": int64,
  "volumeId": string,
  "regionId": int64,
  "userId"?: string,
  "clientType": string,
  "osName": string,
  "osVersion"?: string,
  "appVersion"?: string,
  "hostname"?: string,
  "ipAddr": string,
  "mountMode"?: string,
  "mountPath"?: string,
  "metadata"?: object,
  "metrics"?: object,
  "status": string,
  "lastHeartbeat"?: RFC3339,
  "connectedAt"?: RFC3339,
  "disconnectedAt"?: RFC3339,
  "isActive": bool
}
```

---

## Discover

### GET /api/v1/discover/meta
Query: `access_key_id=string(required)`
Response data: `DiscoverMetaResponse`

---

## Dashboard

### GET /api/v1/dashboard/stats
Query: `accountId=int64(required)`
Response data: `DashboardStats`

---

## Cache

### POST /api/v1/cache/refresh

---

## JWT Construction

```
Header:  { "alg": "EdDSA", "typ": "JWT" }
Payload: {
  "sub": "vendor",
  "aud": ["mountos/appserv"],
  "iat": unix_now,
  "nbf": unix_now,
  "exp": unix_now + 3600,
  "jti": "<nanosecond_timestamp_string>",
  "scope": "service",
  "kfp": "<hex(sha256(ed25519_pubkey)[:16])>"
}
Signature: ED25519 sign(header.payload, privateKey)
```

Key format: raw 64-byte ED25519 private key, base64-encoded (standard encoding).

## PaginationMeta Type
```
{ "page": int, "limit": int, "total": int64, "totalPages": int64 }
```
