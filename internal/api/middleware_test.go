package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_NoAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &SecurityConfig{
		RequireAuth: false,
	}
	
	r := gin.New()
	r.Use(AuthMiddleware(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_APIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &SecurityConfig{
		RequireAuth:  true,
		EnableAPIKey: true,
		APIKey:       "test-api-key",
	}
	
	r := gin.New()
	r.Use(AuthMiddleware(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	tests := []struct {
		name       string
		header     string
		value      string
		expectCode int
	}{
		{
			name:       "valid API key in Authorization header",
			header:     "Authorization",
			value:      "Bearer test-api-key",
			expectCode: http.StatusOK,
		},
		{
			name:       "valid API key in X-API-Key header",
			header:     "X-API-Key",
			value:      "test-api-key",
			expectCode: http.StatusOK,
		},
		{
			name:       "invalid API key",
			header:     "X-API-Key",
			value:      "invalid-key",
			expectCode: http.StatusUnauthorized,
		},
		{
			name:       "no API key",
			header:     "",
			value:      "",
			expectCode: http.StatusUnauthorized,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.header != "" {
				req.Header.Set(tt.header, tt.value)
			}
			w := httptest.NewRecorder()
			
			r.ServeHTTP(w, req)
			
			assert.Equal(t, tt.expectCode, w.Code)
		})
	}
}

func TestAuthMiddleware_HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &SecurityConfig{
		RequireAuth:  true,
		EnableAPIKey: true,
		APIKey:       "test-api-key",
	}
	
	r := gin.New()
	r.Use(AuthMiddleware(config))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	req := httptest.NewRequest("GET", "/health", nil)
	// No authentication headers
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	// Health check should work without authentication
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	r := gin.New()
	r.Use(SecurityHeadersMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	config := &SecurityConfig{
		RateLimitRPM: 2, // Very low limit for testing
	}
	
	r := gin.New()
	r.Use(RateLimitMiddleware(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Make requests up to the limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		
		r.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	}
	
	// This request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestNewSecurityConfig(t *testing.T) {
	// Test with environment variables
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("API_KEY", "test-key")
	os.Setenv("REQUIRE_AUTH", "false")
	defer func() {
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("API_KEY")
		os.Unsetenv("REQUIRE_AUTH")
	}()
	
	config := NewSecurityConfig()
	
	assert.Equal(t, "test-secret", config.JWTSecret)
	assert.Equal(t, "test-key", config.APIKey)
	assert.False(t, config.RequireAuth)
	assert.True(t, config.EnableAPIKey) // Default value
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")
	
	result := getEnvOrDefault("TEST_VAR", "default")
	assert.Equal(t, "test_value", result)
	
	// Test with non-existing environment variable
	result = getEnvOrDefault("NON_EXISTING_VAR", "default")
	assert.Equal(t, "default", result)
}