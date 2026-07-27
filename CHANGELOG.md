# Changelog

All notable changes to the Go Apito SDK will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.6.7] - 2026-07-27

### Fixed

- **`CanonicalizeModelName`** — already-canonical snake_case ids singularize last segment only (parity with open-core `apito_naming.go` / long singles like `indication`).

## [2.6.6] - 2026-07-21

### Added

- **Project-scoped access-token requests** — `Config.ProjectID` supplies `X-Apito-Project-Id` for GraphQL and REST requests. Methods with an explicit `projectID` override the configured value with the same request scope while retaining context-based tenant headers.

### Changed

- **Unified `apt_` access tokens (hard cut)** — `applyAuthCredential` sends `apt_` tokens as `Authorization: Bearer` + `X-Use-Cookies: false` only; dropped the compatibility `X-Apito-Key` dual header. `Config.AccessToken` is the preferred field (alias: `APIKey` set to an `apt_...` value). Legacy `cli-`/`sdk-`/`mcp-` prefixed keys now make `executeGraphQL` return a `TOKEN_FORMAT_RETIRED` error before hitting the network. Plain project API keys (no recognized prefix) are unaffected and still use `X-Apito-Key`.

## [2.6.5] - 2026-07-14

### Added

- **`GetTenant(ctx, projectID, tenantID, status)`** — load one SaaS catalog tenant by exact id via `SearchTenants` (default `status`: `active`). Returns `(nil, nil)` when no exact id match. Parity with `js-admin-sdk` `getTenant` and `flutter_admin_sdk`.

## [2.6.4] - 2026-07-13

### Changed

- **`DeleteTenant`** — soft delete only (`status=deleted`); content and mirror remain until Console hard delete.
- **`SearchTenants`** — optional `status` argument (`active`, `deleted`, `all`).

## [2.6.3] - 2026-07-13

### Added

- **`SearchTenants(ctx, projectID, limit, offset, q, status)`** — paginated SaaS catalog search with `count` and optional free-text filter (`name`, `id`, `domain`, `data`). Optional `status`: `active` (default), `deleted`, or `all`. Parity with engine `searchTenants`, `js-admin-sdk`, and `flutter_admin_sdk`.
- **`SearchTenantsResponse`** type; **`TenantCatalogSearchRow`** now includes optional `icon` and `created_at` from search payloads.

## [2.6.2] - 2026-07-11

### Added

- **`SearchUsers` optional `q`** — 6th argument filters email, username, phone, or id (case-insensitive contains). Parity with engine `searchUsers` GraphQL and `js-admin-sdk`.

## [2.6.1] - 2026-06-22

### Changed

- **Cloudflare Workers v1 (`cloudflare_full`)** — Document engine compatibility: `GenerateTenantToken` and related tenant catalog mutations return `tenant management is not available on Cloudflare Workers v1`. `LoginUser` password/general path is unchanged on Workers; Google paths (`AuthMethod: "google"` / `"google_id_token"`, plus `GoogleOAuthState`) return `google login is not available on Cloudflare Workers v1`. No SDK API or signature changes — use the native/pro engine for tenant provisioning and Google end-user login, or handle these GraphQL errors when targeting a Workers URL.

## [2.6.0] - 2026-06-11

### Added

- **`TenantID` on user CRUD** — optional `tenantID` on `SearchUsers` (5th arg), `CreateUserParams.TenantID`, and `UpdateUserParams.TenantID`; sent as GraphQL `tenant_id` on pro SaaS engines. Omit on general projects.

### Changed

- **Docs** — README auth table documents tenant-aware user ops.

## [2.5.1] - 2026-06-15

### Changed

- **`LoginUser` Google auth (engine behavior)** — When `AuthMethod` is `google` or `google_id_token` and no user matches `google_sub`, the engine requires a verified Google email, then auto-links an existing project user with the same normalized email (password unchanged) or creates a new Google user. New GraphQL errors: `google email not verified`, `google account already linked to another user`, `multiple users matched this email`. No SDK signature changes.
- **`CreateUser` / `UpdateUser` uniqueness (engine behavior)** — Open-core projects now reject duplicate email and phone project-wide (aligned with pro). Stable GraphQL errors: `email already exists for this project`, `phone already exists for this project`. No schema or SDK API changes.

## [2.5.0] - 2026-06-08

### Added

- **`LoginUser` `TenantID`** — optional `TenantID` on `LoginUserParams`; sent as GraphQL `tenant_id` on system `loginUser`. Required by engine for SaaS projects with per-tenant separate databases.

### Changed

- **Docs** — README auth table notes per-tenant DB login requires `TenantID`.

## [2.4.0] - 2026-06-05

### Added

- **`LoginUser` `google_id_token`** — native mobile Google sign-in via `IDToken` on `LoginUserParams`.

### Changed

- **Project files REST** — default `RestBaseURL` resolves to `/secured` when GraphQL uses `/system/graphql`. Full paths: `/secured/files/upload|list|delete`.

## [2.3.0] - 2026-06-05

### Added

- **Naming engine** (`naming.go`) — parity with `flutter_admin_sdk` / refine-apito; golden vectors in `test/fixtures/naming_vectors.json`
- **Operation emitter** (`make gen-operations` / `cmd/apito-gen`) → `codegen/operations/*.graphql` + `schema.graphql`
- **genqlient** config (`genqlient.yaml`, `make gen-types`) for typed GraphQL operations
- **GraphQL doer** + **TypedModelOps** context-aware helpers (`TypedOps()`, `ListModel`, `GetModel`, `ExecuteRaw`)
- **DocumentBuilder** — secured-endpoint operation string generation
- **Schema reader** — introspection parsing + SDL export
- Shared [CONTRACT.md](CONTRACT.md) with JS/Flutter admin SDKs

## [2.2.0] - 2026-05-28

### Changed

- **Project files storage** — Document exported path constants (`FilesUploadPath`, `FilesListPath`, `FilesDeletePath`). File metadata is persisted in the **project DB** `files` table (engine migration from system DB); REST URLs remain `/system/files/*`. SaaS callers should set `tenant_id` on the request context.

## [2.1.0] - 2026-05-17

### Changed (breaking)

- **Removed storage settings GraphQL** — `GetProjectStorageSettings` and `UpdateProjectStorageSettings` dropped from the SDK (configure storage in the console or via raw GraphQL).
- **File API renamed** — action-oriented names aligned with the User API: `UploadFile`, `ListFiles`, `DeleteFiles` (was `*SystemFile*`).

### Migration

| v2.0.0 | v2.1.0 |
|--------|--------|
| `GetProjectStorageSettings` | removed |
| `UpdateProjectStorageSettings` | removed |
| `UploadSystemFile` | `UploadFile` |
| `ListSystemFiles` | `ListFiles` |
| `DeleteSystemFiles` | `DeleteFiles` |
| `SystemFile` | `File` |
| `SystemFileUploadParams` | `UploadFileParams` |
| `SystemFilesListResponse` | `FilesListResponse` |
| `DeleteSystemFilesResponse` | `DeleteFilesResponse` |

## [2.0.0] - 2026-05-17

### Changed (breaking)

- **Tenant-user GraphQL renamed to User API** — aligned with engine open-core migration (`users` table, `UserItem` type). All `*TenantUser*` types and methods renamed to `*User*` (e.g. `LoginUser`, `SearchUsers`, `CreateUser`, `UpdateUser`, `DeleteUser`).
- **`googleOAuthState`** replaces `tenantGoogleOAuthState`.
- **`UpdateUser`** no longer accepts `password`; use **`ResetUserPassword`**.

### Added

- **`ResetUserPassword(ctx, userID, password)`** — admin password reset mutation.
- **`GetProjectStorageSettings`**, **`UpdateProjectStorageSettings`** — project S3/storage settings GraphQL.
- **`UploadSystemFile`**, **`ListSystemFiles`**, **`DeleteSystemFiles`** — `/system/files` REST API (`Config.RestBaseURL` optional; derived from GraphQL base URL).
- Examples: `examples/users/`, `examples/system_files/` (replaces `examples/tenant_users/`).

### Migration

| v1.7.x | v2.0.0 |
|--------|--------|
| `LoginTenantUser` | `LoginUser` |
| `TenantGoogleOAuthState` | `GoogleOAuthState` |
| `SearchTenantUsers` | `SearchUsers` |
| `CreateTenantUser` | `CreateUser` |
| `UpdateTenantUser` (+ `Password`) | `UpdateUser` + `ResetUserPassword` |
| `DeleteTenantUser` | `DeleteUser` |
| `TenantUser` | `User` |

## [1.7.0] - 2026-05-08

### Changed (breaking)

- **`LoginTenantUserGoogle` removed** — engine dropped **`loginTenantUserGoogle`**. Use **`LoginTenantUser`** with **`AuthMethod: "google"`**, **`Code`**, **`State`**.

### Added

- **`TenantGoogleOAuthState(ctx, projectID)`** → **`TenantGoogleOAuthStateResponse`** (**`State`**) for the Google authorize redirect.

### `LoginTenantUserParams`

- **`Code`**, **`State`** for Google flow.
- **`Password`** / **`Email`** / **`Phone`** only required when **`AuthMethod`** is empty or **`general`**.

## [1.6.0] - 2026-05-14

### Changed (breaking)

- **Tenant catalog users** aligned with engine Pro GraphQL: **`TenantUser`** now has **`Phone`** (no **`Username`**). **`LoginTenantUser`** is now **`LoginTenantUser(ctx, projectID, LoginTenantUserParams)`** with **`Password`**, optional **`Email`** / **`Phone`**, optional **`AuthMethod`**. **`CreateTenantUser`** is **`CreateTenantUser(ctx, projectID, CreateTenantUserParams)`** (**`Password`**, optional **`Role`**, **`Email`**, **`Phone`**).
- Added **`UpdateTenantUser`** and **`DeleteTenantUser`** (arguments match system GraphQL; project scope comes from the API key).

### Migration

- Replace `LoginTenantUser(ctx, pid, username, password)` with  
  `LoginTenantUser(ctx, pid, LoginTenantUserParams{Password: password, Email: "..."})` or `Phone: "..."` per project sign-in mode.
- Replace `CreateTenantUser(ctx, pid, username, email, password, role)` with  
  `CreateTenantUser(ctx, pid, CreateTenantUserParams{Password: password, Role: role, Email: email, Phone: phone})`.

## [1.5.2] - 2026-05-13

### Changed (breaking — module path only)

- **Module path** is now **`github.com/apito-io/go-admin-sdk`** (was `github.com/apito-io/go-internal-sdk`). Update imports and `go get github.com/apito-io/go-admin-sdk@v1.5.2`. The Git remote may still point at a repository named `go-internal-sdk` until it is renamed on GitHub; the **module path** in `go.mod` is what `go get` uses.

### Fixed

- **Examples** and **README** use the **`go-admin-sdk`** import path consistently.

## [1.5.1] - 2026-05-13

### Changed (breaking)

- **`GenerateTenantToken`**: signature is now **`(ctx, tenantID, duration, role string)`**, aligned with engine `generateTenantToken` (`tenant_id`, `duration`, optional `role`). Removed the unused legacy **`token`** first argument. Empty **`duration`** still selects the default one-year-ahead expiry in UTC.
- **`github.com/apito-io/types`** `InternalSDKOperation` updated in lockstep. **`go.mod`** requires **`github.com/apito-io/types v0.1.10`** or newer.

## [1.5.0] - 2026-05-09

### Changed (breaking)

- **`SearchTenantsByDomain`**: signature is now `(ctx, projectID, domain)` only (no pagination). Response type renamed to **`TenantByDomainResponse`** with a single nullable **`Tenant`** field (exact per-project domain match; was list + count).

### Engine parity (documented)

- System GraphQL **`searchTenantsByDomain`** returns a single nullable **`tenant`** (no list/count). **`createTenant`** optional **`domain`** is rejected when that domain is already taken in the project; **`updateTenant`** enforces the same when **`domain`** is set to a non-empty value.

## [1.4.0] - 2026-05-09

### Added

- **Pro tenant catalog search by domain**: `SearchTenantsByDomain`; types `TenantCatalogSearchRow`, `TenantsByDomainResponse` (engine `searchTenantsByDomain` on system GraphQL).

## [1.3.0] - 2026-05-08

### Added

- **Pro tenant catalog users** (Apito Pro system GraphQL): `LoginTenantUser`, `LoginTenantUserGoogle`, `SearchTenantUsers`, `CreateTenantUser`; types `TenantUser`, `TenantLoginResponse`, `TenantUsersResponse`.
- **`examples/tenant_users`**: minimal runnable sample using env `APITO_BASE_URL`, `APITO_API_KEY`, `APITO_PROJECT_ID` (optional `APITO_TENANT_USERNAME` / `APITO_TENANT_PASSWORD` for login).

### Tests

- **`TestTenantUserProIntegration`**: optional live checks when `APITO_PROJECT_ID` is set; skipped otherwise.

## [1.2.0] - 2024-12-30

### Added

- 🎯 **Type-Safe Operations**: Complete generic typed methods for all operations
  - `GetSingleResourceTyped[T]()` for type-safe single resource retrieval
  - `SearchResourcesTyped[T]()` for type-safe search operations
  - `GetRelationDocumentsTyped[T]()` for type-safe relation queries
  - `CreateNewResourceTyped[T]()` for type-safe resource creation
  - `UpdateResourceTyped[T]()` for type-safe resource updates
- 🚀 **Comprehensive Todo Example**: Complete practical example demonstrating all SDK features
  - Authentication & tenant token generation
  - Resource creation (todos, users, categories)
  - Both typed and untyped search operations
  - Single resource retrieval
  - Resource updates with connections
  - Relation document queries
  - Audit logging
  - Debug functionality
  - Resource cleanup
- 📚 **Enhanced Documentation**: Completely rewritten README with comprehensive examples
  - Quick start guide
  - Complete API reference
  - Type system documentation
  - Plugin integration examples
  - Production deployment guides
  - Performance optimization tips
  - Error handling best practices
- 🔧 **Improved Request Structure**: New `CreateAndUpdateRequest` struct for cleaner API
- 📊 **Version Tracking**: Added `version.go` with `GetVersion()` function

### Changed

- 🔄 **Updated Client Interface**: Enhanced all methods to use the new request structure
- 📖 **Documentation**: Complete rewrite with practical examples and comprehensive coverage
- 🎨 **Example Structure**: Replaced basic example with comprehensive todo application

### Fixed

- 🐛 **Type Conversion**: Improved JSON marshaling/unmarshaling for typed operations
- 🔧 **Error Handling**: Enhanced GraphQL and HTTP error reporting

### Technical Details

- All generic functions follow the pattern: `OperationTyped[T](client, ctx, ...params)`
- Backward compatibility maintained for all existing non-typed methods
- Enhanced context support with tenant ID handling
- Improved connection pooling and performance optimizations

## [1.1.3] - Previous Version

- Previous features and bug fixes

## [1.1.2] - Previous Version

- Previous features and bug fixes

## [1.1.1] - Previous Version

- Previous features and bug fixes

## [1.1.0] - Previous Version

- Previous features and bug fixes

## [1.0.0] - Initial Release

- Initial SDK implementation
- Basic GraphQL communication
- API key authentication
- Core CRUD operations
