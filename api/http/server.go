package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/controllers"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
)

func StartAPIServer() {
	logger := utils.NewLogger(utils.DebugLevel, true)
	//[TODO] pass same logger through argument throughout.

	r := newRouter()
	h := controllers.NewHandler(logger)
	initRoutes(r, h)

	if err := r.Run(":7007"); err != nil {
		logger.LogError(err)
	}
}

func newRouter() *gin.Engine {
	r := gin.Default()

	r.SetTrustedProxies(nil)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, //[TODO] tighten allowed origins
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.MaxMultipartMemory = 64 << 20

	return r
}
