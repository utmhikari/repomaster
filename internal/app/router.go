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
		v1.GET("/health", handler.HealthCheck)
		repos := v1.Group("/repos")
		{
			repos.GET("/:id", handler.Repo.GetByID)

		}
		repo := v1.Group("/repo")
		{
			repo.POST("/git", handler.Repo.CreateGit)
			repo.PUT("/git", handler.Repo.UpdateGit)
		}
	}
	return r
}
