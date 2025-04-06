# Clean up test data first
DELETE http://localhost:8888/api/v1/test/cleanup
HTTP 200

# Create a new item
POST http://localhost:8888/api/v1/items/
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
GET http://localhost:8888/api/v1/items/{{item_id}}
HTTP 200
[Asserts]
jsonpath "$.title" == "Test Item"
jsonpath "$.description" == "This is a test item"
jsonpath "$.id" == {{item_id}}

# Update the item
PUT http://localhost:8888/api/v1/items/{{item_id}}
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
GET http://localhost:8888/api/v1/items/
[QueryStringParams]
skip: 0
limit: 10

HTTP 200
[Asserts]
jsonpath "$.items" isCollection
jsonpath "$.items[0]" exists
jsonpath "$.meta.skip" == 0
jsonpath "$.meta.limit" == 10

# Delete the item
DELETE http://localhost:8888/api/v1/items/{{item_id}}
HTTP 200
[Asserts]
jsonpath "$.message" == "item deleted successfully"

# Verify item is deleted
GET http://localhost:8888/api/v1/items/{{item_id}}
HTTP 400
[Asserts]
jsonpath "$.error" == "item not found"