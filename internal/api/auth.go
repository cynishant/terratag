package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// AuthService handles authentication operations
type AuthService struct {
	jwtSecret string
	apiKey    string
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret, apiKey string) *AuthService {
	return &AuthService{
		jwtSecret: jwtSecret,
		apiKey:    apiKey,
	}
}

// JWTClaims represents the JWT payload
type JWTClaims struct {
	UserID string   `json:"sub"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a new JWT token
func (a *AuthService) GenerateJWT(userID string, roles []string, duration time.Duration) (string, error) {
	if a.jwtSecret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	now := time.Now()
	claims := JWTClaims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			Issuer:    "terratag-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"roles":      roles,
		"expires_at": claims.ExpiresAt.Time,
	}).Info("JWT token generated")

	return tokenString, nil
}

// ValidateJWT validates and parses a JWT token
func (a *AuthService) ValidateJWT(tokenString string) (*JWTClaims, error) {
	if a.jwtSecret == "" {
		return nil, fmt.Errorf("JWT secret not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid JWT token")
}

// GenerateAPIKey creates a new secure API key
func GenerateAPIKey() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64 URL-safe string
	apiKey := base64.URLEncoding.EncodeToString(bytes)
	
	logrus.Info("New API key generated")
	return apiKey, nil
}

// HashAPIKey creates a hash of an API key for secure storage
func HashAPIKey(apiKey string) string {
	// In a real implementation, use bcrypt or similar
	// For now, this is a placeholder
	return base64.StdEncoding.EncodeToString([]byte(apiKey))
}

// AuthInfo represents authentication information
type AuthInfo struct {
	Method   string    `json:"method"`
	UserID   string    `json:"user_id,omitempty"`
	Roles    []string  `json:"roles,omitempty"`
	IssuedAt time.Time `json:"issued_at,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// GetAuthInfo extracts authentication information from gin context
func GetAuthInfo(c *gin.Context) *AuthInfo {
	info := &AuthInfo{}
	
	if method, exists := c.Get("auth_method"); exists {
		info.Method = method.(string)
	}
	
	if userID, exists := c.Get("user_id"); exists {
		info.UserID = userID.(string)
	}
	
	if roles, exists := c.Get("user_roles"); exists {
		if roleSlice, ok := roles.([]interface{}); ok {
			for _, role := range roleSlice {
				if roleStr, ok := role.(string); ok {
					info.Roles = append(info.Roles, roleStr)
				}
			}
		}
	}
	
	return info
}

// RequireRole middleware to check user roles
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authInfo := GetAuthInfo(c)
		
		// If using API key, allow all operations (admin access)
		if authInfo.Method == "api_key" {
			c.Next()
			return
		}
		
		// Check if user has required role
		for _, role := range authInfo.Roles {
			if role == requiredRole || role == "admin" {
				c.Next()
				return
			}
		}
		
		logrus.WithFields(logrus.Fields{
			"user_id":       authInfo.UserID,
			"user_roles":    authInfo.Roles,
			"required_role": requiredRole,
			"path":          c.Request.URL.Path,
		}).Warn("Insufficient permissions for operation")
		
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "insufficient permissions",
			"message": fmt.Sprintf("Role '%s' required for this operation", requiredRole),
		})
		c.Abort()
	}
}