//go:build !dev
// +build !dev

package main

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//go:embed all:dist
var staticFiles embed.FS

// isAPIPath checks if a path is an API path
func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/api/") || path == "/health"
}

func setupStaticFiles(router *gin.Engine) {
	staticFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		logrus.WithError(err).Warn("Failed to setup static files, frontend may not be available")
		return
	}
	
	router.NoRoute(func(c *gin.Context) {
		// For SPA routing, serve index.html for all non-API routes
		if c.Request.URL.Path == "/" || !isAPIPath(c.Request.URL.Path) {
			c.FileFromFS("/", http.FS(staticFS))
		} else {
			c.FileFromFS(c.Request.URL.Path, http.FS(staticFS))
		}
	})
}