package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/models/gen"
	"github.com/wangfenjin/mojito/internal/app/utils"
)

// RegisterItemsRoutes registers all item related routes
func RegisterItemsRoutes(r chi.Router) {
	r.Route("/api/v1/items", func(r chi.Router) {
		// Apply auth middleware to all item routes
		r.Use(middleware.RequireAuth())

		r.Post("/", middleware.WithHandler(createItemHandler))
		r.Get("/{id}", middleware.WithHandler(getItemHandler))
		r.Put("/{id}", middleware.WithHandler(updateItemHandler))
		r.Delete("/{id}", middleware.WithHandler(deleteItemHandler))
		r.Get("/", middleware.WithHandler(listItemsHandler))
	})
}

// CreateItemRequest represents the request body for creating an item
type CreateItemRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// UpdateItemRequest represents the request body for updating an item
type UpdateItemRequest struct {
	ID          string `uri:"id" binding:"required,uuid"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// GetItemRequest represents the request parameters for getting an item
type GetItemRequest struct {
	ID string `uri:"id" binding:"required,uuid"`
}

// ListItemsRequest represents the request parameters for listing items
type ListItemsRequest struct {
	Skip  int64 `query:"skip" binding:"min=0" default:"0"`
	Limit int64 `query:"limit" binding:"min=1,max=100" default:"10"`
}

// ItemResponse represents a single item in the response
type ItemResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ItemsResponse represents a collection of items in the response
type ItemsResponse struct {
	Items []ItemResponse `json:"items"`
	Meta  struct {
		Skip  int64 `json:"skip"`
		Limit int64 `json:"limit"`
	} `json:"meta"`
}

// Update handlers to use the new response types
func createItemHandler(ctx context.Context, req CreateItemRequest) (*ItemResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}

	item, err := db.CreateItem(ctx, gen.CreateItemParams{
		Title:       req.Title,
		Description: pgtype.Text{String: req.Description, Valid: true},
		OwnerID:     ownerID,
		ID:          uuid.New(),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating item: %w", err)
	}

	return &ItemResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description.String,
		CreatedAt:   item.CreatedAt.Time,
		UpdatedAt:   item.UpdatedAt.Time,
	}, nil
}

func getItemHandler(ctx context.Context, req GetItemRequest) (*ItemResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid item ID format")
	}

	item, err := db.GetItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting item: %w", err)
	}
	if item.OwnerID != ownerID && !claims.IsSuperUser {
		return nil, middleware.NewBadRequestError("item not found or access denied")
	}

	return &ItemResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description.String,
		CreatedAt:   item.CreatedAt.Time,
		UpdatedAt:   item.UpdatedAt.Time,
	}, nil
}

func updateItemHandler(ctx context.Context, req UpdateItemRequest) (*ItemResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid item ID format")
	}

	item, err := db.GetItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting item: %w", err)
	}
	if item.OwnerID != ownerID && !claims.IsSuperUser {
		return nil, middleware.NewBadRequestError("item not found or access denied")
	}

	item, err = db.UpdateItem(ctx, gen.UpdateItemParams{
		Title:       req.Title,
		Description: pgtype.Text{String: req.Description, Valid: true},
		ID:          id,
	})
	if err != nil {
		return nil, fmt.Errorf("error updating item: %w", err)
	}

	return &ItemResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description.String,
		CreatedAt:   item.CreatedAt.Time,
		UpdatedAt:   item.UpdatedAt.Time,
	}, nil
}

func deleteItemHandler(ctx context.Context, req GetItemRequest) (*MessageResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	// Get user_id from context
	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid item ID format")
	}

	// Check if item exists and belongs to the user
	item, err := db.GetItemByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting item: %w", err)
	}
	if item.OwnerID != ownerID && !claims.IsSuperUser {
		return nil, middleware.NewBadRequestError("item not found or access denied")
	}

	err = db.DeleteItem(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error deleting item: %w", err)
	}

	return &MessageResponse{
		Message: "item deleted successfully",
	}, nil
}

func listItemsHandler(ctx context.Context, req ListItemsRequest) (*ItemsResponse, error) {
	claims := ctx.Value("claims").(*utils.Claims)
	db := ctx.Value("database").(*models.DB)

	// Get user_id from context
	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}

	items, err := db.ListItemsByOwner(ctx, gen.ListItemsByOwnerParams{
		OwnerID: ownerID,
		Limit:   req.Limit,
		Offset:  req.Skip,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing items: %w", err)
	}

	itemList := make([]ItemResponse, len(items))
	for i, item := range items {
		itemList[i] = ItemResponse{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description.String,
			CreatedAt:   item.CreatedAt.Time,
			UpdatedAt:   item.UpdatedAt.Time,
		}
	}

	return &ItemsResponse{
		Items: itemList,
		Meta: struct {
			Skip  int64 `json:"skip"`
			Limit int64 `json:"limit"`
		}{
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	}, nil
}
