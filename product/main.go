package main

import (
	"log"
	"net/http"

	"product/config"
	"product/controller"
	"product/database"
	"product/repository"
	"product/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}

	repo := repository.NewProductRepository(db)
	svc := service.NewProductService(repo)
	ctrl := controller.NewProductController(svc)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "product", "status": "UP"})
	})

	r.POST("/products", ctrl.CreateProduct)
	r.GET("/products/:id", ctrl.GetProduct)
	r.GET("/products", ctrl.ListProducts)

	log.Printf("Product service running on :%s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal(err)
	}
}
