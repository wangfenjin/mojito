# Clean up test data first
DELETE http://localhost:8888/api/v1/test/cleanup
HTTP 200

# Create a new user
POST http://localhost:8888/api/v1/users/
Content-Type: application/json
{
    "email": "test_user@example.com",
    "password": "password123",
    "full_name": "Test User",
    "is_active": true,
    "is_superuser": false
}

HTTP 200
[Captures]
user_id: jsonpath "$.id"
[Asserts]
jsonpath "$.email" == "test_user@example.com"
jsonpath "$.full_name" == "Test User"
jsonpath "$.is_active" == true
jsonpath "$.is_superuser" == false
jsonpath "$.id" exists
jsonpath "$.created_at" exists
jsonpath "$.password" not exists

# Get the created user
GET http://localhost:8888/api/v1/users/{{user_id}}
HTTP 200

# Update the user (email update not allowed)
PUT http://localhost:8888/api/v1/users/{{user_id}}
Content-Type: application/json
{
    "full_name": "Updated User",
    "is_active": true,
    "is_superuser": true
}

HTTP 200
[Asserts]
jsonpath "$.email" == "test_user@example.com"
jsonpath "$.full_name" == "Updated User"
jsonpath "$.is_active" == true
jsonpath "$.is_superuser" == true
jsonpath "$.id" == {{user_id}}

# Try to update email (should fail)
PUT http://localhost:8888/api/v1/users/{{user_id}}
Content-Type: application/json
{
    "email": "new_email@example.com"
}

HTTP 400
[Asserts]
jsonpath "$.error" == "email updates are not allowed"

# Update the user
PUT http://localhost:8888/api/v1/users/{{user_id}}
Content-Type: application/json
{
    "full_name": "Updated User",
    "is_active": true,
    "is_superuser": true
}

HTTP 200
[Asserts]
jsonpath "$.full_name" == "Updated User"
jsonpath "$.is_active" == true
jsonpath "$.is_superuser" == true
jsonpath "$.id" == {{user_id}}

# List users
GET http://localhost:8888/api/v1/users/
[QueryStringParams]
skip: 0
limit: 10

HTTP 200
[Asserts]
jsonpath "$.users" isCollection
jsonpath "$.users[0]" exists
jsonpath "$.meta.skip" == 0
jsonpath "$.meta.limit" == 10

# Try to create user with existing email (should fail)
POST http://localhost:8888/api/v1/users/
Content-Type: application/json
{
    "email": "test_user@example.com",
    "password": "password123",
    "full_name": "Test User",
    "is_active": true,
    "is_superuser": false
}

HTTP 400
[Asserts]
jsonpath "$.error" exists