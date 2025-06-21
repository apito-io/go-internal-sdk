package goapitosdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
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
	req.Header.Set("X-APITO-KEY", c.apiKey)

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

// GetProjectDetails retrieves project details for the given project ID
func (c *Client) GetProjectDetails(ctx context.Context, projectID string) (*Project, error) {
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
				database_credentials {
					host
					port
					db
					user
					password
					access_key
					secret_key
					driver
				}
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

	var project Project
	if err := json.Unmarshal(projectJSON, &project); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project data: %w", err)
	}

	return &project, nil
}

// GetSingleResource retrieves a single resource by model and ID, with optional single page data
func (c *Client) GetSingleResource(ctx context.Context, model, _id string, singlePageData bool) (interface{}, error) {
	query := `
		query GetSingleData($model: String, $_id: String!, $single_page_data: Boolean) {
			getSingleData(model: $model, _id: $_id, single_page_data: $single_page_data) {
				_id
				_key
				created_at
				updated_at
				data
				meta {
					source_id
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

	return data["getSingleData"], nil
}

// SearchResources searches for resources in the specified model using the provided filter
func (c *Client) SearchResources(ctx context.Context, model string, filter map[string]interface{}, aggregate bool) (interface{}, error) {
	query := `
		query GetModelData($model: String!, $page: Int, $limit: Int, $where: JSON, $search: String) {
			getModelData(model: $model, page: $page, limit: $limit, where: $where, search: $search) {
				results {
					_id
					_key
					created_at
					updated_at
					data
					meta {
						source_id
						created_at
						updated_at
						status
						revision
						revision_at
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

	return data["getModelData"], nil
}

// GetRelationDocuments retrieves related documents for the given ID and connection parameters
func (c *Client) GetRelationDocuments(ctx context.Context, _id string, connection map[string]interface{}) (interface{}, error) {
	query := `
		query GetModelData($model: String!, $connection: ListAllDataDetailedOfAModelConnectionPayload) {
			getModelData(model: $model, connection: $connection) {
				results {
					_id
					_key
					created_at
					updated_at
					data
					meta {
						source_id
						created_at
						updated_at
						status
						revision
						revision_at
					}
				}
				count
			}
		}
	`

	variables := map[string]interface{}{
		"model":      connection["model"],
		"connection": connection,
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get relation documents: %w", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	return data["getModelData"], nil
}

// CreateNewResource creates a new resource in the specified model with the given data and connections
func (c *Client) CreateNewResource(ctx context.Context, model string, data map[string]interface{}, connection map[string]interface{}) (interface{}, error) {
	// Note: This is a placeholder implementation as the exact mutation wasn't found in the schema
	// You would need to implement the actual createNewData mutation based on your GraphQL schema
	query := `
		mutation CreateNewData($model: String!, $data: JSON!, $connection: JSON) {
			createNewData(model: $model, data: $data, connection: $connection) {
				_id
				_key
				created_at
				updated_at
				data
				meta {
					source_id
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
		"model": model,
		"data":  data,
	}

	if connection != nil {
		variables["connection"] = connection
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to create new resource: %w", err)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	return responseData["createNewData"], nil
}

// UpdateResource updates an existing resource by model and ID, with optional single page data, data updates, and connection changes
func (c *Client) UpdateResource(ctx context.Context, model, _id string, singlePageData bool, data map[string]interface{}, connect map[string]interface{}, disconnect map[string]interface{}) (interface{}, error) {
	// Note: This is a placeholder implementation as the exact mutation wasn't found in the schema
	// You would need to implement the actual updateData mutation based on your GraphQL schema
	query := `
		mutation UpdateData($model: String!, $_id: String!, $data: JSON!, $connect: JSON, $disconnect: JSON) {
			updateData(model: $model, _id: $_id, data: $data, connect: $connect, disconnect: $disconnect) {
				_id
				_key
				created_at
				updated_at
				data
				meta {
					source_id
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
		"model": model,
		"_id":   _id,
		"data":  data,
	}

	if connect != nil {
		variables["connect"] = connect
	}
	if disconnect != nil {
		variables["disconnect"] = disconnect
	}

	response, err := c.executeGraphQL(ctx, query, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}

	responseData, ok := response.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	return responseData["updateData"], nil
}

// DeleteResource deletes a resource by model and ID
func (c *Client) DeleteResource(ctx context.Context, model, _id string) error {
	// Note: This is a placeholder implementation as the exact mutation wasn't found in the schema
	// You would need to implement the actual deleteData mutation based on your GraphQL schema
	query := `
		mutation DeleteData($model: String!, $_id: String!) {
			deleteData(model: $model, _id: $_id) {
				message
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
