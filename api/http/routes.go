package api

import (
	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/controllers"
	m "github.com/kunalvirwal/shogun-cd/api/http/middlewares"
)

func initRoutes(router *gin.Engine, h *controllers.Handler) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", h.Login)
	}
	webhook := router.Group("/webhook")
	{
		webhook.POST("/:token", h.HandleWebhook)
	}

	admin := router.Group("/admin")
	admin.Use(m.AuthRequired(), m.VerifyAdmin)
	{
		key := admin.Group("/api-keys")
		{
			key.POST("", h.CreateAPIKey)
			key.GET("", h.ListAllAPIKeys)
			key.DELETE("/:key", h.DeleteAPIKey)
		}
		secret := admin.Group("/secrets")
		{
			secret.GET("", h.ListAllSecrets)
			secret.POST("", h.CreateSecret)
			secret.PATCH("/:secret", h.UpdateSecret)
			secret.DELETE("/:secret", h.DeleteSecret)
		}
	}

	system := router.Group("/system")
	system.Use(m.AuthRequired())
	{
		system.GET("/targets", h.ListAllTargets)
		system.GET("/pipelines", h.ListAllPipelines)
	}

	pipelines := router.Group("/pipelines/:pipeline")
	pipelines.Use(m.AuthRequired())
	{
		webhooks := pipelines.Group("/webhooks")
		{
			webhooks.GET("", h.ListPipelineWebhooks)
			webhooks.POST("", h.CreateWebhook)
			webhooks.DELETE("/:webhook_id", h.DeleteWebhook)
		}

		runs := pipelines.Group("/runs")
		{
			runs.GET("", h.ListPipelineRuns)
			runs.GET("/:run_id/logs", h.StreamPipelineRunLogs)
			runs.GET("/:run_id/steps", h.ListPipelineRunSteps)
		}
	}
}
