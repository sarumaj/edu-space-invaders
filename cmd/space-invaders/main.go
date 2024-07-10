//go:generate bash -c "../../src/build.sh -d \"../../dist\""
package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gin "github.com/gin-gonic/gin"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	dist "github.com/sarumaj/edu-space-invaders/dist"
)

const envVarPrefix = "SPACE_INVADERS_"

//go:embed boot.yaml
var bootRaw []byte

var port = flag.Int("port", func() int {
	parsed, err := strconv.Atoi(os.Getenv("PORT"))
	if err == nil {
		return parsed
	}
	return 8080
}(), "port to listen on")

// cacheControl is a middleware that sets the Cache-Control header.
// It uses the ETag from the dist package to determine if the file has changed.
func cacheControlMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := filepath.Base(ctx.Request.URL.Path)
		if path == "/" {
			path = "index.html"
		}

		eTag := dist.Hashes[path]
		if eTag != "" {
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

// handleEnvVars is a handler that sets the environment variables.
// It reads the body of the request and sets the environment variables.
// It returns the body of the request as a response.
func handleEnvVars(ctx *gin.Context) {
	body := gin.H{}
	_ = ctx.ShouldBind(&body)

	// Set the environment variables based on the request body.
	// Communication from WASM to Go is done through the request body.
	for k, v := range body {
		if strings.HasPrefix(k, envVarPrefix) {
			_ = os.Setenv(k, fmt.Sprintf("%v", v))
		}
	}

	// Return the environment variables as a response.
	// Communication from Go to WASM is done through the response body.
	for _, pair := range os.Environ() {
		k, v, _ := strings.Cut(pair, "=")
		if strings.HasPrefix(k, envVarPrefix) {
			body[k] = v
		}
	}

	ctx.JSON(http.StatusOK, body)
}

// main is the entry point of the game server.
func main() {
	flag.Parse()

	// Set the port based on the environment variable (necessary for Heroku).
	_ = os.Setenv("RK_GIN_0_PORT", fmt.Sprint(*port))

	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(bootRaw))

	entry := rkgin.GetGinEntry("space-invaders")
	entry.Router.Use(cacheControlMiddleware())

	entry.Router.StaticFS("/", dist.HttpFS)
	entry.Router.POST("/.env", handleEnvVars)

	boot.Bootstrap(context.Background())

	boot.WaitForShutdownSig(context.Background())
}
