package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"product/controller"
	"product/database"
	"product/repository"
	"product/service"
)

func main() {
	db := database.ConnectPostgres()
	prodRepo := repository.NewProductRepository(db)
	prodService := service.NewProductService(prodRepo)
	prodCtrl := controller.NewProductController(prodService)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "product"})
	})

	api := r.Group("/api/v1/products")
	{
		api.GET("", prodCtrl.GetAllProducts)
		api.GET("/:id", prodCtrl.GetProductByID)
		api.POST("", prodCtrl.CreateProduct)
		api.PUT("/:id", prodCtrl.UpdateProduct)
		api.DELETE("/:id", prodCtrl.DeleteProduct)
	}

	port := os.Getenv("PRODUCT_PORT")
	if port == "" {
		port = "8084"
	}
	port = strings.TrimPrefix(port, ":")

	log.Printf("Starting Product Service on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Product service failed to start: %v", err)
	}
}
