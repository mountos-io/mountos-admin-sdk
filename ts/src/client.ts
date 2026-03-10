import { TokenSigner } from './auth.js'
import { MountOSError } from './errors.js'
import type {
  Config, StandardResponse, ListOptions,
  PaginatedResponse, CursorPaginatedResponse,
  Account, CreateAccountRequest, EditAccountRequest,
  User, AddUserRequest, EditUserRequest, UserListOptions,
  Region, CreateRegionRequest, EditRegionRequest,
  Storage, CreateStorageRequest, EditStorageRequest, StorageListOptions,
  UpdateVolumeQuotaRequest,
  AuditLog, AuditLogListOptions,
  ServiceNode,
  DiscoverMetaResponse,
} from './types.js'

function queryString(params: Record<string, string | number | undefined>): string {
  const entries = Object.entries(params).filter(([, v]) => v !== undefined)
  if (entries.length === 0) return ''
  return '?' + entries.map(([k, v]) => `${k}=${encodeURIComponent(v!)}`).join('&')
}

export class MountOSAdmin {
  private readonly baseUrl: string
  private readonly signer: TokenSigner

  private _accounts?: AccountsResource
  private _users?: UsersResource
  private _regions?: RegionsResource
  private _storages?: StoragesResource
  private _volumes?: VolumesResource
  private _auditLogs?: AuditLogsResource
  private _serviceNodes?: ServiceNodesResource
  private _discover?: DiscoverResource

  constructor(config: Config) {
    this.baseUrl = config.baseUrl.replace(/\/+$/, '')
    this.signer = new TokenSigner(config.privateKey)
  }

  get accounts(): AccountsResource {
    return (this._accounts ??= new AccountsResource(this))
  }

  get users(): UsersResource {
    return (this._users ??= new UsersResource(this))
  }

  get regions(): RegionsResource {
    return (this._regions ??= new RegionsResource(this))
  }

  get storages(): StoragesResource {
    return (this._storages ??= new StoragesResource(this))
  }

  get volumes(): VolumesResource {
    return (this._volumes ??= new VolumesResource(this))
  }

  get auditLogs(): AuditLogsResource {
    return (this._auditLogs ??= new AuditLogsResource(this))
  }

  get serviceNodes(): ServiceNodesResource {
    return (this._serviceNodes ??= new ServiceNodesResource(this))
  }

  get discover(): DiscoverResource {
    return (this._discover ??= new DiscoverResource(this))
  }

  async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const token = await this.signer.getToken()
    const headers: Record<string, string> = {
      'Authorization': `Bearer ${token}`,
    }

    const init: RequestInit = { method, headers }
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json'
      init.body = JSON.stringify(body)
    }

    const res = await fetch(`${this.baseUrl}${path}`, init)

    let json: StandardResponse<T>
    try {
      json = await res.json() as StandardResponse<T>
    } catch {
      throw new MountOSError(res.statusText || 'request failed', res.status)
    }

    if (json.status !== 'success') {
      throw new MountOSError(json.message, res.status, json.errorCode)
    }
    return json.data as T
  }
}

class AccountsResource {
  constructor(private client: MountOSAdmin) {}

  create(req: CreateAccountRequest): Promise<{ id: number }> {
    return this.client.request('POST', '/api/v1/accounts/create', req)
  }

  list(opts?: ListOptions): Promise<PaginatedResponse<Account>> {
    return this.client.request('GET', `/api/v1/accounts/list${queryString({ page: opts?.page, limit: opts?.limit })}`)
  }

  get(accountId: number): Promise<Account> {
    return this.client.request('GET', `/api/v1/accounts/${accountId}`)
  }

  edit(accountId: number, req: EditAccountRequest): Promise<{ id: number }> {
    return this.client.request('PUT', `/api/v1/accounts/${accountId}/edit`, req)
  }

  lock(accountId: number): Promise<{ id: number }> {
    return this.client.request('POST', `/api/v1/accounts/${accountId}/lock`)
  }

  unlock(accountId: number): Promise<{ id: number }> {
    return this.client.request('POST', `/api/v1/accounts/${accountId}/unlock`)
  }

  activate(accountId: number): Promise<{ id: number }> {
    return this.client.request('POST', `/api/v1/accounts/${accountId}/activate`)
  }

  deactivate(accountId: number): Promise<{ id: number }> {
    return this.client.request('POST', `/api/v1/accounts/${accountId}/deactivate`)
  }
}

class UsersResource {
  constructor(private client: MountOSAdmin) {}

  add(req: AddUserRequest): Promise<{ id: number }> {
    return this.client.request('POST', '/api/v1/users/add', req)
  }

  list(opts: UserListOptions): Promise<PaginatedResponse<User>> {
    return this.client.request('GET', `/api/v1/users/list${queryString({ accountId: opts.accountId, page: opts.page, limit: opts.limit })}`)
  }

  get(userId: number): Promise<User> {
    return this.client.request('GET', `/api/v1/users/${userId}`)
  }

  edit(userId: number, req: EditUserRequest): Promise<{ id: number }> {
    return this.client.request('PUT', `/api/v1/users/${userId}/edit`, req)
  }

  activate(userId: number): Promise<{ id: number }> {
    return this.client.request('POST', `/api/v1/users/${userId}/activate`)
  }

  deactivate(userId: number): Promise<{ id: number }> {
    return this.client.request('POST', `/api/v1/users/${userId}/deactivate`)
  }
}

class RegionsResource {
  constructor(private client: MountOSAdmin) {}

  create(req: CreateRegionRequest): Promise<{ id: number }> {
    return this.client.request('POST', '/api/v1/regions/create', req)
  }

  list(opts?: ListOptions): Promise<PaginatedResponse<Region>> {
    return this.client.request('GET', `/api/v1/regions/list${queryString({ page: opts?.page, limit: opts?.limit })}`)
  }

  get(regionId: number): Promise<Region> {
    return this.client.request('GET', `/api/v1/regions/${regionId}`)
  }

  edit(regionId: number, req: EditRegionRequest): Promise<{ id: number }> {
    return this.client.request('PUT', `/api/v1/regions/${regionId}/edit`, req)
  }

  activate(regionId: number): Promise<{ id: number }> {
    return this.client.request('POST', `/api/v1/regions/${regionId}/activate`)
  }

  deactivate(regionId: number): Promise<{ id: number }> {
    return this.client.request('POST', `/api/v1/regions/${regionId}/deactivate`)
  }
}

class StoragesResource {
  constructor(private client: MountOSAdmin) {}

  create(req: CreateStorageRequest): Promise<{ id: string; shardId: number }> {
    return this.client.request('POST', '/api/v1/storages/create', req)
  }

  list(opts: StorageListOptions): Promise<PaginatedResponse<Storage>> {
    return this.client.request('GET', `/api/v1/storages/list${queryString({ accountId: opts.accountId, page: opts.page, limit: opts.limit })}`)
  }

  get(storageId: string): Promise<Storage> {
    return this.client.request('GET', `/api/v1/storages/${storageId}`)
  }

  edit(storageId: string, req: EditStorageRequest): Promise<{ id: string }> {
    return this.client.request('PUT', `/api/v1/storages/${storageId}/edit`, req)
  }

  activate(storageId: string): Promise<{ id: string }> {
    return this.client.request('POST', `/api/v1/storages/${storageId}/activate`)
  }

  deactivate(storageId: string): Promise<{ id: string }> {
    return this.client.request('POST', `/api/v1/storages/${storageId}/deactivate`)
  }
}

class VolumesResource {
  constructor(private client: MountOSAdmin) {}

  updateQuota(volumeId: string, req: UpdateVolumeQuotaRequest): Promise<{ id: string }> {
    return this.client.request('PUT', `/api/v1/volumes/${volumeId}/quota`, req)
  }
}

class AuditLogsResource {
  constructor(private client: MountOSAdmin) {}

  list(opts?: AuditLogListOptions): Promise<CursorPaginatedResponse<AuditLog>> {
    return this.client.request('GET', `/api/v1/audit-logs/list${queryString({
      accountId: opts?.accountId,
      cursor: opts?.cursor,
      limit: opts?.limit,
      subject: opts?.subject,
    })}`)
  }
}

class ServiceNodesResource {
  constructor(private client: MountOSAdmin) {}

  list(regionId: number): Promise<ServiceNode[]> {
    return this.client.request('GET', `/api/v1/regions/${regionId}/nodes`)
  }

  drain(regionId: number, nodeId: string): Promise<void> {
    return this.client.request('POST', `/api/v1/regions/${regionId}/nodes/${encodeURIComponent(nodeId)}/drain`)
  }

  activate(regionId: number, nodeId: string): Promise<void> {
    return this.client.request('POST', `/api/v1/regions/${regionId}/nodes/${encodeURIComponent(nodeId)}/activate`)
  }

  remove(regionId: number, nodeId: string): Promise<void> {
    return this.client.request('DELETE', `/api/v1/regions/${regionId}/nodes/${encodeURIComponent(nodeId)}`)
  }
}

class DiscoverResource {
  constructor(private client: MountOSAdmin) {}

  meta(accessKeyId: string): Promise<DiscoverMetaResponse> {
    return this.client.request('GET', `/api/v1/discover/meta${queryString({ access_key_id: accessKeyId })}`)
  }
}
