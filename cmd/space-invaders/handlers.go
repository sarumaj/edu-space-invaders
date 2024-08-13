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
	gorm "gorm.io/gorm"
	clause "gorm.io/gorm/clause"
)

// GetScores returns the scores as a response.
// It returns the scores in descending order of the score.
// If the scores have the same score, they are ordered in ascending order of the name.
func GetScores(scoreBoardDatabase *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		scores := make([]Score, 0)
		if err := scoreBoardDatabase.
			Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).
			Order(clause.OrderBy{
				Columns: []clause.OrderByColumn{
					{Column: clause.Column{Name: "score"}, Desc: true},
					{Column: clause.Column{Name: "CASE WHEN updated_at > created_at THEN updated_at ELSE created_at END", Raw: true}},
				},
			}).
			Find(&scores).
			Error; err != nil {

			logger.Error("Failed to get scores", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Scores retrieved", zapcore.Field{Key: "scores", Interface: scores, Type: zapcore.ReflectType})
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
func HandleHealth(metricsDatabase *gorm.DB) gin.HandlerFunc {
	bootTime := time.Now()

	return func(ctx *gin.Context) {
		var metrics []MetricsEntry
		if err := metricsDatabase.
			Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).
			Select("endpoint", "method", "count").
			Find(&metrics).
			Error; err != nil {

			logger.Error("Failed to get metrics", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
			return
		}

		if ctx.Request.Method == http.MethodHead {
			ctx.Status(http.StatusOK)
			return
		}

		metricsObject := make(map[string]int)
		for _, metric := range metrics {
			switch metric.Endpoint {
			case "/health":

			default:
				metricsObject[metric.Method+" "+metric.Endpoint] = metric.Count

			}
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
			if pattern.MatchString(path) {
				logger.Info("Matched conflicting path", zapcore.Field{Key: "path", Type: zapcore.StringType, String: path})
				for _, handler := range handlers {
					handler(ctx)
				}
				return
			}
		}

		logger.Info("Serving file", zapcore.Field{Key: "path", Type: zapcore.StringType, String: path})
		ctx.FileFromFS("/"+strings.TrimLeft(path, "/"), dist.HttpFS)
	}
}

// SaveScores saves the scores to the database.
func SaveScores(scoreBoardDatabase *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var scores []Score
		if err := ctx.ShouldBind(&scores); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := scoreBoardDatabase.
			Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "name"}},
				DoUpdates: clause.Assignments(map[string]any{
					"score":      gorm.Expr("excluded.score"),
					"updated_at": gorm.Expr("?", time.Now()),
				}),
				Where: clause.Where{Exprs: []clause.Expression{gorm.Expr("excluded.score > score")}},
			}).
			Create(&scores).
			Error; err != nil {

			logger.Error("Failed to save scores", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Scores saved", zapcore.Field{Key: "scores", Interface: scores, Type: zapcore.ReflectType})
		ctx.Status(http.StatusOK)
	}
}
