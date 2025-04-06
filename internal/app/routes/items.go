package routes

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/wangfenjin/mojito/internal/app/middleware"
)

// Request structs for items routes
type CreateItemRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type UpdateItemRequest struct {
	ID          string `path:"id" binding:"required"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GetItemRequest struct {
	ID string `path:"id" binding:"required"`
}

type ListItemsRequest struct {
	Skip  int `query:"skip"`
	Limit int `query:"limit"`
}

// RegisterItemsRoutes registers all item related routes
func RegisterItemsRoutes(h *server.Hertz) {
	itemsGroup := h.Group("/api/v1/items")
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
	return nil, errors.New("Not implemented: createItemHandler - Title: " + req.Title)
}

func getItemHandler(ctx context.Context, req GetItemRequest) (interface{}, error) {
	return nil, errors.New("Not implemented: getItemHandler - ID: " + req.ID)
}

func updateItemHandler(ctx context.Context, req UpdateItemRequest) (interface{}, error) {
	return nil, errors.New("Not implemented: updateItemHandler - ID: " + req.ID)
}

func deleteItemHandler(ctx context.Context, req GetItemRequest) (interface{}, error) {
	return nil, errors.New("Not implemented: deleteItemHandler - ID: " + req.ID)
}

func listItemsHandler(ctx context.Context, req ListItemsRequest) (interface{}, error) {
	return nil, errors.New("Not implemented: listItemsHandler")
}
