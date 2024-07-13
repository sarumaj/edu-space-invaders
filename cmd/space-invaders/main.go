//go:generate bash -c "../../src/build.sh -d \"../../dist\""
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
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	zapcore "go.uber.org/zap/zapcore"
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

// main is the entry point of the game server.
func main() {
	flag.Parse()

	// Set the port based on the environment variable (necessary for Heroku).
	_ = os.Setenv("RK_GIN_0_PORT", fmt.Sprint(*port))

	envData := os.Environ()
	slices.Sort(envData)

	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(bootRaw))

	entry := rkgin.GetGinEntry("space-invaders")

	entry.LoggerEntry.Info("Booting up", zapcore.Field{Key: "environ", Interface: envData, Type: zapcore.ReflectType})

	entry.Router.Use(CacheControlMiddleware())
	entry.Router.POST("/.env", HandleEnv())
	entry.Router.Match([]string{http.MethodHead, http.MethodGet}, "/*filepath", ServerFileSystem(map[*regexp.Regexp]gin.HandlerFunc{
		regexp.MustCompile(`^/?health/?$`): HandleHealth(),
	}))

	boot.Bootstrap(context.Background())

	boot.WaitForShutdownSig(context.Background())
	entry.LoggerEntry.Warn("Unexpected shut down")
}
