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

	gin "github.com/gin-gonic/gin"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkmidjwt "github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkginjwt "github.com/rookie-ninja/rk-gin/v2/middleware/jwt"
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
	signer  rkentry.SignerJwt

	port        = flag.Int("port", parsePort(), "port to listen on")
	private_key = flag.String("private-key", "private_key.pem", "path to the private key")
	public_key  = flag.String("public-key", "public_key.pem", "path to the public key")
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

	privKey, err := os.ReadFile(*private_key)
	if err != nil {
		logger.Error("Failed to read private key", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
	}

	pubKey, err := os.ReadFile(*public_key)
	if err != nil {
		logger.Error("Failed to read public key", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
	}

	// Register the asymmetric JWT signer.
	signer = rkentry.RegisterAsymmetricJwtSigner(ginEntryName, "RS256", privKey, pubKey)

	entry.Router.Use(CacheControlMiddleware())
	entry.Router.POST("/.env", rkginjwt.Middleware(rkmidjwt.WithSigner(signer)), HandleEnv())
	entry.Router.Match([]string{http.MethodHead, http.MethodGet}, "/*filepath", ServerFileSystem(map[*regexp.Regexp]gin.HandlerFunc{
		regexp.MustCompile(`^/?health/?$`): HandleHealth(),
		regexp.MustCompile(`^/?\.env/?$`):  HandleEnv(),
	}))

	boot.Bootstrap(context.Background())

	boot.WaitForShutdownSig(context.Background())
	logger.Warn("Unexpected shut down")
}

// parsePort parses the port from the environment variable.
// Heroku sets the port in the environment variable.
func parsePort() int {
	raw := os.Getenv("PORT")
	if raw == "" {
		return 8080
	}

	parsed, err := strconv.Atoi(raw)
	if err == nil {
		return parsed
	}

	return 8080
}
