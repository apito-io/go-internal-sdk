package goapitosdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/apito-io/types"
	"github.com/apito-io/types/interfaces"
)

// Client represents the Apito SDK client
type Client struct {
	baseURL     string
	restBaseURL string
	apiKey      string
	projectID   string
	httpClient  *http.Client
}

// Config represents the SDK configuration
type Config struct {
	BaseURL     string        // Base URL of the Apito GraphQL endpoint (e.g. http://host:5050/system/graphql)
	RestBaseURL string        // Optional REST base (e.g. http://host:5050/system); derived from BaseURL when empty
	APIKey      string        // Project API key (ak_…) or legacy field; prefer AccessToken for apt_ automation
	AccessToken string        // Unified system access token (apt_…) → Authorization: Bearer
	ProjectID   string        // Optional default project scope → X-Apito-Project-Id
	Timeout     time.Duration // HTTP client timeout (default: 30 seconds)
	HTTPClient  *http.Client  // Custom HTTP client (optional)
}

// NewClient creates a new Apito SDK client
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	restBase := strings.TrimSpace(config.RestBaseURL)
	if restBase == "" {
		restBase = deriveRestBaseURL(config.BaseURL)
	}

	key := strings.TrimSpace(config.AccessToken)
	if key == "" {
		key = strings.TrimSpace(config.APIKey)
	}

	return &Client{
		baseURL:     config.BaseURL,
		restBaseURL: restBase,
		apiKey:      key,
		projectID:   strings.TrimSpace(config.ProjectID),
		httpClient:  httpClient,
	}
}

// errTokenFormatRetired mirrors the engine's TOKEN_FORMAT_RETIRED rejection so
// callers get the same guidance whether the client or the server catches the
// legacy cli-/sdk-/mcp- token format.
var errTokenFormatRetired = fmt.Errorf(
	"TOKEN_FORMAT_RETIRED: cli-/sdk-/mcp- prefixed keys are no longer accepted. " +
		"Generate a unified apt_ access token in Console → Access Token and set it as Config.AccessToken",
)

// isRetiredTokenPrefix reports legacy cli-/sdk-/mcp- token formats.
func isRetiredTokenPrefix(key string) bool {
	return strings.HasPrefix(key, "cli-") || strings.HasPrefix(key, "sdk-") || strings.HasPrefix(key, "mcp-")
}

// applyAuthCredential sets the request's auth header from key. Unified apt_
// access tokens are sent as Authorization: Bearer only (X-Use-Cookies: false
// tells the engine this is a headless API call, no browser session cookies) —
// hard cut, no compatibility X-Apito-Key fallback. Legacy project keys
// without a recognized prefix still use X-Apito-Key. Retired cli-/sdk-/mcp-
// prefixed keys are rejected by the caller before this is reached.
func applyAuthCredential(req *http.Request, key string) {
	key = strings.TrimSpace(key)
	if key == "" {
		return
	}
	if strings.HasPrefix(key, "apt_") {
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("X-Use-Cookies", "false")
		return
	}
	req.Header.Set("X-Apito-Key", key)
}

func deriveRestBaseURL(graphqlURL string) string {
	u := strings.TrimSuffix(strings.TrimSpace(graphqlURL), "/")
	if strings.HasSuffix(u, "/graphql") {
		base := strings.TrimSuffix(u, "/graphql")
		// Project file REST lives on /secured even when GraphQL uses /system/graphql.
		if strings.HasSuffix(base, "/system") {
			return strings.TrimSuffix(base, "/system") + "/secured"
		}
		return base
	}
	return u
}

const headerApitoProjectID = "X-Apito-Project-Id"

// executeGraphQL executes a GraphQL query or mutation using the configured
// default project scope, when present.
func (c *Client) executeGraphQL(ctx context.Context, query string, variables map[string]interface{}) (*types.GraphQLResponse, error) {
	return c.executeGraphQLScoped(ctx, query, variables, "")
}

// executeGraphQLScoped sends the canonical project header. A non-empty
// projectID overrides Config.ProjectID for this request.
func (c *Client) executeGraphQLScoped(ctx context.Context, query string, variables map[string]interface{}, projectID string) (*types.GraphQLResponse, error) {

	if isRetiredTokenPrefix(c.apiKey) {
		return nil, errTokenFormatRetired
	}

	var tenantID string
	if ctx != nil && ctx.Value("tenant_id") != nil {
		tenantID, _ = ctx.Value("tenant_id").(string)
	}

	payload := map[string]interface{}{
		"query": query,
	}

	if variables != nil {
		payload["variables"] = variables
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	applyAuthCredential(req, c.apiKey)
	effectiveProjectID := strings.TrimSpace(projectID)
	if effectiveProjectID == "" {
		effectiveProjectID = c.projectID
	}
	if effectiveProjectID != "" {
		req.Header.Set(headerApitoProjectID, effectiveProjectID)
	}
	if tenantID != "" {
		req.Header.Set("X-Apito-Tenant-ID", tenantID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var response types.GraphQLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
	}

	if len(response.Errors) > 0 {
		return &response, fmt.Errorf("GraphQL errors: %v", response.Errors)
	}

	return &response, nil
}

// GenerateTenantToken generates a new tenant-scoped API key for the given tenant_id.
//
// Authentication uses the client's Config.AccessToken/APIKey (unified apt_
// tokens send Authorization: Bearer; legacy project keys send X-Apito-Key).
//
// duration is the token expiry calendar day (YYYY-MM-DD), matching the engine mutation.
// If duration is empty, a default of one calendar year ahead in UTC is used.
//
// role is optional; when empty the engine defaults the token role to "admin".
//
// Not available on Cloudflare Workers v1 ("tenant management is not available on Cloudflare Workers v1").
func (c *Client) GenerateTenantToken(ctx context.Context, tenantID, duration, role string) (string, error) {
	if strings.TrimSpace(tenantID) == "" {
		return "", fmt.Errorf("tenantID is required")
	}
	if strings.TrimSpace(duration) == "" {
		duration = time.Now().UTC().AddDate(1, 0, 0).Format("2006-01-02")
	}

	query := `
		mutation GenerateTenantToken($tenantId: String!, $duration: String!, $role: String) {
			generateTenantToken(tenant_id: $tenantId, duration: $duration, role: $role) {
				token
			}
		}
	`

	variables := map[string]interface{}{
		"tenantId": tenantID,
		"duration": duration,
	}
	if strings.TrimSpace(role) != "" {
		variables["role"] = role
	} else {
		variables["role"] = nil
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return "", fmt.Errorf("failed to generate tenant token: %w", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	result, ok := data["generateTenantToken"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	tokenStr, ok := result["token"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected token format")
	}

	return tokenStr, nil
}

func withTenantCtx(ctx context.Context, tenantID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(tenantID) == "" {
		return ctx
	}
	return context.WithValue(ctx, "tenant_id", tenantID)
}

func mapToUser(m map[string]interface{}) *User {
	if m == nil {
		return nil
	}
	u := &User{}
	if v, ok := m["id"].(string); ok {
		u.ID = v
	}
	if v, ok := m["email"].(string); ok {
		u.Email = v
	}
	if v, ok := m["phone"].(string); ok {
		u.Phone = v
	}
	if v, ok := m["role"].(string); ok {
		u.Role = v
	}
	if v, ok := m["tenant_id"].(string); ok {
		u.TenantID = v
	}
	if v, ok := m["provider"].(string); ok {
		u.Provider = v
	}
	if v, ok := m["status"].(string); ok {
		u.Status = v
	}
	if v, ok := m["created_at"].(string); ok {
		u.CreatedAt = v
	}
	if v, ok := m["updated_at"].(string); ok {
		u.UpdatedAt = v
	}
	return u
}

func mapToTenantCatalogSearchRow(m map[string]interface{}) *TenantCatalogSearchRow {
	if m == nil {
		return nil
	}
	r := &TenantCatalogSearchRow{}
	if v, ok := m["id"].(string); ok {
		r.ID = v
	}
	if v, ok := m["name"].(string); ok {
		r.Name = v
	}
	if v, ok := m["status"].(string); ok {
		r.Status = v
	}
	if v, ok := m["domain"].(string); ok {
		r.Domain = v
	}
	if v, ok := m["data"].(string); ok {
		r.Data = v
	}
	if v, ok := m["icon"].(string); ok {
		r.Icon = v
	}
	if v, ok := m["created_at"].(string); ok {
		r.CreatedAt = v
	}
	if v, ok := m["plan_tier"].(string); ok {
		r.PlanTier = v
	}
	return r
}

// LoginUser runs loginUser (password or Google OAuth code / id_token flow).
// Google paths may auto-link a verified email to an existing user; handle engine errors
// "google email not verified", "google account already linked to another user", "multiple users matched this email".
// On Cloudflare Workers v1, Google paths are unavailable; password login is supported.
func (c *Client) LoginUser(ctx context.Context, projectID string, params LoginUserParams) (*LoginUserResponse, error) {
	authMethod := strings.ToLower(strings.TrimSpace(params.AuthMethod))
	if authMethod == "" {
		authMethod = "general"
	}

	query := `
		query LoginUser($project_id: String!, $tenant_id: String, $password: String, $auth_method: String, $email: String, $phone: String, $code: String, $state: String, $id_token: String) {
			loginUser(project_id: $project_id, tenant_id: $tenant_id, password: $password, auth_method: $auth_method, email: $email, phone: $phone, code: $code, state: $state, id_token: $id_token) {
				token
				user {
					id
					email
					phone
					role
					provider
					tenant_id
					status
					created_at
					updated_at
				}
			}
		}
	`
	variables := map[string]interface{}{
		"project_id": projectID,
	}
	if tid := strings.TrimSpace(params.TenantID); tid != "" {
		variables["tenant_id"] = tid
	}
	if authMethod == "google" {
		if strings.TrimSpace(params.Code) == "" || strings.TrimSpace(params.State) == "" {
			return nil, fmt.Errorf("loginUser: code and state are required for google auth_method")
		}
		variables["auth_method"] = "google"
		variables["code"] = strings.TrimSpace(params.Code)
		variables["state"] = strings.TrimSpace(params.State)
	} else if authMethod == "facebook" || authMethod == "github" || authMethod == "x" || authMethod == "linkedin" {
		if strings.TrimSpace(params.Code) == "" || strings.TrimSpace(params.State) == "" {
			return nil, fmt.Errorf("loginUser: code and state are required for %s auth_method", authMethod)
		}
		variables["auth_method"] = authMethod
		variables["code"] = strings.TrimSpace(params.Code)
		variables["state"] = strings.TrimSpace(params.State)
	} else if authMethod == "google_id_token" {
		if strings.TrimSpace(params.IDToken) == "" {
			return nil, fmt.Errorf("loginUser: id_token is required for google_id_token auth_method")
		}
		variables["auth_method"] = "google_id_token"
		variables["id_token"] = strings.TrimSpace(params.IDToken)
	} else {
		if strings.TrimSpace(params.Password) == "" {
			return nil, fmt.Errorf("loginUser: password is required")
		}
		if strings.TrimSpace(params.Email) == "" && strings.TrimSpace(params.Phone) == "" {
			return nil, fmt.Errorf("loginUser: email or phone is required")
		}
		variables["password"] = params.Password
		if strings.TrimSpace(params.Email) != "" {
			variables["email"] = strings.TrimSpace(params.Email)
		}
		if strings.TrimSpace(params.Phone) != "" {
			variables["phone"] = strings.TrimSpace(params.Phone)
		}
	}

	response, err := c.executeGraphQLScoped(ctx, query, variables, projectID)
	if err != nil {
		return nil, fmt.Errorf("loginUser: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["loginUser"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected loginUser response")
	}
	token, _ := raw["token"].(string)
	var user *User
	if um, ok := raw["user"].(map[string]interface{}); ok {
		user = mapToUser(um)
	}
	return &LoginUserResponse{Token: token, User: user}, nil
}

// GoogleOAuthState fetches signed OAuth state for building the Google authorize URL (googleOAuthState query).
func (c *Client) GoogleOAuthState(ctx context.Context, projectID string) (*GoogleOAuthStateResponse, error) {
	query := `
		query GoogleOAuthState($project_id: String!) {
			googleOAuthState(project_id: $project_id) {
				state
			}
		}
	`
	variables := map[string]interface{}{"project_id": projectID}
	response, err := c.executeGraphQLScoped(ctx, query, variables, projectID)
	if err != nil {
		return nil, fmt.Errorf("googleOAuthState: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["googleOAuthState"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected googleOAuthState response")
	}
	state, _ := raw["state"].(string)
	if strings.TrimSpace(state) == "" {
		return nil, fmt.Errorf("googleOAuthState: empty state")
	}
	return &GoogleOAuthStateResponse{State: state}, nil
}

// OAuthState fetches signed OAuth state for a provider (oauthState query).
func (c *Client) OAuthState(ctx context.Context, projectID, provider string) (*GoogleOAuthStateResponse, error) {
	query := `
		query OAuthState($project_id: String!, $provider: String!) {
			oauthState(project_id: $project_id, provider: $provider) {
				state
			}
		}
	`
	variables := map[string]interface{}{"project_id": projectID, "provider": provider}
	response, err := c.executeGraphQLScoped(ctx, query, variables, projectID)
	if err != nil {
		return nil, fmt.Errorf("oauthState: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["oauthState"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected oauthState response")
	}
	state, _ := raw["state"].(string)
	if strings.TrimSpace(state) == "" {
		return nil, fmt.Errorf("oauthState: empty state")
	}
	return &GoogleOAuthStateResponse{State: state}, nil
}

// SearchUsers lists project end-users. tenantID is optional (pro SaaS catalog tenant filter).
// q is optional free-text filter on email, username, phone, or id (case-insensitive contains).
func (c *Client) SearchUsers(ctx context.Context, projectID string, limit, offset int, tenantID, q string) (*UsersResponse, error) {
	query := `
		query SearchUsers($project_id: String!, $limit: Int, $offset: Int, $tenant_id: String, $q: String) {
			searchUsers(project_id: $project_id, limit: $limit, offset: $offset, tenant_id: $tenant_id, q: $q) {
				count
				users {
					id
					email
					phone
					role
					provider
					tenant_id
					status
					created_at
					updated_at
				}
			}
		}
	`
	variables := map[string]interface{}{
		"project_id": projectID,
		"limit":      limit,
		"offset":     offset,
	}
	if tid := strings.TrimSpace(tenantID); tid != "" {
		variables["tenant_id"] = tid
	}
	if needle := strings.TrimSpace(q); needle != "" {
		variables["q"] = needle
	}
	response, err := c.executeGraphQLScoped(ctx, query, variables, projectID)
	if err != nil {
		return nil, fmt.Errorf("searchUsers: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["searchUsers"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected searchUsers response")
	}
	count := 0
	if v, ok := raw["count"].(float64); ok {
		count = int(v)
	}
	if v, ok := raw["count"].(int); ok {
		count = v
	}
	var users []*User
	if arr, ok := raw["users"].([]interface{}); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]interface{}); ok {
				users = append(users, mapToUser(m))
			}
		}
	}
	return &UsersResponse{Users: users, Count: count}, nil
}

// SearchTenantsByDomain returns the single SaaS catalog tenant for an exact domain match in the project, or nil tenant if none.
func (c *Client) SearchTenantsByDomain(ctx context.Context, projectID, domain string) (*TenantByDomainResponse, error) {
	query := `
		query SearchTenantsByDomain($project_id: String!, $domain: String!) {
			searchTenantsByDomain(project_id: $project_id, domain: $domain) {
				tenant {
					id
					name
					status
					domain
					data
				}
			}
		}
	`
	variables := map[string]interface{}{
		"project_id": projectID,
		"domain":     domain,
	}
	response, err := c.executeGraphQLScoped(ctx, query, variables, projectID)
	if err != nil {
		return nil, fmt.Errorf("searchTenantsByDomain: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["searchTenantsByDomain"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected searchTenantsByDomain response")
	}
	if raw["tenant"] == nil {
		return &TenantByDomainResponse{Tenant: nil}, nil
	}
	tm, ok := raw["tenant"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected searchTenantsByDomain.tenant response")
	}
	return &TenantByDomainResponse{Tenant: mapToTenantCatalogSearchRow(tm)}, nil
}

// SearchTenants lists SaaS catalog tenants with optional pagination, free-text filter, and status filter (system GraphQL only).
// q filters name, id, domain, and data (case-insensitive contains).
// status: active (default), deleted, or all.
func (c *Client) SearchTenants(ctx context.Context, projectID string, limit, offset int, q, status string) (*SearchTenantsResponse, error) {
	pid := strings.TrimSpace(projectID)
	if pid == "" {
		return nil, fmt.Errorf("searchTenants: projectID is required")
	}
	query := `
		query SearchTenants($project_id: String!, $limit: Int, $offset: Int, $q: String, $status: String) {
			searchTenants(project_id: $project_id, limit: $limit, offset: $offset, q: $q, status: $status) {
				count
				tenants {
					id
					name
					status
					domain
					icon
					data
					created_at
					plan_tier
				}
			}
		}
	`
	variables := map[string]interface{}{
		"project_id": pid,
	}
	if limit > 0 {
		variables["limit"] = limit
	}
	if offset > 0 {
		variables["offset"] = offset
	}
	if needle := strings.TrimSpace(q); needle != "" {
		variables["q"] = needle
	}
	if statusFilter := strings.TrimSpace(status); statusFilter != "" {
		variables["status"] = statusFilter
	}
	response, err := c.executeGraphQLScoped(ctx, query, variables, pid)
	if err != nil {
		return nil, fmt.Errorf("searchTenants: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["searchTenants"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected searchTenants response")
	}
	count := 0
	if v, ok := raw["count"].(float64); ok {
		count = int(v)
	}
	if v, ok := raw["count"].(int); ok {
		count = v
	}
	var tenants []*TenantCatalogSearchRow
	if arr, ok := raw["tenants"].([]interface{}); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]interface{}); ok {
				tenants = append(tenants, mapToTenantCatalogSearchRow(m))
			}
		}
	}
	return &SearchTenantsResponse{Tenants: tenants, Count: count}, nil
}

// GetTenant loads one SaaS catalog tenant by exact id (system GraphQL only).
// Uses SearchTenants with q = tenantID; returns nil when no exact id match.
// status defaults to "active" when empty.
func (c *Client) GetTenant(ctx context.Context, projectID, tenantID, status string) (*TenantCatalogSearchRow, error) {
	pid := strings.TrimSpace(projectID)
	tid := strings.TrimSpace(tenantID)
	if pid == "" {
		return nil, fmt.Errorf("getTenant: projectID is required")
	}
	if tid == "" {
		return nil, fmt.Errorf("getTenant: tenantID is required")
	}
	statusFilter := strings.TrimSpace(status)
	if statusFilter == "" {
		statusFilter = "active"
	}
	res, err := c.SearchTenants(ctx, pid, 5, 0, tid, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("getTenant: %w", err)
	}
	if res == nil {
		return nil, nil
	}
	for _, row := range res.Tenants {
		if row != nil && strings.TrimSpace(row.ID) == tid {
			return row, nil
		}
	}
	return nil, nil
}

func mapToTenantCatalogListItem(m map[string]interface{}) *TenantCatalogListItem {
	if m == nil {
		return nil
	}
	r := &TenantCatalogListItem{}
	if v, ok := m["id"].(string); ok {
		r.ID = v
	}
	if v, ok := m["name"].(string); ok {
		r.Name = v
	}
	if v, ok := m["domain"].(string); ok {
		r.Domain = v
	}
	if v, ok := m["icon"].(string); ok {
		r.Icon = v
	}
	if v, ok := m["data"].(string); ok {
		r.Data = v
	}
	return r
}

// GetTenants lists SaaS catalog tenants for the authenticated project (system GraphQL only).
func (c *Client) GetTenants(ctx context.Context) (*GetTenantsResponse, error) {
	query := `
		query GetTenants {
			getTenants {
				tenants {
					id
					name
					domain
					icon
					data
				}
			}
		}
	`
	response, err := c.executeGraphQL(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("getTenants: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["getTenants"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected getTenants response")
	}
	var tenants []*TenantCatalogListItem
	if arr, ok := raw["tenants"].([]interface{}); ok {
		for _, it := range arr {
			if m, ok := it.(map[string]interface{}); ok {
				tenants = append(tenants, mapToTenantCatalogListItem(m))
			}
		}
	}
	return &GetTenantsResponse{Tenants: tenants}, nil
}

// MyTenant loads public GraphQL myTenant for the authenticated app-user token (includes plan_tier).
// Call against /secured/graphql (or public) with the JWT from LoginUser.
func (c *Client) MyTenant(ctx context.Context) (*MyTenant, error) {
	query := `
		query MyTenant {
			myTenant {
				id
				name
				domain
				status
				plan_tier
			}
		}
	`
	response, err := c.executeGraphQL(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("myTenant: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["myTenant"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected myTenant response")
	}
	out := &MyTenant{}
	if v, ok := raw["id"].(string); ok {
		out.ID = v
	}
	if strings.TrimSpace(out.ID) == "" {
		return nil, fmt.Errorf("invalid myTenant response: missing id")
	}
	if v, ok := raw["name"].(string); ok {
		out.Name = v
	}
	if v, ok := raw["domain"].(string); ok {
		out.Domain = v
	}
	if v, ok := raw["status"].(string); ok {
		out.Status = v
	}
	if v, ok := raw["plan_tier"].(string); ok {
		out.PlanTier = v
	}
	return out, nil
}

// MyEffectivePermissions loads public GraphQL myEffectivePermissions (role ∩ plan ceiling).
func (c *Client) MyEffectivePermissions(ctx context.Context) (*EffectivePermissionsSnapshot, error) {
	query := `
		query MyEffectivePermissions {
			myEffectivePermissions {
				plan_slug
				role_id
				plan_clamped
				api_permissions
				logic_executions
				quotas
				usage
				grace_models
				is_admin
			}
		}
	`
	response, err := c.executeGraphQL(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("myEffectivePermissions: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["myEffectivePermissions"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected myEffectivePermissions response")
	}
	return ParseEffectivePermissionsSnapshot(raw)
}

// CreateTenant provisions a SaaS catalog tenant (system GraphQL only; not /secured/graphql).
func (c *Client) CreateTenant(ctx context.Context, params CreateTenantParams) (*TenantCatalogSearchRow, error) {
	name := strings.TrimSpace(params.Name)
	if name == "" {
		return nil, fmt.Errorf("createTenant: name is required")
	}
	query := `
		mutation CreateTenant($name: String!, $data: String, $domain: String) {
			createTenant(name: $name, data: $data, domain: $domain) {
				id
				name
				status
				domain
				data
			}
		}
	`
	variables := map[string]interface{}{"name": name}
	if d := strings.TrimSpace(params.Data); d != "" {
		variables["data"] = d
	}
	if dom := strings.TrimSpace(params.Domain); dom != "" {
		variables["domain"] = dom
	}
	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("createTenant: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	row, ok := data["createTenant"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected createTenant response")
	}
	return mapToTenantCatalogSearchRow(row), nil
}

// UpdateTenant updates name/data/domain on a catalog tenant row.
func (c *Client) UpdateTenant(ctx context.Context, tenantID string, params UpdateTenantParams) (*TenantCatalogSearchRow, error) {
	tid := strings.TrimSpace(tenantID)
	if tid == "" {
		return nil, fmt.Errorf("updateTenant: tenantID is required")
	}
	if params.Name == nil && params.Data == nil && params.Domain == nil {
		return nil, fmt.Errorf("updateTenant: at least one field must be provided")
	}
	query := `
		mutation UpdateTenant($tenant_id: String!, $name: String, $data: String, $domain: String) {
			updateTenant(tenant_id: $tenant_id, name: $name, data: $data, domain: $domain) {
				id
				name
				status
				domain
				data
			}
		}
	`
	variables := map[string]interface{}{"tenant_id": tid}
	if params.Name != nil {
		variables["name"] = *params.Name
	}
	if params.Data != nil {
		variables["data"] = *params.Data
	}
	if params.Domain != nil {
		variables["domain"] = *params.Domain
	}
	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("updateTenant: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	row, ok := data["updateTenant"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected updateTenant response")
	}
	return mapToTenantCatalogSearchRow(row), nil
}

// DeleteTenant soft-deletes a tenant from the system catalog (status=deleted).
func (c *Client) DeleteTenant(ctx context.Context, tenantID string) (bool, error) {
	tid := strings.TrimSpace(tenantID)
	if tid == "" {
		return false, fmt.Errorf("deleteTenant: tenantID is required")
	}
	query := `
		mutation DeleteTenant($tenant_id: String!) {
			deleteTenant(tenant_id: $tenant_id)
		}
	`
	variables := map[string]interface{}{"tenant_id": tid}
	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return false, fmt.Errorf("deleteTenant: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected response format")
	}
	okVal, ok := data["deleteTenant"].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected deleteTenant response")
	}
	return okVal, nil
}

// CreateUser creates a local-password project user via system GraphQL mutation createUser.
// Duplicate email/phone project-wide returns stable engine validation errors.
func (c *Client) CreateUser(ctx context.Context, projectID string, params CreateUserParams) (*User, error) {
	if strings.TrimSpace(params.Password) == "" {
		return nil, fmt.Errorf("createUser: password is required")
	}
	query := `
		mutation CreateUser($project_id: String!, $password: String!, $role: String, $email: String, $phone: String, $tenant_id: String) {
			createUser(project_id: $project_id, password: $password, role: $role, email: $email, phone: $phone, tenant_id: $tenant_id) {
				id
				email
				phone
				role
				provider
				tenant_id
				status
				created_at
				updated_at
			}
		}
	`
	variables := map[string]interface{}{
		"project_id": projectID,
		"password":   params.Password,
	}
	if strings.TrimSpace(params.Role) != "" {
		variables["role"] = strings.TrimSpace(params.Role)
	}
	if strings.TrimSpace(params.Email) != "" {
		variables["email"] = strings.TrimSpace(params.Email)
	}
	if strings.TrimSpace(params.Phone) != "" {
		variables["phone"] = strings.TrimSpace(params.Phone)
	}
	if tid := strings.TrimSpace(params.TenantID); tid != "" {
		variables["tenant_id"] = tid
	}
	response, err := c.executeGraphQLScoped(ctx, query, variables, projectID)
	if err != nil {
		return nil, fmt.Errorf("createUser: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["createUser"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected createUser response")
	}
	return mapToUser(raw), nil
}

// UpdateUser updates a project user by id (system GraphQL updateUser). Project scope comes from the API key.
// Duplicate email/phone project-wide returns stable engine validation errors.
func (c *Client) UpdateUser(ctx context.Context, userID string, params UpdateUserParams) (*User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("updateUser: user id is required")
	}
	has := params.Email != nil || params.Phone != nil || params.Role != nil || params.TenantID != nil
	if !has {
		return nil, fmt.Errorf("updateUser: at least one field must be set")
	}
	query := `
		mutation UpdateUser($user_id: String!, $email: String, $phone: String, $role: String, $tenant_id: String) {
			updateUser(user_id: $user_id, email: $email, phone: $phone, role: $role, tenant_id: $tenant_id) {
				id
				email
				phone
				role
				provider
				tenant_id
				status
				created_at
				updated_at
			}
		}
	`
	variables := map[string]interface{}{"user_id": userID}
	if params.Email != nil {
		variables["email"] = *params.Email
	}
	if params.Phone != nil {
		variables["phone"] = *params.Phone
	}
	if params.Role != nil {
		variables["role"] = *params.Role
	}
	if params.TenantID != nil {
		variables["tenant_id"] = *params.TenantID
	}
	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("updateUser: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	raw, ok := data["updateUser"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected updateUser response")
	}
	return mapToUser(raw), nil
}

// ResetUserPassword sets a new password for a project user (admin mutation resetUserPassword).
func (c *Client) ResetUserPassword(ctx context.Context, userID, password string) (bool, error) {
	if strings.TrimSpace(userID) == "" {
		return false, fmt.Errorf("resetUserPassword: user id is required")
	}
	if strings.TrimSpace(password) == "" {
		return false, fmt.Errorf("resetUserPassword: password is required")
	}
	query := `
		mutation ResetUserPassword($user_id: String!, $password: String!) {
			resetUserPassword(user_id: $user_id, password: $password)
		}
	`
	response, err := c.executeGraphQL(ctx, query, map[string]interface{}{
		"user_id":  userID,
		"password": password,
	})
	if err != nil {
		return false, fmt.Errorf("resetUserPassword: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected response format")
	}
	okOut, ok := data["resetUserPassword"].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected resetUserPassword response")
	}
	return okOut, nil
}

// DeleteUser removes a project user by id (system GraphQL deleteUser). Project scope comes from the API key.
func (c *Client) DeleteUser(ctx context.Context, userID string) (bool, error) {
	if strings.TrimSpace(userID) == "" {
		return false, fmt.Errorf("deleteUser: user id is required")
	}
	query := `
		mutation DeleteUser($user_id: String!) {
			deleteUser(user_id: $user_id)
		}
	`
	response, err := c.executeGraphQL(ctx, query, map[string]interface{}{"user_id": userID})
	if err != nil {
		return false, fmt.Errorf("deleteUser: %w", err)
	}
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected response format")
	}
	okOut, ok := data["deleteUser"].(bool)
	if !ok {
		return false, fmt.Errorf("unexpected deleteUser response")
	}
	return okOut, nil
}

// =============================================================================
// TYPED GENERIC FUNCTIONS
// =============================================================================

// GetSingleResourceTyped retrieves a single resource by model and ID with typed data
func GetSingleResourceTyped[T any](c *Client, ctx context.Context, model, _id string, singlePageData bool) (*types.TypedDocumentStructure[T], error) {
	rawDocument, err := c.GetSingleResource(ctx, model, _id, singlePageData)
	if err != nil {
		return nil, err
	}
	return convertToTypedDocument[T](rawDocument)
}

// SearchResourcesTyped searches for resources with typed results
func SearchResourcesTyped[T any](c *Client, ctx context.Context, model string, filter map[string]interface{}, aggregate bool) (*types.TypedSearchResult[T], error) {
	rawResults, err := c.SearchResources(ctx, model, filter, aggregate)
	if err != nil {
		return nil, err
	}
	return convertToTypedSearchResult[T](rawResults)
}

// GetRelationDocumentsTyped retrieves related documents with typed results
func GetRelationDocumentsTyped[T any](c *Client, ctx context.Context, _id string, connection map[string]interface{}) (*types.TypedSearchResult[T], error) {
	rawResults, err := c.GetRelationDocuments(ctx, _id, connection)
	if err != nil {
		return nil, err
	}
	return convertToTypedSearchResult[T](rawResults)
}

// CreateNewResourceTyped creates a new resource with typed result
func CreateNewResourceTyped[T any](c *Client, ctx context.Context, request *types.CreateAndUpdateRequest) (*types.TypedDocumentStructure[T], error) {
	rawDocument, err := c.CreateNewResource(ctx, request)
	if err != nil {
		return nil, err
	}
	return convertToTypedDocument[T](rawDocument)
}

// UpdateResourceTyped updates a resource with typed result
func UpdateResourceTyped[T any](c *Client, ctx context.Context, request *types.CreateAndUpdateRequest) (*types.TypedDocumentStructure[T], error) {
	rawDocument, err := c.UpdateResource(ctx, request)
	if err != nil {
		return nil, err
	}
	return convertToTypedDocument[T](rawDocument)
}

// =============================================================================
// HELPER FUNCTIONS FOR TYPE CONVERSION
// =============================================================================

// convertToTypedDocument converts a raw DefaultDocumentStructure to a typed document
func convertToTypedDocument[T any](rawDoc *types.DefaultDocumentStructure) (*types.TypedDocumentStructure[T], error) {
	dataJSON, err := json.Marshal(rawDoc.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw data: %w", err)
	}

	var typedData T
	if err := json.Unmarshal(dataJSON, &typedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to typed data: %w", err)
	}

	return &types.TypedDocumentStructure[T]{
		Key:           rawDoc.Key,
		Data:          typedData,
		Meta:          rawDoc.Meta,
		ID:            rawDoc.ID,
		ExpireAt:      parseExpireAt(rawDoc.ExpireAt),
		RelationDocID: rawDoc.RelationDocID,
		Type:          rawDoc.Type,
	}, nil
}

// convertToTypedSearchResult converts a raw SearchResult to a typed search result
func convertToTypedSearchResult[T any](rawResults *types.SearchResult) (*types.TypedSearchResult[T], error) {
	typedResults := &types.TypedSearchResult[T]{
		Count:   rawResults.Count,
		Results: make([]*types.TypedDocumentStructure[T], len(rawResults.Results)),
	}

	for i, rawDoc := range rawResults.Results {
		typedDoc, err := convertToTypedDocument[T](rawDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert document at index %d: %w", i, err)
		}
		typedResults.Results[i] = typedDoc
	}

	return typedResults, nil
}

// parseExpireAt converts string expire_at to int64
func parseExpireAt(expireAt string) int64 {
	if expireAt == "" {
		return 0
	}
	return 0 // Could implement actual parsing logic here
}

// =============================================================================
// BACKWARD COMPATIBLE METHODS (Non-generic versions)
// =============================================================================

/* // GetProjectDetails retrieves project details for the given project ID
func (c *Client) GetProjectDetails(ctx context.Context, projectID string) (*protobuff.Project, error) {
	query := `
		query GetProject($_id: String) {
			getProject(_id: $_id) {
				id
				name
				description
				created_at
				updated_at
				settings
				tenant_model_name
				project_secret_key
				status
				organization_id
			}
		}
	`

	variables := map[string]interface{}{}
	if projectID != "" {
		variables["_id"] = projectID
	}

	response, err := c.executeGraphQLScoped(ctx, query, variables, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project details: %w", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	projectData, ok := data["getProject"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("project not found or unexpected response format")
	}

	// Convert the response to Project struct
	projectJSON, err := json.Marshal(projectData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project data: %w", err)
	}

	var project protobuff.Project
	if err := json.Unmarshal(projectJSON, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project data: %w", err)
	}

	return &project, nil
} */

// GetSingleResource retrieves a single resource by model and ID, with optional single page data
func (c *Client) GetSingleResource(ctx context.Context, model, _id string, singlePageData bool) (*types.DefaultDocumentStructure, error) {
	query := `
		query GetSingleData($model: String, $_id: String!, $single_page_data: Boolean) {
			getSingleData(model: $model, _id: $_id, single_page_data: $single_page_data) {
				_key
				data
				meta {
				created_at
				updated_at
				status
				revision
				revision_at
				}
				id
				expire_at
				relation_doc_id
				type
			}
		}
	`
	variables := map[string]interface{}{
		"model":            model,
		"_id":              _id,
		"single_page_data": singlePageData,
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get single resource: %w", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	singleDataRaw, ok := data["getSingleData"]
	if !ok {
		return nil, fmt.Errorf("getSingleData not found in response")
	}

	// Convert interface{} to *shared.DefaultDocumentStructure
	singleDataJSON, err := json.Marshal(singleDataRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal getSingleData: %w", err)
	}

	var document types.DefaultDocumentStructure
	if err := json.Unmarshal(singleDataJSON, &document); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getSingleData: %w", err)
	}

	return &document, nil
}

// SearchResources searches for resources in the specified model using the provided filter
func (c *Client) SearchResources(ctx context.Context, model string, filter map[string]interface{}, aggregate bool) (*types.SearchResult, error) {
	query := `
		query GetModelData($model: String!, $page: Int, $limit: Int, $_key: JSON, $where: JSON, $search: String) {
			getModelData(model: $model, page: $page, limit: $limit, _key: $_key, where: $where, search: $search) {
				results {
					id
					relation_doc_id
					data
					type
					expire_at
					meta {
						created_at
						updated_at
						status
						root_revision_id
					}
				}
				count
			}
		}
	`

	variables := map[string]interface{}{
		"model": model,
	}

	// Add filter parameters if provided
	if filter != nil {
		if _key, ok := filter["_key"]; ok {
			variables["_key"] = _key
		}
		if page, ok := filter["page"]; ok {
			variables["page"] = page
		}
		if limit, ok := filter["limit"]; ok {
			variables["limit"] = limit
		}
		if where, ok := filter["where"]; ok {
			variables["where"] = where
		}
		if search, ok := filter["search"]; ok {
			variables["search"] = search
		}
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to search resources: %w", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	modelDataRaw, ok := data["getModelData"]
	if !ok {
		return nil, fmt.Errorf("getModelData not found in response")
	}

	// Convert interface{} to SearchResult
	modelDataJSON, err := json.Marshal(modelDataRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal getModelData: %w", err)
	}

	var searchResult types.SearchResult
	if err := json.Unmarshal(modelDataJSON, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getModelData: %w", err)
	}

	return &searchResult, nil
}

// GetRelationDocuments retrieves related documents for the given ID and connection parameters
func (c *Client) GetRelationDocuments(ctx context.Context, _id string, connection map[string]interface{}) (*types.SearchResult, error) {
	query := `
		query GetModelData($model: String!, $page: Int, $limit: Int, $where: JSON, $search: String, $connection : ListAllDataDetailedOfAModelConnectionPayload) {
			getModelData(model: $model, page: $page, limit: $limit, where: $where, search: $search, connection: $connection) {
				results {
					id
					relation_doc_id
					data
					type
					expire_at
					meta {
						created_at
						updated_at
						status
						root_revision_id
					}
				}
				count
			}
		}
	`

	variables := map[string]interface{}{
		"connection": connection,
	}

	// Extract model from connection if available
	if model, ok := connection["model"].(string); ok {
		variables["model"] = model
	} else {
		return nil, fmt.Errorf("model is required in connection parameters")
	}

	// Add filter parameters if provided in connection
	if filter, ok := connection["filter"].(map[string]interface{}); ok {
		if page, ok := filter["page"]; ok {
			variables["page"] = page
		}
		if limit, ok := filter["limit"]; ok {
			variables["limit"] = limit
		}
		if where, ok := filter["where"]; ok {
			variables["where"] = where
		}
		if search, ok := filter["search"]; ok {
			variables["search"] = search
		}
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get relation documents: %w", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	modelDataRaw, ok := data["getModelData"]
	if !ok {
		return nil, fmt.Errorf("getModelData not found in response")
	}

	// Convert interface{} to SearchResult
	modelDataJSON, err := json.Marshal(modelDataRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal getModelData: %w", err)
	}

	var searchResult types.SearchResult
	if err := json.Unmarshal(modelDataJSON, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getModelData: %w", err)
	}

	return &searchResult, nil
}

// CreateNewResource creates a new resource in the specified model with the given data and connections
func (c *Client) CreateNewResource(ctx context.Context, request *types.CreateAndUpdateRequest) (*types.DefaultDocumentStructure, error) {

	if request.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if request.Payload == nil {
		return nil, fmt.Errorf("payload is required")
	}

	query := `
		mutation CreateNewData($model: String!, $single_page_data: Boolean, $payload: JSON!, $connect: JSON) {
			upsertModelData(
				connect: $connect
				model_name: $model
				single_page_data: $single_page_data
				payload: $payload
			) {
				id
				type
				data
				meta {
					created_at
					updated_at
					status
					revision
					revision_at
				}
			}
		}
	`

	variables := map[string]interface{}{
		"model":            request.Model,
		"payload":          request.Payload,
		"single_page_data": request.SinglePageData,
	}

	if request.Connect != nil {
		variables["connect"] = request.Connect
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to create new resource: %w", err)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	singleDataRaw, ok := responseData["upsertModelData"]
	if !ok {
		return nil, fmt.Errorf("upsertModelData not found in response")
	}

	// Convert interface{} to *shared.DefaultDocumentStructure
	singleDataJSON, err := json.Marshal(singleDataRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal getSingleData: %w", err)
	}

	var document types.DefaultDocumentStructure
	if err := json.Unmarshal(singleDataJSON, &document); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getSingleData: %w", err)
	}

	return &document, nil
}

// UpdateResource updates an existing resource by model and ID, with optional single page data, data updates, and connection changes
func (c *Client) UpdateResource(ctx context.Context, request *types.CreateAndUpdateRequest) (*types.DefaultDocumentStructure, error) {
	// fetch tenant_id from data if available

	if request.ID == "" {
		return nil, fmt.Errorf("id is required")
	}

	if request.Model == "" {
		return nil, fmt.Errorf("model is required")
	}

	if request.Payload == nil {
		return nil, fmt.Errorf("payload is required")
	}

	query := `
		mutation UpdateModelData($_id: String!, $model: String!, $single_page_data: Boolean, $force_update: Boolean, $payload: JSON!, $connect: JSON, $disconnect: JSON) {
			upsertModelData(
				connect: $connect
				model_name: $model
				single_page_data: $single_page_data
				force_update: $force_update
				disconnect: $disconnect
				_id: $_id
				payload: $payload
			) {
				id
				type
				data
				meta {
					created_at
					updated_at
					status
					revision
					revision_at
				}
			}
		}
	`

	variables := map[string]interface{}{
		"_id":              request.ID,
		"model":            request.Model,
		"payload":          request.Payload,
		"single_page_data": request.SinglePageData,
		"force_update":     request.ForceUpdate,
	}

	if request.Connect != nil {
		variables["connect"] = request.Connect
	}
	if request.Disconnect != nil {
		variables["disconnect"] = request.Disconnect
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	singleDataRaw, ok := responseData["upsertModelData"]
	if !ok {
		return nil, fmt.Errorf("upsertModelData not found in response")
	}

	// Convert interface{} to *shared.DefaultDocumentStructure
	singleDataJSON, err := json.Marshal(singleDataRaw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal getSingleData: %w", err)
	}

	var document types.DefaultDocumentStructure
	if err := json.Unmarshal(singleDataJSON, &document); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getSingleData: %w", err)
	}

	return &document, nil
}

// DeleteResource deletes a resource by model and ID
func (c *Client) DeleteResource(ctx context.Context, model, _id string) error {
	// Note: This is a placeholder implementation as the exact mutation wasn't found in the schema
	// You would need to implement the actual deleteData mutation based on your GraphQL schema
	query := `
		mutation DeleteData($model: String!, $_id: String!) {
			deleteModelData(model_name: $model, _id: $_id) {
				id
			}
		}
	`

	variables := map[string]interface{}{
		"model": model,
		"_id":   _id,
	}

	_, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

// Debug is used to debug the plugin, you can pass data here to debug the plugin
func (c *Client) Debug(ctx context.Context, stage string, data ...interface{}) (interface{}, error) {
	// Note: This is a placeholder implementation as the exact mutation wasn't found in the schema
	// You would need to implement the actual debug mutation based on your GraphQL schema
	query := `
		mutation Debug($stage: String!, $data: JSON) {
			debug(stage: $stage, data: $data) {
				message
				data
			}
		}
	`

	variables := map[string]interface{}{
		"stage": stage,
		"data":  data,
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to debug: %w", err)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	return responseData["debug"], nil
}

// Verify that Client implements InjectedDBOperationInterface
var _ interfaces.InternalSDKOperation = (*Client)(nil)
