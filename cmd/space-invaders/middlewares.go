package main

import (
	"net/http"
	"net/url"
	"strings"

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

// HttpsRedirectMiddleware redirects HTTP requests to HTTPS
func HttpsRedirectMiddleware(enabled bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if enabled && ctx.Request.URL.Scheme != "https" && ctx.Request.Header.Get("X-Forwarded-Proto") != "https" {
			location := &url.URL{
				Scheme:      "https",
				Host:        ctx.Request.Header.Get("X-Forwarded-Host"),
				Path:        ctx.Request.URL.Path,
				RawPath:     ctx.Request.URL.RawPath,
				RawQuery:    ctx.Request.URL.RawQuery,
				ForceQuery:  false,
				User:        ctx.Request.URL.User,
				Fragment:    ctx.Request.URL.Fragment,
				RawFragment: ctx.Request.URL.RawFragment,
				Opaque:      ctx.Request.URL.Opaque,
				OmitHost:    false,
			}

			if location.Host == "" {
				location.Host = ctx.Request.Host
			}

			ctx.Redirect(http.StatusMovedPermanently, location.String())
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
