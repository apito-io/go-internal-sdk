package goapitosdk

import (
	"context"

	"gitlab.com/apito.io/buffers/protobuff"
	"gitlab.com/apito.io/buffers/shared"
)

// TypedSearchResult represents a search result with typed data
type TypedSearchResult[T any] struct {
	Results []*TypedDocumentStructure[T] `json:"results"`
	Count   int                          `json:"count"`
}

// TypedDocumentStructure represents a document with typed data
type TypedDocumentStructure[T any] struct {
	Key           string               `json:"_key,omitempty" firestore:"_key,omitempty" bson:"_key,omitempty"`
	Data          T                    `json:"data,omitempty" firestore:"data,omitempty" bson:"data,omitempty"`
	Meta          *protobuff.MetaField `json:"meta,omitempty" firestore:"meta,omitempty" bson:"meta,omitempty"`
	ID            string               `json:"id,omitempty" firestore:"id,omitempty" bson:"id,omitempty"`
	ExpireAt      int64                `json:"expire_at,omitempty" firestore:"expire_at,omitempty" bson:"expire_at,omitempty"`
	RelationDocID string               `json:"relation_doc_id,omitempty" firestore:"relation_doc_id,omitempty" bson:"relation_doc_id,omitempty"`
	Type          string               `json:"type,omitempty" firestore:"type,omitempty" bson:"type,omitempty"`
	TenantID      string               `json:"tenant_id,omitempty" firestore:"tenant_id,omitempty" bson:"tenant_id,omitempty"`
	TenantModel   string               `json:"tenant_model,omitempty" firestore:"tenant_model,omitempty" bson:"tenant_model,omitempty"`
}

type SearchResult struct {
	Results []*shared.DefaultDocumentStructure `json:"results"`
	Count   int                                `json:"count"`
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
	//GetProjectDetails(ctx context.Context, projectID string) (*protobuff.Project, error)

	// GetSingleResource retrieves a single resource by model and ID, with optional single page data
	GetSingleResource(ctx context.Context, model, _id string, singlePageData bool) (*shared.DefaultDocumentStructure, error)

	// SearchResources searches for resources in the specified model using the provided filter
	SearchResources(ctx context.Context, model string, filter map[string]interface{}, aggregate bool) (*SearchResult, error)

	// GetRelationDocuments retrieves related documents for the given ID and connection parameters
	GetRelationDocuments(ctx context.Context, _id string, connection map[string]interface{}) (*SearchResult, error)

	// CreateNewResource creates a new resource in the specified model with the given data and connections
	CreateNewResource(ctx context.Context, model string, data map[string]interface{}, connection map[string]interface{}) (*shared.DefaultDocumentStructure, error)

	// UpdateResource updates an existing resource by model and ID, with optional single page data, data updates, and connection changes
	UpdateResource(ctx context.Context, model, _id string, singlePageData bool, data map[string]interface{}, connect map[string]interface{}, disconnect map[string]interface{}) (*shared.DefaultDocumentStructure, error)

	// DeleteResource deletes a resource by model and ID
	DeleteResource(ctx context.Context, model, _id string) error

	// SendAuditLog sends an audit log entry to the audit log service
	SendAuditLog(ctx context.Context, auditData AuditData) error

	// Debug is used to debug the plugin, you can pass data here to debug the plugin
	Debug(ctx context.Context, stage string, data ...interface{}) (interface{}, error)
}

// TypedOperationsInterface defines the typed operations interface
// Since Go interfaces cannot have generic methods, we define these as separate generic functions
// that will be implemented as methods on the Client struct
type TypedOperationsInterface interface {
	// Note: These will be implemented as generic methods on the Client struct
	// GetSingleResourceTyped[T any](ctx context.Context, model, _id string, singlePageData bool) (*TypedDocumentStructure[T], error)
	// SearchResourcesTyped[T any](ctx context.Context, model string, filter map[string]interface{}, aggregate bool) (*TypedSearchResult[T], error)
	// GetRelationDocumentsTyped[T any](ctx context.Context, _id string, connection map[string]interface{}) (*TypedSearchResult[T], error)
	// CreateNewResourceTyped[T any](ctx context.Context, model string, data map[string]interface{}, connection map[string]interface{}) (*TypedDocumentStructure[T], error)
	// UpdateResourceTyped[T any](ctx context.Context, model, _id string, singlePageData bool, data map[string]interface{}, connect map[string]interface{}, disconnect map[string]interface{}) (*TypedDocumentStructure[T], error)
}
