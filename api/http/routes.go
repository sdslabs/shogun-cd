package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/controllers"
	"github.com/kunalvirwal/shogun-cd/api/http/middlewares"
)

func initRoutes(router *gin.Engine, m *middlewares.Manager, h *controllers.Handler) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
	}

	webhook := router.Group("/hook")
	{
		webhook.GET("/:pipeline", m.AuthRequired, h.ListWebhooks)
		webhook.GET("", m.AuthRequired, h.ListWebhooks)
		webhook.POST("/:slug", h.HandleWebhook)
	}

	admin := router.Group("/admin")
	admin.Use(m.AuthRequired, m.VerifyAdmin)
	{
		webhook := admin.Group("/hook")
		{
			webhook.POST("", h.CreateWebhook)
			webhook.DELETE("/:slug", h.DeleteWebhook)
			// [TODO] deactivate a webhook
		}

		// [TODO] api key routes are not implemented yet
		key := admin.Group("/api-keys")
		{
			key.POST("", h.CreateAPIKey)
			key.GET("", h.ListAllAPIKeys)
			key.DELETE("/:key", h.DeleteAPIKey)
		}
		// [TODO] secrets routes are not implemented yet
		secret := admin.Group("/secrets")
		{
			secret.GET("", h.ListAllSecrets)
			secret.POST("", h.CreateSecret)
			secret.PATCH("/:secret", h.UpdateSecret)
			secret.DELETE("/:secret", h.DeleteSecret)
		}
	}

	// [TODO] system routes are not implemented yet
	system := router.Group("/system")
	system.Use(m.AuthRequired)
	{
		system.GET("/targets", h.ListAllTargets)
		system.GET("/pipelines", h.ListAllPipelines)
	}

	// [TODO] pipeline data routes are not implemented yet
	pipelines := router.Group("/pipelines/:pipeline")
	pipelines.Use(m.AuthRequired)
	{
		runs := pipelines.Group("/runs")
		{
			runs.GET("", h.ListPipelineRuns)
			runs.GET("/:run_id/logs", h.StreamPipelineRunLogs)
			runs.GET("/:run_id/steps", h.ListPipelineRunSteps)
		}
	}
}
