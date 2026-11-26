package api

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kunalvirwal/shogun-cd/api/http/controllers"
	"github.com/kunalvirwal/shogun-cd/api/http/middlewares"
	"github.com/kunalvirwal/shogun-cd/api/http/response"
	"github.com/kunalvirwal/shogun-cd/docs"
	"github.com/kunalvirwal/shogun-cd/internal/config"
	"github.com/kunalvirwal/shogun-cd/internal/crons"
	"github.com/kunalvirwal/shogun-cd/internal/orchestrator"
	"github.com/kunalvirwal/shogun-cd/internal/secrets"
	"github.com/kunalvirwal/shogun-cd/internal/store"
	"github.com/kunalvirwal/shogun-cd/internal/users"
	"github.com/kunalvirwal/shogun-cd/internal/utils"
	"github.com/kunalvirwal/shogun-cd/internal/webhooks"
	"github.com/kunalvirwal/shogun-cd/web"
)

func StartAPIServer(logger utils.Logger, cfg *config.Config, w webhooks.WebhookService, u users.Service, secrets secrets.SecretService, store *store.Store, orch orchestrator.Orchestrator, cronService crons.CronService) {
	r := newRouter()

	responder := response.NewResponder(cfg.Debug)
	m := middlewares.NewManager(logger, cfg, responder, w)
	h := controllers.NewHandler(logger, cfg, responder, w, u, secrets, store, orch, cronService)
	initRoutes(r.Group("/api"), m, h)

	if cfg.ApiConfig.EnableSpec {
		docs.Serve(r)
	}

	serveUI(r)

	if err := r.Run(fmt.Sprintf(":%v", cfg.ApiConfig.Port)); err != nil {
		logger.LogError(err)
	}
	// [TODO] implement server shutdown using context
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

	return r
}

func serveUI(r *gin.Engine) {
	uiFS, err := fs.Sub(web.UI, "static")
	if err != nil {
		return
	}
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.JSON(http.StatusNotFound, gin.H{"success": false})
			return
		}
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if path == "api" || strings.HasPrefix(path, "api/") {
			c.JSON(http.StatusNotFound, gin.H{"success": false})
			return
		}
		if path != "" {
			if f, err := uiFS.Open(path); err == nil {
				f.Close()
				c.FileFromFS(path, http.FS(uiFS))
				return
			}
		}
		c.FileFromFS("index.html", http.FS(uiFS))
	})
}
