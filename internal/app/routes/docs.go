package routes

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/wangfenjin/mojito/pkg/openapi"
)

// RegisterDocsRoutes registers routes for API documentation
func RegisterDocsRoutes(h *server.Hertz) {
	// Serve Swagger UI
	docsGroup := h.Group("/docs")
	{
		// Serve the OpenAPI spec
		docsGroup.GET("/openapi.json", func(ctx context.Context, c *app.RequestContext) {
			c.File("./api/openapi.json")
		})

		// Serve Swagger UI using CDN
		docsGroup.GET("/swagger/*any", func(ctx context.Context, c *app.RequestContext) {
			// Generate OpenAPI spec
			err := openapi.GenerateSwaggerJSON("./api/openapi.json")
			if err != nil {
				c.AbortWithError(consts.StatusInternalServerError, err)
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
			c.Data(200, "text/html; charset=utf-8", []byte(swaggerHTML))
		})
	}
}
