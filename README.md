# Go Apito SDK

A Go SDK for communicating with Apito GraphQL API endpoints. This SDK implements the `InjectedDBOperationInterface` and provides a simple, idiomatic Go interface for interacting with Apito's backend services.

## Features

- ✅ Full implementation of `InjectedDBOperationInterface`
- ✅ GraphQL-based communication with Apito backend
- ✅ API key authentication via `X-APITO-KEY` header
- ✅ Context-aware operations with timeout support
- ✅ Comprehensive error handling
- ✅ Type-safe operations with Go structs
- ✅ Suitable for use in HashiCorp Go plugins and any Go environment

## Installation

```bash
go get github.com/apito-io/go-apito-sdk
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    goapitosdk "github.com/apito-io/go-apito-sdk"
)

func main() {
    // Create a new client
    client := goapitosdk.NewClient(goapitosdk.Config{
        BaseURL: "https://api.apito.io/graphql",
        APIKey:  "your-api-key-here",
        Timeout: 30 * time.Second,
    })

    ctx := context.Background()

    // Get project details
    project, err := client.GetProjectDetails(ctx, "project-id")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Project: %s (%s)\n", project.Name, project.ID)

    // Search for resources
    filter := map[string]interface{}{
        "limit": 10,
        "page":  1,
    }

    results, err := client.SearchResources(ctx, "users", filter, false)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Search results: %+v\n", results)
}
```

## Configuration

The SDK is configured using the `Config` struct:

```go
type Config struct {
    BaseURL    string        // Base URL of the Apito GraphQL endpoint
    APIKey     string        // API key for authentication (X-APITO-KEY header)
    Timeout    time.Duration // HTTP client timeout (default: 30 seconds)
    HTTPClient *http.Client  // Custom HTTP client (optional)
}
```

### Example with Custom HTTP Client

```go
customClient := &http.Client{
    Timeout: 60 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
    },
}

client := goapitosdk.NewClient(goapitosdk.Config{
    BaseURL:    "https://api.apito.io/graphql",
    APIKey:     "your-api-key-here",
    HTTPClient: customClient,
})
```

## API Reference

### Core Operations

#### GetProjectDetails

Retrieves project details for the given project ID.

```go
project, err := client.GetProjectDetails(ctx, "project-id")
```

#### GetSingleResource

Retrieves a single resource by model and ID.

```go
resource, err := client.GetSingleResource(ctx, "users", "user-id", false)
```

#### SearchResources

Searches for resources using filters.

```go
filter := map[string]interface{}{
    "page":   1,
    "limit":  10,
    "where":  map[string]interface{}{"status": "active"},
    "search": "john@example.com",
}

results, err := client.SearchResources(ctx, "users", filter, false)
```

#### CreateNewResource

Creates a new resource with data and optional connections.

```go
data := map[string]interface{}{
    "name":  "John Doe",
    "email": "john@example.com",
}

connection := map[string]interface{}{
    "organization_id": "org-123",
}

result, err := client.CreateNewResource(ctx, "users", data, connection)
```

#### UpdateResource

Updates an existing resource.

```go
data := map[string]interface{}{
    "name": "Jane Doe",
}

connect := map[string]interface{}{
    "role_id": "role-456",
}

result, err := client.UpdateResource(ctx, "users", "user-id", false, data, connect, nil)
```

#### DeleteResource

Deletes a resource by model and ID.

```go
err := client.DeleteResource(ctx, "users", "user-id")
```

### Utility Operations

#### GenerateTenantToken

Generates a new tenant token.

```go
token, err := client.GenerateTenantToken(ctx, "current-token", "tenant-id")
```

#### SendAuditLog

Sends audit log data.

```go
auditData := goapitosdk.AuditData{
    Resource: "users",
    Action:   "create",
    Author:   map[string]interface{}{"user_id": "user-123"},
    Data:     map[string]interface{}{"email": "john@example.com"},
    Meta:     map[string]interface{}{"timestamp": time.Now()},
}

err := client.SendAuditLog(ctx, auditData)
```

#### Debug

Debug utility for troubleshooting.

```go
result, err := client.Debug(ctx, "plugin-stage", "debug", "data", map[string]interface{}{"key": "value"})
```

## Usage in HashiCorp Go Plugins

The SDK is designed to work seamlessly with HashiCorp Go plugins:

```go
package main

import (
    "context"
    "fmt"

    goapitosdk "github.com/apito-io/go-apito-sdk"
    "github.com/hashicorp/go-plugin"
)

type MyPlugin struct {
    client goapitosdk.InjectedDBOperationInterface
}

func (p *MyPlugin) Initialize(apiKey, baseURL string) {
    p.client = goapitosdk.NewClient(goapitosdk.Config{
        BaseURL: baseURL,
        APIKey:  apiKey,
    })
}

func (p *MyPlugin) DoSomething(ctx context.Context) error {
    // Use the client
    project, err := p.client.GetProjectDetails(ctx, "")
    if err != nil {
        return err
    }

    fmt.Printf("Working with project: %s\n", project.Name)
    return nil
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: handshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "myPlugin": &MyPluginRPC{Impl: &MyPlugin{}},
        },
    })
}
```

## Error Handling

The SDK provides comprehensive error handling:

```go
result, err := client.GetSingleResource(ctx, "users", "invalid-id", false)
if err != nil {
    // Handle different types of errors
    switch {
    case strings.Contains(err.Error(), "HTTP error"):
        fmt.Println("Network or server error:", err)
    case strings.Contains(err.Error(), "GraphQL errors"):
        fmt.Println("GraphQL query error:", err)
    case strings.Contains(err.Error(), "unexpected response format"):
        fmt.Println("Response parsing error:", err)
    default:
        fmt.Println("Unknown error:", err)
    }
}
```

## Authentication

The SDK uses API key authentication via the `X-APITO-KEY` header. Make sure to:

1. Obtain a valid API key from your Apito dashboard
2. Keep your API key secure and never commit it to version control
3. Use environment variables or secure configuration management

```go
import "os"

client := goapitosdk.NewClient(goapitosdk.Config{
    BaseURL: os.Getenv("APITO_BASE_URL"),
    APIKey:  os.Getenv("APITO_API_KEY"),
})
```

## Best Practices

### Context Usage

Always pass a context with appropriate timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.GetSingleResource(ctx, "users", "user-id", false)
```

### Error Handling

Check for errors and handle them appropriately:

```go
if err != nil {
    // Log the error
    log.Printf("Failed to fetch resource: %v", err)

    // Return a user-friendly error
    return fmt.Errorf("unable to fetch user data: %w", err)
}
```

### Resource Management

The SDK manages HTTP connections efficiently, but you can customize the HTTP client if needed:

```go
client := goapitosdk.NewClient(goapitosdk.Config{
    BaseURL: "https://api.apito.io/graphql",
    APIKey:  "your-api-key",
    HTTPClient: &http.Client{
        Timeout: 60 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:       100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:    90 * time.Second,
        },
    },
})
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, please open an issue on GitHub or contact the Apito team.

## Changelog

### v1.0.0

- Initial release
- Full implementation of `InjectedDBOperationInterface`
- GraphQL-based API communication
- API key authentication support
- Context-aware operations
- Comprehensive error handling
