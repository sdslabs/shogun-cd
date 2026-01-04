package api

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/controllers"
	"github.com/kunalvirwal/shogun-cd/api/http/middlewares"
	"github.com/kunalvirwal/shogun-cd/api/http/response"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
)

func StartAPIServer(logger utils.Logger, cfg *config.Config, w webhooks.WebhookService) {

	r := newRouter()

	responder := response.NewResponder(cfg.Debug)
	m := middlewares.NewManager(logger, cfg, responder, w)
	h := controllers.NewHandler(logger, cfg, responder, w)

	initRoutes(r, m, h)

	if err := r.Run(fmt.Sprintf(":%v", cfg.ApiConfig.Port)); err != nil {
		logger.LogError(err)
	}

}

func newRouter() *gin.Engine {
	r := gin.Default()

	r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, //[TODO] tighten allowed origins
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, //[TODO] add HMAC headers for browser
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	return r
}
