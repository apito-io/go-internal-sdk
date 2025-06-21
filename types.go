package goapitosdk

import (
	"context"
	"time"
)

// Project represents the project details
type Project struct {
	ID                  string                 `json:"id,omitempty"`
	Name                string                 `json:"name,omitempty"`
	Description         string                 `json:"description,omitempty"`
	CreatedAt           time.Time              `json:"created_at,omitempty"`
	UpdatedAt           time.Time              `json:"updated_at,omitempty"`
	Settings            map[string]interface{} `json:"settings,omitempty"`
	TenantModelName     string                 `json:"tenant_model_name,omitempty"`
	ProjectSecretKey    string                 `json:"project_secret_key,omitempty"`
	Status              string                 `json:"status,omitempty"`
	OrganizationID      string                 `json:"organization_id,omitempty"`
	DatabaseCredentials *DriverCredentials     `json:"database_credentials,omitempty"`
}

// DriverCredentials represents database connection credentials
type DriverCredentials struct {
	Host      string   `json:"host,omitempty"`
	Port      string   `json:"port,omitempty"`
	Database  []string `json:"db,omitempty"`
	User      []string `json:"user,omitempty"`
	Password  []string `json:"password,omitempty"`
	AccessKey []string `json:"access_key,omitempty"`
	SecretKey []string `json:"secret_key,omitempty"`
	Driver    string   `json:"driver,omitempty"`
}

// AuditData represents audit log data
type AuditData struct {
	Resource         string                 `json:"resource"`
	Action           string                 `json:"action"`
	Author           map[string]interface{} `json:"author"`
	Data             map[string]interface{} `json:"data"`
	Meta             map[string]interface{} `json:"meta"`
	AdditionalFields map[string]interface{} `json:"-"` // Fields to be added directly to the flattened log
}

// Filter represents query filter parameters
type Filter struct {
	Page     int    `json:"page,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Order    string `json:"order,omitempty"`
	Min      int    `json:"min,omitempty"`
	Max      int    `json:"max,omitempty"`
	Category string `json:"category,omitempty"`
}

// GraphQLResponse represents a generic GraphQL response
type GraphQLResponse struct {
	Data   interface{}    `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents the location of a GraphQL error
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// InjectedDBOperationInterface defines the interface that this SDK implements
// This matches the interface from the main Apito Engine
type InjectedDBOperationInterface interface {
	// GenerateTenantToken generates a new tenant token for the specified tenant ID
	GenerateTenantToken(ctx context.Context, token string, tenantID string) (string, error)

	// GetProjectDetails retrieves project details for the given project ID
	GetProjectDetails(ctx context.Context, projectID string) (*Project, error)

	// GetSingleResource retrieves a single resource by model and ID, with optional single page data
	GetSingleResource(ctx context.Context, model, _id string, singlePageData bool) (interface{}, error)

	// SearchResources searches for resources in the specified model using the provided filter
	SearchResources(ctx context.Context, model string, filter map[string]interface{}, aggregate bool) (interface{}, error)

	// GetRelationDocuments retrieves related documents for the given ID and connection parameters
	GetRelationDocuments(ctx context.Context, _id string, connection map[string]interface{}) (interface{}, error)

	// CreateNewResource creates a new resource in the specified model with the given data and connections
	CreateNewResource(ctx context.Context, model string, data map[string]interface{}, connection map[string]interface{}) (interface{}, error)

	// UpdateResource updates an existing resource by model and ID, with optional single page data, data updates, and connection changes
	UpdateResource(ctx context.Context, model, _id string, singlePageData bool, data map[string]interface{}, connect map[string]interface{}, disconnect map[string]interface{}) (interface{}, error)

	// DeleteResource deletes a resource by model and ID
	DeleteResource(ctx context.Context, model, _id string) error

	// SendAuditLog sends an audit log entry to the audit log service
	SendAuditLog(ctx context.Context, auditData AuditData) error

	// Debug is used to debug the plugin, you can pass data here to debug the plugin
	Debug(ctx context.Context, stage string, data ...interface{}) (interface{}, error)
}
