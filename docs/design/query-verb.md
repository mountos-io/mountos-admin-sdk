# Using the HTTP QUERY method in api.yaml

QUERY (RFC 10008) is a safe, idempotent method like GET, but unlike GET it is
defined to carry a request body. appserv (mountos-servers) already runs on
Echo v5, which has native `QUERY` support, and registers it as a same-handler
alias next to `GET` on ~30 list/filter endpoints (accounts, users, regions,
storages, volumes, audit logs, alerts, gc-worker-events, discover, dashboard,
client-sessions, raft). Those aliases exist for REST/curl callers whose filter
set is too long or awkward for a URL query string; the handler reads scalar
params from either the URL or a flat JSON body via `utils.QueryParam`.

None of that was reflected in `api.yaml` until this doc's companion change,
which is why every generated SDK method for those ~30 endpoints still speaks
GET. This is intentional, not an oversight to "catch up" wholesale — see
below.

## When to declare `method: QUERY` in the spec

Reach for QUERY only when an endpoint is:

1. **Read-only / safe** — it must not mutate anything. If it writes, it's
   POST/PUT, full stop.
2. **Body-shaped by necessity, not by convenience** — the parameters don't
   fit a URL query string: an array of ids, a filter object with nested
   structure, or anything whose size is caller-controlled and can grow past
   what a query string comfortably carries.

`Users.bulk` (`ids: int64[]!` → `UserLite[]`) is the model case: resolving a
page's worth of creator/updater ids for the admin tree-browse UI never fit a
query string, so it was a POST-shaped read from the start. It's now
`method: QUERY` instead — same semantics, correctly labeled as safe.

## Why the ~30 existing appserv GET+QUERY aliases are *not* mirrored into the spec

Those endpoints (`accounts.list`, `volumes.list`, etc.) take a handful of
scalar filters — `accountId`, `isActive`, `page`, `limit`, and similar. A URL
query string handles that fine, and every SDK already builds one correctly.
Adding a second `method: QUERY` action per endpoint would:

- Double the generated method surface (`list` and `listQuery`, or similar)
  for zero behavioral difference a typed SDK caller would ever notice.
- Encourage spec authors to reach for QUERY out of habit rather than because
  an endpoint actually needs a body.

If a specific list endpoint's filter set grows into genuinely
query-string-unfriendly territory (a large `ids: int64[]` filter, a complex
nested search DSL), convert *that* endpoint's action to `method: QUERY` at
that point, following the `Users.bulk` pattern. Until then, GET stays the
right transport for scalar filters.

## Declaring a QUERY endpoint

```yaml
- action: bulk
  method: QUERY
  path: /bulk
  request:
    - "ids: int64[]!"
  response:
    - "users: UserLite[]"
```

`make gen` rejects `method: QUERY` on an endpoint with neither `request` nor
`query` fields (`gen/main.go`'s `validateSpec`) — a parameterless QUERY
endpoint is just GET with extra steps, so the generator fails loudly instead
of letting Rust's toggle/void writers silently mis-map it to a bare POST.

## What the generators and runtimes do

- **Go**: `goHTTPMethod` maps `QUERY` → `query`; `go/http.go` has a
  `query(ctx, path, body)` wrapper (own `MethodQuery` constant — the method
  isn't in `net/http`'s constant set). The two body-carrying method writers
  (`writeGoBodyResponseMethod`, `writeGoBodyResponseTypeMethod`) treat
  `query` the same as `put`/`post` for attaching the body — a naive switch
  would have silently dropped the request body for QUERY calls, since only
  PUT/POST were originally recognized as body-carrying.
- **TypeScript**: no changes needed. The transport (`ts/src/server.ts`)
  interpolates the method string straight into `fetch()` and only special-
  cases GET/HEAD to omit a body; QUERY passes through untouched, and the
  generator's body-shaped writer (`writeTSBodyMethod`) already emits
  `ep.Method` verbatim.
- **Rust**: `rust/src/http.rs` gained a `query<T, B>()` transport method
  (reqwest has no built-in `Method::QUERY` constant, so it's built via
  `Method::from_bytes(b"QUERY")`). `writeRustBodyMethod` lowercases
  `ep.Method` and calls `self.inner.<verb>(...)` directly, so `query` needed
  nothing else on the generator side for the body-shaped case. The
  toggle/void writers (bodyless GET/DELETE/POST-only shapes) still don't
  understand QUERY and would fall through to `post_empty` — the spec
  validator above exists specifically to keep any real endpoint from landing
  there.

## appserv wiring notes

- Register with `g.QUERY(path, handler)`. For a body-only read with no GET
  fallback (like bulk resolve), that's the only registration needed — reuse
  the same JSON-body decode a POST handler would use; Echo's request body
  reading doesn't care which method produced the request.
- `internal/middlewares/license_write_guard.go` treats every non-GET/HEAD/
  OPTIONS method as a write and blocks it when the license is read-only.
  QUERY needed adding to that safe-method allowlist — it's the same class of
  bug any new safe-but-non-GET method would hit, worth checking again if
  another such method is ever introduced.
- The admin-client's own authorization layer
  (`mountos-admin-client/server/authz.ts`) maps HTTP method to a required
  capability bit; QUERY needed a `Cap.R` case alongside GET/HEAD, or it falls
  through to an automatic 403. The proxy layer (`server/proxy.ts`) needed no
  changes — its GET/HEAD body-stripping check and write-classification logic
  are both allowlists that already treat QUERY as a normal, body-carrying
  read.

## Checklist for adding a new QUERY endpoint

1. Confirm it's genuinely read-only and genuinely needs a body (see above).
2. Add it to `api.yaml` with `method: QUERY` and either `request` or `query`
   fields.
3. `make gen && make check` — confirms all three languages generate and
   build correctly.
4. Register `g.QUERY(path, handler)` in the relevant `routes_*.go`, reusing
   the handler's existing JSON-body decode.
5. If the route sits behind `LicenseWriteGuard` or any other method-based
   middleware, confirm QUERY is treated as safe there too.
6. If the admin-client proxies it, confirm `authz.ts`'s `requiredCap` maps
   QUERY to `Cap.R` for the resource (it does, generically, as of this
   change — only a new *method*, not a new *resource*, would need a fresh
   case there).
