package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"order/controller"
	"order/database"
	"order/middleware"
	"order/repository"
	"order/service"
)

func main() {
	db := database.ConnectPostgres()
	orderRepo := repository.NewOrderRepository(db)
	productClient := service.NewProductClient()
	orderService := service.NewOrderService(orderRepo, productClient)
	orderCtrl := controller.NewOrderController(orderService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.StructuredLogger())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "order"})
	})

	// Order REST routes
	api := r.Group("/api/v1/orders")
	{
		api.POST("", orderCtrl.CreateOrder)
		api.GET("", orderCtrl.ListOrders)
		api.GET("/:id", orderCtrl.GetOrderByID)
		api.PUT("/:id/status", orderCtrl.UpdateOrderStatus)
	}

	// GraphQL endpoint (passes db *gorm.DB properly)
	r.POST("/graphql", controller.GraphQLHandler(db))

	port := os.Getenv("ORDER_PORT")
	if port == "" {
		port = "8082"
	}
	port = strings.TrimPrefix(port, ":")

	log.Printf("Starting Order Service on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Order service failed to start: %v", err)
	}
}
