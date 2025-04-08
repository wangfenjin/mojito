package middleware

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/wangfenjin/mojito/internal/pkg/logger"
)

// RouteInfo stores information about a route
type RouteInfo struct {
	Method       string
	Path         string
	Handler      interface{}
	HandlerName  string
	Middlewares  []string
	Tags         []string
	Summary      string
	Description  string
	RequestType  reflect.Type
	ResponseType reflect.Type
}

// Global registry to store route information
var RouteRegistry = make(map[string][]RouteInfo)

// RegisterRoute adds route information to the registry
func RegisterRoute(method, path string, handler interface{}, middlewares ...string) {
	// Get handler name using reflection
	handlerName := getFunctionName(handler)

	// Extract request and response types if possible
	var requestType, responseType reflect.Type

	// Try to infer types from handler signature
	if handlerFunc := reflect.ValueOf(handler); handlerFunc.Kind() == reflect.Func {
		handlerType := handlerFunc.Type()

		// Check if it has a request parameter (index 1, after context)
		if handlerType.NumIn() > 1 {
			requestType = handlerType.In(1)
		}

		// Check if it has a response type (first return value)
		if handlerType.NumOut() > 0 {
			responseType = handlerType.Out(0)
		}
		logger.GetLogger().Info("handler params", "in", requestType, "out", responseType)
	}

	// Generate tag from path
	pathParts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")
	tag := pathParts[0]

	// Generate summary based on method and path
	summary := generateSummary(method, path)

	// Create route info
	routeInfo := RouteInfo{
		Method:       method,
		Path:         path,
		Handler:      handler,
		HandlerName:  handlerName,
		Middlewares:  middlewares,
		Tags:         []string{tag},
		Summary:      summary,
		Description:  method + " " + path,
		RequestType:  requestType,
		ResponseType: responseType,
	}

	// Add to registry
	key := method + ":" + path
	RouteRegistry[key] = append(RouteRegistry[key], routeInfo)
}

// Helper functions
func getFunctionName(i interface{}) string {
	if i == nil {
		return "nil"
	}

	// Get the function name
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()

	// Extract just the function name without package path
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

func generateSummary(method, path string) string {
	// Remove base path
	path = strings.TrimPrefix(path, "/api/v1")

	// Split path into segments
	segments := strings.Split(path, "/")

	// Generate summary based on method and path
	switch method {
	case "GET":
		if strings.Contains(path, ":") {
			return "Get " + singularize(segments[1]) + " by ID"
		} else if len(segments) > 2 && segments[2] == "me" {
			return "Get Current User"
		} else {
			return "List " + segments[1]
		}
	case "POST":
		if strings.Contains(path, "signup") {
			return "Register User"
		} else if strings.Contains(path, "login") {
			return "Login"
		} else if strings.Contains(path, "password-recovery") {
			return "Recover Password"
		} else {
			return "Create " + singularize(segments[1])
		}
	case "PUT", "PATCH":
		if strings.Contains(path, "me") {
			return "Update Current User"
		} else {
			return "Update " + singularize(segments[1])
		}
	case "DELETE":
		if strings.Contains(path, "me") {
			return "Delete Current User"
		} else {
			return "Delete " + singularize(segments[1])
		}
	default:
		return method + " " + path
	}
}

func singularize(word string) string {
	if strings.HasSuffix(word, "s") {
		return word[:len(word)-1]
	}
	return word
}
