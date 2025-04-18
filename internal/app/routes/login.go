package routes

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/repository"
	"github.com/wangfenjin/mojito/internal/app/utils"
	"github.com/wangfenjin/mojito/internal/pkg/logger"
)

// RegisterLoginRoutes registers all login related routes
func RegisterLoginRoutes(r *gin.Engine) {
	loginGroup := r.Group("/api/v1")
	{
		loginGroup.POST("/login/access-token",
			middleware.WithHandler(loginAccessTokenHandler))

		loginGroup.GET("/login/test-token",
			middleware.WithHandler(testTokenHandler))

		loginGroup.POST("/password-recovery/:email",
			middleware.WithHandler(recoverPasswordHandler))

		loginGroup.POST("/reset-password/",
			middleware.WithHandler(resetPasswordHandler))

		loginGroup.POST("/password-recovery-html-content/:email",
			middleware.WithHandler(recoverPasswordHTMLContentHandler))
	}
}

// LoginAccessTokenRequest structs
type LoginAccessTokenRequest struct {
	Username     string `form:"username" binding:"required"`
	Password     string `form:"password" binding:"required"`
	GrantType    string `form:"grant_type"`
	Scope        string `form:"scope"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
}

// RecoverPasswordRequest structs
type RecoverPasswordRequest struct {
	Email string `uri:"email" binding:"required,email"`
}

// ResetPasswordRequest structs
type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// RecoverPasswordHTMLContentRequest structs
type RecoverPasswordHTMLContentRequest struct {
	Email string `uri:"email" binding:"required,email"`
}

// TokenResponse structs
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// MessageResponse structs
type MessageResponse struct {
	Message string `json:"message"`
}

// TestTokenResponse structs
type TestTokenResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// HTMLContentResponse structs
type HTMLContentResponse struct {
	HTMLContent string `json:"html_content"`
}

// Login handlers with updated signatures
func loginAccessTokenHandler(ctx context.Context, req LoginAccessTokenRequest) (*TokenResponse, error) {
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

	return &TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
	}, nil
}

// TestTokenRequest represents a request for generating test tokens
type TestTokenRequest struct {
	Token string `header:"Authorization" binding:"required"`
}

// Update handler signatures to use pointer returns
func testTokenHandler(_ context.Context, req TestTokenRequest) (*TestTokenResponse, error) {
	// Get token from Authorization header
	authHeader := req.Token
	logger.GetLogger().Debug("authHeader", "token", authHeader)
	// Extract token from "Bearer <token>"
	tokenString := string(authHeader[7:])
	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		logger.GetLogger().Error("error validating token", "error", err)
		return nil, middleware.NewUnauthorizedError("invalid token")
	}

	return &TestTokenResponse{
		UserID: claims.UserID,
		Email:  claims.Email,
	}, nil
}

func recoverPasswordHandler(ctx context.Context, req RecoverPasswordRequest) (*MessageResponse, error) {
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
	return &MessageResponse{
		Message: "password recovery email sent",
	}, nil
}

func resetPasswordHandler(_ context.Context, _ ResetPasswordRequest) (*MessageResponse, error) {
	// TODO: Implement password reset logic with token validation
	return &MessageResponse{
		Message: "password reset successful",
	}, nil
}

func recoverPasswordHTMLContentHandler(_ context.Context, _ RecoverPasswordHTMLContentRequest) (*HTMLContentResponse, error) {
	return &HTMLContentResponse{
		HTMLContent: "<h1>Reset Your Password</h1><p>Click the link below to reset your password.</p>",
	}, nil
}
