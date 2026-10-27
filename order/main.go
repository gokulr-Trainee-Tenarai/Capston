package main

import (
	"log"
	"net/http"

	"order/config"
	"order/controller"
	"order/database"
	"order/middleware"
	"order/repository"
	"order/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal("database initialization failed: ", err)
	}

	repo := repository.NewOrderRepository(db)
	client := service.NewProductClient(cfg.ProductServiceURL)
	svc := service.NewOrderService(repo, client)
	ctrl := controller.NewOrderController(svc)

	gqlHandler, err := controller.NewGraphQLHandler(svc)
	if err != nil {
		log.Fatal("GraphQL schema initialization failed: ", err)
	}

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.StructuredLogging())
	r.Use(middleware.Recovery())

	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "DOWN", "error": "db down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	})

	// REST routes
	r.POST("/orders", ctrl.CreateOrder)
	r.GET("/orders/:id", ctrl.GetOrder)
	r.GET("/orders", ctrl.ListOrders)
	r.PUT("/orders/:id/status", ctrl.UpdateOrderStatus)

	// GraphQL route
	r.POST("/graphql", gqlHandler.Handle)

	log.Printf("Order service listening on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
