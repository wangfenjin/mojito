package openapi

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
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
	Host:        "localhost:8080",
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

	// Create extra fields for security if middleware includes auth
	var extraFields map[string]interface{}
	for _, middleware := range route.Middlewares {
		if strings.Contains(middleware, "RequireAuth") {
			extraFields = map[string]interface{}{
				"security": []map[string][]string{
					{"OAuth2PasswordBearer": {}},
				},
			}
			break
		}
	}

	// Create operation
	return createOperation(route.Method, summary, description, tag, route.RequestType, route.ResponseType, extraFields)
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
	method string,
	summary string,
	description string,
	tag string,
	requestType reflect.Type,
	responseType reflect.Type,
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

	if requestType != nil {
		fieldTags := getTypeFieldTags(requestType)

		if slices.Contains([]string{"GET", "DELETE"}, method) {
			params := getQueryParameters(method, requestType)
			if len(params) > 0 {
				examples := generateExample(requestType)
				// Add examples to each parameter
				for i, param := range params {
					fieldName := param["name"].(string)
					if example, ok := examples[fieldName]; ok {
						params[i]["example"] = example
					}
				}
				if operation["parameters"] == nil {
					operation["parameters"] = params
				} else {
					existing := operation["parameters"].([]map[string]interface{})
					operation["parameters"] = append(existing, params...)
				}
			}
		} else {
			// Handle form data
			if _, hasForm := fieldTags[TypeForm]; hasForm {
				operation["requestBody"] = map[string]interface{}{
					"content": map[string]interface{}{
						"application/x-www-form-urlencoded": map[string]interface{}{
							"schema": getSchemaRef(requestType),
						},
					},
					"required": true,
				}
			}

			// Handle JSON data
			if _, hasJson := fieldTags[TypeJson]; hasJson {
				operation["requestBody"] = map[string]interface{}{
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": getSchemaRef(requestType),
						},
					},
					"required": true,
				}
			}
		}

		// Add extra fields if provided
		if len(extraFields) > 0 {
			for k, v := range extraFields[0] {
				operation[k] = v
			}
		}
	}
	return operation
}

const (
	TypeJson   = "json"
	TypeUri    = "uri"
	TypeForm   = "form"
	TypeHeader = "header"
	TypeQuery  = "query"
)

type FieldInfo struct {
	Name string
	Type string
	// VD   []string
}

func getTypeFieldTags(t reflect.Type) map[string][]FieldInfo {
	fieldsInfo := make(map[string][]FieldInfo)
	// Handle nil type
	if t == nil {
		return fieldsInfo
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only process struct types
	if t.Kind() != reflect.Struct {
		return fieldsInfo
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if jsonTag, ok := field.Tag.Lookup("json"); ok {
			jsonField := FieldInfo{
				Name: jsonTag,
				Type: TypeJson,
			}
			fieldsInfo[TypeJson] = append(fieldsInfo[TypeJson], jsonField)
		} else if uriTag, ok := field.Tag.Lookup("uri"); ok {
			uriField := FieldInfo{
				Name: uriTag,
				Type: TypeUri,
			}
			fieldsInfo[TypeUri] = append(fieldsInfo[TypeUri], uriField)
		} else if formTag, ok := field.Tag.Lookup("form"); ok {
			formField := FieldInfo{
				Name: formTag,
				Type: TypeForm,
			}
			fieldsInfo[TypeForm] = append(fieldsInfo[TypeForm], formField)
		} else if headerTag, ok := field.Tag.Lookup("header"); ok {
			headerField := FieldInfo{
				Name: headerTag,
				Type: TypeHeader,
			}
			fieldsInfo[TypeHeader] = append(fieldsInfo[TypeHeader], headerField)
		} else if queryTag, ok := field.Tag.Lookup("query"); ok {
			queryField := FieldInfo{
				Name: queryTag,
				Type: TypeQuery,
			}
			fieldsInfo[TypeQuery] = append(fieldsInfo[TypeQuery], queryField)
		}
	}
	return fieldsInfo
}

// getQueryParameters extracts query parameters from a struct type
func getQueryParameters(method string, t reflect.Type) []map[string]interface{} {
	var params []map[string]interface{}
	// Handle nil type
	if t == nil {
		return params
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only process struct types
	if t.Kind() != reflect.Struct {
		return params
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var fieldTag string
		var queryType string
		if _, ok := field.Tag.Lookup("query"); ok {
			fieldTag = field.Tag.Get("query")
			queryType = "query"
		} else if _, ok := field.Tag.Lookup("header"); ok {
			fieldTag = field.Tag.Get("header")
			queryType = "header"
		} else if _, ok := field.Tag.Lookup("uri"); ok {
			fieldTag = field.Tag.Get("uri")
			queryType = "path"
		} else if _, ok := field.Tag.Lookup("cookie"); ok {
			fieldTag = field.Tag.Get("cookie")
			queryType = "cookie"
		} else {
			if strings.ToUpper(method) == "GET" {
				if _, ok := field.Tag.Lookup("form"); ok {
					fieldTag = field.Tag.Get("form")
					queryType = "query"
				}
			}
		}
		if fieldTag == "" {
			continue
		}
		param := map[string]interface{}{
			"name":     fieldTag,
			"in":       queryType,
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

	// Get schema properties and required fields
	properties := getStructProperties(t)
	required := getRequiredFields(t)

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	// Add example if available
	if example := generateExample(t); len(example) > 0 {
		schema["example"] = example
	}

	return schema
}

// generateExample generates an example object for a type
// TODO: make it apply to the validation rules
func generateExample(t reflect.Type) map[string]interface{} {
	example := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Get field name from json or form tag
		var name string
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			name = strings.Split(jsonTag, ",")[0]
		} else if formTag := field.Tag.Get("form"); formTag != "" && formTag != "-" {
			name = strings.Split(formTag, ",")[0]
		} else {
			continue
		}

		// Generate example value based on field type and tags
		var value interface{}
		switch field.Type.Kind() {
		case reflect.String:
			if strings.Contains(field.Tag.Get("binding"), "email") {
				value = "user@example.com"
			} else if strings.Contains(name, "password") {
				value = "password123"
			} else if strings.Contains(name, "name") {
				value = "John Doe"
			} else {
				value = "example"
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			value = 123
		case reflect.Float32, reflect.Float64:
			value = 123.45
		case reflect.Bool:
			value = true
		case reflect.Slice:
			value = []interface{}{}
		}

		if value != nil {
			example[name] = value
		}
	}

	return example
}

// getRequiredFields returns a list of required field names
func getRequiredFields(t reflect.Type) []string {
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		name := strings.Split(jsonTag, ",")[0]

		// Check binding tag for required
		bindingTag := field.Tag.Get("binding")
		if strings.Contains(bindingTag, "required") {
			required = append(required, name)
		}
	}

	return required
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

	// Handle nil type
	if t == nil {
		return properties
	}

	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only process struct types
	if t.Kind() != reflect.Struct {
		return properties
	}

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
		"OAuth2PasswordBearer": map[string]interface{}{
			"type": "oauth2",
			"flows": map[string]interface{}{
				"password": map[string]interface{}{
					"scopes":   map[string]interface{}{},
					"tokenUrl": "/api/v1/login/access-token",
				},
			},
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
	// Only process struct types
	if t.Kind() != reflect.Struct {
		return
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
