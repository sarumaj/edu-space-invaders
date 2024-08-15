//go:generate bash -c "../../src/build.sh --directory \"../../dist\""
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
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

	aesKey      = flag.String("aes-key", "aes_key.pem", "path to the AES key to encrypt and decrypt JWT tokens")
	databaseURL = flag.String("database-url", getenv("DATABASE_URL", "postgres://postgres:pass@db:5432/postgres"), "database address")
	port        = flag.Uint("port", getenv[uint]("PORT", 8080), "port to listen on")
	forceSecure = flag.Bool("force-secure", getenv("FORCE_SECURE", false), "force secure connection over HTTPS")
	limitRPS    = flag.Float64("limit-rps", 90, "requests per second for rate limiting")
	limitBurst  = flag.Uint("limit-burst", 12, "burst size for rate limiting")
	rsaKey      = flag.String("rsa-key", "rsa_key.pem", "path to the RSA key to sign and verify JWT tokens")
)

// init runs the initialization code.
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
		zap.Stringp("aesKey", aesKey),
		zap.Stringp("rsaKey", rsaKey),
		zap.Float64p("limitRPS", limitRPS),
		zap.Uintp("limitBurst", limitBurst),
	)

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
		_ = scoreBoardDatabase.AutoMigrate(&Metric{}, &Score{})
	}

	// Define the skipper function.
	skipper := func(c *gin.Context) bool {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead:
			switch strings.TrimSuffix(c.Request.URL.Path, "/") {
			case "/health", "/.env":
				return true
			}
		}

		return false
	}

	// Configure router.
	router := gin.New(func(e *gin.Engine) {
		e.Use(ginzap.GinzapWithConfig(logger, &ginzap.Config{
			TimeFormat:   time.RFC3339,
			UTC:          true,
			DefaultLevel: zapcore.DebugLevel,
			Skipper:      skipper,
			Context:      func(ctx *gin.Context) []zapcore.Field { return []zapcore.Field{zap.Any("headers", ctx.Request.Header)} },
		}))
		e.Use(ginzap.CustomRecoveryWithZap(logger, true, func(c *gin.Context, err any) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%v", err)})
		}))
	})

	// Register the asymmetric JWT validator.
	raw, err := os.ReadFile(*rsaKey)
	if err != nil {
		logger.Fatal("Failed to read RSA key", zap.Error(err))
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(raw)
	if err != nil {
		logger.Fatal("Failed to parse RSA key", zap.Error(err))
	}

	// Register the symmetric AES cipher.
	raw, err = os.ReadFile(*aesKey)
	if err != nil {
		logger.Fatal("Failed to read AES key", zap.Error(err))
	}

	cryptKey, err := parseAES2GCMKeyFromPem(raw)
	if err != nil {
		logger.Fatal("Failed to parse AES key", zap.Error(err))
	}

	logger.Info("Keys loaded")
	jwtAuthenticator := AuthenticatorMiddleware(&key.PublicKey, cryptKey, map[string]string{
		"header": "Authorization",
		"query":  "token",
		"cookie": "session",
	})

	// Register the routes.
	router.Use(
		SessionMiddleware(key, cryptKey, "session", time.Hour),
		gzip.Gzip(gzip.BestCompression),
		MetricsMiddleware(scoreBoardDatabase, skipper),
		HttpsRedirectMiddleware(*forceSecure),
		CacheControlMiddleware(),
		LimitMiddleware(*limitRPS, *limitBurst, nil),
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
