# Clean up test data first
DELETE {{host}}/api/v1/test/cleanup
HTTP 200

# Register a test user
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

# Create a new item
POST {{host}}/api/v1/items/
Authorization: Bearer {{token}}
Content-Type: application/json
{
    "title": "Test Item",
    "description": "This is a test item"
}

HTTP 200
[Captures]
item_id: jsonpath "$.id"
[Asserts]
jsonpath "$.title" == "Test Item"
jsonpath "$.description" == "This is a test item"
jsonpath "$.id" exists
jsonpath "$.created_at" exists

# Get the created item
GET {{host}}/api/v1/items/{{item_id}}
Authorization: Bearer {{token}}
HTTP 200
[Asserts]
jsonpath "$.title" == "Test Item"
jsonpath "$.description" == "This is a test item"
jsonpath "$.id" == {{item_id}}

# Update the item
PATCH {{host}}/api/v1/items/{{item_id}}
Authorization: Bearer {{token}}
Content-Type: application/json
{
    "title": "Updated Item",
    "description": "This is an updated item"
}

HTTP 200
[Asserts]
jsonpath "$.title" == "Updated Item"
jsonpath "$.description" == "This is an updated item"
jsonpath "$.id" == {{item_id}}

# List items
GET {{host}}/api/v1/items/
Authorization: Bearer {{token}}
[QueryStringParams]
skip: 0
limit: 10

HTTP 200
[Asserts]
jsonpath "$.items" isCollection
jsonpath "$.items[0]" exists
jsonpath "$.meta.skip" == 0
jsonpath "$.meta.limit" == 10

# Register a test user2
POST {{host}}/api/v1/users/signup
Content-Type: application/json
{
    "email": "test2@example.com",
    "password": "password123",
    "full_name": "Test User2"
}

HTTP 200
[Captures]
user_id2: jsonpath "$.id"

# Login to get token
POST {{host}}/api/v1/login/access-token
Content-Type: application/x-www-form-urlencoded
[FormParams]
username: test2@example.com
password: password123
grant_type: password
scope: 
client_id: 
client_secret: 

HTTP 200
[Captures]
token2: jsonpath "$.access_token"

# Get user1's items failed
GET {{host}}/api/v1/items/{{item_id}}
Authorization: Bearer {{token2}}
HTTP 403

# Update the user1' item failed
PATCH {{host}}/api/v1/items/{{item_id}}
Authorization: Bearer {{token2}}
Content-Type: application/json
{
    "title": "Updated Item",
    "description": "This is an updated item"
}

HTTP 403

# List items empty
GET {{host}}/api/v1/items/
Authorization: Bearer {{token2}}
[QueryStringParams]
skip: 0
limit: 10

HTTP 200
[Asserts]
jsonpath "$.items" isEmpty
jsonpath "$.meta.skip" == 0
jsonpath "$.meta.limit" == 10

# Delete the item
DELETE {{host}}/api/v1/items/{{item_id}}
Authorization: Bearer {{token}}
HTTP 200
[Asserts]
jsonpath "$.message" == "item deleted successfully"

# Verify item is deleted
GET {{host}}/api/v1/items/{{item_id}}
Authorization: Bearer {{token}}
HTTP 400