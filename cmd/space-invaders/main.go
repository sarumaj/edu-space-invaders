//go:generate bash -c "../../src/build.sh --directory \"../../dist\" --api-key $(../../src/jwt.sh --private-key \"../../secret/private_key.pem\" --ttl 0 --subject \"wasm\")"
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	gzip "github.com/gin-contrib/gzip"
	ginzap "github.com/gin-contrib/zap"
	gin "github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
	postgres "gorm.io/driver/postgres"
	gorm "gorm.io/gorm"
)

const (
	defaultEndpoint = "index.html"      // defaultEndpoint is the default endpoint.
	envVarPrefix    = "SPACE_INVADERS_" // envVarPrefix is the prefix for the environment variables.
)

var (
	environ []string
	logger  *zap.Logger

	databaseURL = flag.String("database-url", getenv("SPACE_INVADERS_DATABASE_URL", "postgres://postgres:pass@db:5432/postgres"), "database address")
	port        = flag.Uint("port", getenv[uint]("PORT", 8080), "port to listen on")
	forceSecure = flag.Bool("force-secure", getenv("SPACE_INVADERS_FORCE_SECURE", false), "force secure connection over HTTPS")
	publicKey   = flag.String("public-key", "public_key.pem", "path to the public key to verify JWT tokens")
	limitRPS    = flag.Float64("limit-rps", 60, "requests per second for rate limiting")
	limitBurst  = flag.Uint("limit-burst", 10, "burst size for rate limiting")
)

func init() {
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)

	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	enc := zapcore.NewConsoleEncoder(cfg)
	logger = zap.New(zapcore.NewTee(
		zapcore.NewCore(enc, zapcore.Lock(os.Stdout), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl < zapcore.ErrorLevel })),
		zapcore.NewCore(enc, zapcore.Lock(os.Stderr), zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl >= zapcore.ErrorLevel })),
	))
}

// main is the entry point of the game server.
func main() {
	defer func() { _ = logger.Sync() }()

	// Log the server start.
	logger.Info("Starting server",
		zap.Uintp("port", port),
		zap.Boolp("forceSecure", forceSecure),
		zap.Stringp("databaseURL", databaseURL),
		zap.Stringp("public_key", publicKey))

	// Load the environment variables.
	environ = os.Environ()
	slices.Sort(environ)
	logger.Info("Environment variables", zap.Strings("environ", environ))

	// Connect to the database.
	dsn, err := parsePostgresURL(*databaseURL)
	if err != nil {
		logger.Fatal("Failed to parse database URL", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
	}

	// Connect to the database.
	scoreBoardDatabase, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Failed to connect to database", zapcore.Field{Key: "error", Interface: err, Type: zapcore.ErrorType})
	}

	// Migrate the database.
	if !scoreBoardDatabase.DryRun {
		_ = scoreBoardDatabase.AutoMigrate(&Metric{})
		_ = scoreBoardDatabase.AutoMigrate(&Score{})
	}

	// Register the routes.
	router := gin.New(func(e *gin.Engine) {
		e.Use(ginzap.Ginzap(logger, time.RFC3339, true))
		e.Use(ginzap.RecoveryWithZap(logger, true))
	})

	// Register the asymmetric JWT validator.
	raw, err := os.ReadFile(*publicKey)
	if err != nil {
		logger.Fatal("Failed to read public key", zap.Error(err))
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(raw)
	if err != nil {
		logger.Fatal("Failed to parse public key", zap.Error(err))
	}

	logger.Info("Public RSA key loaded")
	jwtAuthenticator := AuthenticatorMiddleware(key, map[string]string{"header": "Authorization", "query": "token"})

	// Register the routes.
	router.Use(
		gzip.Gzip(gzip.BestCompression),
		MetricsMiddleware(scoreBoardDatabase),
		HttpsRedirectMiddleware(*forceSecure),
		CacheControlMiddleware(),
		LimitMiddleware(*limitRPS, *limitBurst),
	)

	router.POST("/.env", jwtAuthenticator, HandleEnv())
	router.POST("/scores.db", jwtAuthenticator, SaveScores(scoreBoardDatabase))
	router.Match([]string{http.MethodHead, http.MethodGet}, "/*filepath", ServeFileSystem(map[*regexp.Regexp]gin.HandlersChain{
		regexp.MustCompile(`^/?health/?$`):     {HandleHealth(scoreBoardDatabase)},
		regexp.MustCompile(`^/?\.env/?$`):      {jwtAuthenticator, HandleEnv()},
		regexp.MustCompile(`^/?scores\.db/?$`): {GetScores(scoreBoardDatabase)},
	}))

	if err := router.Run(fmt.Sprintf(":%d", *port)); err != nil {
		logger.Fatal("Unexpected server error", zap.Error(err))
	}
}

// getenv returns the value of the environment variable with the given key.
func getenv[T any](key string, fallback T) (out T) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}

	target := reflect.ValueOf(&out).Elem()

	switch target.Kind() {
	case reflect.String:
		target.SetString(raw)
		return

	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err == nil {
			target.SetBool(parsed)
		}
		return

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err == nil {
			target.SetInt(parsed)
		}
		return

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err == nil {
			target.SetUint(parsed)
		}
		return

	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(raw, 64)
		if err == nil {
			target.SetFloat(parsed)
		}
		return

	}

	return fallback
}

// parsePostgresURL parses the database URL and returns the DSN.
func parsePostgresURL(databaseUrl string) (string, error) {
	out := map[string]string{
		"host":     "localhost",
		"port":     "5432",
		"dbname":   "postgres",
		"user":     "postgres",
		"password": "pass",
		"sslmode":  "disable",
		"timezone": "Europe/Berlin",
	}

	// Parse the database URL
	databaseAddress, err := url.Parse(databaseUrl)
	if err != nil {
		return "", err
	}

	// Helper function to update the `out` map
	write := func(key string, value string, force bool) {
		if value != "" || force {
			out[key] = value
		}
	}

	// Write the host, port, and dbname
	write("host", databaseAddress.Hostname(), false)
	write("port", databaseAddress.Port(), false)
	write("dbname", strings.TrimPrefix(databaseAddress.Path, "/"), false)

	// Handle user credentials
	if databaseAddress.User != nil {
		write("user", databaseAddress.User.Username(), false)
		password, ok := databaseAddress.User.Password()
		write("password", password, ok)
	}

	// Handle query parameters (e.g., sslmode)
	for key, value := range databaseAddress.Query() {
		switch key {
		case "host", "port", "dbname", "user", "password":

		default:
			write(key, value[0], true)
		}
	}

	var parts []string
	for key, value := range out {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	slices.Sort(parts)
	return strings.Join(parts, " "), nil
}
