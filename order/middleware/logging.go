package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func StructuredLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID, _ := c.Get("request_id")

		c.Next()

		log.Printf("request_id=%v method=%s path=%s status=%d duration=%s error=%s",
			reqID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
			c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
}