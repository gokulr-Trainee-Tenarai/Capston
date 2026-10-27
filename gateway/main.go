package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	consulapi "github.com/hashicorp/consul/api"
)

var secret = []byte("lab2-secret-key-change-me")

func discoverService(name string) (string, error) {
	config := consulapi.DefaultConfig()
	client, err := consulapi.NewClient(config)
	if err != nil {
		return "", err
	}

	entries, _, err := client.Health().Service(name, "", false, nil)
	if err != nil || len(entries) == 0 {
		return "", fmt.Errorf("service %s not found in consul", name)
	}

	svc := entries[0].Service
	return fmt.Sprintf("http://%s:%d", svc.Address, svc.Port), nil
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

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user", claims["sub"])
			c.Set("role", claims["role"])
		}
		c.Next()
	}
}

func requireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: requires " + role + " privileges"})
			return
		}
		c.Next()
	}
}

func proxy(serviceName string, pathPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetHost, err := discoverService(serviceName)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		targetPath := strings.TrimPrefix(c.Request.URL.Path, pathPrefix)
		targetURL := targetHost + targetPath
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}

		for k, v := range c.Request.Header {
			for _, val := range v {
				req.Header.Add(k, val)
			}
		}

		if user, exists := c.Get("user"); exists {
			req.Header.Set("X-User", fmt.Sprintf("%v", user))
		}
		if role, exists := c.Get("role"); exists {
			req.Header.Set("X-Role", fmt.Sprintf("%v", role))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": serviceName + " unavailable"})
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			c.Header(k, strings.Join(v, ","))
		}
		c.Status(resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		c.Writer.Write(body)
	}
}

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "gateway", "status": "UP"})
	})

	r.POST("/login", func(c *gin.Context) {
		proxy("auth-service", "")(c)
	})

	orders := r.Group("/api/v1/orders", jwtMiddleware())
	{
		orders.Any("", proxy("order-service", "/api/v1"))
		orders.Any("/*path", proxy("order-service", "/api/v1"))
	}

	products := r.Group("/api/v1/products", jwtMiddleware())
	{
		products.POST("", requireRole("ADMIN"), proxy("product-service", "/api/v1"))
		products.GET("", proxy("product-service", "/api/v1"))
		products.GET("/:id", proxy("product-service", "/api/v1"))
	}

	// Direct GraphQL proxy
	r.POST("/graphql", jwtMiddleware(), proxy("order-service", ""))
	port := os.Getenv("PORT")
	if port == "" {
	    port = "10006" // local development
	}

	r.Run(":" + port)
}
