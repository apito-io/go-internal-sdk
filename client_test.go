package goapitosdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := Config{
		BaseURL: "https://api.example.com/graphql",
		APIKey:  "test-api-key",
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
		BaseURL: "https://api.example.com/graphql",
		APIKey:  "test-api-key",
		// No timeout specified - should use default
	}

	client := NewClient(config)

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestExecuteGraphQL_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("X-APITO-KEY") != "test-api-key" {
			t.Errorf("Expected X-APITO-KEY test-api-key, got %s", r.Header.Get("X-APITO-KEY"))
		}

		// Return a valid GraphQL response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"test": "success"}}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	})

	ctx := context.Background()
	query := "query { test }"
	variables := map[string]interface{}{"key": "value"}

	response, err := client.executeGraphQL(ctx, query, variables)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Data == nil {
		t.Fatal("Expected data in response, got nil")
	}
}

func TestExecuteGraphQL_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	})

	ctx := context.Background()
	query := "query { test }"

	_, err := client.executeGraphQL(ctx, query, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "HTTP error 500") {
		t.Errorf("Expected HTTP error 500 in error message, got %s", err.Error())
	}
}

func TestExecuteGraphQL_GraphQLError(t *testing.T) {
	// Create a test server that returns GraphQL errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errors": [{"message": "Field not found"}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	})

	ctx := context.Background()
	query := "query { nonexistentField }"

	_, err := client.executeGraphQL(ctx, query, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !contains(err.Error(), "GraphQL errors") {
		t.Errorf("Expected GraphQL errors in error message, got %s", err.Error())
	}
}

func TestInterface_Implementation(t *testing.T) {
	// This test ensures that Client implements InjectedDBOperationInterface
	var _ InjectedDBOperationInterface = (*Client)(nil)
}

// Helper function to check if a string contains a substring
func contains(str, substr string) bool {
	return len(str) >= len(substr) && (str == substr || len(substr) == 0 ||
		(len(substr) <= len(str) && str[:len(substr)] == substr) ||
		(len(str) > len(substr) && containsHelper(str, substr)))
}

func containsHelper(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
