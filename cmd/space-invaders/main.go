//go:generate bash -c "../../src/build.sh -d \"../../dist\""
//make:deploy deployment-with-ingress, port:8080, subdomain:space-invaders
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	gzip "github.com/gin-contrib/gzip"
	gin "github.com/gin-gonic/gin"
	dist "github.com/sarumaj/edu-space-invaders/dist"
)

var port = flag.Int("port", 8080, "port to listen on")
var debug = flag.Bool("debug", false, "enable debug mode")

// cacheControl is a middleware that sets the Cache-Control header.
// It uses the ETag from the dist package to determine if the file has changed.
func cacheControl(ctx *gin.Context) {
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
	gin.SetMode(gin.ReleaseMode)

	// Set the mode based on the environment variable.
	// Can be used in the future to set the mode based on the command line argument
	// and alternate the game environment.
	_ = os.Setenv("SPACE_INVADERS_MODE", "PRODUCTION")
	if *debug {
		_ = os.Setenv("SPACE_INVADERS_MODE", "DEVELOPMENT")
		gin.SetMode(gin.DebugMode)
	}

	server := &http.Server{Addr: fmt.Sprintf(":%d", *port)}

	router := gin.New()
	router.Use(gin.Logger(), gzip.Gzip(gzip.BestCompression), cacheControl)
	router.StaticFS("/", dist.HttpFS)
	router.POST("/.env", handleEnvVars)

	server.Handler = router
	log.Fatal(server.ListenAndServe())
}
