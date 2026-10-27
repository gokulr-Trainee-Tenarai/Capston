package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	consulapi "github.com/hashicorp/consul/api"
)

var secret = []byte("lab2-secret-key-change-me")

var defaultServicePorts = map[string]string{
	"order-service":   "http://127.0.0.1:8082",
	"product-service": "http://127.0.0.1:8084",
	"auth-service":    "http://127.0.0.1:8083",
}

func discoverService(name string) string {
	envKey := strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "_URL"
	if customURL := os.Getenv(envKey); customURL != "" {
		return strings.TrimRight(customURL, "/")
	}

	config := consulapi.DefaultConfig()
	client, err := consulapi.NewClient(config)
	if err == nil {
		entries, _, err := client.Health().Service(name, "", true, nil)
		if err == nil && len(entries) > 0 {
			svc := entries[0].Service
			addr := svc.Address
			if addr == "" {
				addr = "127.0.0.1"
			}
			return fmt.Sprintf("http://%s:%d", addr, svc.Port)
		}
	}

	if fallback, ok := defaultServicePorts[name]; ok {
		return fallback
	}
	return "http://127.0.0.1:8080"
}

func jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Bearer token"})
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		if sub, ok := claims["sub"]; ok {
			c.Set("userID", sub)
		}
		if role, ok := claims["role"]; ok {
			c.Set("role", role)
		}

		c.Next()
	}
}

func proxyRequest(getTargetURL func() string) gin.HandlerFunc {
	client := &http.Client{}

	return func(c *gin.Context) {
		baseURL := getTargetURL()
		targetURL := fmt.Sprintf("%s%s", baseURL, c.Request.URL.RequestURI())

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create proxy request"})
			return
		}

		for k, v := range c.Request.Header {
			req.Header[k] = v
		}

		traceID := c.GetHeader("X-Request-ID")
		if traceID == "" {
			traceID = c.GetHeader("X-Trace-ID")
		}
		if traceID != "" {
			req.Header.Set("X-Request-ID", traceID)
		}

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": fmt.Sprintf("downstream service unavailable: %v", err)})
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			c.Writer.Header()[k] = v
		}
		c.Status(resp.StatusCode)
		_, _ = io.Copy(c.Writer, resp.Body)
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// ALB health checks
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "api-gateway"})
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ALIVE"})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	})

	// Public Auth routes
	r.POST("/login", proxyRequest(func() string { return discoverService("auth-service") }))
	r.POST("/register", proxyRequest(func() string { return discoverService("auth-service") }))

	// Orders routes
	orderProxy := proxyRequest(func() string { return discoverService("order-service") })
	orders := r.Group("/api/v1/orders")
	{
		orders.Any("", orderProxy)
		orders.Any("/*path", orderProxy)
	}

	// Products routes
	prodProxy := proxyRequest(func() string { return discoverService("product-service") })
	products := r.Group("/api/v1/products")
	{
		products.GET("", prodProxy)
		products.GET("/:id", prodProxy)
		products.POST("", jwtMiddleware(), prodProxy)
		products.PUT("/:id", jwtMiddleware(), prodProxy)
		products.DELETE("/:id", jwtMiddleware(), prodProxy)
	}

	// GraphQL
	r.POST("/graphql", prodProxy)

	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "5000"
	}
	port = strings.TrimPrefix(port, ":")

	log.Printf("Starting API Gateway on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("API Gateway run error: %v", err)
	}
}
