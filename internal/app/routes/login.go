package routes

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/repository"
	"github.com/wangfenjin/mojito/internal/app/utils"
)

// Request structs for login routes
type LoginAccessTokenRequest struct {
	Username     string `form:"username" binding:"required"`
	Password     string `form:"password" binding:"required"`
	GrantType    string `form:"grant_type"`
	Scope        string `form:"scope"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
}

type RecoverPasswordRequest struct {
	Email string `path:"email" binding:"required"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RecoverPasswordHtmlContentRequest struct {
	Email string `path:"email" binding:"required"`
}

// Response structs
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// RegisterLoginRoutes registers all login related routes
func RegisterLoginRoutes(h *server.Hertz) {
	loginGroup := h.Group("/api/v1")
	{
		loginGroup.POST("/login/access-token",
			middleware.WithRequest(LoginAccessTokenRequest{}),
			middleware.WithResponse(loginAccessTokenHandler))

		// Change POST to GET and wrap with middleware
		loginGroup.GET("/login/test-token",
			middleware.WithResponse(testTokenHandler))

		loginGroup.POST("/password-recovery/:email",
			middleware.WithRequest(RecoverPasswordRequest{}),
			middleware.WithResponse(recoverPasswordHandler))

		loginGroup.POST("/reset-password/",
			middleware.WithRequest(ResetPasswordRequest{}),
			middleware.WithResponse(resetPasswordHandler))

		loginGroup.POST("/password-recovery-html-content/:email",
			middleware.WithRequest(RecoverPasswordHtmlContentRequest{}),
			middleware.WithResponse(recoverPasswordHtmlContentHandler))
	}
}

// Login handlers with updated signatures
func loginAccessTokenHandler(ctx context.Context, req LoginAccessTokenRequest) (interface{}, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	// Get user by email
	user, err := userRepo.GetByEmail(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, middleware.NewBadRequestError("invalid credentials")
	}

	// Check password using utils package
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, middleware.NewBadRequestError("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, middleware.NewBadRequestError("inactive user")
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	return TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
	}, nil
}

// Update testTokenHandler signature to match middleware
func testTokenHandler(ctx context.Context, _ interface{}) (interface{}, error) {
	c := ctx.Value("requestContext").(*app.RequestContext)

	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) == 0 {
		return nil, middleware.NewUnauthorizedError("missing authorization header")
	}

	// Extract token from "Bearer <token>"
	tokenString := string(authHeader[7:])
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		return nil, middleware.NewUnauthorizedError("invalid token")
	}

	return map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
	}, nil
}

func recoverPasswordHandler(ctx context.Context, req RecoverPasswordRequest) (interface{}, error) {
	userRepo := ctx.Value("userRepository").(*repository.UserRepository)

	// Check if user exists
	user, err := userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return nil, middleware.NewBadRequestError("user not found")
	}

	// TODO: Generate password reset token and send email
	// For now, just return success message
	return MessageResponse{
		Message: "password recovery email sent",
	}, nil
}

func resetPasswordHandler(ctx context.Context, req ResetPasswordRequest) (interface{}, error) {
	// TODO: Implement password reset logic with token validation
	return MessageResponse{
		Message: "password reset successful",
	}, nil
}

func recoverPasswordHtmlContentHandler(ctx context.Context, req RecoverPasswordHtmlContentRequest) (interface{}, error) {
	// TODO: Generate HTML content for password recovery email
	return map[string]string{
		"html_content": "<h1>Reset Your Password</h1><p>Click the link below to reset your password.</p>",
	}, nil
}
