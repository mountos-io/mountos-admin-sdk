import { TokenSigner } from './auth.js'
import { MountOSError } from './errors.js'
import { signDashboardUser } from './dashboard_user.js'
import { createClient, type RequestFn, type AdminClient } from './client_gen.js'
import type { Config, StandardResponse } from './types_gen.js'

export function createServerClient(config: Config): AdminClient {
  const baseUrl = config.baseUrl.replace(/\/+$/, '')
  const signer = new TokenSigner(config.privateKey)
  const privateKey = config.privateKey
  const dashboardUser = config.dashboardUser

  const request: RequestFn = async <T>(method: string, path: string, body?: unknown, signal?: AbortSignal): Promise<T> => {
    const token = await signer.getToken()
    const headers: Record<string, string> = { Authorization: `Bearer ${token}` }
    if (dashboardUser) {
      headers['X-MountOS-Dashboard-User'] = await signDashboardUser(dashboardUser, privateKey)
    }

    const init: RequestInit = { method, headers, signal }
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json'
      init.body = JSON.stringify(body)
    }

    const res = await fetch(`${baseUrl}${path}`, init)

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

  return createClient(request)
}
