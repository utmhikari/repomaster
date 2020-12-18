package app

import (
	"github.com/gin-gonic/gin"

	handler "github.com/utmhikari/repomaster/internal/handler"
)

// getWebHandler get gin web handler
func getWebHandler() *gin.Engine {
	r := gin.Default()
	api := r.Group("/api")
	v1 := api.Group("/v1")
	{
		base := v1.Group("/base")
		{
			base.GET("/health", handler.Base.HealthCheck)
		}
		repo := v1.Group("/repo/:id")
		{
			repo.GET("/info", handler.Repo.GetRepoInfoByID)
		}
	}
	return r
}
