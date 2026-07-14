# mountOS Admin API Reference

Base path: `/api/v1`
Auth: `Authorization: Bearer <JWT>` (ED25519/EdDSA, sub=mountos:provider, aud=mountos/appserv)

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

## Enums

`ClientSessionStatus`: "connected" | "active" | "unhealthy" | "disconnected" | "expired" | "unknown"

`LicenseQuotaState`: "ok" | "exceeded"

`LicenseStatus`: "valid" | "expiring" | "grace" | "expired_access" | "expired"

---

## Accounts

### POST /api/v1/accounts/create
Request:
```
{
  "name": string(required),
  "description"?: string,
  "iconUrl"?: string,
  "providerInfo"?: object
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/accounts/list
Query: `isActive=bool`, `page=int(default 1)`, `limit=int(default 10)`
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
  "providerInfo"?: object
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

### PUT /api/v1/accounts/:accountId/quota
Param: `accountId`
Request:
```
{
  "quotaLimit": int64(required),
  "quotaExcessPct"?: int32
}
```
Response data: `{ "id": int64 }`

### Account Type
```
{
  "id": int64,
  "name": string,
  "description": string,
  "iconUrl"?: string,
  "providerInfo"?: object,
  "liveVolume": int64,
  "totalVolume": int64,
  "quotaLimit": int64,
  "quotaExcessPct": int32,
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
  "providerInfo"?: object
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/users/list
Query: `accountId=int64(required)`, `search=string`, `isActive=bool`, `page=int(default 1)`, `limit=int(default 10)`
Response data: `{ "items": User[], "pagination": PaginationMeta }`

### GET /api/v1/users/:userId
Param: `userId`
Response data: `User`

### POST /api/v1/users/bulk
Request:
```
{
  "ids": int64[](required)
}
```
Response data: `{ "users": UserLite[] }`

### PUT /api/v1/users/:userId/edit
Param: `userId`
Request:
```
{
  "username": string(required),
  "email": string(required),
  "name"?: string,
  "providerInfo"?: object
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
Query: `accountId=int64(required)`, `isActive=bool`, `page=int(default 1)`, `limit=int(default 10)`
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

## Clusters

### GET /api/v1/clusters/list
Query: `accountId=int64(required)`, `regionId=int64`, `isActive=bool`, `page=int(default 1)`, `limit=int(default 100)`
Response data: `{ "items": RegionCluster[], "pagination": PaginationMeta }`

### RegionCluster Type
```
{
  "id": int64,
  "exportId": string,
  "regionId": int64,
  "name": string,
  "defaultCluster": bool,
  "isReady": bool,
  "isActive": bool,
  "createdAt": RFC3339,
  "updatedAt": RFC3339
}
```

---

## RegionClusters

### POST /api/v1/regions/:regionId/clusters/create
Param: `regionId`
Request:
```
{
  "name": string(required)
}
```
Response data: `{ "id": int64 }`

### GET /api/v1/regions/:regionId/clusters/list
Param: `regionId`
Query: `isActive=bool`, `page=int(default 1)`, `limit=int(default 20)`
Response data: `{ "items": RegionCluster[], "pagination": PaginationMeta }`

### GET /api/v1/regions/:regionId/clusters/:clusterId
Param: `regionId`
Param: `clusterId`
Response data: `RegionCluster`

### PUT /api/v1/regions/:regionId/clusters/:clusterId/edit
Param: `regionId`
Param: `clusterId`
Request:
```
{
  "name": string(required)
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/regions/:regionId/clusters/:clusterId/set-default
Param: `regionId`
Param: `clusterId`
Response data: `{ "id": int64 }`

### POST /api/v1/regions/:regionId/clusters/:clusterId/set-ready
Param: `regionId`
Param: `clusterId`
Request:
```
{
  "ready": bool(required)
}
```
Response data: `{ "id": int64, "ready": bool }`

### POST /api/v1/regions/:regionId/clusters/:clusterId/deactivate
Param: `regionId`
Param: `clusterId`
Response data: `{ "id": int64 }`

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
  "blockSize"?: int32,
  "members"?: BlockMember[],
  "accessKey"?: string,
  "secretKey"?: string
}
```
Response data: `{ "id": int64, "blockVolumeIds?": string[] }`

### GET /api/v1/storages/list
Query: `accountId=int64(required)`, `search=string`, `regionId=int64`, `storageType=string`, `providerType=string`, `isActive=bool`, `directAccess=bool`, `page=int(default 1)`, `limit=int(default 10)`
Response data: `{ "items": Storage[], "pagination": PaginationMeta }`

### GET /api/v1/storages/:storageId
Param: `storageId`
Response data: `Storage`

### GET /api/v1/storages/:storageId/block-volumes
Param: `storageId`
Response data: `BlockVolume[]`

### PUT /api/v1/storages/:storageId/edit
Param: `storageId`
Request:
```
{
  "name": string(required),
  "description"?: string,
  "endpoint"?: string,
  "accessKey"?: string,
  "secretKey"?: string,
  "directAccess"?: bool
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

### GET /api/v1/storages/:storageId/compatible
Param: `storageId`
Response data: `{ "storages": CompatibleStorage[] }`

### POST /api/v1/storages/:storageId/move-volumes
Param: `storageId`
Request:
```
{
  "volumeIds": string[](required)
}
```
Response data: `{ "moved": string[], "failures": MoveVolumeFailure[] }`

### POST /api/v1/storages/backfill-fingerprints
Response data: `{ "scanned": int32, "updated": int32, "failures": BackfillFailure[], "hasMore": bool }`

### Storage Type
```
{
  "id": int64,
  "uuid": string,
  "account": Ref,
  "regionInfo": Ref,
  "name": string,
  "description"?: string,
  "storageType": string,
  "providerType": string,
  "endpoint": string,
  "region"?: string,
  "bucket"?: string,
  "base"?: string,
  "physicalFingerprint"?: string,
  "blockRegion"?: string,
  "blockSize"?: int32,
  "directAccess"?: bool,
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
  "forkGracePeriod"?: int32,
  "eventLogRetentionPeriod"?: int32,
  "quotaLimit"?: int64,
  "regionClusterId"?: int64,
  "regionClusterUuid"?: string
}
```
Response data: `{ "id": int64, "encryptionKey?": string }`

### GET /api/v1/volumes/list
Query: `accountId=int64(required)`, `regionId=int64`, `regionClusterId=int64`, `storageId=int64`, `volumeType=string`, `locked=bool`, `isActive=bool`, `page=int(default 1)`, `limit=int(default 10)`
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
  "gracePeriod"?: int32,
  "forkGracePeriod"?: int32,
  "eventLogRetentionPeriod"?: int32
}
```
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/lock
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/unlock
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/move-cluster
Param: `volumeId`
Request:
```
{
  "targetClusterId"?: int64,
  "targetClusterUuid"?: string
}
```
Response data: `{ "id": int64, "sourceClusterId": int64, "targetClusterId": int64, "handoverUntil": int64 }`

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

### POST /api/v1/volumes/:volumeId/activate
Param: `volumeId`
Response data: `{ "id": int64 }`

### POST /api/v1/volumes/:volumeId/api-keys/generate
Param: `volumeId`
Request:
```
{
  "userId": int64(required),
  "name"?: string
}
```
Response data: `{ "apiKey": string, "apiSecret": string, "evictedApiKeys?": string[] }`

### GET /api/v1/volumes/:volumeId/api-keys
Param: `volumeId`
Response data: `{ "keys": VolumeApiKey[] }`

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

### POST /api/v1/volumes/:volumeId/stt-key/generate
Param: `volumeId`
Request:
```
{
  "userId"?: int64,
  "expirySeconds": int64(required)
}
```
Response data: `{ "apiKey": string, "apiSecret": string, "expiresAt": string }`

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
Response data: `{ "volumeId": string, "liveVolume": int64, "totalVolume": int64, "pendingVolume": int64, "liveInactiveVolume": int64 }`

### GET /api/v1/volumes/:volumeId/size-history
Param: `volumeId`
Query: `from=string`, `to=string`
Response data: `{ "points": VolumeSizePoint[] }`

### POST /api/v1/volumes/:volumeId/forks/create
Param: `volumeId`
Request:
```
{
  "name": string(required),
  "parentName"?: string,
  "asOf"?: int64,
  "volumeType"?: string
}
```
Response data: `Fork`

### GET /api/v1/volumes/:volumeId/forks
Param: `volumeId`
Query: `volumeType=string`
Response data: `Fork[]`

### GET /api/v1/volumes/:volumeId/forks?include_inactive=true
Param: `volumeId`
Query: `volumeType=string`
Response data: `Fork[]`

### POST /api/v1/volumes/:volumeId/forks/:forkName/delete
Param: `volumeId`
Param: `forkName`
Request:
```
{
  "force"?: bool,
  "volumeType"?: string
}
```
Response data: `{ "inactivatedFids": int32[] }`

### POST /api/v1/volumes/:volumeId/forks/:forkName/restore
Param: `volumeId`
Param: `forkName`
Request:
```
{
  "volumeType"?: string
}
```
Response data: `Fork`

### Volume Type
```
{
  "id": int64,
  "account": Ref,
  "storage": Ref,
  "region": Ref,
  "regionCluster"?: Ref,
  "name": string,
  "description"?: string,
  "volumeType": string,
  "storageType"?: string,
  "encryption": bool,
  "quotaLimit": int64,
  "liveVolume": int64,
  "totalVolume": int64,
  "pendingVolume": int64,
  "liveInactiveVolume": int64,
  "locked": bool,
  "retentionPeriod": int32,
  "gracePeriod": int32,
  "forkGracePeriod": int32,
  "eventLogRetentionPeriod": int32,
  "isActive": bool,
  "isCleanupMetaEnabled": bool,
  "isCleanupStorageEnabled": bool,
  "isCleanupVaultEnabled": bool,
  "createdAt": RFC3339,
  "updatedAt": RFC3339
}
```

---

## VolumeForkTrees

### GET /api/v1/volumes/:volumeId/forks/:forkName/tree
Param: `volumeId`
Param: `forkName`
Query: `path=string`, `asOf=int64`, `cursor=int64`, `limit=int(default 20)`, `sort=string`, `kind=string`
Response data: `{ "items": ForkTreeEntry[], "nextCursor": int64|null }`

### ForkTreeEntry Type
```
{
  "inode": int64,
  "name": string,
  "kind": string,
  "size": int64,
  "mtime": int64,
  "ctime": int64,
  "creatorId"?: int64,
  "updaterId"?: int64
}
```

---

## VolumeForkEntries

### GET /api/v1/volumes/:volumeId/forks/:forkName/entry
Param: `volumeId`
Param: `forkName`
Query: `path=string`, `inode=int64`, `asOf=int64`
Response data: `ForkEntryDetail`

### GET /api/v1/volumes/:volumeId/forks/:forkName/entry/versions
Param: `volumeId`
Param: `forkName`
Query: `path=string`, `cursor=int64`, `limit=int(default 20)`
Response data: `{ "items": ForkEntryVersion[], "nextCursor": int64|null }`

### ForkEntryDetail Type
```
{
  "inode": int64,
  "path": string,
  "name": string,
  "kind": string,
  "size": int64,
  "mtime": int64,
  "ctime": int64,
  "generation": int64,
  "owner"?: string,
  "mode"?: int32,
  "xattrs"?: object,
  "creatorId"?: int64,
  "updaterId"?: int64
}
```

---

## VolumeForkSearches

### GET /api/v1/volumes/:volumeId/forks/:forkName/search
Param: `volumeId`
Param: `forkName`
Query: `q=string`, `path=string`, `asOf=int64`, `exact=bool`, `cursor=int64`, `limit=int(default 20)`, `kind=string`
Response data: `{ "items": ForkTreeMatch[], "nextCursor": int64|null }`

### ForkTreeMatch Type
```
{
  "path": string,
  "inode": int64,
  "kind": string,
  "size": int64,
  "mtime": int64
}
```

---

## AuditLogs

### GET /api/v1/audit-logs/list
Query: `accountId=int64(required)`, `regionId=int64`, `regionClusterId=int64`, `cursor=int64`, `limit=int(default 20)`, `subject=string`
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
  "node"?: string,
  "accountId"?: int64,
  "regionId"?: int64,
  "regionClusterId"?: int64,
  "createdAt"?: RFC3339,
  "updatedAt"?: RFC3339
}
```

---

## RegionAuditLogs

### GET /api/v1/regions/:regionId/audit-logs/list
Param: `regionId`
Query: `regionClusterId=int64`, `cursor=int64`, `limit=int(default 20)`, `subject=string`, `node=string`
Response data: `{ "items": AuditLog[], "nextCursor": int64|null }`

---

## ServiceNodes

### GET /api/v1/regions/:regionId/nodes
Param: `regionId`
Query: `serviceType=string`, `status=string`, `inactiveHours=int`, `regionClusterId=int64`
Response data: `ServiceNode[]`

### GET /api/v1/regions/:regionId/nodes/:nodeId/stats
Param: `regionId`
Param: `nodeId`
Response data: `string`

### GET /api/v1/regions/:regionId/nodes/:nodeId/stats/history
Param: `regionId`
Param: `nodeId`
Response data: `{ "intervalMs": int64, "samples": NodeStatsSample[] }`

### ServiceNode Type
```
{
  "id": int64,
  "regionId": int64,
  "regionClusterId"?: int64,
  "serviceType": string,
  "nodeId": string,
  "advertiseAddr": string,
  "rpcAddr"?: string,
  "metadata"?: object,
  "instanceId"?: string,
  "instanceInfo"?: object,
  "status": string,
  "lastHeartbeat"?: int64,
  "isActive": bool,
  "memUsage"?: float,
  "sysLoad"?: int,
  "binaryVersion"?: int32
}
```

---

## Nodes

### GET /api/v1/nodes
Query: `accountId=int64(required)`, `serviceType=string`, `status=string`, `inactiveHours=int`
Response data: `ServiceNode[]`

---

## ClientSessions

### GET /api/v1/client-sessions/list
Query: `accountId=int64(required)`, `regionId=int64`, `regionClusterId=int64`, `volumeId=int64`, `userId=int64`, `clientType=string`, `status=ClientSessionStatus`, `isActive=string`, `osName=string`, `platform=string`, `search=string`, `page=int(default 1)`, `limit=int(default 20)`
Response data: `{ "items": ClientSession[], "pagination": PaginationMeta }`

### GET /api/v1/client-sessions/:sessionId
Param: `sessionId`
Response data: `ClientSession`

### GET /api/v1/client-sessions/summary
Query: `accountId=int64(required)`, `regionId=int64`, `regionClusterId=int64`, `volumeId=int64`, `userId=int64`
Response data: `SessionSummary`

### ClientSession Type
```
{
  "id": int64,
  "account": Ref,
  "region": Ref,
  "regionCluster"?: Ref,
  "volume": VolumeRef,
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
  "status": ClientSessionStatus,
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

### GET /api/v1/license/terms
Response data: `LicenseTerms`

### POST /api/v1/license/load
Request:
```
{
  "payloads": string[](required)
}
```
Response data: `LicenseLoadResult`

### GET /api/v1/license/list
Response data: `LicenseList`

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
  "expiredAccessDays": int,
  "maxNodes": int64,
  "maxVolumes": int64,
  "maxUsers": int64,
  "maxAccounts": int64,
  "maxRegions": int64,
  "maxStorageBytes": int64,
  "status": LicenseStatus,
  "daysRemaining": int,
  "graceEndsAt": string,
  "graceDaysLeft": int,
  "expiredAccessEndsAt": string,
  "expiredAccessDaysLeft": int,
  "quota": LicenseQuota,
  "distribution"?: string,
  "distributionRef"?: string[],
  "unlimitedStorage"?: bool
}
```

---

## Alerts

### GET /api/v1/alerts/list
Query: `active=bool(default true)`, `accountId=int64`, `regionId=int64`, `severity=int`, `category=string`, `since=string`, `page=int(default 1)`, `limit=int(default 20)`
Response data: `{ "items": ServiceAlert[], "pagination": PaginationMeta }`

### GET /api/v1/alerts/count
Response data: `AlertCountResponse`

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
  "region"?: Ref,
  "account"?: Ref,
  "eventTime": RFC3339,
  "resolvedAt"?: RFC3339,
  "createdAt"?: RFC3339
}
```

---

## RegionAlerts

### GET /api/v1/regions/:regionId/alerts/list
Param: `regionId`
Query: `active=bool(default true)`, `severity=int`, `category=string`, `nodeId=string`, `regionClusterId=int64`, `since=string`, `page=int(default 1)`, `limit=int(default 20)`
Response data: `{ "items": RegionAlert[], "pagination": PaginationMeta }`

### GET /api/v1/regions/:regionId/alerts/count
Param: `regionId`
Query: `regionClusterId=int64`
Response data: `AlertCountResponse`

### POST /api/v1/regions/:regionId/alerts/:alertId/resolve
Param: `regionId`
Param: `alertId`

### RegionAlert Type
```
{
  "id": int64,
  "alertId": string,
  "source": string,
  "nodeId": string,
  "regionClusterId"?: int64,
  "severity": int,
  "category": string,
  "title": string,
  "description"?: string,
  "eventTime": RFC3339,
  "resolvedAt"?: RFC3339,
  "createdAt"?: RFC3339
}
```

---

## GCWorkerEvents

### GET /api/v1/regions/:regionId/gc-worker-events/list
Param: `regionId`
Query: `nodeId=string`, `goal=string`, `sid=int64`, `regionClusterId=int64`, `since=string`, `page=int(default 1)`, `limit=int(default 20)`
Response data: `{ "items": GCWorkerEvent[], "pagination": PaginationMeta }`

### GET /api/v1/regions/:regionId/gc-worker-events/histogram
Param: `regionId`
Query: `nodeId=string`, `goal=string`, `sid=int64`, `regionClusterId=int64`, `since=string`, `bucketSeconds=int64(default 900)`
Response data: `GCWorkerEventHistogramResponse`

### GCWorkerEvent Type
```
{
  "id": int64,
  "nodeId": string,
  "regionClusterId"?: int64,
  "goal": string,
  "sid"?: int64,
  "subject"?: string,
  "ops": object,
  "durationMs": int64,
  "eventTime": RFC3339,
  "createdAt"?: RFC3339
}
```

---

## Vault

### POST /api/v1/vault/resync

---

## Types

### AlertCountResponse Type
```
{
  "active": int64,
  "recent": int64,
  "infoCount": int64,
  "warningCount": int64,
  "criticalCount": int64
}
```

### BackfillFailure Type
```
{
  "shardId": int64,
  "error": string
}
```

### BlockMember Type
```
{
  "name"?: string,
  "regionClusterId": int64
}
```

### BlockVolume Type
```
{
  "id": string,
  "name"?: string,
  "clusterName"?: string,
  "clusterUuid"?: string,
  "shardId": int64,
  "regionClusterId": int64,
  "clusterReady": bool,
  "isActive": bool,
  "createdAt": RFC3339,
  "updatedAt": RFC3339
}
```

### CompatibleStorage Type
```
{
  "id": int64,
  "uuid": string,
  "name": string,
  "storageType": string,
  "providerType": string,
  "volumes": CompatibleVolume[]
}
```

### CompatibleVolume Type
```
{
  "id": string,
  "name": string
}
```

### DashboardStats Type
```
{
  "userCount": int64,
  "volumeCount": int64,
  "regionCount": int64,
  "storageCount": int64,
  "totalVolumeUsed": int64,
  "totalQuotaLimit": int64,
  "activeSessionCount": int64,
  "regionBreakdown": RegionVolumeMetrics[]
}
```

### DashboardUser Type
```
{
  "id": string,
  "name": string,
  "email"?: string,
  "role": string,
  "username"?: string,
  "accountId"?: int64,
  "userId"?: int64,
  "volumeId"?: int64,
  "exp"?: int64
}
```

### DiscoverEndpoint Type
```
{
  "nodeId": string,
  "addr": string,
  "status": string
}
```

### DiscoverMetaResponse Type
```
{
  "regionId": int64,
  "region": string,
  "endpoints": DiscoverEndpoint[]
}
```

### Fork Type
```
{
  "fid": int32,
  "name": string,
  "parentFid": int32,
  "parentName": string,
  "snapshotTs": int64,
  "createdBy"?: int64,
  "createdAt": int64,
  "childrenCount": int32,
  "inactive"?: bool,
  "inactiveAt"?: int64,
  "status": string,
  "size": int64
}
```

### ForkEntryVersion Type
```
{
  "generation": int64,
  "size": int64,
  "mtime": int64,
  "updaterId"?: int64,
  "contentHash"?: string
}
```

### GCWorkerEventBucket Type
```
{
  "bucketStart": RFC3339,
  "goal": string,
  "count": int64
}
```

### GCWorkerEventHistogramResponse Type
```
{
  "buckets": GCWorkerEventBucket[]
}
```

### LicenseList Type
```
{
  "items": LicenseRecord[]
}
```

### LicenseLoadResult Type
```
{
  "loaded": int,
  "ignored": int
}
```

### LicenseQuota Type
```
{
  "state": LicenseQuotaState,
  "liveVolume": int64,
  "totalVolume": int64,
  "generation": int64,
  "lastTransitionAtMs": int64
}
```

### LicenseRecord Type
```
{
  "key": string,
  "licensee": string,
  "status": LicenseStatus,
  "issuedAt": string,
  "expiresAt": string,
  "maxStorageBytes": int64,
  "insertedAt": RFC3339
}
```

### LicenseTerms Type
```
{
  "terms": string
}
```

### MoveVolumeFailure Type
```
{
  "volumeId": string,
  "error": string
}
```

### NodeStatsSample Type
```
{
  "timestampMs": int64,
  "intervalMs": int64,
  "loadAvg1": float,
  "loadAvg5": float,
  "loadAvg15": float,
  "memUsage": float,
  "readIops": float,
  "writeIops": float,
  "netRxBytesPerSec": float,
  "netTxBytesPerSec": float,
  "processCount": int,
  "diskUsedBytes"?: int64,
  "diskTotalBytes"?: int64
}
```

### Ref Type
```
{
  "id": int64,
  "name": string
}
```

### RegionVolumeMetrics Type
```
{
  "regionId": int64,
  "regionName": string,
  "volumeCount": int64,
  "totalVolumeUsed": int64,
  "totalQuotaLimit": int64
}
```

### SessionSummary Type
```
{
  "byStatus": SessionSummaryStatusEntry[],
  "byPlatform": SessionSummaryFacet[],
  "byOsName": SessionSummaryFacet[],
  "regionCount": int64,
  "volumeCount": int64,
  "hostCount": int64,
  "unhealthyCount": int64
}
```

### SessionSummaryFacet Type
```
{
  "label": string,
  "count": int64
}
```

### SessionSummaryStatusEntry Type
```
{
  "clientType": string,
  "status": string,
  "count": int64
}
```

### UserLite Type
```
{
  "id": int64,
  "username": string,
  "name": string
}
```

### VolumeApiKey Type
```
{
  "apiKey": string,
  "name"?: string,
  "createdAt"?: RFC3339,
  "lastUsedAt"?: RFC3339
}
```

### VolumeRef Type
```
{
  "id": int64,
  "name": string,
  "type"?: string
}
```

### VolumeSizePoint Type
```
{
  "bucketEnd": RFC3339,
  "liveVolume": int64,
  "totalVolume": int64,
  "pendingVolume": int64,
  "liveInactiveVolume": int64
}
```

---

## JWT Construction

```
Header:  { "alg": "EdDSA", "typ": "JWT" }
Payload: {
  "sub": "mountos:provider",
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
