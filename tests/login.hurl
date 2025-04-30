# Clean up test data first
DELETE {{host}}/api/v1/test/cleanup
HTTP 200

# Create a test user
POST {{host}}/api/v1/users/signup
Content-Type: application/json
{
    "email": "test@example.com",
    "password": "password123",
    "full_name": "Test User"
}

HTTP 200

# Test successful login
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
access_token: jsonpath "$.access_token"
[Asserts]
jsonpath "$.access_token" exists
jsonpath "$.token_type" == "bearer"

# Test token validation
GET {{host}}/api/v1/login/test-token
Authorization: Bearer {{access_token}}
HTTP 200
[Asserts]
jsonpath "$.user_id" exists
jsonpath "$.email" == "test@example.com"

# Test invalid credentials
POST {{host}}/api/v1/login/access-token
Content-Type: application/x-www-form-urlencoded
[FormParams]
username: test@example.com
password: wrongpassword
grant_type: password

HTTP 400
[Asserts]
jsonpath "$.message" == "invalid credentials"

# Test password recovery request
POST {{host}}/api/v1/password-recovery/test@example.com
HTTP 200
[Asserts]
jsonpath "$.message" == "password recovery email sent"

# Test password recovery for non-existent user
POST {{host}}/api/v1/password-recovery/nonexistent@example.com
HTTP 400

# Test reset password
POST {{host}}/api/v1/reset-password/
Content-Type: application/json
{
    "token": "some-reset-token",
    "password": "newpassword123"
}
HTTP 200
[Asserts]
jsonpath "$.message" == "password reset successful"

# Test password recovery HTML content
POST {{host}}/api/v1/password-recovery-html-content/test@example.com
HTTP 200
[Asserts]
jsonpath "$.html_content" exists