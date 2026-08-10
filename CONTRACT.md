# Apito Admin SDK — Shared Contract

All three admin SDKs (`flutter_admin_sdk`, `js-admin-sdk`, `go-admin-sdk`) share:

## 1. Naming engine

Golden vectors: `test/fixtures/naming_vectors.json` (canonical; copy verbatim to JS/Go).

Reference implementations:
- Dart: `lib/src/runtime/naming.dart`
- TypeScript: `refine-apito/src/apitoGraphqlNames.ts` (vendored in JS as `src/naming/apitoGraphqlNames.ts`)
- Go: `naming.go`

## 2. Introspection snapshot

Offline codegen uses a checked-in introspection JSON:
- Flutter: `example/apito_introspection.json`
- JS: `schema/apito_introspection.json`
- Go: `schema/apito_introspection.json`

Live fetch fallback: POST introspection query with `X-Apito-Key` + optional `X-Apito-Tenant-ID`.

## 3. Operation doc format (5 ops per model)

Each model emits:
1. `Get{Model}List` — list + count
2. `Get{Model}` — single by `_id`
3. `Create{Model}` — create mutation
4. `Update{Model}` — update mutation
5. `Delete{Model}` — delete mutation

Uses Apito composed type names (`{Model}List_Input_Where_Payload`, `{Model}_Create_Payload`, etc.).

Reference output: `example/lib/generated/operations/loan.graphql`.

## 4. Admin client surface

All SDKs expose:
- **GraphQL CRUD** (secured endpoint chainable builder + system fallback)
- **REST storage**: `uploadFile`, `listFiles`, `deleteFiles` at `/files/upload|list|delete`
- **Auth/admin**: `generateTenantToken`, `getTenants`, `getTenant`, `searchTenants`, `createTenant`, `updateTenant`, `deleteTenant`, `loginUser`, `googleOAuthState`, `searchUsers`, `searchTenantsByDomain`, `createUser`, `updateUser`, `resetUserPassword`, `deleteUser`
- **Tenant plan access (app-user JWT on secured/public)**: `myTenant` (includes `plan_tier`), `myEffectivePermissions` (role ∩ plan ceiling, quotas, grace). Helpers: plan tier parse/rank (`parsePlanTier` / `PlanAtLeast`), `canFromSnapshot`, `isPlanQuotaError`. JS also exports Casbin policy builder + React `CasbinProvider` via `@apito-io/js-admin-sdk/access` and `/react`.

Pro SaaS user ops accept optional **`TenantID`** / GraphQL `tenant_id` on `SearchUsers`, `CreateUser`, and `UpdateUser` (in addition to `LoginUser`). Omit on general projects. **`SearchUsers`** also accepts optional **`q`** (6th arg) for free-text filter on email, username, phone, or id.

### Cloudflare Workers v1 (`cloudflare_full`)

When the engine URL is a Cloudflare Worker (`-tags cloudflare`), the SDK contract is unchanged but some operations are **not implemented** on Workers v1:

| Operation | Workers v1 |
|-----------|------------|
| `GenerateTenantToken`, tenant catalog mutations | GraphQL error: `tenant management is not available on Cloudflare Workers v1` |
| `LoginUser` (password / general) | Supported |
| `LoginUser` (`google`, `google_id_token`), `GoogleOAuthState` | GraphQL error: `google login is not available on Cloudflare Workers v1` |

Use the native/pro engine for tenant lifecycle and Google end-user login, or handle these errors in client code.

## 5. Codegen outputs

| SDK | Tool | Hooks |
|-----|------|-------|
| JS | graphql-codegen | TanStack React Query |
| Go | genqlient | Context-aware client wrappers |
| Flutter | build_runner (custom) | Riverpod providers |
