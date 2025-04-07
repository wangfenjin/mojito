package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/google/uuid"
)

// RegisterDocsRoutes registers routes for API documentation
func RegisterDocsRoutes(h *server.Hertz) {
	// Generate OpenAPI spec
	err := GenerateSwaggerJSON("./api/openapi.json")
	if err != nil {
		panic("Failed to generate OpenAPI spec: " + err.Error())
	}

	// Serve Swagger UI
	docsGroup := h.Group("/docs")
	{
		// Serve the OpenAPI spec
		docsGroup.GET("/openapi.json", func(ctx context.Context, c *app.RequestContext) {
			c.File("./api/openapi.json")
		})

		// Serve Swagger UI using CDN
		docsGroup.GET("/*any", func(ctx context.Context, c *app.RequestContext) {
			// HTML for Swagger UI using CDN
			swaggerHTML := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Mojito API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.20.7/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.20.7/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.20.7/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/docs/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
				queryConfigEnabled: true, // enables the reading of URL params
                defaultModelsExpandDepth: -1
            });
            window.ui = ui;
        };
    </script>
</body>
</html>
`
			c.Data(200, "text/html; charset=utf-8", []byte(swaggerHTML))
		})
	}
}

// SwaggerInfo holds the API information used by the OpenAPI spec
type SwaggerInfo struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
}

// OpenAPISpec represents the OpenAPI specification structure
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       map[string]interface{} `json:"info"`
	Servers    []map[string]string    `json:"servers,omitempty"`
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components"`
}

// Default swagger documentation info
var SwaggerDoc = SwaggerInfo{
	Title:       "Mojito API",
	Description: "API documentation for Mojito backend",
	Version:     "1.0.0",
	Host:        "localhost:8888",
	BasePath:    "/api/v1",
}

// GenerateSwaggerJSON generates the OpenAPI specification JSON file
func GenerateSwaggerJSON(outputPath string) error {
	spec := OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: map[string]interface{}{
			"title":       SwaggerDoc.Title,
			"description": SwaggerDoc.Description,
			"version":     SwaggerDoc.Version,
		},
		Servers: []map[string]string{
			{
				"url": fmt.Sprintf("http://%s%s", SwaggerDoc.Host, SwaggerDoc.BasePath),
			},
		},
		Paths:      generatePaths(),
		Components: generateComponents(),
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI spec to file: %w", err)
	}

	return nil
}

// generatePaths creates the paths section of the OpenAPI spec
func generatePaths() map[string]interface{} {
	paths := make(map[string]interface{})

	// Login routes
	paths["/login/access-token"] = map[string]interface{}{
		"post": createOperation(
			"Login",
			"Get access token",
			"login",
			reflect.TypeOf(LoginAccessTokenRequest{}),
			reflect.TypeOf(TokenResponse{}),
			[]string{"form"},
		),
	}

	paths["/login/test-token"] = map[string]interface{}{
		"get": createOperation(
			"Test Token",
			"Test if the access token is valid",
			"login",
			nil,
			reflect.TypeOf(TestTokenResponse{}),
			nil,
			map[string]interface{}{
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
	}

	paths["/password-recovery/{email}"] = map[string]interface{}{
		"post": createOperation(
			"Recover Password",
			"Recover user password",
			"login",
			reflect.TypeOf(RecoverPasswordRequest{}),
			reflect.TypeOf(MessageResponse{}),
			nil,
			map[string]interface{}{
				"parameters": []map[string]interface{}{
					{
						"name":     "email",
						"in":       "path",
						"required": true,
						"schema": map[string]string{
							"type": "string",
						},
					},
				},
			},
		),
	}

	// User routes
	paths["/users"] = map[string]interface{}{
		"get": createOperation(
			"List Users",
			"Get list of users",
			"users",
			reflect.TypeOf(ListUsersRequest{}),
			reflect.TypeOf(UsersResponse{}),
			[]string{"query"},
			map[string]interface{}{
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
	}

	paths["/users/me"] = map[string]interface{}{
		"get": createOperation(
			"Get Current User",
			"Get current user information",
			"users",
			nil,
			reflect.TypeOf(UserResponse{}),
			nil,
			map[string]interface{}{
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
		"patch": createOperation(
			"Update Current User",
			"Update current user information",
			"users",
			reflect.TypeOf(UpdateUserMeRequest{}),
			reflect.TypeOf(UserResponse{}),
			nil,
			map[string]interface{}{
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
		"delete": createOperation(
			"Delete Current User",
			"Delete current user",
			"users",
			nil,
			reflect.TypeOf(MessageResponse{}),
			nil,
			map[string]interface{}{
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
	}

	paths["/users/signup"] = map[string]interface{}{
		"post": createOperation(
			"Register User",
			"Register a new user",
			"users",
			reflect.TypeOf(RegisterUserRequest{}),
			reflect.TypeOf(UserResponse{}),
			nil,
		),
	}

	// Item routes
	paths["/items"] = map[string]interface{}{
		"get": createOperation(
			"List Items",
			"Get list of items",
			"items",
			reflect.TypeOf(ListItemsRequest{}),
			reflect.TypeOf(ItemsResponse{}),
			[]string{"query"},
			map[string]interface{}{
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
		"post": createOperation(
			"Create Item",
			"Create a new item",
			"items",
			reflect.TypeOf(CreateItemRequest{}),
			reflect.TypeOf(ItemResponse{}),
			nil,
			map[string]interface{}{
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
	}

	paths["/items/{id}"] = map[string]interface{}{
		"get": createOperation(
			"Get Item",
			"Get item by ID",
			"items",
			reflect.TypeOf(GetItemRequest{}),
			reflect.TypeOf(ItemResponse{}),
			nil,
			map[string]interface{}{
				"parameters": []map[string]interface{}{
					{
						"name":     "id",
						"in":       "path",
						"required": true,
						"schema": map[string]string{
							"type":   "string",
							"format": "uuid",
						},
					},
				},
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
		"put": createOperation(
			"Update Item",
			"Update an item",
			"items",
			reflect.TypeOf(UpdateItemRequest{}),
			reflect.TypeOf(ItemResponse{}),
			nil,
			map[string]interface{}{
				"parameters": []map[string]interface{}{
					{
						"name":     "id",
						"in":       "path",
						"required": true,
						"schema": map[string]string{
							"type":   "string",
							"format": "uuid",
						},
					},
				},
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
		"delete": createOperation(
			"Delete Item",
			"Delete an item",
			"items",
			reflect.TypeOf(GetItemRequest{}),
			reflect.TypeOf(MessageResponse{}),
			nil,
			map[string]interface{}{
				"parameters": []map[string]interface{}{
					{
						"name":     "id",
						"in":       "path",
						"required": true,
						"schema": map[string]string{
							"type":   "string",
							"format": "uuid",
						},
					},
				},
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			},
		),
	}

	// Utils routes
	paths["/utils/health-check"] = map[string]interface{}{
		"get": createOperation(
			"Health Check",
			"Check if the API is running",
			"utils",
			nil,
			reflect.TypeOf(HealthCheckResponse{}),
			nil,
		),
	}

	return paths
}

// createOperation creates an operation object for the OpenAPI spec
func createOperation(
	summary string,
	description string,
	tag string,
	requestType reflect.Type,
	responseType reflect.Type,
	paramIn []string,
	extraFields ...map[string]interface{},
) map[string]interface{} {
	operation := map[string]interface{}{
		"summary":     summary,
		"description": description,
		"tags":        []string{tag},
		"responses": map[string]interface{}{
			"200": map[string]interface{}{
				"description": "Successful Response",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": getSchemaRef(responseType),
					},
				},
			},
			"422": map[string]interface{}{
				"description": "Validation Error",
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/HTTPValidationError",
						},
					},
				},
			},
		},
	}

	// Add request body if provided
	if requestType != nil && len(paramIn) == 0 {
		operation["requestBody"] = map[string]interface{}{
			"content": map[string]interface{}{
				"application/json": map[string]interface{}{
					"schema": getSchemaRef(requestType),
				},
			},
			"required": true,
		}
	}

	// Add form parameters if specified
	if requestType != nil && len(paramIn) > 0 && contains(paramIn, "form") {
		operation["requestBody"] = map[string]interface{}{
			"content": map[string]interface{}{
				"application/x-www-form-urlencoded": map[string]interface{}{
					"schema": getSchemaRef(requestType),
				},
			},
			"required": true,
		}
	}

	// Add query parameters if specified
	if requestType != nil && len(paramIn) > 0 && contains(paramIn, "query") {
		params := getQueryParameters(requestType)
		if len(params) > 0 {
			operation["parameters"] = params
		}
	}

	// Add extra fields if provided
	if len(extraFields) > 0 {
		for k, v := range extraFields[0] {
			operation[k] = v
		}
	}

	return operation
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getQueryParameters extracts query parameters from a struct type
func getQueryParameters(t reflect.Type) []map[string]interface{} {
	var params []map[string]interface{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		queryTag := field.Tag.Get("query")
		if queryTag == "" {
			continue
		}

		param := map[string]interface{}{
			"name":     queryTag,
			"in":       "query",
			"required": field.Tag.Get("binding") != "" && strings.Contains(field.Tag.Get("binding"), "required"),
			"schema":   getTypeSchema(field.Type),
		}

		// Add default value if specified
		defaultTag := field.Tag.Get("default")
		if defaultTag != "" {
			param["schema"].(map[string]interface{})["default"] = defaultTag
		}

		params = append(params, param)
	}

	return params
}

// getSchemaRef returns a schema reference for a type
func getSchemaRef(t reflect.Type) map[string]interface{} {
	if t == nil {
		return map[string]interface{}{
			"type": "object",
		}
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return map[string]interface{}{
		"$ref": fmt.Sprintf("#/components/schemas/%s", t.Name()),
	}
}

// getTypeSchema returns a schema for a specific type
func getTypeSchema(t reflect.Type) map[string]interface{} {
	schema := make(map[string]interface{})

	switch t.Kind() {
	case reflect.String:
		schema["type"] = "string"
	case reflect.Bool:
		schema["type"] = "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema["type"] = "integer"
	case reflect.Float32, reflect.Float64:
		schema["type"] = "number"
	case reflect.Slice, reflect.Array:
		schema["type"] = "array"
		schema["items"] = getTypeSchema(t.Elem())
	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			schema["type"] = "string"
			schema["format"] = "date-time"
		} else if t == reflect.TypeOf(uuid.UUID{}) {
			schema["type"] = "string"
			schema["format"] = "uuid"
		} else {
			schema["type"] = "object"
			schema["properties"] = getStructProperties(t)
		}
	case reflect.Map:
		schema["type"] = "object"
		if t.Key().Kind() == reflect.String {
			schema["additionalProperties"] = getTypeSchema(t.Elem())
		}
	default:
		schema["type"] = "object"
	}

	return schema
}

// getStructProperties extracts properties from a struct type
func getStructProperties(t reflect.Type) map[string]interface{} {
	properties := make(map[string]interface{})
	requiredFields := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse the json tag to get the field name and options
		parts := strings.Split(jsonTag, ",")
		name := parts[0]

		// Check if the field is required
		bindingTag := field.Tag.Get("binding")
		if bindingTag != "" && strings.Contains(bindingTag, "required") {
			requiredFields = append(requiredFields, name)
		}

		properties[name] = getTypeSchema(field.Type)
	}

	return properties
}

// generateComponents creates the components section of the OpenAPI spec
func generateComponents() map[string]interface{} {
	components := make(map[string]interface{})
	schemas := make(map[string]interface{})

	// Add all request and response schemas
	addSchemas(schemas, reflect.TypeOf(LoginAccessTokenRequest{}))
	addSchemas(schemas, reflect.TypeOf(TokenResponse{}))
	addSchemas(schemas, reflect.TypeOf(TestTokenResponse{}))
	addSchemas(schemas, reflect.TypeOf(RecoverPasswordRequest{}))
	addSchemas(schemas, reflect.TypeOf(ResetPasswordRequest{}))
	addSchemas(schemas, reflect.TypeOf(MessageResponse{}))
	addSchemas(schemas, reflect.TypeOf(HTMLContentResponse{}))

	addSchemas(schemas, reflect.TypeOf(CreateUserRequest{}))
	addSchemas(schemas, reflect.TypeOf(UpdateUserRequest{}))
	addSchemas(schemas, reflect.TypeOf(RegisterUserRequest{}))
	addSchemas(schemas, reflect.TypeOf(UpdateUserMeRequest{}))
	addSchemas(schemas, reflect.TypeOf(GetUserRequest{}))
	addSchemas(schemas, reflect.TypeOf(ListUsersRequest{}))
	addSchemas(schemas, reflect.TypeOf(UserResponse{}))
	addSchemas(schemas, reflect.TypeOf(UsersResponse{}))

	addSchemas(schemas, reflect.TypeOf(CreateItemRequest{}))
	addSchemas(schemas, reflect.TypeOf(UpdateItemRequest{}))
	addSchemas(schemas, reflect.TypeOf(GetItemRequest{}))
	addSchemas(schemas, reflect.TypeOf(ListItemsRequest{}))
	addSchemas(schemas, reflect.TypeOf(ItemResponse{}))
	addSchemas(schemas, reflect.TypeOf(ItemsResponse{}))

	addSchemas(schemas, reflect.TypeOf(HealthCheckResponse{}))

	// Add validation error schemas
	schemas["HTTPValidationError"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"error": map[string]interface{}{
				"type": "string",
			},
		},
	}

	components["schemas"] = schemas

	// Add security schemes
	components["securitySchemes"] = map[string]interface{}{
		"bearerAuth": map[string]interface{}{
			"type":         "http",
			"scheme":       "bearer",
			"bearerFormat": "JWT",
		},
	}

	return components
}

// addSchemas adds a schema to the schemas map
func addSchemas(schemas map[string]interface{}, t reflect.Type) {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Skip if already added
	if _, exists := schemas[t.Name()]; exists {
		return
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": getStructProperties(t),
	}

	// Add required fields
	var required []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		parts := strings.Split(jsonTag, ",")
		name := parts[0]

		bindingTag := field.Tag.Get("binding")
		if bindingTag != "" && strings.Contains(bindingTag, "required") {
			required = append(required, name)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	schemas[t.Name()] = schema
}
