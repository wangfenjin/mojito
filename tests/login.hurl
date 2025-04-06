# Clean up test data first
DELETE http://localhost:8888/api/v1/test/cleanup
HTTP 200

# Create a test user
POST http://localhost:8888/api/v1/users/
Content-Type: application/json
{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User",
    "is_active": true,
    "is_superuser": false
}

HTTP 200

# Test successful login
POST http://localhost:8888/api/v1/login/access-token
Content-Type: application/x-www-form-urlencoded
[FormParams]
username: test@example.com
password: password123

HTTP 200
[Captures]
access_token: jsonpath "$.access_token"
[Asserts]
jsonpath "$.access_token" exists
jsonpath "$.token_type" == "bearer"

# Test token validation
GET http://localhost:8888/api/v1/login/test-token
Authorization: Bearer {{access_token}}

HTTP 200
[Asserts]
jsonpath "$.email" == "test@example.com"

# Test invalid credentials
POST http://localhost:8888/api/v1/login/access-token
Content-Type: application/x-www-form-urlencoded
[FormParams]
username: test@example.com
password: wrongpassword

HTTP 400
[Asserts]
jsonpath "$.error" == "invalid credentials"

# Test password recovery request
POST http://localhost:8888/api/v1/password-recovery/test@example.com
HTTP 200
[Asserts]
jsonpath "$.message" == "password recovery email sent"

# Test password recovery for non-existent user
POST http://localhost:8888/api/v1/password-recovery/nonexistent@example.com
HTTP 400
[Asserts]
jsonpath "$.error" == "user not found"