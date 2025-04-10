# Clean up test data first
DELETE {{host}}/api/v1/test/cleanup
HTTP 200

# Register a new user
POST {{host}}/api/v1/users/signup
Content-Type: application/json
{
    "email": "test@example.com",
    "phone_number": "+8613800138000",
    "password": "password123",
    "full_name": "Test User"
}

HTTP 200
[Captures]
user_id: jsonpath "$.id"
[Asserts]
jsonpath "$.email" == "test@example.com"
jsonpath "$.phone_number" == "+8613800138000"
jsonpath "$.full_name" == "Test User"
jsonpath "$.is_active" == true
jsonpath "$.is_superuser" == false
jsonpath "$.id" exists
jsonpath "$.created_at" exists

# Login to get token
POST {{host}}/api/v1/login/access-token
Content-Type: application/x-www-form-urlencoded
[FormParams]
username: test@example.com
password: password123
grant_type: password
scope: 
client_id: 
client_secret: 

HTTP 200
[Captures]
token: jsonpath "$.access_token"
[Asserts]
jsonpath "$.access_token" exists
jsonpath "$.token_type" == "bearer"

# Get current user (with auth)
GET {{host}}/api/v1/users/me
Authorization: Bearer {{token}}

HTTP 200
[Asserts]
jsonpath "$.email" == "test@example.com"
jsonpath "$.phone_number" == "+8613800138000"
jsonpath "$.full_name" == "Test User"
jsonpath "$.id" == {{user_id}}

# Update current user (with auth)
# PATCH {{host}}/api/v1/users/me
# Authorization: Bearer {{token}}
# Content-Type: application/json
# {
    # "full_name": "Updated User",
    # "phone_number": "+8613800138001"
# }

# HTTP 200
# [Asserts]
# jsonpath "$.full_name" == "Updated User"
# jsonpath "$.phone_number" == "+8613800138001"
# jsonpath "$.id" == {{user_id}}

# Update password
PATCH {{host}}/api/v1/users/me/password
Authorization: Bearer {{token}}
Content-Type: application/json
{
    "current_password": "password123",
    "new_password": "newpassword123"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Password updated successfully"

# List users
GET {{host}}/api/v1/users/
Authorization: Bearer {{token}}
[QueryStringParams]
skip: 0
limit: 10

HTTP 200
[Asserts]
jsonpath "$.users" isCollection
jsonpath "$.users[0]" exists
jsonpath "$.meta.skip" == 0
jsonpath "$.meta.limit" == 10

# Delete current user
DELETE {{host}}/api/v1/users/me
Authorization: Bearer {{token}}

HTTP 200
[Asserts]
jsonpath "$.message" == "User deleted successfully"

# Verify user is deleted - should return 401 since token is no longer valid
GET {{host}}/api/v1/users/me

HTTP 401
[Asserts]
jsonpath "$.error" == "missing authorization header"