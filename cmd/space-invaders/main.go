//go:generate bash -c "../../src/build.sh --directory \"../../dist\""
package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	zapcore "go.uber.org/zap/zapcore"
)

const (
	defaultEndpoint = "index.html"      // defaultEndpoint is the default endpoint.
	envVarPrefix    = "SPACE_INVADERS_" // envVarPrefix is the prefix for the environment variables.
	ginEntryName    = "space-invaders"  // ginEntryName is the name of the Gin entry.
)

var (
	//go:embed boot.yaml
	bootRaw []byte

	environ []string
	logger  *rkentry.LoggerEntry

	port = flag.Int("port", func() int {
		parsed, err := strconv.Atoi(os.Getenv("PORT"))
		if err == nil {
			return parsed
		}
		return 8080
	}(), "port to listen on")
)

// main is the entry point of the game server.
func main() {
	flag.Parse()

	// Set the port based on the environment variable (necessary for Heroku).
	_ = os.Setenv("RK_GIN_0_PORT", fmt.Sprint(*port))

	environ = os.Environ()
	slices.Sort(environ)

	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(bootRaw))

	entry := rkgin.GetGinEntry(ginEntryName)
	logger = entry.LoggerEntry

	logger.Info("Booting up", zapcore.Field{Key: "environ", Interface: environ, Type: zapcore.ReflectType})

	entry.Router.Use(CacheControlMiddleware())
	entry.Router.POST("/.env", HandleEnv())
	entry.Router.Match([]string{http.MethodHead, http.MethodGet}, "/*filepath", ServerFileSystem(map[*regexp.Regexp]gin.HandlerFunc{
		regexp.MustCompile(`^/?health/?$`): HandleHealth(),
	}))

	boot.Bootstrap(context.Background())

	boot.WaitForShutdownSig(context.Background())
	logger.Warn("Unexpected shut down")
}
