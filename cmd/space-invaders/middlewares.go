package main

import (
	"crypto/cipher"
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	gin "github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	dist "github.com/sarumaj/edu-space-invaders/dist"
	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
	rate "golang.org/x/time/rate"
	gorm "gorm.io/gorm"
)

// nopWriter is a writer that does not write anything.
type nopWriter struct{ gin.ResponseWriter }

// Write writes nothing.
func (*nopWriter) Write([]byte) (n int, err error) { return }

// WriteString writes nothing.
func (*nopWriter) WriteString(string) (n int, err error) { return }

// AuthenticatorMiddleware is a middleware that authenticates the request using JWT.
// It uses the public key to verify the JWT token.
// The sources map contains the sources of the JWT token.
// The key is the source and the value is the key.
// The sources can be "cookie", "header", or "query".
// The key is the name of the cookie, header, or query parameter.
// If the token is not found, the middleware will return a 401 status code.
// If the token is invalid, the middleware will return a 401 status code.
func AuthenticatorMiddleware(publicKey *rsa.PublicKey, cryptKey cipher.AEAD, sources map[string]string) gin.HandlerFunc {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}),
		jwt.WithIssuer("space-invaders"),
		jwt.WithAudience("space-invaders"),
		jwt.WithIssuedAt(),
		jwt.WithPaddingAllowed(),
	)

	validatorFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing algorithm: %s", token.Method.Alg())
		}

		return publicKey, nil
	}

	return func(ctx *gin.Context) {
		var jwtToken string
		for source, key := range sources {
			switch source {
			case "cookie":
				if cookie, _ := ctx.Request.Cookie(key); cookie != nil && cookie.Valid() == nil {
					var err error
					jwtToken, err = decodeB64AndDecryptWithAES(cryptKey, cookie.Value)
					if err != nil {
						logger.Error("Failed to decrypt cookie", zap.String("source", source+":"+key), zap.Error(err))
					}
				}

			case "header":
				jwtToken = strings.TrimPrefix(ctx.GetHeader(key), "Bearer ")

			case "query":
				jwtToken = ctx.Query(key)

			default:
				logger.Warn("Unknown source", zap.String("source", source))

			}

			if jwtToken != "" {
				logger.Debug("Found JWT token", zap.String("source", source+":"+key), zap.String("token", jwtToken))
				break
			}
		}

		if jwtToken == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		token, err := parser.Parse(jwtToken, validatorFunc)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx.Set("claims", token.Claims)
		ctx.Next()
	}
}

// BeHeadMiddleware is a middleware that handles the HEAD method.
// It converts the HEAD method to a GET method and writes the headers.
// The middleware is useful for APIs that do not support the HEAD method.
func BeHeadMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodHead {
			ctx.Request.Method = http.MethodGet
			ctx.Next()
			ctx.Request.Method = http.MethodHead

			ctx.Writer.WriteHeaderNow()
			ctx.Writer = &nopWriter{ctx.Writer}
			return
		}

		ctx.Next()
	}
}

// CacheControlMiddleware is a middleware that sets the cache control headers.
// It also handles the ETag header.
func CacheControlMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := strings.Trim(ctx.Request.URL.Path, "/")
		if path == "/" {
			path = defaultEndpoint
		}

		fields := []zapcore.Field{zap.String("path", path)}
		if eTag, ok := dist.LookupHash(path); ok {
			fields = append(fields, zap.String("eTag", eTag))
			ctx.Header("Cache-Control", "public, must-revalidate")
			ctx.Header("ETag", eTag)

			if strings.Contains(ctx.GetHeader("If-None-Match"), eTag) {
				logger.Debug("ETag matched", fields...)
				ctx.Status(http.StatusNotModified)
				return
			}
		}

		logger.Debug("ETag not matched", fields...)
		ctx.Next()
	}
}

// HttpsRedirectMiddleware redirects HTTP requests to HTTPS
func HttpsRedirectMiddleware(enabled bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if enabled && selectValue(ctx.Request.Header.Get("X-Forwarded-Proto"), ctx.Request.URL.Scheme) != "https" {
			location := &url.URL{
				Scheme:      "https",
				Host:        selectValue(ctx.Request.Header.Get("X-Forwarded-Host"), ctx.Request.Host),
				Path:        ctx.Request.URL.Path,
				RawPath:     ctx.Request.URL.RawPath,
				RawQuery:    ctx.Request.URL.RawQuery,
				ForceQuery:  false,
				User:        ctx.Request.URL.User,
				Fragment:    ctx.Request.URL.Fragment,
				RawFragment: ctx.Request.URL.RawFragment,
				Opaque:      ctx.Request.URL.Opaque,
				OmitHost:    false,
			}

			logger.Debug("Redirecting to HTTPS", zap.String("location", location.String()))
			ctx.Redirect(http.StatusMovedPermanently, location.String())
			return
		}

		ctx.Next()
	}
}

// LimitMiddleware is a middleware that limits the number of requests per second.
// It uses a token bucket algorithm to limit the number of requests.
// The rate is the number of requests per second and the bursts is the number of requests that can be bursted.
// If the limit is reached, the middleware will return a 429 status code.
func LimitMiddleware(rps float64, bursts uint, skip gin.Skipper) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rps), int(bursts))

	return func(ctx *gin.Context) {
		if skip != nil && skip(ctx) {
			ctx.Next()
			return
		}

		now := time.Now()
		reservation := limiter.ReserveN(now, 1)
		defer reservation.Cancel()

		if delay := reservation.DelayFrom(now); delay > 0 {
			logger.Debug("Rate limit exceeded", zap.Duration("delay", delay))
			ctx.Header("Retry-After", now.Add(delay).Format(time.RFC1123))
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		ctx.Next()
	}
}

// MetricsMiddleware is a middleware that logs metrics.
func MetricsMiddleware(database *gorm.DB, skip gin.Skipper) gin.HandlerFunc {
	queueLock := sync.Mutex{}

	return func(ctx *gin.Context) {
		ctx.Next()

		if skip != nil && skip(ctx) {
			return
		}

		queueLock.Lock()
		defer queueLock.Unlock()

		fields := []zapcore.Field{zap.String("endpoint", ctx.Request.URL.Path), zap.String("method", ctx.Request.Method)}
		if err := Helper(database).SaveMetric(Metric{
			Endpoint: ctx.Request.URL.Path,
			Method:   ctx.Request.Method,
			Count:    1,
		}); err != nil {
			logger.Error("Failed to save metrics", append(fields, zap.Error(err))...)
			return
		}

		logger.Debug("Metrics saved", fields...)
	}
}

// SessionMiddleware is a middleware that creates a session cookie.
// It uses the private key to sign the JWT token.
// The session name is the name of the cookie.
// The session duration is the duration of the session.
// If the cookie is not found or invalid, the middleware will create a new session.
// If the token is invalid, the middleware will return a 500 status code.
func SessionMiddleware(privateKey *rsa.PrivateKey, cryptKey cipher.AEAD, sessionName string, sessionDuration time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cookie, _ := ctx.Request.Cookie(sessionName); cookie != nil && cookie.Valid() == nil {
			ctx.Next()
			return
		}

		now := time.Now()
		jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
			Issuer:    "space-invaders",
			Audience:  jwt.ClaimStrings{"space-invaders"},
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   "internal",
			ExpiresAt: jwt.NewNumericDate(now.Add(sessionDuration)),
		}).SignedString(privateKey)
		if err != nil {
			logger.Error("Failed to sign token", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		encrypted, err := encryptAndEncodeB64WithAES(cryptKey, jwtToken)
		if err != nil {
			logger.Error("Failed to encrypt token", zap.Error(err))
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt token"})
			return
		}

		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:     sessionName,
			Value:    encrypted,
			MaxAge:   int((sessionDuration - time.Since(now)).Seconds()),
			Path:     "/",
			Domain:   selectValue(ctx.GetHeader("X-Forwarded-Host"), ctx.Request.URL.Hostname()),
			Secure:   selectValue(ctx.GetHeader("X-Forwarded-Proto"), ctx.Request.URL.Scheme) == "https",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
}
