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
  "liveVolume": int64,
  "totalVolume": int64,
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
Query: `accountId=int64(required)`, `search=string`, `page=int(default 1)`, `limit=int(default 10)`
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
  "liveVolume": int64,
  "totalVolume": int64,
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
Query: `accountId=int64(required)`, `search=string`, `regionId=int64`, `storageType=string`, `providerType=string`, `page=int(default 1)`, `limit=int(default 10)`
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
  "secretKey": string(required),
  "providerType"?: string
}
```
Response data: `{ "bucketExists": bool, "list": bool, "write": bool, "read": bool, "delete": bool, "multipart": bool }`

### POST /api/v1/storages/:storageId/test-bucket
Param: `storageId`
Response data: `{ "bucketExists": bool, "list": bool, "write": bool, "read": bool, "delete": bool, "multipart": bool }`

### Storage Type
```
{
  "id": int64,
  "account": Ref,
  "regionInfo": Ref,
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
  "quotaLimit"?: int64
}
```
Response data: `{ "id": int64, "encryptionKey": string }`

### GET /api/v1/volumes/list
Query: `accountId=int64(required)`, `regionId=int64`, `storageId=int64`, `page=int(default 1)`, `limit=int(default 10)`
Response data: `{ "items": Volume[], "pagination": PaginationMeta }`

### GET /api/v1/volumes/:volumeId
Param: `volumeId`
Response data: `Volume`

### PUT /api/v1/volumes/:volumeId/edit
Param: `volumeId`
Request:
```
{
  "description"?: string,
  "retentionPeriod"?: int32,
  "gracePeriod"?: int32
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/lock
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/unlock
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/deactivate
Param: `volumeId`
Request:
```
{
  "isCleanupMetaEnabled"?: bool,
  "isCleanupStorageEnabled"?: bool,
  "isCleanupVaultEnabled"?: bool
}
```
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

### POST /api/v1/volumes/:volumeId/api-keys/revoke-by-user
Param: `volumeId`
Request:
```
{
  "userId": int64(required)
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
Response data: `{ "volumeId": string, "liveVolume": int64, "totalVolume": int64, "pendingVolume": int64 }`

### GET /api/v1/volumes/:volumeId/forks
Param: `volumeId`
Response data: `Fork[]`

### Volume Type
```
{
  "id": int64,
  "account": Ref,
  "storage": Ref,
  "region": Ref,
  "name": string,
  "description"?: string,
  "encryption": bool,
  "quotaLimit": int64,
  "liveVolume": int64,
  "totalVolume": int64,
  "pendingVolume": int64,
  "locked": bool,
  "retentionPeriod": int32,
  "gracePeriod": int32,
  "isActive": bool,
  "isCleanupMetaEnabled": bool,
  "isCleanupStorageEnabled": bool,
  "isCleanupVaultEnabled": bool,
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

## RegionAuditLogs

### GET /api/v1/regions/:regionId/audit-logs/list
Param: `regionId`
Query: `cursor=int64`, `limit=int(default 20)`, `subject=string`
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
Query: `serviceType=string`, `status=string`, `inactiveHours=int`
Response data: `ServiceNode[]`

### GET /api/v1/regions/:regionId/nodes/:nodeId/stats
Param: `regionId`
Param: `nodeId`
Response data: `string`

### ServiceNode Type
```
{
  "id": int64,
  "regionId": int64,
  "serviceType": string,
  "nodeId": string,
  "advertiseAddr": string,
  "rpcAddr"?: string,
  "metadata"?: object,
  "status": string,
  "lastHeartbeat"?: int64,
  "isActive": bool
}
```

---

## Nodes

### GET /api/v1/nodes
Query: `serviceType=string`, `status=string`, `inactiveHours=int`
Response data: `ServiceNode[]`

### ServiceNode Type
```
{
  "id": int64,
  "regionId": int64,
  "serviceType": string,
  "nodeId": string,
  "advertiseAddr": string,
  "rpcAddr"?: string,
  "metadata"?: object,
  "status": string,
  "lastHeartbeat"?: int64,
  "isActive": bool
}
```

---

## ClientSessions

### GET /api/v1/client-sessions/list
Query: `accountId=int64`, `regionId=int64`, `volumeId=int64`, `userId=int64`, `clientType=string`, `status=string`, `isActive=string`, `page=int(default 1)`, `limit=int(default 20)`
Response data: `{ "items": ClientSession[], "pagination": PaginationMeta }`

### GET /api/v1/client-sessions/:sessionId
Param: `sessionId`
Response data: `ClientSession`

### GET /api/v1/client-sessions/summary
Query: `accountId=int64`, `volumeId=int64`
Response data: `SessionSummary[]`

### ClientSession Type
```
{
  "id": int64,
  "account": Ref,
  "region": Ref,
  "volume": Ref,
  "user"?: Ref,
  "clientType": string,
  "osName": string,
  "osVersion"?: string,
  "appVersion"?: string,
  "hostname"?: string,
  "ipAddr": string,
  "mountMode"?: string,
  "mountPath"?: string,
  "forkName"?: string,
  "isTemporaryFork": bool,
  "metadata"?: object,
  "metrics"?: object,
  "status": string,
  "lastHeartbeat"?: int64,
  "connectedAt"?: int64,
  "disconnectedAt"?: int64,
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

## License

### GET /api/v1/license
Response data: `LicenseDetails`

### LicenseDetails Type
```
{
  "licenseId": string,
  "licensee": string,
  "contact": string,
  "licenseType": string,
  "issuedAt": string,
  "expiresAt": string,
  "gracePeriodDays": int,
  "maxNodes": int64,
  "maxVolumes": int64,
  "maxUsers": int64,
  "maxStorageBytes": int64,
  "status": LicenseStatus,
  "daysRemaining": int,
  "graceEndsAt": string,
  "graceDaysLeft": int
}
```

---

## Alerts

### GET /api/v1/alerts/list
Query: `active=bool(default true)`, `page=int(default 1)`, `limit=int(default 20)`
Response data: `{ "items": ServiceAlert[], "pagination": PaginationMeta }`

### POST /api/v1/alerts/:alertId/resolve
Param: `alertId`

### ServiceAlert Type
```
{
  "id": int64,
  "alertId": string,
  "source": string,
  "nodeId": string,
  "severity": int,
  "category": string,
  "title": string,
  "description"?: string,
  "regionId"?: int64,
  "accountId"?: int64,
  "eventTime": RFC3339,
  "resolvedAt"?: RFC3339,
  "createdAt"?: RFC3339
}
```

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
