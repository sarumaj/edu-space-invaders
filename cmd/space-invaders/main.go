//go:generate bash -c "../../src/build.sh --directory \"../../dist\" --api-key $(../../src/jwt.sh --private-key \"../../secret/private_key.pem\" --ttl 0 --subject \"wasm\")"
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
	rksqlite "github.com/rookie-ninja/rk-db/sqlite"
	rkentry "github.com/rookie-ninja/rk-entry/v2/entry"
	rkmidjwt "github.com/rookie-ninja/rk-entry/v2/middleware/jwt"
	rkgin "github.com/rookie-ninja/rk-gin/v2/boot"
	rkginjwt "github.com/rookie-ninja/rk-gin/v2/middleware/jwt"
	zapcore "go.uber.org/zap/zapcore"
)

const (
	defaultEndpoint        = "index.html"      // defaultEndpoint is the default endpoint.
	envVarPrefix           = "SPACE_INVADERS_" // envVarPrefix is the prefix for the environment variables.
	ginEntryName           = "space-invaders"  // ginEntryName is the name of the Gin entry.
	scoreBoardDatabaseName = "space-invaders"  // scoreBoardDatabaseName is the name of the score board database.
	sqliteEntryName        = "space-invaders"  // sqliteEntryName is the name of the SQLite entry.
)

var (
	//go:embed boot.yaml
	bootRaw []byte

	environ []string
	logger  *rkentry.LoggerEntry

	port        = flag.Int("port", parsePort(), "port to listen on")
	forceSecure = flag.Bool("force-secure", os.Getenv("SPACE_INVADERS_FORCE_SECURE") == "true", "force secure connection")
	private_key = flag.String("private-key", "private_key.pem", "path to the private key")
	public_key  = flag.String("public-key", "public_key.pem", "path to the public key")
)

// main is the entry point of the game server.
func main() {
	flag.Parse()

	// Set the port based on the environment variable (necessary for Heroku).
	_ = os.Setenv("RK_GIN_0_PORT", fmt.Sprint(*port))

	// Load the environment variables.
	environ = os.Environ()
	slices.Sort(environ)

	// Bootstrap the application.
	boot := rkboot.NewBoot(rkboot.WithBootConfigRaw(bootRaw))
	boot.Bootstrap(context.Background())

	// Get the Gin entry and logger.
	ginEntry := rkgin.GetGinEntry(ginEntryName)
	logger = ginEntry.LoggerEntry

	logger.Info("Booting up", zapcore.Field{Key: "environ", Interface: environ, Type: zapcore.ReflectType})

	// Get the SQLite entry and database.
	sqliteEntry := rksqlite.GetSqliteEntry(sqliteEntryName)
	scoreBoardDatabase := sqliteEntry.GetDB(scoreBoardDatabaseName)
	if !scoreBoardDatabase.DryRun {
		_ = scoreBoardDatabase.AutoMigrate(&Score{})
	}

	privKey, err := os.ReadFile(*private_key)
	if err != nil {
		logger.Error("Failed to read private key", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
	}

	pubKey, err := os.ReadFile(*public_key)
	if err != nil {
		logger.Error("Failed to read public key", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
	}

	// Register the asymmetric JWT signer.
	signer := rkentry.RegisterAsymmetricJwtSigner(ginEntryName, "RS256", privKey, pubKey)

	// Register the routes.
	jwtAuthenticator := rkginjwt.Middleware(rkmidjwt.WithSigner(signer))
	ginEntry.Router.Use(HttpsRedirectMiddleware(*forceSecure), CacheControlMiddleware())
	ginEntry.Router.POST("/.env", jwtAuthenticator, HandleEnv())
	ginEntry.Router.POST("/scores", jwtAuthenticator, SaveScores(scoreBoardDatabase))
	ginEntry.Router.Match([]string{http.MethodHead, http.MethodGet}, "/*filepath", ServeFileSystem(map[*regexp.Regexp]gin.HandlersChain{
		regexp.MustCompile(`^/?health/?$`): {HandleHealth()},
		regexp.MustCompile(`^/?\.env/?$`):  {HandleEnv()},
		regexp.MustCompile(`^/?scores/?$`): {GetScores(scoreBoardDatabase)},
	}))

	// Start the server.
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
