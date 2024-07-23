package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	gin "github.com/gin-gonic/gin"
	dist "github.com/sarumaj/edu-space-invaders/dist"
)

// CacheControlMiddleware is a middleware that sets the cache control headers.
// It also handles the ETag header.
func CacheControlMiddleware() gin.HandlerFunc {
	defaultEntrypoint := "index.html"

	return func(ctx *gin.Context) {
		path := strings.Trim(ctx.Request.URL.Path, "/")
		if path == "/" {
			path = defaultEntrypoint
		}

		if eTag, ok := dist.LookupHash(path); ok {
			if eTag == ctx.GetHeader("If-None-Match") {
				ctx.Status(http.StatusNotModified)
				return
			}

			ctx.Header("Cache-Control", "public, must-revalidate")
			ctx.Header("ETag", eTag)
		}

		ctx.Next()
	}
}

// HandleEnv handles the environment variables.
// It sets the environment variables based on the request body.
// It returns the environment variables as a response.
func HandleEnv() gin.HandlerFunc {
	environ := os.Environ()

	return func(ctx *gin.Context) {
		body := gin.H{}
		_ = ctx.ShouldBind(&body)

		// Set the environment variables based on the request body.
		// Communication from WASM to Go is done through the request body.
		altered := false
		for k, v := range body {
			if strings.HasPrefix(k, envVarPrefix) {
				_ = os.Setenv(k, fmt.Sprintf("%v", v))
				altered = true
			}
		}

		// Update the environment variables if they were altered.
		if altered {
			environ = os.Environ()
		}

		// Return the environment variables as a response.
		// Communication from Go to WASM is done through the response body.
		for _, pair := range environ {
			k, v, _ := strings.Cut(pair, "=")
			if strings.HasPrefix(k, envVarPrefix) {
				body[k] = v
			}
		}

		ctx.JSON(http.StatusOK, body)
	}
}

func HandleHealth() gin.HandlerFunc {
	bootTime := time.Now()

	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodHead {
			ctx.Status(http.StatusOK)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"BootTime":  bootTime.Format(time.RFC3339),
			"BuildTime": dist.BuildTime(),
			"Current":   time.Now().Format(time.RFC3339),
			"Status":    "ok",
			"UpTime":    time.Since(bootTime).String(),
		})
	}
}

// Redirect redirects the client to the specified location.
func Redirect[L interface {
	~string | func(*gin.Context) string
}](location L) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		switch location := any(location).(type) {
		case string:
			ctx.Redirect(http.StatusMovedPermanently, location)
		case func(*gin.Context) string:
			ctx.Redirect(http.StatusMovedPermanently, location(ctx))
		}
	}
}

// ServerFileSystem serves the files from the embedded file system.
func ServerFileSystem(conflicting map[*regexp.Regexp]gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Param("filepath")
		for pattern, handler := range conflicting {
			if pattern.MatchString(path) {
				handler(ctx)
				return
			}
		}

		ctx.FileFromFS("/"+strings.TrimLeft(path, "/"), dist.HttpFS)
	}
}
