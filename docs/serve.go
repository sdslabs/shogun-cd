package docs

import (
	"net/http"

	_ "embed"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.yaml
var spec []byte

func Serve(r *gin.Engine) {
	d := r.Group("/docs")

	d.GET("/spec", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", spec)
	})

	d.GET("/", func(c *gin.Context) {
		html := getScalarHTML("/docs/spec")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})
}

func getScalarHTML(specUrl string) string {
	return `
        <!doctype html>
        <html>
          <head>
            <title>Shogun-CD API Reference</title>
            <meta charset="utf-8" />
            <meta name="viewport" content="width=device-width, initial-scale=1" />
            <style>
              body { margin: 0; }
            </style>
          </head>
          <body>
            <script
              id="api-reference"
              data-url="` + specUrl + `">
            </script>
            <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
          </body>
        </html>
    `
}
