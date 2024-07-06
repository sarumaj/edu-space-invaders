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

	gin "github.com/gin-gonic/gin"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	dist "github.com/sarumaj/edu-space-invaders/dist"
)

//go:embed boot.yaml
var bootRaw []byte

var debug = flag.Bool("debug", false, "enable debug mode")
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

	body["SPACE_INVADERS_MODE"] = os.Getenv("SPACE_INVADERS_MODE")
	for k, v := range body {
		_ = os.Setenv(k, fmt.Sprintf("%v", v))
	}

	ctx.JSON(http.StatusOK, body)
}

// main is the entry point of the game server.
func main() {
	flag.Parse()

	// Set the mode based on the environment variable.
	// Can be used in the future to set the mode based on the command line argument
	// and alternate the game environment.
	_ = os.Setenv("SPACE_INVADERS_MODE", map[bool]string{true: "DEVELOPMENT", false: "PRODUCTION"}[*debug])

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
