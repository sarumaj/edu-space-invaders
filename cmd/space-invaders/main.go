//go:generate bash -c "../../src/build.sh --directory \"../../dist\""
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
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

	aesKey      = flag.String("aes-key", "aes_key.pem", "path to the AES key to encrypt and decrypt JWT tokens")
	databaseURL = flag.String("database-url", getenv("DATABASE_URL", "postgres://postgres:pass@db:5432/postgres"), "database address")
	port        = flag.Uint("port", getenv[uint]("PORT", 8080), "port to listen on")
	forceSecure = flag.Bool("force-secure", getenv("FORCE_SECURE", false), "force secure connection over HTTPS")
	limitRPS    = flag.Float64("limit-rps", 60, "requests per second for rate limiting")
	limitBurst  = flag.Uint("limit-burst", 10, "burst size for rate limiting")
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
		_ = scoreBoardDatabase.AutoMigrate(&Metric{})
		_ = scoreBoardDatabase.AutoMigrate(&Score{})
	}

	// Configure router.
	skipper := func(c *gin.Context) bool {
		switch {
		case
			c.Request.Method == http.MethodGet && c.Request.URL.Path == "/health",
			c.Request.Method == http.MethodGet && c.Request.URL.Path == "/.env":

			return true
		}

		return false
	}

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
		logger.Fatal("Failed to read public key", zap.Error(err))
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(raw)
	if err != nil {
		logger.Fatal("Failed to parse private key", zap.Error(err))
	}

	raw, err = os.ReadFile(*aesKey)
	if err != nil {
		logger.Fatal("Failed to read encryption key", zap.Error(err))
	}

	cryptKey, err := ParseAES2GCMKeyFromPem(raw)
	if err != nil {
		logger.Fatal("Failed to parse encryption key", zap.Error(err))
	}

	logger.Info("RSA key loaded")
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

// DecryptWithAES decrypts the encrypted message using the AES key.
// It returns the decrypted message or an error if decryption fails.
// The encrypted message is expected to be a hex-encoded string.
func DecryptWithAES(keyCipher cipher.AEAD, encryptedMessage string) (string, error) {
	ciphertext, err := hex.DecodeString(encryptedMessage)
	if err != nil {
		return "", err
	}

	nonceSize := keyCipher.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := keyCipher.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptWithAES encrypts the plaintext using the AES key.
// It returns the encrypted message as a hex-encoded string or an error if encryption fails.
// The nonce is prepended to the ciphertext.
func EncryptWithAES(keyCipher cipher.AEAD, plaintext string) (string, error) {
	nonce := make([]byte, keyCipher.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := keyCipher.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// ParseAES2GCMKeyFromPem parses the AES key from the PEM-encoded data.
// It returns the AES GCM cipher or an error if parsing fails.
// The PEM-encoded data is expected to contain the AES key.
func ParseAES2GCMKeyFromPem(raw []byte) (cipher.AEAD, error) {
	decoded, _ := pem.Decode(raw)
	if decoded == nil || decoded.Type != "AES PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode encryption key")
	}

	block, err := aes.NewCipher(decoded.Bytes)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}
