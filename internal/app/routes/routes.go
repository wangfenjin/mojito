package routes

import "github.com/gin-gonic/gin"

// RegisterRoutes registers all application routes
func RegisterRoutes(r *gin.Engine) {
	RegisterUtilRoutes(r)
	RegisterLoginRoutes(r)
	RegisterUsersRoutes(r)
	RegisterItemsRoutes(r)
	RegisterDocsRoutes(r)
}
