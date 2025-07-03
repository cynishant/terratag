package api

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// SecurityConfig holds security configuration
type SecurityConfig struct {
	JWTSecret       string
	APIKey          string
	EnableAPIKey    bool
	EnableJWT       bool
	RequireAuth     bool
	TrustedProxies  []string
	RateLimitRPM    int // Requests per minute
}

// NewSecurityConfig creates security configuration from environment variables
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		JWTSecret:      getEnvOrDefault("JWT_SECRET", ""),
		APIKey:         getEnvOrDefault("API_KEY", ""),
		EnableAPIKey:   getEnvOrDefault("ENABLE_API_KEY", "true") == "true",
		EnableJWT:      getEnvOrDefault("ENABLE_JWT", "false") == "true",
		RequireAuth:    getEnvOrDefault("REQUIRE_AUTH", "true") == "true",
		TrustedProxies: strings.Split(getEnvOrDefault("TRUSTED_PROXIES", "127.0.0.1,::1"), ","),
		RateLimitRPM:   getEnvIntOrDefault("RATE_LIMIT_RPM", 60),
	}
}

// AuthMiddleware provides authentication and authorization
func AuthMiddleware(config *SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		// Skip auth if not required (development mode)
		if !config.RequireAuth {
			logrus.WithField("path", c.Request.URL.Path).Debug("Authentication disabled")
			c.Next()
			return
		}

		// Try API key authentication first
		if config.EnableAPIKey && config.APIKey != "" {
			if authenticated := checkAPIKey(c, config.APIKey); authenticated {
				c.Set("auth_method", "api_key")
				c.Next()
				return
			}
		}

		// Try JWT authentication
		if config.EnableJWT && config.JWTSecret != "" {
			if authenticated := checkJWT(c, config.JWTSecret); authenticated {
				c.Set("auth_method", "jwt")
				c.Next()
				return
			}
		}

		// No valid authentication found
		logrus.WithFields(logrus.Fields{
			"ip":     c.ClientIP(),
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"agent":  c.Request.UserAgent(),
		}).Warn("Unauthorized API access attempt")

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "authentication required",
			"message": "Valid API key or JWT token required",
		})
		c.Abort()
	}
}

// checkAPIKey validates API key from header or query parameter
func checkAPIKey(c *gin.Context, expectedKey string) bool {
	// Check Authorization header (Bearer format)
	if auth := c.GetHeader("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(token), []byte(expectedKey)) == 1 {
				return true
			}
		}
	}

	// Check X-API-Key header
	if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) == 1 {
			return true
		}
	}

	// Check query parameter (less secure, for development only)
	if apiKey := c.Query("api_key"); apiKey != "" {
		logrus.WithField("ip", c.ClientIP()).Warn("API key provided via query parameter - insecure")
		if subtle.ConstantTimeCompare([]byte(apiKey), []byte(expectedKey)) == 1 {
			return true
		}
	}

	return false
}

// checkJWT validates JWT token
func checkJWT(c *gin.Context, secret string) bool {
	auth := c.GetHeader("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return false
	}

	tokenString := strings.TrimPrefix(auth, "Bearer ")
	
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		}).Debug("JWT validation failed")
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Set user information in context
		if sub, ok := claims["sub"].(string); ok {
			c.Set("user_id", sub)
		}
		if roles, ok := claims["roles"].([]interface{}); ok {
			c.Set("user_roles", roles)
		}
		return true
	}

	return false
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		// Remove server header
		c.Header("Server", "")
		
		c.Next()
	}
}

// RateLimitMiddleware implements basic rate limiting
func RateLimitMiddleware(config *SecurityConfig) gin.HandlerFunc {
	// Simple in-memory store for rate limiting
	// In production, use Redis or similar
	clients := make(map[string][]time.Time)
	
	return func(c *gin.Context) {
		if config.RateLimitRPM <= 0 {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		now := time.Now()
		
		// Clean old entries and count recent requests
		var recentRequests []time.Time
		if requests, exists := clients[clientIP]; exists {
			for _, requestTime := range requests {
				if now.Sub(requestTime) < time.Minute {
					recentRequests = append(recentRequests, requestTime)
				}
			}
		}
		
		// Check rate limit
		if len(recentRequests) >= config.RateLimitRPM {
			logrus.WithFields(logrus.Fields{
				"ip":           clientIP,
				"requests":     len(recentRequests),
				"limit":        config.RateLimitRPM,
			}).Warn("Rate limit exceeded")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate limit exceeded",
				"message": "Too many requests. Please try again later.",
				"limit":   config.RateLimitRPM,
			})
			c.Abort()
			return
		}
		
		// Add current request
		recentRequests = append(recentRequests, now)
		clients[clientIP] = recentRequests
		
		c.Next()
	}
}

// LoggingMiddleware provides detailed API logging
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			logrus.WithFields(logrus.Fields{
				"ip":         param.ClientIP,
				"method":     param.Method,
				"path":       param.Path,
				"status":     param.StatusCode,
				"latency":    param.Latency,
				"user_agent": param.Request.UserAgent(),
			}).Info("API request")
			return ""
		},
	})
}

// TrustedProxyMiddleware validates trusted proxies
func TrustedProxyMiddleware(trustedProxies []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// Check if request comes from trusted proxy
		if len(trustedProxies) > 0 {
			trusted := false
			for _, proxy := range trustedProxies {
				proxy = strings.TrimSpace(proxy)
				if proxy == clientIP {
					trusted = true
					break
				}
			}
			
			if !trusted {
				logrus.WithFields(logrus.Fields{
					"ip":            clientIP,
					"trusted_proxies": trustedProxies,
				}).Debug("Request from untrusted proxy")
			}
		}
		
		c.Next()
	}
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}