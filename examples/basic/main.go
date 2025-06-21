package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	goapitosdk "github.com/apito-io/go-apito-sdk"
)

func main() {
	// Initialize the client with environment variables
	client := goapitosdk.NewClient(goapitosdk.Config{
		BaseURL: getEnv("APITO_BASE_URL", "https://api.apito.io/graphql"),
		APIKey:  getEnv("APITO_API_KEY", ""),
		Timeout: 30 * time.Second,
	})

	ctx := context.Background()

	fmt.Println("=== Apito SDK Basic Example ===")

	// Example 1: Get project details
	fmt.Println("\n1. Getting project details...")
	project, err := client.GetProjectDetails(ctx, "")
	if err != nil {
		log.Printf("Error getting project details: %v", err)
	} else {
		fmt.Printf("✅ Project: %s (ID: %s)\n", project.Name, project.ID)
		fmt.Printf("   Status: %s\n", project.Status)
		fmt.Printf("   Organization: %s\n", project.OrganizationID)
	}

	// Example 2: Search for resources
	fmt.Println("\n2. Searching for resources...")
	filter := map[string]interface{}{
		"limit": 5,
		"page":  1,
	}

	results, err := client.SearchResources(ctx, "users", filter, false)
	if err != nil {
		log.Printf("Error searching resources: %v", err)
	} else {
		fmt.Printf("✅ Search completed successfully\n")
		fmt.Printf("   Results: %+v\n", results)
	}

	fmt.Println("\n=== Example completed ===")
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
