# Test health check endpoint
GET {{host}}/api/v1/utils/health-check/
HTTP 200
[Asserts]
jsonpath "$.status" == true

# Test Swagger UI endpoint
GET {{host}}/docs/swagger/
HTTP 200

# Test OpenAPI JSON endpoint
GET {{host}}/docs/openapi.json
HTTP 200