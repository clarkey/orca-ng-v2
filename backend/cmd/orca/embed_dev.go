//go:build dev
// +build dev

package main

import (
	"github.com/gin-gonic/gin"
)

func setupStaticFiles(router *gin.Engine) {
	// In development, static files are served by the frontend dev server
	// This is just a no-op function to satisfy the build
}