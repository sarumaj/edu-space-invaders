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
	config "github.com/sarumaj/edu-space-invaders/src/pkg/config"
	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
	gorm "gorm.io/gorm"
)

// GetConfig returns the configuration as a response.
func GetConfig() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.TOML(http.StatusOK, config.Config)
	}
}

// GetScores returns the scores as a response.
// It returns the scores in descending order of the score.
// If the scores have the same score, they are ordered in ascending order of the name.
func GetScores(database *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		scores, err := Helper(database).GetScores()
		if err != nil {
			logger.Error("Failed to get scores", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Debug("Scores retrieved", zap.Any("scores", scores))
		ctx.SecureJSON(http.StatusOK, scores)
	}
}

// HandleEnv handles the environment variables.
// It sets the environment variables based on the request body.
// It returns the environment variables as a response.
func HandleEnv() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		body := make(gin.H, 0)
		if ctx.Request.Method == http.MethodPost {
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
					fields = append(fields, zap.String(k, "unset"))
				} else {
					_ = os.Setenv(k, fmt.Sprintf("%v", v))
					fields = append(fields, zap.String(k, fmt.Sprintf("%v", v)))
				}
			}

			// Update the environment variables if they were altered.
			if len(fields) > 0 {
				logger.Debug("Environment variables updated", fields...)
				environ = os.Environ()
			} else {
				logger.Debug("No environment variables updated")
			}
		}

		// Return the environment variables as a response.
		// Communication from Go to WASM is done through the response body.
		for _, pair := range environ {
			k, v, _ := strings.Cut(pair, "=")
			if strings.HasPrefix(k, envVarPrefix) {
				body[k] = v
			}
		}

		body["_size"] = len(body)
		body["_prefix"] = envVarPrefix
		ctx.JSON(http.StatusOK, body)
	}
}

// HandleHealth handles the health check.
// It returns the boot time, build time, current time, status, and uptime as a response.
// It also returns the database size, database utilization, and table sizes as metrics.
// If the database size exceeds the threshold, it deletes the oldest scores.
// The database size is considered to have exceeded the threshold if it is greater than or equal to 1 GB.
// The threshold is approximately 93% of the maximum size, which is 1 GiB.
// The top 100 scores are kept in the database.
func HandleHealth(database *gorm.DB) gin.HandlerFunc {
	bootTime := time.Now()

	return func(ctx *gin.Context) {
		var metrics []Metric
		if err := database.Omit("created_at", "updated_at").Find(&metrics).Error; err != nil {
			logger.Error("Failed to get metrics", zap.Error(err))
			return
		}

		metricsObject := make(map[string]any)
		for _, metric := range metrics {
			metricsObject[metric.Method+" "+metric.Endpoint] = metric.Count
		}

		size, err := Helper(database).GetDatabaseSize()
		if err != nil {
			logger.Error("Failed to get database size", zap.Error(err))
			return
		}

		if size >= sizeThreshold {
			if err := Helper(database).ClearMetrics(10); err != nil {
				ctx.AbortWithStatusJSON(http.StatusServiceUnavailable,
					gin.H{"error": fmt.Sprintf("database unavailable, size exceeded: %s, err: %s", Size(size), err.Error())})
				return
			}

			if err := Helper(database).ClearScores(10); err != nil {
				ctx.AbortWithStatusJSON(http.StatusServiceUnavailable,
					gin.H{"error": fmt.Sprintf("database unavailable, size exceeded: %s, err: %s", Size(size), err.Error())})
				return
			}
		}

		metricsObject["STATS /database/limit"] = Size(maximumSize).String()
		metricsObject["STATS /database/size"] = size.String()
		metricsObject["STATS /database/utilization"] = fmt.Sprintf("%.2f%%", float64(size)/float64(maximumSize)*100)

		tables, err := Helper(database).GetTableSizes()
		if err != nil {
			logger.Error("Failed to get table sizes", zap.Error(err))
		}

		for table, tableSize := range tables {
			metricsObject["STATS /database/tables/"+table+"/size"] = tableSize.String()
			metricsObject["STATS /database/tables/"+table+"/utilization"] = fmt.Sprintf("%.2f%%", float64(tableSize)/float64(maximumSize)*100)
		}

		ctx.JSON(http.StatusOK, gin.H{
			"BootTime":  bootTime.Format(time.RFC3339),
			"BuildTime": dist.BuildTime(),
			"Current":   time.Now().Format(time.RFC3339),
			"Metrics":   metricsObject,
			"Status":    "ok",
			"UpTime":    time.Since(bootTime).String(),
		})
	}
}

// ServeFileSystem serves the files from the embedded file system.
func ServeFileSystem(conflicting map[*regexp.Regexp]gin.HandlersChain) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Param("filepath")
		for pattern, handlers := range conflicting {
			if pattern.MatchString(path) && len(handlers) > 0 {
				logger.Debug("Matched conflicting path", zap.String("path", path))
				// Execute the middleware handlers.
				for _, handler := range handlers[:len(handlers)-1] {
					if handler(ctx); ctx.IsAborted() {
						return
					}
				}
				// Execute the last handler.
				handlers.Last()(ctx)
				return
			}
		}

		logger.Debug("Serving file", zap.String("path", path))
		ctx.FileFromFS("/"+strings.TrimLeft(path, "/"), dist.HttpFS)
	}
}

// SaveScores saves the scores to the database.
func SaveScores(database *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var scores []Score
		if err := ctx.ShouldBind(&scores); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := Helper(database).SaveScores(scores); err != nil {
			logger.Error("Failed to save scores", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Debug("Scores saved", zap.Any("scores", scores))
		ctx.Status(http.StatusOK)
	}
}
