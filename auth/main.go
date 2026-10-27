package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("lab2-secret-key-change-me")

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "auth", "status": "UP"})
	})

	r.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		var role string
		if req.Username == "admin" && req.Password == "admin123" {
			role = "ADMIN"
		} else if req.Username == "user1" && req.Password == "user123" {
			role = "USER"
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		claims := jwt.MapClaims{
			"sub":  req.Username,
			"role": role,
			"exp":  time.Now().Add(2 * time.Hour).Unix(),
			"iat":  time.Now().Unix(),
		}

		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token creation failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token, "role": role})
	})

	r.Run(":8083")
}