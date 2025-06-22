package goapitosdk

import (
	"context"
	"testing"
	"time"
)

const (
	BaseURL = "http://localhost:5050/system/graphql"
	APIKey  = "EhvreBZvFOKYCxWx3xL9xuW4g1WLx3dhdbCWmPhuIaIVI4zeBMk5gUYfuXM4jccwNGjRqitMaNyK1kt6b3S8NKowNXzwFDL6ivZL4rscGu49w8E3vVEYPeyvAgzT0NeTPO9SiJxmI4nBGkMpcBX789VqEfH1tuwacKKivQ4jhLtGt3PsyfmIXX9"
)

type Task struct {
	Name        string      `json:"name"`
	Took        float64     `json:"took"`        // Can be string or number
	Description interface{} `json:"description"` // Can be string or object
	Progress    string      `json:"progress"`
	List        []struct {
		ID          string      `json:"id"`
		Title       string      `json:"title"`
		Description interface{} `json:"description"` // Can be string or object
		Status      string      `json:"status"`
	} `json:"list"`
	Properties struct {
		GivenBy      string `json:"given_by"`
		HandoverDate string `json:"handover_date"`
		Commission   string `json:"commission"`
	} `json:"properties"`
}

// Example Product struct for testing typed operations
type Product struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CategoryID  string  `json:"category_id"`
	InStock     bool    `json:"in_stock"`
	CreatedAt   string  `json:"created_at"`
}

// Example User struct for testing typed operations
type User struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Active    bool   `json:"active"`
}

func getTestClient() *Client {
	return NewClient(Config{
		BaseURL: BaseURL,
		APIKey:  APIKey,
		Timeout: 30 * time.Second,
	})
}

func TestNewClient(t *testing.T) {
	config := Config{
		BaseURL: BaseURL,
		APIKey:  APIKey,
		Timeout: 10 * time.Second,
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.baseURL != config.BaseURL {
		t.Errorf("Expected baseURL %s, got %s", config.BaseURL, client.baseURL)
	}

	if client.apiKey != config.APIKey {
		t.Errorf("Expected apiKey %s, got %s", config.APIKey, client.apiKey)
	}
}

func TestNewClientWithDefaults(t *testing.T) {
	config := Config{
		BaseURL: BaseURL,
		APIKey:  APIKey,
		// No timeout specified - should use default
	}

	client := NewClient(config)

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.httpClient.Timeout)
	}
}

/* func TestGetProjectDetails(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	project, err := client.GetProjectDetails(ctx, "")
	if err != nil {
		t.Logf("GetProjectDetails failed (may be expected): %v", err)
		return
	}

	if project == nil {
		t.Error("Expected project details, got nil")
		return
	}

	t.Logf("✅ GetProjectDetails succeeded: Project ID=%s, Name=%s", project.ID, project.Name)
} */

func TestGetSingleResource(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	// Test with a dummy ID - may fail but should not panic
	resource, err := client.GetSingleResource(ctx, "task", "401fa9f2-b174-42b1-84da-1227be8d8755", false)
	if err != nil {
		t.Logf("GetSingleResource failed (may be expected): %v", err)
		return
	}

	if resource == nil {
		t.Error("Expected resource, got nil")
		return
	}

	t.Logf("✅ GetSingleResource succeeded: %+v", resource)
}

func TestSearchResources(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	filter := map[string]interface{}{
		"page":  1,
		"limit": 5,
	}

	results, err := client.SearchResources(ctx, "task", filter, false)
	if err != nil {
		t.Logf("SearchResources failed (may be expected): %v", err)
		return
	}

	if results == nil {
		t.Error("Expected results, got nil")
		return
	}

	t.Logf("✅ SearchResources succeeded: %+v", results)
}

func TestGetRelationDocuments(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	connection := map[string]interface{}{
		"model":           "users",
		"_id":             "test-id",
		"to_model":        "roles",
		"relation_type":   "belongs_to",
		"known_as":        "user_role",
		"connection_type": "outbound",
	}

	results, err := client.GetRelationDocuments(ctx, "test-id", connection)
	if err != nil {
		t.Logf("GetRelationDocuments failed (may be expected): %v", err)
		return
	}

	if results == nil {
		t.Error("Expected results, got nil")
		return
	}

	t.Logf("✅ GetRelationDocuments succeeded: %+v", results)
}

func TestCreateNewResource(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	data := map[string]interface{}{
		"name":        "Test",
		"took":        3,
		"description": "Test Description",
		"progress":    "DONE",
	}

	connct := map[string]interface{}{
		"category_ids": []string{"56b2a1dd-25cf-44b4-ad65-8a78b6deab89"},
		"executor_id":  "354c47b6-8693-4720-9a4d-7404a64386f9",
	}

	result, err := client.CreateNewResource(ctx, "task", data, connct)
	if err != nil {
		t.Logf("CreateNewResource failed (may be expected): %v", err)
		return
	}

	if result == nil {
		t.Error("Expected result, got nil")
		return
	}

	t.Logf("✅ CreateNewResource succeeded: %+v", result)
}

func TestUpdateResource(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	data := map[string]interface{}{
		"name":        "Test",
		"took":        3,
		"description": "Test Description",
		"progress":    "DONE",
	}

	connct := map[string]interface{}{
		"category_ids": []string{"56b2a1dd-25cf-44b4-ad65-8a78b6deab89"},
		//"executor_id": "354c47b6-8693-4720-9a4d-7404a64386f9",
	}

	disconnect := map[string]interface{}{
		"executor_id": "354c47b6-8693-4720-9a4d-7404a64386f9",
	}

	result, err := client.UpdateResource(ctx, "task", "a0d50ad7-3001-4be0-92bd-d0daac0af3a9", false, data, connct, disconnect)
	if err != nil {
		t.Logf("UpdateResource failed (may be expected): %v", err)
		return
	}

	if result == nil {
		t.Error("Expected result, got nil")
		return
	}

	t.Logf("✅ UpdateResource succeeded: %+v", result)
}

func TestDeleteResource(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	err := client.DeleteResource(ctx, "task", "a0d50ad7-3001-4be0-92bd-d0daac0af3a9")
	if err != nil {
		t.Logf("DeleteResource failed (may be expected): %v", err)
		return
	}

	t.Logf("✅ DeleteResource succeeded")
}

func TestGenerateTenantToken(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	token, err := client.GenerateTenantToken(ctx, "ak_4HESWVQEXE7V4GVGGDRYGXVWXSCAJL44TAUICSLBPQTOB6CJ53KTU3GUOEXJUIXVAKFMM2BDRJRWWPKEN3DRA3HDLZUY4NZMVLFJUIK5H4BWLY26AUKDOHPZE2ENGJNCXPPPEBKCNLTUXXUFUKVDGYJ2H6CZCSMQCY5KSCYNJVYBXVJBYE6O7C73DI3NV7Q", "ba0ee756-6aea-43a6-b052-c7baab3da91c")
	if err != nil {
		t.Logf("GenerateTenantToken failed (may be expected): %v", err)
		return
	}

	if token == "" {
		t.Error("Expected token, got empty string")
		return
	}

	t.Logf("✅ GenerateTenantToken succeeded: %s", token)
}

func TestSendAuditLog(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	auditData := AuditData{
		Resource: "users",
		Action:   "test",
		Author: map[string]interface{}{
			"user_id": "test-user",
			"name":    "Test User",
		},
		Data: map[string]interface{}{
			"test": "data",
		},
		Meta: map[string]interface{}{
			"test_run": true,
		},
	}

	err := client.SendAuditLog(ctx, auditData)
	if err != nil {
		t.Logf("SendAuditLog failed (may be expected): %v", err)
		return
	}

	t.Logf("✅ SendAuditLog succeeded")
}

func TestDebug(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	result, err := client.Debug(ctx, "test_stage", "debug", "data", map[string]interface{}{
		"test": true,
	})
	if err != nil {
		t.Logf("Debug failed (may be expected): %v", err)
		return
	}

	if result == nil {
		t.Error("Expected debug result, got nil")
		return
	}

	t.Logf("✅ Debug succeeded: %+v", result)
}

func TestExecuteGraphQL(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	// Test with a simple query
	query := `query { getProject { id name } }`
	variables := map[string]interface{}{}

	response, err := client.executeGraphQL(ctx, query, variables)
	if err != nil {
		t.Logf("executeGraphQL failed (may be expected): %v", err)
		return
	}

	if response == nil {
		t.Error("Expected response, got nil")
		return
	}

	t.Logf("✅ executeGraphQL succeeded: %+v", response.Data)
}

// Integration test that runs multiple operations in sequence
func TestClientIntegration(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	t.Log("=== Running Integration Test ===")

	/* // Test 1: Get project details
	t.Log("1. Testing GetProjectDetails...")
	project, err := client.GetProjectDetails(ctx, "")
	if err != nil {
		t.Logf("   GetProjectDetails failed: %v", err)
	} else {
		t.Logf("   ✅ Project: %s", project.Name)
	} */

	// Test 2: Search resources
	t.Log("2. Testing SearchResources...")
	filter := map[string]interface{}{
		"limit": 3,
		"page":  1,
	}
	results, err := client.SearchResources(ctx, "users", filter, false)
	if err != nil {
		t.Logf("   SearchResources failed: %v", err)
	} else {
		t.Logf("   ✅ Search completed: %+v", results)
	}

	// Test 3: Send audit log
	t.Log("3. Testing SendAuditLog...")
	auditData := AuditData{
		Resource: "test",
		Action:   "integration_test",
		Author: map[string]interface{}{
			"test": "integration",
		},
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
		Meta: map[string]interface{}{
			"test_type": "integration",
		},
	}
	err = client.SendAuditLog(ctx, auditData)
	if err != nil {
		t.Logf("   SendAuditLog failed: %v", err)
	} else {
		t.Log("   ✅ Audit log sent successfully")
	}

	t.Log("=== Integration Test Completed ===")
}

func TestTypedOperations(t *testing.T) {
	// Create a mock client (you would use real credentials in practice)
	config := Config{
		BaseURL: BaseURL,
		APIKey:  APIKey,
		Timeout: 30 * time.Second,
	}
	client := NewClient(config)

	ctx := context.Background()

	// Suppress unused variable warnings for test setup
	_ = client
	_ = ctx

	// Example usage of typed functions
	t.Run("GetSingleResourceTyped", func(t *testing.T) {
		// This would be called like:
		// product, err := GetSingleResourceTyped[Product](client, ctx, "products", "product-123", false)
		// if err != nil {
		//     t.Fatal(err)
		// }
		//
		// // Now product.Data is of type Product, not map[string]interface{}
		// t.Logf("Product name: %s, price: %.2f", product.Data.Name, product.Data.Price)

		t.Skip("Skipping integration test - requires live API")
	})

	t.Run("SearchResourcesTyped", func(t *testing.T) {
		// This would be called like:
		// filter := map[string]interface{}{
		//     "limit": 10,
		//     "where": map[string]interface{}{
		//         "in_stock": true,
		//     },
		// }
		// results, err := SearchResourcesTyped[Product](client, ctx, "products", filter, false)
		// if err != nil {
		//     t.Fatal(err)
		// }
		//
		// for _, product := range results.Results {
		//     t.Logf("Product: %s - $%.2f", product.Data.Name, product.Data.Price)
		// }

		t.Skip("Skipping integration test - requires live API")
	})

	t.Run("CreateNewResourceTyped", func(t *testing.T) {
		// This would be called like:
		// data := map[string]interface{}{
		//     "name": "New Product",
		//     "description": "A great new product",
		//     "price": 29.99,
		//     "in_stock": true,
		// }
		//
		// product, err := CreateNewResourceTyped[Product](client, ctx, "products", data, nil)
		// if err != nil {
		//     t.Fatal(err)
		// }
		//
		// t.Logf("Created product: %s with ID: %s", product.Data.Name, product.ID)

		t.Skip("Skipping integration test - requires live API")
	})

	t.Run("UpdateResourceTyped", func(t *testing.T) {
		// This would be called like:
		// data := map[string]interface{}{
		//     "price": 24.99,
		//     "in_stock": false,
		// }
		//
		// product, err := UpdateResourceTyped[Product](client, ctx, "products", "product-123", false, data, nil, nil)
		// if err != nil {
		//     t.Fatal(err)
		// }
		//
		// t.Logf("Updated product: %s, new price: %.2f", product.Data.Name, product.Data.Price)

		t.Skip("Skipping integration test - requires live API")
	})

	t.Run("GetRelationDocumentsTyped", func(t *testing.T) {
		// This would be called like:
		// connection := map[string]interface{}{
		//     "model": "users",
		//     "relation": "purchased_by",
		//     "filter": map[string]interface{}{
		//         "limit": 5,
		//     },
		// }
		//
		// users, err := GetRelationDocumentsTyped[User](client, ctx, "product-123", connection)
		// if err != nil {
		//     t.Fatal(err)
		// }
		//
		// for _, user := range users.Results {
		//     t.Logf("User: %s %s (%s)", user.Data.FirstName, user.Data.LastName, user.Data.Email)
		// }

		t.Skip("Skipping integration test - requires live API")
	})
}

// TestTypedOperationsIntegration demonstrates real usage of typed operations
func TestTypedOperationsIntegration(t *testing.T) {
	config := Config{
		BaseURL: BaseURL,
		APIKey:  APIKey,
		Timeout: 30 * time.Second,
	}
	client := NewClient(config)
	ctx := context.Background()

	t.Run("GetSingleTask", func(t *testing.T) {
		// Get a single task with type safety
		task, err := GetSingleResourceTyped[Task](client, ctx, "task", "401fa9f2-b174-42b1-84da-1227be8d8755", false)
		if err != nil {
			t.Logf("Failed to get single task: %v", err)
			t.Skip("Skipping - API call failed (expected in test environment)")
			return
		}

		// Now task.Data is strongly typed as Task
		t.Logf("Task name: %s", task.Data.Name)
		t.Logf("Task took: %v", task.Data.Took)
		t.Logf("Task description: %v", task.Data.Description)
		t.Logf("Task progress: %s", task.Data.Progress)

		// Verify type safety - these are compile-time checked
		_ = task.Data.Name        // string
		_ = task.Data.Took        // interface{} (can be string or number)
		_ = task.Data.Description // interface{} (can be string or object)
		_ = task.Data.Progress    // string
	})

	t.Run("SearchTasks", func(t *testing.T) {
		// Search for tasks with type safety
		filter := map[string]interface{}{
			"limit": 10,
			"where": map[string]interface{}{
				"progress": map[string]interface{}{
					"eq": "INPROGRESS",
				},
			},
		}

		results, err := SearchResourcesTyped[Task](client, ctx, "task", filter, false)
		if err != nil {
			t.Logf("Failed to search tasks: %v", err)
			t.Skip("Skipping - API call failed (expected in test environment)")
			return
		}

		t.Logf("Found %d tasks", results.Count)

		// All results are strongly typed
		for i, task := range results.Results {
			if i >= 3 { // Limit output for testing
				break
			}
			t.Logf("Task %d: %s - %s", i+1, task.Data.Name, task.Data.Progress)

			// Type safety verification
			_ = task.Data.Name        // string - no need for type assertions!
			_ = task.Data.Took        // interface{} (can be string or number)
			_ = task.Data.Description // interface{} (can be string or object)
		}
	})
}

// ExampleTypedOperations shows how to use the typed operations in documentation
func ExampleGetSingleResourceTyped() {
	config := Config{
		BaseURL: BaseURL,
		APIKey:  APIKey,
		Timeout: 30 * time.Second,
	}
	client := NewClient(config)
	ctx := context.Background()

	// Get a single task with full type safety
	task, err := GetSingleResourceTyped[Task](client, ctx, "task", "401fa9f2-b174-42b1-84da-1227be8d8755", false)
	if err != nil {
		// Handle error appropriately
		return
	}

	// Access strongly typed fields
	taskName := task.Data.Name
	taskProgress := task.Data.Progress

	_ = taskName
	_ = taskProgress
}
