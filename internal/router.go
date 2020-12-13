package app

import (
	"github.com/gin-gonic/gin"

	handler "github.com/utmhikari/repomaster/internal/handler"
)

// Router gin router
func Router() *gin.Engine {
	r := gin.Default()
	api := r.Group("/api")
	v1 := api.Group("/v1")
	{
		base := v1.Group("/base")
		{
			base.GET("/health", handler.Base.HealthCheck)
		}
	}
	return r
}
