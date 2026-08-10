package goapitosdk

// User is a project end-user from the engine system DB (table project_users).
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	TenantID  string `json:"tenant_id,omitempty"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// LoginUserParams configures login via system GraphQL loginUser.
// Password path (AuthMethod empty or "general"): set Password plus Email or Phone per project Authentication.
// Google path (AuthMethod "google"): set Code and State from OAuth callback; optionally use GoogleOAuthState first.
// Native mobile (AuthMethod "google_id_token"): set IDToken from google_sign_in (server client id).
// SaaS per-tenant separate DB: set TenantID (required by engine).
// Google paths: engine may auto-link a verified email to an existing user; errors include
// "google email not verified", "google account already linked to another user", "multiple users matched this email".
// On Cloudflare Workers v1, Google paths are unavailable ("google login is not available on Cloudflare Workers v1"); password login is supported.
type LoginUserParams struct {
	TenantID   string
	Password   string
	Email      string
	Phone      string
	AuthMethod string // optional; "", "general", "google", or "google_id_token"
	Code       string // OAuth authorization code (Google)
	State      string // OAuth state (from GoogleOAuthState or callback)
	IDToken    string // Google ID token (native sign-in)
}

// GoogleOAuthStateResponse is returned by googleOAuthState.
type GoogleOAuthStateResponse struct {
	State string
}

// CreateUserParams configures createUser. The engine requires an email or phone
// according to the project's general authentication identifier mode.
// Duplicate email/phone project-wide returns "email already exists for this project" or
// "phone already exists for this project".
// Pro SaaS: set TenantID (catalog tenant); sent as GraphQL tenant_id.
type CreateUserParams struct {
	Password string
	Role     string // optional; engine defaults when empty
	Email    string
	Phone    string
	TenantID string
}

// UpdateUserParams lists optional fields for updateUser. Nil pointers are omitted from the mutation.
// Duplicate email/phone project-wide returns stable engine validation errors.
// Pro SaaS: TenantID scopes the admin operation (GraphQL tenant_id).
type UpdateUserParams struct {
	Email    *string
	Phone    *string
	Role     *string
	TenantID *string
}

// LoginUserResponse is returned by loginUser (general or Google code flow).
type LoginUserResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// UsersResponse is returned by searchUsers.
type UsersResponse struct {
	Users []*User `json:"users"`
	Count int     `json:"count"`
}

// TenantCatalogSearchRow is one catalog tenant row from searchTenants or searchTenantsByDomain.
type TenantCatalogSearchRow struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Domain    string `json:"domain"`
	Data      string `json:"data"`
	Icon      string `json:"icon,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	PlanTier  string `json:"plan_tier,omitempty"`
}

// MyTenant is public GraphQL myTenant for the authenticated app-user token.
type MyTenant struct {
	ID       string `json:"id"`
	Name     string `json:"name,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Status   string `json:"status,omitempty"`
	PlanTier string `json:"plan_tier,omitempty"`
}

// EffectiveModelPermission is one api_permissions entry from myEffectivePermissions.
type EffectiveModelPermission struct {
	Read   string `json:"read,omitempty"`
	Create string `json:"create,omitempty"`
	Update string `json:"update,omitempty"`
	Delete string `json:"delete,omitempty"`
	Grace  bool   `json:"grace,omitempty"`
}

// EffectivePermissionsSnapshot mirrors engine myEffectivePermissions.
type EffectivePermissionsSnapshot struct {
	PlanSlug        string                               `json:"plan_slug,omitempty"`
	RoleID          string                               `json:"role_id,omitempty"`
	PlanClamped     bool                                 `json:"plan_clamped,omitempty"`
	APIPermissions  map[string]*EffectiveModelPermission `json:"api_permissions,omitempty"`
	LogicExecutions []string                             `json:"logic_executions,omitempty"`
	Quotas          map[string]float64                   `json:"quotas,omitempty"`
	Usage           map[string]float64                   `json:"usage,omitempty"`
	GraceModels     []string                             `json:"grace_models,omitempty"`
	IsAdmin         bool                                 `json:"is_admin,omitempty"`
}

// TenantByDomainResponse is returned by searchTenantsByDomain (at most one match per project).
type TenantByDomainResponse struct {
	Tenant *TenantCatalogSearchRow `json:"tenant"`
}

// SearchTenantsResponse is returned by searchTenants.
type SearchTenantsResponse struct {
	Tenants []*TenantCatalogSearchRow `json:"tenants"`
	Count   int                       `json:"count"`
}

// TenantCatalogListItem is one row from getTenants (includes icon from catalog data JSON).
type TenantCatalogListItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain,omitempty"`
	Icon   string `json:"icon,omitempty"`
	Data   string `json:"data,omitempty"`
}

// GetTenantsResponse is returned by getTenants.
type GetTenantsResponse struct {
	Tenants []*TenantCatalogListItem `json:"tenants"`
}

// CreateTenantParams configures createTenant on system GraphQL (SaaS catalog only).
type CreateTenantParams struct {
	Name   string
	Data   string // optional JSON string stored on catalog row
	Domain string
}

// UpdateTenantParams configures updateTenant. Empty strings are omitted except when clearing domain via explicit empty if needed.
type UpdateTenantParams struct {
	Name   *string
	Data   *string
	Domain *string
}

// File is metadata for a project file returned by the /secured/files REST API (stored in the project DB).
type File struct {
	ID            string `json:"id"`
	FileType      string `json:"file_type"`
	FileName      string `json:"file_name"`
	FileExtension string `json:"file_extension,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	Size          int64  `json:"size"`
	URL           string `json:"url"`
	CreatedBy     string `json:"created_by,omitempty"`
	CreatedAt     string `json:"created_at,omitempty"`
}

// FilesListResponse is returned by ListFiles.
type FilesListResponse struct {
	Files []File `json:"files"`
	Total int    `json:"total"`
}

// UploadFileParams configures UploadFile.
type UploadFileParams struct {
	FileName string
	Content  []byte
	FileType string // optional; inferred from content type when empty
}

// DeleteFilesResponse is returned by DeleteFiles.
type DeleteFilesResponse struct {
	Success        bool     `json:"success"`
	DeletedIDs     []string `json:"deleted_ids"`
	StorageFailed  []string `json:"storage_failed,omitempty"`
	Message        string   `json:"message,omitempty"`
}
