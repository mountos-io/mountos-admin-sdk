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
| 10002 | INVALID_SESSION         |
| 10003 | INVALID_CREDENTIALS     |
| 10004 | SESSION_EXPIRED         |
| 10200 | INVALID_REQUEST_FORMAT  |
| 10201 | VALIDATION_FAILED       |
| 10202 | MISSING_PARAMETER       |
| 10900 | INTERNAL_ERROR          |
| 10901 | SERVICE_UNAVAILABLE     |
| 10902 | DATABASE_ERROR          |

---

## Accounts

### POST /api/v1/accounts/create
Request:
```
{ "name": string(required,1-255), "description"?: string(max 1000), "vendorInfo"?: object }
```
Response data: `{ "id": int64 }`

### GET /api/v1/accounts/list
Query: `page=int(default 1)`, `limit=int(default 10, max 100)`
Response data: `{ "items": Account[], "pagination": PaginationMeta }`

### GET /api/v1/accounts/:accountId
Param: `accountId` (int64)
Response data: `Account`

### PUT /api/v1/accounts/:accountId/edit
Param: `accountId` (int64)
Request:
```
{ "name": string(required,1-255), "description"?: string(max 1000), "vendorInfo"?: object }
```
Response data: `{ "id": int64 }`

### POST /api/v1/accounts/:accountId/lock
Param: `accountId` (int64)
Response data: `{ "id": int64 }`

### POST /api/v1/accounts/:accountId/unlock
Param: `accountId` (int64)
Response data: `{ "id": int64 }`

### POST /api/v1/accounts/:accountId/activate
Param: `accountId` (int64)
Response data: `{ "id": int64 }`

### POST /api/v1/accounts/:accountId/deactivate
Param: `accountId` (int64)
Response data: `{ "id": int64 }`

### Account Type
```
{
  "id": int64, "name": string, "description": string,
  "vendorInfo"?: object, "isActive": bool, "locked": bool,
  "createdAt": RFC3339, "updatedAt": RFC3339
}
```

---

## Users

### POST /api/v1/users/add
Request:
```
{
  "accountId": int64(required), "username": string(required,1-255),
  "email": string(required,email,max 255), "name"?: string(max 255),
  "vendorInfo"?: object
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/users/list
Query: `accountId=int64(required)`, `page=int(default 1)`, `limit=int(default 10, max 100)`
Response data: `{ "items": User[], "pagination": PaginationMeta }`

### GET /api/v1/users/:userId
Param: `userId` (int64)
Response data: `User`

### PUT /api/v1/users/:userId/edit
Param: `userId` (int64)
Request:
```
{
  "username": string(required,1-255), "email": string(required,email,max 255),
  "name"?: string(max 255), "vendorInfo"?: object
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/users/:userId/activate
Param: `userId` (int64)
Response data: `{ "id": int64 }`

### POST /api/v1/users/:userId/deactivate
Param: `userId` (int64)
Response data: `{ "id": int64 }`

### User Type (returned by get/list)
```
{
  "id": int64, "accountId": int64, "username": string,
  "email": string, "name": string
}
```

---

## Regions

### POST /api/v1/regions/create
Request:
```
{ "accountId": int64(required), "name": string(required,1-255), "dns": string(required,1-512) }
```
Response data: `{ "id": int64 }`

### GET /api/v1/regions/list
Query: `page=int(default 1)`, `limit=int(default 10, max 100)`
Response data: `{ "items": Region[], "pagination": PaginationMeta }`

### GET /api/v1/regions/:regionId
Param: `regionId` (int64)
Response data: `Region`

### PUT /api/v1/regions/:regionId/edit
Param: `regionId` (int64)
Request:
```
{ "accountId": int64(required), "name": string(required,1-255), "dns": string(required,1-512) }
```
Response data: `{ "id": int64 }`

### POST /api/v1/regions/:regionId/activate
Param: `regionId` (int64)
Response data: `{ "id": int64 }`

### POST /api/v1/regions/:regionId/deactivate
Param: `regionId` (int64)
Response data: `{ "id": int64 }`

### Region Type
```
{
  "id": int64, "accountId": int64, "name": string, "dns": string,
  "isActive": bool, "createdAt": RFC3339, "updatedAt": RFC3339
}
```

---

## Storages

### POST /api/v1/storages/create
Request:
```
{
  "accountId": int64(required), "regionId": int64(required),
  "name": string(required,1-100), "description"?: string(max 1000),
  "storageType": "object"|"block"(required), "providerType": string(required,max 50),
  "endpoint": string(required,max 255), "region"?: string(max 100),
  "bucket"?: string(max 100), "base"?: string(max 255),
  "blockRegion"?: string(max 255), "blockType"?: "standard"|"passthrough",
  "blockSize"?: int32, "accessKey"?: string, "secretKey"?: string
}
```
Response data: `{ "id": string(UUID), "shardId": int64 }`

### GET /api/v1/storages/list
Query: `accountId=int64(required)`, `page=int(default 1)`, `limit=int(default 10, max 100)`
Response data: `{ "items": Storage[], "pagination": PaginationMeta }`

### GET /api/v1/storages/:storageId
Param: `storageId` (UUID string)
Response data: `Storage`

### PUT /api/v1/storages/:storageId/edit
Param: `storageId` (UUID string)
Request:
```
{
  "name": string(required,1-100), "description"?: string(max 1000),
  "endpoint"?: string(max 255), "accessKey"?: string, "secretKey"?: string
}
```
Response data: `{ "id": string(UUID) }`

### POST /api/v1/storages/:storageId/activate
Param: `storageId` (UUID string)
Response data: `{ "id": string(UUID) }`

### POST /api/v1/storages/:storageId/deactivate
Param: `storageId` (UUID string)
Response data: `{ "id": string(UUID) }`

### Storage Type
```
{
  "id": string(UUID), "shardId": int64, "accountId": int64, "regionId": int64,
  "name": string, "description"?: string, "storageType": string,
  "providerType": string, "blockType"?: string, "endpoint": string,
  "region"?: string, "bucket"?: string, "base"?: string,
  "blockRegion"?: string, "blockSize"?: int32,
  "isActive": bool, "createdAt": RFC3339, "updatedAt": RFC3339
}
```

---

## Volumes (stubs — return 501)

### POST /api/v1/volumes/create
Status: 501 Not Implemented

### GET /api/v1/volumes/list
Status: 501 Not Implemented

### GET /api/v1/volumes/:volumeId
Param: `volumeId` (UUID string)
Status: 501 Not Implemented

### PUT /api/v1/volumes/:volumeId/edit
Param: `volumeId` (UUID string)
Status: 501 Not Implemented

### POST /api/v1/volumes/:volumeId/lock
Param: `volumeId` (UUID string)
Status: 501 Not Implemented

### POST /api/v1/volumes/:volumeId/unlock
Param: `volumeId` (UUID string)
Status: 501 Not Implemented

### POST /api/v1/volumes/:volumeId/activate
Param: `volumeId` (UUID string)
Status: 501 Not Implemented

### POST /api/v1/volumes/:volumeId/deactivate
Param: `volumeId` (UUID string)
Status: 501 Not Implemented

### PUT /api/v1/volumes/:volumeId/quota
Param: `volumeId` (UUID string)
Request:
```
{ "quotaLimit": int64(required, >=0) }
```
Response data: `{ "id": string(UUID) }`

---

## Audit Logs

### GET /api/v1/audit-logs/list
Query: `accountId=int64`, `cursor=int64`, `limit=int(default 20, max 100)`, `subject=string`
Response data: `{ "items": AuditLog[], "nextCursor": int64|null }`

### AuditLog Type
```
{
  "id": int64, "title": string, "description"?: string, "subject"?: string,
  "success": bool, "data"?: object, "createdBy"?: string, "accountId"?: string,
  "createdAt"?: RFC3339, "updatedAt"?: RFC3339
}
```

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
  "scope": "service"
}
Signature: ED25519 sign(header.payload, privateKey)
```

Key format: raw 64-byte ED25519 private key, base64-encoded (standard encoding).

## PaginationMeta Type
```
{ "page": int, "limit": int, "total": int64, "totalPages": int64 }
```
