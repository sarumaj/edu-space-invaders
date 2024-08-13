package main

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	gin "github.com/gin-gonic/gin"
	dist "github.com/sarumaj/edu-space-invaders/dist"
	zapcore "go.uber.org/zap/zapcore"
	gorm "gorm.io/gorm"
	clause "gorm.io/gorm/clause"
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

// MetricsMiddleware is a middleware that logs metrics.
func MetricsMiddleware[I interface{ ~int | ~uint }](metricsDatabase *gorm.DB, queueSize I) gin.HandlerFunc {
	queue := make([]MetricsEntry, 0, int(queueSize))
	var queueLock sync.Mutex

	return func(ctx *gin.Context) {
		ctx.Next()

		entry := MetricsEntry{
			Endpoint: ctx.Request.URL.Path,
			Method:   ctx.Request.Method,
			Count:    1,
		}

		queueLock.Lock()
		queue = append(queue, entry)
		queueLock.Unlock()

		if len(queue) < cap(queue) {
			return
		}

		if err := metricsDatabase.
			Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "endpoint"}, {Name: "method"}},
				DoUpdates: clause.Assignments(map[string]any{
					"count":      gorm.Expr("count + ?", 1), // Increment the count.
					"updated_at": gorm.Expr("?", time.Now()),
				}),
				Where: clause.Where{Exprs: []clause.Expression{
					gorm.Expr("excluded.endpoint = endpoint"),
					gorm.Expr("excluded.method = method"),
				}},
			}).
			Create(queue).
			Error; err == nil {

			queue = queue[:0:cap(queue)]
		} else {
			logger.Error("Failed to save metrics", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
		}
	}
}
