package routes

import (
	"github.com/cloudwego/hertz/pkg/app/server"
)

// RegisterRoutes registers all application routes
func RegisterRoutes(h *server.Hertz) {
	RegisterUtilRoutes(h)
	RegisterLoginRoutes(h)
	RegisterUsersRoutes(h)
	RegisterItemsRoutes(h)
}
