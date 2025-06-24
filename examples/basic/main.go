package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	goapitosdk "github.com/apito-io/go-apito-sdk"
)

// Todo represents a todo item structure
type Todo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	DueDate     time.Time `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User represents a user structure
type User struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
}

// Category represents a todo category
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func main() {
	// Initialize the client with environment variables
	client := goapitosdk.NewClient(goapitosdk.Config{
		BaseURL: getEnv("APITO_BASE_URL", "https://api.apito.io/graphql"),
		APIKey:  getEnv("APITO_API_KEY", ""),
		Timeout: 30 * time.Second,
	})

	// Set up context with tenant ID if available
	ctx := context.Background()
	if tenantID := getEnv("APITO_TENANT_ID", ""); tenantID != "" {
		ctx = context.WithValue(ctx, "tenant_id", tenantID)
	}

	fmt.Println("🚀 Apito SDK Comprehensive Todo Example")
	fmt.Println("========================================")

	// =============================================================================
	// 1. AUTHENTICATION & TENANT TOKEN GENERATION
	// =============================================================================
	fmt.Println("\n🔐 1. Authentication & Tenant Token Generation")
	if authToken := getEnv("APITO_AUTH_TOKEN", ""); authToken != "" && getEnv("APITO_TENANT_ID", "") != "" {
		tenantToken, err := client.GenerateTenantToken(ctx, authToken, getEnv("APITO_TENANT_ID", ""))
		if err != nil {
			log.Printf("❌ Error generating tenant token: %v", err)
		} else {
			fmt.Printf("✅ Generated tenant token: %s...\n", tenantToken[:20])
		}
	} else {
		fmt.Println("ℹ️  Skipping tenant token generation (missing auth token or tenant ID)")
	}

	// =============================================================================
	// 2. CREATE RESOURCES (Categories, Users, Todos)
	// =============================================================================
	fmt.Println("\n📝 2. Creating Resources")

	// Create a category
	categoryData := map[string]interface{}{
		"name":        "Work",
		"description": "Work-related tasks",
		"color":       "#3498db",
	}

	categoryRequest := &goapitosdk.CreateAndUpdateRequest{
		Model:          "categories",
		Payload:        categoryData,
		SinglePageData: false,
	}

	createdCategory, err := client.CreateNewResource(ctx, categoryRequest)
	if err != nil {
		log.Printf("❌ Error creating category: %v", err)
	} else {
		fmt.Printf("✅ Created category: %s\n", createdCategory.ID)
	}

	// Create a user
	userData := map[string]interface{}{
		"name":   "John Doe",
		"email":  "john.doe@example.com",
		"role":   "developer",
		"active": true,
	}

	userRequest := &goapitosdk.CreateAndUpdateRequest{
		Model:          "users",
		Payload:        userData,
		SinglePageData: false,
	}

	createdUser, err := client.CreateNewResource(ctx, userRequest)
	if err != nil {
		log.Printf("❌ Error creating user: %v", err)
	} else {
		fmt.Printf("✅ Created user: %s\n", createdUser.ID)
	}

	// Create multiple todos
	todos := []map[string]interface{}{
		{
			"title":       "Implement user authentication",
			"description": "Add JWT-based authentication to the application",
			"status":      "todo",
			"priority":    "high",
			"due_date":    time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339),
		},
		{
			"title":       "Write unit tests",
			"description": "Create comprehensive test suite for all modules",
			"status":      "in_progress",
			"priority":    "medium",
			"due_date":    time.Now().Add(14 * 24 * time.Hour).Format(time.RFC3339),
		},
		{
			"title":       "Update documentation",
			"description": "Refresh API documentation and user guides",
			"status":      "todo",
			"priority":    "low",
			"due_date":    time.Now().Add(21 * 24 * time.Hour).Format(time.RFC3339),
		},
	}

	var createdTodos []string
	for i, todoData := range todos {
		todoRequest := &goapitosdk.CreateAndUpdateRequest{
			Model:          "todos",
			Payload:        todoData,
			SinglePageData: false,
		}

		// Connect to user and category if they were created successfully
		if createdUser != nil && createdCategory != nil {
			todoRequest.Connect = map[string]interface{}{
				"user_id":     createdUser.ID,
				"category_id": createdCategory.ID,
			}
		}

		createdTodo, err := client.CreateNewResource(ctx, todoRequest)
		if err != nil {
			log.Printf("❌ Error creating todo %d: %v", i+1, err)
		} else {
			fmt.Printf("✅ Created todo: %s (%s)\n", createdTodo.ID, todoData["title"])
			createdTodos = append(createdTodos, createdTodo.ID)
		}
	}

	// =============================================================================
	// 3. SEARCH RESOURCES (Both typed and untyped examples)
	// =============================================================================
	fmt.Println("\n🔍 3. Searching Resources")

	// Search todos with filters (untyped)
	todoFilter := map[string]interface{}{
		"limit": 10,
		"page":  1,
		"where": map[string]interface{}{
			"status": "todo",
		},
		"search": "authentication",
	}

	todoResults, err := client.SearchResources(ctx, "todos", todoFilter, false)
	if err != nil {
		log.Printf("❌ Error searching todos: %v", err)
	} else {
		fmt.Printf("✅ Found %d todos (untyped search)\n", todoResults.Count)
		for i, todo := range todoResults.Results {
			if i < 3 { // Show first 3 results
				fmt.Printf("   - %s: %v\n", todo.ID, todo.Data)
			}
		}
	}

	// Search todos with typed results
	typedTodoResults, err := goapitosdk.SearchResourcesTyped[Todo](client, ctx, "todos", todoFilter, false)
	if err != nil {
		log.Printf("❌ Error searching todos (typed): %v", err)
	} else {
		fmt.Printf("✅ Found %d todos (typed search)\n", typedTodoResults.Count)
		for i, todoDoc := range typedTodoResults.Results {
			if i < 3 { // Show first 3 results
				fmt.Printf("   - %s: %s (Status: %s, Priority: %s)\n",
					todoDoc.ID, todoDoc.Data.Title, todoDoc.Data.Status, todoDoc.Data.Priority)
			}
		}
	}

	// Search users
	userFilter := map[string]interface{}{
		"limit": 5,
		"page":  1,
		"where": map[string]interface{}{
			"active": true,
		},
	}

	userResults, err := goapitosdk.SearchResourcesTyped[User](client, ctx, "users", userFilter, false)
	if err != nil {
		log.Printf("❌ Error searching users: %v", err)
	} else {
		fmt.Printf("✅ Found %d active users\n", userResults.Count)
		for _, userDoc := range userResults.Results {
			fmt.Printf("   - %s: %s (%s)\n", userDoc.ID, userDoc.Data.Name, userDoc.Data.Email)
		}
	}

	// =============================================================================
	// 4. GET SINGLE RESOURCES
	// =============================================================================
	fmt.Println("\n📄 4. Getting Single Resources")

	if len(createdTodos) > 0 {
		todoID := createdTodos[0]

		// Get single todo (untyped)
		singleTodo, err := client.GetSingleResource(ctx, "todos", todoID, false)
		if err != nil {
			log.Printf("❌ Error getting single todo: %v", err)
		} else {
			fmt.Printf("✅ Retrieved todo (untyped): %s\n", singleTodo.ID)
		}

		// Get single todo (typed)
		typedSingleTodo, err := goapitosdk.GetSingleResourceTyped[Todo](client, ctx, "todos", todoID, false)
		if err != nil {
			log.Printf("❌ Error getting single todo (typed): %v", err)
		} else {
			fmt.Printf("✅ Retrieved todo (typed): %s - %s\n", typedSingleTodo.ID, typedSingleTodo.Data.Title)
		}
	}

	// =============================================================================
	// 5. UPDATE RESOURCES
	// =============================================================================
	fmt.Println("\n✏️  5. Updating Resources")

	if len(createdTodos) > 0 {
		todoID := createdTodos[0]

		// Update todo status
		updateData := map[string]interface{}{
			"status":     "in_progress",
			"updated_at": time.Now().Format(time.RFC3339),
		}

		updateRequest := &goapitosdk.CreateAndUpdateRequest{
			ID:             todoID,
			Model:          "todos",
			Payload:        updateData,
			SinglePageData: false,
			ForceUpdate:    false,
		}

		updatedTodo, err := client.UpdateResource(ctx, updateRequest)
		if err != nil {
			log.Printf("❌ Error updating todo: %v", err)
		} else {
			fmt.Printf("✅ Updated todo status: %s\n", updatedTodo.ID)
		}

		// Update with typed result
		typedUpdatedTodo, err := goapitosdk.UpdateResourceTyped[Todo](client, ctx, updateRequest)
		if err != nil {
			log.Printf("❌ Error updating todo (typed): %v", err)
		} else {
			fmt.Printf("✅ Updated todo (typed): %s - Status: %s\n",
				typedUpdatedTodo.ID, typedUpdatedTodo.Data.Status)
		}
	}

	// =============================================================================
	// 6. GET RELATION DOCUMENTS
	// =============================================================================
	fmt.Println("\n🔗 6. Getting Related Documents")

	if createdUser != nil {
		// Get todos related to a specific user
		relationConnection := map[string]interface{}{
			"model": "todos",
			"filter": map[string]interface{}{
				"limit": 5,
				"where": map[string]interface{}{
					"user_id": createdUser.ID,
				},
			},
		}

		relatedTodos, err := client.GetRelationDocuments(ctx, createdUser.ID, relationConnection)
		if err != nil {
			log.Printf("❌ Error getting related todos: %v", err)
		} else {
			fmt.Printf("✅ Found %d todos related to user %s\n", relatedTodos.Count, createdUser.ID)
		}

		// Get related documents with typed results
		typedRelatedTodos, err := goapitosdk.GetRelationDocumentsTyped[Todo](client, ctx, createdUser.ID, relationConnection)
		if err != nil {
			log.Printf("❌ Error getting related todos (typed): %v", err)
		} else {
			fmt.Printf("✅ Found %d todos related to user (typed)\n", typedRelatedTodos.Count)
			for _, todoDoc := range typedRelatedTodos.Results {
				fmt.Printf("   - %s: %s\n", todoDoc.ID, todoDoc.Data.Title)
			}
		}
	}

	// =============================================================================
	// 7. AUDIT LOGGING
	// =============================================================================
	fmt.Println("\n📊 7. Audit Logging")

	auditData := goapitosdk.AuditData{
		Resource: "todos",
		Action:   "bulk_create",
		Author: map[string]interface{}{
			"user_id": "system",
			"name":    "SDK Example",
		},
		Data: map[string]interface{}{
			"todos_created": len(createdTodos),
			"timestamp":     time.Now().Format(time.RFC3339),
		},
		Meta: map[string]interface{}{
			"source":     "go-apito-sdk-example",
			"version":    "1.0.0",
			"ip_address": "127.0.0.1",
		},
	}

	err = client.SendAuditLog(ctx, auditData)
	if err != nil {
		log.Printf("❌ Error sending audit log: %v", err)
	} else {
		fmt.Printf("✅ Audit log sent successfully\n")
	}

	// =============================================================================
	// 8. DEBUG FUNCTIONALITY
	// =============================================================================
	fmt.Println("\n🐛 8. Debug Functionality")

	debugData := map[string]interface{}{
		"stage":         "todo_management_example",
		"todos_created": len(createdTodos),
		"operations":    []string{"create", "search", "update", "relations"},
		"timestamp":     time.Now().Format(time.RFC3339),
		"performance": map[string]interface{}{
			"total_operations": 8,
			"success_rate":     "95%",
		},
	}

	debugResult, err := client.Debug(ctx, "example_completion", debugData)
	if err != nil {
		log.Printf("❌ Error sending debug info: %v", err)
	} else {
		fmt.Printf("✅ Debug info sent: %+v\n", debugResult)
	}

	// =============================================================================
	// 9. CLEANUP (DELETE RESOURCES)
	// =============================================================================
	fmt.Println("\n🗑️  9. Cleanup (Deleting Resources)")

	// Delete created todos
	for i, todoID := range createdTodos {
		err := client.DeleteResource(ctx, "todos", todoID)
		if err != nil {
			log.Printf("❌ Error deleting todo %d: %v", i+1, err)
		} else {
			fmt.Printf("✅ Deleted todo: %s\n", todoID)
		}
	}

	// Delete created user
	if createdUser != nil {
		err := client.DeleteResource(ctx, "users", createdUser.ID)
		if err != nil {
			log.Printf("❌ Error deleting user: %v", err)
		} else {
			fmt.Printf("✅ Deleted user: %s\n", createdUser.ID)
		}
	}

	// Delete created category
	if createdCategory != nil {
		err := client.DeleteResource(ctx, "categories", createdCategory.ID)
		if err != nil {
			log.Printf("❌ Error deleting category: %v", err)
		} else {
			fmt.Printf("✅ Deleted category: %s\n", createdCategory.ID)
		}
	}

	fmt.Println("\n🎉 Todo Example Completed Successfully!")
	fmt.Println("=====================================")
	fmt.Println("\nThis example demonstrated:")
	fmt.Println("• Authentication & tenant token generation")
	fmt.Println("• Creating resources (todos, users, categories)")
	fmt.Println("• Searching with both typed and untyped methods")
	fmt.Println("• Getting single resources")
	fmt.Println("• Updating resources")
	fmt.Println("• Getting related documents")
	fmt.Println("• Audit logging")
	fmt.Println("• Debug functionality")
	fmt.Println("• Resource cleanup")
	fmt.Println("\nFor more examples, check the README.md file.")
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
