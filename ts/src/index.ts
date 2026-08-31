export { createClient, type RequestFn, type AdminClient } from './client_gen.js'
export { createServerClient } from './server.js'
export { MountOSError } from './errors.js'
export { TokenSigner } from './auth.js'
export { signDashboardUser } from './dashboard_user.js'
export { assertSafeIntegers, UnsafeIntegerError } from './safe_int.js'
export type * from './types_gen.js'
export type * from './types.js'
// Deprecated pre-1.14.0 "Pair" naming aliases (VE-017). Imported for its
// side effect (patches StoragesResource/VolumesResource with the old
// method names) as well as its type/value exports.
export * from './compat.js'
