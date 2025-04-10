package routes

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/repository"
)

// RegisterItemsRoutes registers all item related routes
func RegisterItemsRoutes(r *gin.Engine) {
	itemsGroup := r.Group("/api/v1/items")
	{
		itemsGroup.POST("/",
			middleware.WithHandler(createItemHandler))

		itemsGroup.GET("/:id",
			middleware.WithHandler(getItemHandler))

		itemsGroup.PUT("/:id",
			middleware.WithHandler(updateItemHandler))

		itemsGroup.DELETE("/:id",
			middleware.WithHandler(deleteItemHandler))

		itemsGroup.GET("/",
			middleware.WithHandler(listItemsHandler))
	}
}

// Request structs for items routes
type CreateItemRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateItemRequest struct {
	ID          string `path:"id" binding:"required,uuid"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type GetItemRequest struct {
	ID string `path:"id" binding:"required,uuid"`
}

type ListItemsRequest struct {
	Skip  int `query:"skip" binding:"min=0" default:"0"`
	Limit int `query:"limit" binding:"min=1,max=100" default:"100"`
}

// Item handlers
// Add response structs
type ItemResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ItemsResponse struct {
	Items []ItemResponse `json:"items"`
	Meta  struct {
		Skip  int `json:"skip"`
		Limit int `json:"limit"`
	} `json:"meta"`
}

// Update handlers to use the new response types
func createItemHandler(ctx context.Context, req CreateItemRequest) (*ItemResponse, error) {
	itemRepo := ctx.Value("itemRepository").(*repository.ItemRepository)

	// Get user_id from context instead of claims
	userID := ctx.Value("user_id").(string)
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}

	item := &models.Item{
		Title:       req.Title,
		Description: req.Description,
		OwnerID:     id,
	}

	if err := itemRepo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("error creating item: %w", err)
	}

	return &ItemResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}, nil
}

func getItemHandler(ctx context.Context, req GetItemRequest) (*ItemResponse, error) {
	itemRepo := ctx.Value("itemRepository").(*repository.ItemRepository)

	// Get user_id from context
	userID := ctx.Value("user_id").(string)
	ownerID, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid item ID format")
	}

	// Check if item exists and belongs to the user
	item, err := itemRepo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error getting item: %w", err)
	}
	if item == nil {
		return nil, middleware.NewBadRequestError("item not found or access denied")
	}

	return &ItemResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}, nil
}

func updateItemHandler(ctx context.Context, req UpdateItemRequest) (*ItemResponse, error) {
	itemRepo := ctx.Value("itemRepository").(*repository.ItemRepository)

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid item ID format")
	}

	item, err := itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error getting item: %w", err)
	}
	if item == nil {
		return nil, middleware.NewBadRequestError("item not found")
	}

	if req.Title != "" {
		item.Title = req.Title
	}
	if req.Description != "" {
		item.Description = req.Description
	}

	err = itemRepo.Update(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("error updating item: %w", err)
	}

	return &ItemResponse{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}, nil
}

func deleteItemHandler(ctx context.Context, req GetItemRequest) (*MessageResponse, error) {
	itemRepo := ctx.Value("itemRepository").(*repository.ItemRepository)

	// Get user_id from context
	userID := ctx.Value("user_id").(string)
	ownerID, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid item ID format")
	}

	// Check if item exists and belongs to the user
	item, err := itemRepo.GetByIDAndOwner(ctx, id, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error getting item: %w", err)
	}
	if item == nil {
		return nil, middleware.NewBadRequestError("item not found or access denied")
	}

	err = itemRepo.Delete(ctx, id, ownerID)
	if err != nil {
		return nil, fmt.Errorf("error deleting item: %w", err)
	}

	return &MessageResponse{
		Message: "item deleted successfully",
	}, nil
}

func listItemsHandler(ctx context.Context, req ListItemsRequest) (*ItemsResponse, error) {
	itemRepo := ctx.Value("itemRepository").(*repository.ItemRepository)

	// Get user_id from context
	userID := ctx.Value("user_id").(string)
	ownerID, err := uuid.Parse(userID)
	if err != nil {
		return nil, middleware.NewBadRequestError("invalid owner ID")
	}

	items, err := itemRepo.List(ctx, ownerID, req.Skip, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("error listing items: %w", err)
	}

	itemList := make([]ItemResponse, len(items))
	for i, item := range items {
		itemList[i] = ItemResponse{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	return &ItemsResponse{
		Items: itemList,
		Meta: struct {
			Skip  int `json:"skip"`
			Limit int `json:"limit"`
		}{
			Skip:  req.Skip,
			Limit: req.Limit,
		},
	}, nil
}
