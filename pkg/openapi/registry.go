package openapi

import (
	"reflect"
	"regexp"
	"strings"
)

// handlerInfo stores information about a route
type handlerInfo struct {
	Method       string
	Path         string
	HandlerName  string
	Middlewares  []string
	Tags         []string
	Summary      string
	Description  string
	RequestType  reflect.Type
	ResponseType reflect.Type
}

// Global registry to store route information
var handlerRegistry = make(map[string]handlerInfo)

func Registered(method, pattern string) bool {
	_, ok := handlerRegistry[method+":"+pattern]
	return ok
}

// RegisterRoute adds route information to the registry
func RegisterHandler(method, pattern, handlerName string, req, resp reflect.Type, middlewares ...string) {
	// Generate tag from path
	re := regexp.MustCompile(`/api/v\d+/([^/]+)`)
	matches := re.FindStringSubmatch(pattern)
	tag := "default"
	if len(matches) > 1 {
		tag = matches[1]
	}

	// Create route info
	info := handlerInfo{
		Method:       method,
		Path:         pattern,
		HandlerName:  handlerName,
		Middlewares:  middlewares,
		Tags:         []string{tag},
		Summary:      handlerPathToTitle(handlerName),
		RequestType:  req,
		ResponseType: resp,
	}

	// Add to registry
	key := method + ":" + pattern
	handlerRegistry[key] = info
}

// Convert full package path handler name to human readable title
func handlerPathToTitle(fullPath string) string {
	// Get the last part after the last dot
	parts := strings.Split(fullPath, ".")
	name := parts[len(parts)-1]

	// Remove "Handler" suffix
	name = strings.TrimSuffix(name, "Handler")

	// Handle special cases like "Html", "Content"
	name = strings.TrimSuffix(name, "HtmlContent")
	name = strings.TrimSuffix(name, "Content")

	// Split by uppercase letters
	var words []string
	var start int

	for i := 0; i < len(name); i++ {
		if i > 0 && name[i] >= 'A' && name[i] <= 'Z' {
			words = append(words, name[start:i])
			start = i
		}
	}
	words = append(words, name[start:])

	// Capitalize first letter of each word
	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(words, " ")
}
