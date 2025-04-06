package routes

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/google/uuid"
	"github.com/wangfenjin/mojito/internal/app/middleware"
	"github.com/wangfenjin/mojito/internal/app/models"
	"github.com/wangfenjin/mojito/internal/app/repository"
)

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

// RegisterItemsRoutes registers all item related routes
func RegisterItemsRoutes(h *server.Hertz) {
	itemsGroup := h.Group("/api/v1/items", middleware.RequireAuth())
	{
		itemsGroup.POST("/",
			middleware.WithRequest(CreateItemRequest{}),
			middleware.WithResponse(createItemHandler))

		itemsGroup.GET("/:id",
			middleware.WithRequest(GetItemRequest{}),
			middleware.WithResponse(getItemHandler))

		itemsGroup.PUT("/:id",
			middleware.WithRequest(UpdateItemRequest{}),
			middleware.WithResponse(updateItemHandler))

		itemsGroup.DELETE("/:id",
			middleware.WithRequest(GetItemRequest{}),
			middleware.WithResponse(deleteItemHandler))

		itemsGroup.GET("/",
			middleware.WithRequest(ListItemsRequest{}),
			middleware.WithResponse(listItemsHandler))
	}
}

// Item handlers
func createItemHandler(ctx context.Context, req CreateItemRequest) (interface{}, error) {
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

	return map[string]interface{}{
		"id":          item.ID,
		"title":       item.Title,
		"description": item.Description,
		"created_at":  item.CreatedAt,
	}, nil
}

func getItemHandler(ctx context.Context, req GetItemRequest) (interface{}, error) {
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

	return map[string]interface{}{
		"id":          item.ID,
		"title":       item.Title,
		"description": item.Description,
		"created_at":  item.CreatedAt,
		"updated_at":  item.UpdatedAt,
	}, nil
}

func updateItemHandler(ctx context.Context, req UpdateItemRequest) (interface{}, error) {
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

	return map[string]interface{}{
		"id":          item.ID,
		"title":       item.Title,
		"description": item.Description,
		"created_at":  item.CreatedAt,
		"updated_at":  item.UpdatedAt,
	}, nil
}

func deleteItemHandler(ctx context.Context, req GetItemRequest) (interface{}, error) {
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

	return map[string]string{
		"message": "item deleted successfully",
	}, nil
}

func listItemsHandler(ctx context.Context, req ListItemsRequest) (interface{}, error) {
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

	itemList := make([]map[string]interface{}, len(items))
	for i, item := range items {
		itemList[i] = map[string]interface{}{
			"id":          item.ID,
			"title":       item.Title,
			"description": item.Description,
			"created_at":  item.CreatedAt,
			"updated_at":  item.UpdatedAt,
		}
	}

	return map[string]interface{}{
		"items": itemList,
		"meta": map[string]interface{}{
			"skip":  req.Skip,
			"limit": req.Limit,
		},
	}, nil
}
