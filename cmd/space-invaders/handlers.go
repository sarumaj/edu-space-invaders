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
	zapcore "go.uber.org/zap/zapcore"
)

// CacheControlMiddleware is a middleware that sets the cache control headers.
// It also handles the ETag header.
func CacheControlMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := strings.Trim(ctx.Request.URL.Path, "/")
		if path == "/" {
			path = defaultEndpoint
		}

		fields := []zapcore.Field{{Key: "path", Type: zapcore.StringType, String: path}}

		if eTag, ok := dist.LookupHash(path); ok {
			fields = append(fields, zapcore.Field{Key: "eTag", Type: zapcore.StringType, String: eTag})
			ctx.Header("Cache-Control", "public, must-revalidate")
			ctx.Header("ETag", eTag)

			if strings.Contains(ctx.GetHeader("If-None-Match"), eTag) {
				logger.Info("ETag matched", fields...)
				ctx.Status(http.StatusNotModified)
				return
			}
		}

		logger.Info("ETag not matched", fields...)
		ctx.Next()
	}
}

// HandleEnv handles the environment variables.
// It sets the environment variables based on the request body.
// It returns the environment variables as a response.
func HandleEnv() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var body gin.H
		if err := ctx.ShouldBind(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set the environment variables based on the request body.
		// Communication from WASM to Go is done through the request body.
		var fields []zapcore.Field
		for k, v := range body {
			if !strings.HasPrefix(k, envVarPrefix) {
				continue
			}

			if v == nil {
				_ = os.Unsetenv(k)
				fields = append(fields, zapcore.Field{Key: k, Type: zapcore.StringType, String: "unset"})
			} else {
				_ = os.Setenv(k, fmt.Sprintf("%v", v))
				fields = append(fields, zapcore.Field{Key: k, Type: zapcore.StringType, String: fmt.Sprintf("%v", v)})
			}
		}

		// Update the environment variables if they were altered.
		if len(fields) > 0 {
			logger.Info("Environment variables updated", fields...)
			environ = os.Environ()
		} else {
			logger.Info("No environment variables updated")
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
				logger.Info("Matched conflicting path", zapcore.Field{Key: "path", Type: zapcore.StringType, String: path})
				handler(ctx)
				return
			}
		}

		logger.Info("Serving file", zapcore.Field{Key: "path", Type: zapcore.StringType, String: path})
		ctx.FileFromFS("/"+strings.TrimLeft(path, "/"), dist.HttpFS)
	}
}
