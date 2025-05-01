# Clean up test data first
DELETE {{host}}/api/v1/test/cleanup
HTTP 200

# Register a new user
POST {{host}}/api/v1/users/signup
Content-Type: application/json
{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User"
}

HTTP 200
[Captures]
user_id: jsonpath "$.id"
[Asserts]
jsonpath "$.email" == "test@example.com"
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
jsonpath "$.full_name" == "Test User"
jsonpath "$.id" == {{user_id}}

# Get user not allowed for normal user
GET {{host}}/api/v1/users/{{user_id}}
Authorization: Bearer {{token}}

HTTP 403

# Update current user (with auth)
PATCH {{host}}/api/v1/users/me
Authorization: Bearer {{token}}
Content-Type: application/json
{
    "full_name": "Updated User"
}

HTTP 200
[Asserts]
jsonpath "$.full_name" == "Updated User"
jsonpath "$.id" == {{user_id}}

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

# Get current user (with auth)
GET {{host}}/api/v1/users/me
Authorization: Bearer {{token}}

HTTP 200
[Asserts]
jsonpath "$.email" == "test@example.com"
jsonpath "$.full_name" == "Updated User"
jsonpath "$.id" == {{user_id}}

# List users
GET {{host}}/api/v1/users/
Authorization: Bearer {{token}}
[QueryStringParams]
skip: 0
limit: 10

HTTP 403

# Create a super user for List API
POST {{host}}/api/v1/test/superuser
Content-Type: application/json
{
    "email": "admin@example.com",
    "password": "admin123456",
    "full_name": "Admin User"
}

HTTP 200
[Asserts]
jsonpath "$.message" == "Super user created successfully"

# Login as super user
POST {{host}}/api/v1/login/access-token
Content-Type: application/x-www-form-urlencoded
[FormParams]
username: admin@example.com
password: admin123456
grant_type: password
scope: 
client_id: 
client_secret: 

HTTP 200
[Captures]
admin_token: jsonpath "$.access_token"
[Asserts]
jsonpath "$.access_token" exists
jsonpath "$.token_type" == "bearer"

# Update user is ok for super user
PATCH {{host}}/api/v1/users/{{user_id}}
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
    "full_name": "Updated User By Admin"
}

HTTP 200
[Asserts]
jsonpath "$.full_name" == "Updated User By Admin"
jsonpath "$.id" == {{user_id}}

# Get user is ok for super user
GET {{host}}/api/v1/users/{{user_id}}
Authorization: Bearer {{admin_token}}

HTTP 200
[Asserts]
jsonpath "$.email" == "test@example.com"
jsonpath "$.full_name" == "Updated User By Admin"
jsonpath "$.id" == {{user_id}}

# List users as super user
GET {{host}}/api/v1/users/
Authorization: Bearer {{admin_token}}
[QueryStringParams]
skip: 0
limit: 10

HTTP 200
[Asserts]
jsonpath "$.users" isCollection
jsonpath "$.users[0]" exists
jsonpath "$.users[1]" exists
jsonpath "$.meta.skip" == 0
jsonpath "$.meta.limit" == 10

# Login normal user again
POST {{host}}/api/v1/login/access-token
Content-Type: application/x-www-form-urlencoded
[FormParams]
username: test@example.com
password: newpassword123
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
jsonpath "$.message" == "Authorization header is required"