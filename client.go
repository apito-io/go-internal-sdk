package goapitosdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitlab.com/apito.io/buffers/shared"
)

// Client represents the Apito SDK client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Config represents the SDK configuration
type Config struct {
	BaseURL    string        // Base URL of the Apito GraphQL endpoint
	APIKey     string        // API key for authentication (X-APITO-KEY header)
	Timeout    time.Duration // HTTP client timeout (default: 30 seconds)
	HTTPClient *http.Client  // Custom HTTP client (optional)
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

	return &Client{
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		httpClient: httpClient,
	}
}

// executeGraphQL executes a GraphQL query or mutation
func (c *Client) executeGraphQL(ctx context.Context, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	
	var tenantID string
	if ctx.Value("tenant_id") != nil {
		tenantID = ctx.Value("tenant_id").(string)
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
	req.Header.Set("X-Apito-Key", c.apiKey)
	req.Header.Set("X-Apito-Tenant-ID", tenantID)


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

	var response GraphQLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal GraphQL response: %w", err)
	}

	if len(response.Errors) > 0 {
		return &response, fmt.Errorf("GraphQL errors: %v", response.Errors)
	}

	return &response, nil
}

// GenerateTenantToken generates a new tenant token for the specified tenant ID
func (c *Client) GenerateTenantToken(ctx context.Context, token string, tenantID string) (string, error) {
	query := `
		mutation GenerateTenantToken($token: String!, $tenantId: String!) {
			generateTenantToken(token: $token, tenant_id: $tenantId) {
				token
			}
		}
	`

	variables := map[string]interface{}{
		"token":    token,
		"tenantId": tenantID,
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

// =============================================================================
// TYPED GENERIC FUNCTIONS
// =============================================================================

// GetSingleResourceTyped retrieves a single resource by model and ID with typed data
func GetSingleResourceTyped[T any](c *Client, ctx context.Context, model, _id string, singlePageData bool) (*TypedDocumentStructure[T], error) {
	rawDocument, err := c.GetSingleResource(ctx, model, _id, singlePageData)
	if err != nil {
		return nil, err
	}
	return convertToTypedDocument[T](rawDocument)
}

// SearchResourcesTyped searches for resources with typed results
func SearchResourcesTyped[T any](c *Client, ctx context.Context, model string, filter map[string]interface{}, aggregate bool) (*TypedSearchResult[T], error) {
	rawResults, err := c.SearchResources(ctx, model, filter, aggregate)
	if err != nil {
		return nil, err
	}
	return convertToTypedSearchResult[T](rawResults)
}

// GetRelationDocumentsTyped retrieves related documents with typed results
func GetRelationDocumentsTyped[T any](c *Client, ctx context.Context, _id string, connection map[string]interface{}) (*TypedSearchResult[T], error) {
	rawResults, err := c.GetRelationDocuments(ctx, _id, connection)
	if err != nil {
		return nil, err
	}
	return convertToTypedSearchResult[T](rawResults)
}

// CreateNewResourceTyped creates a new resource with typed result
func CreateNewResourceTyped[T any](c *Client, ctx context.Context, request *CreateAndUpdateRequest) (*TypedDocumentStructure[T], error) {
	rawDocument, err := c.CreateNewResource(ctx, request)
	if err != nil {
		return nil, err
	}
	return convertToTypedDocument[T](rawDocument)
}

// UpdateResourceTyped updates a resource with typed result
func UpdateResourceTyped[T any](c *Client, ctx context.Context, request *CreateAndUpdateRequest) (*TypedDocumentStructure[T], error) {
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
func convertToTypedDocument[T any](rawDoc *shared.DefaultDocumentStructure) (*TypedDocumentStructure[T], error) {
	dataJSON, err := json.Marshal(rawDoc.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw data: %w", err)
	}

	var typedData T
	if err := json.Unmarshal(dataJSON, &typedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to typed data: %w", err)
	}

	return &TypedDocumentStructure[T]{
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
func convertToTypedSearchResult[T any](rawResults *SearchResult) (*TypedSearchResult[T], error) {
	typedResults := &TypedSearchResult[T]{
		Count:   rawResults.Count,
		Results: make([]*TypedDocumentStructure[T], len(rawResults.Results)),
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

	response, err := c.executeGraphQL(ctx, query, variables)
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
func (c *Client) GetSingleResource(ctx context.Context, model, _id string, singlePageData bool) (*shared.DefaultDocumentStructure, error) {
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

	var document shared.DefaultDocumentStructure
	if err := json.Unmarshal(singleDataJSON, &document); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getSingleData: %w", err)
	}

	return &document, nil
}

// SearchResources searches for resources in the specified model using the provided filter
func (c *Client) SearchResources(ctx context.Context, model string, filter map[string]interface{}, aggregate bool) (*SearchResult, error) {
	query := `
		query GetModelData($model: String!, $page: Int, $limit: Int, $where: JSON, $search: String) {
			getModelData(model: $model, page: $page, limit: $limit, where: $where, search: $search) {
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

	var searchResult SearchResult
	if err := json.Unmarshal(modelDataJSON, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getModelData: %w", err)
	}

	return &searchResult, nil
}

// GetRelationDocuments retrieves related documents for the given ID and connection parameters
func (c *Client) GetRelationDocuments(ctx context.Context, _id string, connection map[string]interface{}) (*SearchResult, error) {
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

	var searchResult SearchResult
	if err := json.Unmarshal(modelDataJSON, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getModelData: %w", err)
	}

	return &searchResult, nil
}

// CreateNewResource creates a new resource in the specified model with the given data and connections
func (c *Client) CreateNewResource(ctx context.Context, request *CreateAndUpdateRequest) (*shared.DefaultDocumentStructure, error) {
	
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
		"model": request.Model,
		"payload":  request.Payload,
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

	var document shared.DefaultDocumentStructure
	if err := json.Unmarshal(singleDataJSON, &document); err != nil {
		return nil, fmt.Errorf("failed to unmarshal getSingleData: %w", err)
	}

	return &document, nil
}

// UpdateResource updates an existing resource by model and ID, with optional single page data, data updates, and connection changes
func (c *Client) UpdateResource(ctx context.Context, request *CreateAndUpdateRequest) (*shared.DefaultDocumentStructure, error) {
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
		"_id":   request.ID,
		"model": request.Model,
		"payload":  request.Payload,
		"single_page_data": request.SinglePageData,
		"force_update": request.ForceUpdate,
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

	var document shared.DefaultDocumentStructure
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

// SendAuditLog sends an audit log entry to the audit log service
func (c *Client) SendAuditLog(ctx context.Context, auditData AuditData) error {
	// Note: This is a placeholder implementation as the exact mutation wasn't found in the schema
	// You would need to implement the actual sendAuditLog mutation based on your GraphQL schema
	query := `
		mutation SendAuditLog($auditData: JSON!) {
			sendAuditLog(auditData: $auditData) {
				message
			}
		}
	`

	variables := map[string]interface{}{
		"auditData": auditData,
	}

	_, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return fmt.Errorf("failed to send audit log: %w", err)
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
var _ InjectedDBOperationInterface = (*Client)(nil)
