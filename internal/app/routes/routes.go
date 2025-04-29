package routes

import "github.com/go-chi/chi/v5"

// RegisterRoutes registers all application routes
func RegisterRoutes(r chi.Router) {
	RegisterUtilRoutes(r)
	RegisterLoginRoutes(r)
	RegisterUsersRoutes(r)
	RegisterItemsRoutes(r)
	RegisterDocsRoutes(r)
}
