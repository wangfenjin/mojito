// Package routes defines the API routes and handlers for the application
package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wangfenjin/mojito/common"
	"github.com/wangfenjin/mojito/openapi"
)

// RegisterDocsRoutes registers routes for API documentation
func RegisterDocsRoutes(r chi.Router) {
	r.Route("/docs", func(r chi.Router) {
		// Serve the OpenAPI spec
		r.Get("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./api/openapi.json")
		})

		// Serve Swagger UI using CDN
		r.Get("/swagger/*", func(w http.ResponseWriter, _ *http.Request) {
			// Generate OpenAPI spec
			err := openapi.GenerateSwaggerJSON("./api/openapi.json")
			if err != nil {
				common.GetLogger().Error("Failed to generate OpenAPI spec", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

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
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(swaggerHTML))
		})
	})
}
