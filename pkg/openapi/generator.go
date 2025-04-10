package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
		Paths:      generatePathsFromRegistry(),
		Components: generateComponents(),
	}

	// security?
	// path params?
	// components schema?
	// get route group name?

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

// generatePathsFromRegistry creates the paths section of the OpenAPI spec from the route registry
func generatePathsFromRegistry() map[string]interface{} {
	// Group routes by path
	routesByPath := make(map[string][]handlerInfo)
	for _, route := range handlerRegistry {
		// Skip docs routes to avoid circular references
		if strings.HasPrefix(route.Path, "/docs") {
			continue
		}
		fmt.Printf("route: %+v\n", route)

		// Normalize path for OpenAPI
		apiPath := strings.TrimPrefix(route.Path, SwaggerDoc.BasePath)
		// Convert Hertz path params (:id) to OpenAPI path params ({id})
		apiPath = convertPathParams(apiPath)

		routesByPath[apiPath] = append(routesByPath[apiPath], route)
	}

	paths := make(map[string]interface{})
	// Process each path
	for path, routes := range routesByPath {
		pathItem := make(map[string]interface{})

		for _, route := range routes {
			// Convert HTTP method to lowercase
			method := strings.ToLower(route.Method)

			// Create operation for this method
			operation := createOperationFromRouteInfo(route)

			// Add to path item
			pathItem[method] = operation
		}
		fmt.Printf("path %v, pathItem: %+v\n", path, pathItem)
		paths[path] = pathItem
	}

	return paths
}

// createOperationFromRouteInfo creates an operation object for a route
func createOperationFromRouteInfo(route handlerInfo) map[string]interface{} {
	// Extract tag from route info
	tag := route.Tags[0]

	// Default values
	summary := route.Summary
	description := route.Description

	var paramIn []string

	// Determine parameter location (query, path, form)
	if strings.Contains(route.Path, ":") {
		paramIn = append(paramIn, "path")
	}
	if route.Method == "GET" && route.RequestType != nil {
		paramIn = append(paramIn, "query")
	}
	// TODO: fix this
	if route.Method == "POST" && strings.Contains(route.Path, "login/access-token") {
		paramIn = append(paramIn, "form")
	}

	// Create extra fields for security if middleware includes auth
	var extraFields map[string]interface{}
	for _, middleware := range route.Middlewares {
		if middleware == "RequireAuth" {
			extraFields = map[string]interface{}{
				"security": []map[string][]string{
					{"bearerAuth": {}},
				},
			}
			break
		}
	}

	// Create operation
	return createOperation(summary, description, tag, route.RequestType, route.ResponseType, paramIn, extraFields)
}

// TODO: make it general
// convertPathParams converts Hertz path params to OpenAPI path params
func convertPathParams(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			// Convert :id to {id}
			paramName := strings.TrimPrefix(segment, ":")
			segments[i] = "{" + paramName + "}"
		}
	}
	return strings.Join(segments, "/")
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
			// TODO: enum http error
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
	// TODO: request and json tag
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
	// TODO: request and form tag
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
	// TODO: request and query tag
	if requestType != nil && len(paramIn) > 0 && contains(paramIn, "query") {
		params := getQueryParameters(requestType)
		if len(params) > 0 {
			operation["parameters"] = params
		}
	}

	// TODO: request and path tag

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

	// Add schemas from route registry
	for _, route := range handlerRegistry {
		// Add request type schema if available
		if route.RequestType != nil {
			addSchemas(schemas, route.RequestType)
		}

		// Add response type schema if available
		if route.ResponseType != nil {
			addSchemas(schemas, route.ResponseType)
		}
	}

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
