package routes

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/wangfenjin/mojito/internal/app/middleware"
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

		loginGroup.POST("/login/test-token", testTokenHandler)

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
	// Now you can use req.Username, req.Password, etc. directly
	return nil, errors.New("Not implemented: loginAccessTokenHandler - Username: " + req.Username)
}

func testTokenHandler(ctx context.Context, c *app.RequestContext) {
	// This endpoint doesn't have a request body or path params
	panic("Not implemented: testTokenHandler")
}

func recoverPasswordHandler(ctx context.Context, req RecoverPasswordRequest) (interface{}, error) {
	// Now you can use req.Email directly
	return nil, errors.New("Not implemented: recoverPasswordHandler - Email: " + req.Email)
}

func resetPasswordHandler(ctx context.Context, req ResetPasswordRequest) (interface{}, error) {
	// Now you can use req.Token and req.Password directly
	return nil, errors.New("Not implemented: resetPasswordHandler - Token: " + req.Token)
}

func recoverPasswordHtmlContentHandler(ctx context.Context, req RecoverPasswordHtmlContentRequest) (interface{}, error) {
	// Now you can use req.Email directly
	return nil, errors.New("Not implemented: recoverPasswordHtmlContentHandler - Email: " + req.Email)
}
