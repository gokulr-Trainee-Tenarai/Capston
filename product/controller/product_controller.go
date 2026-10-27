package controller

import (
	"errors"
	"net/http"
	"strconv"

	"product/model"
	"product/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductController struct {
	svc service.ProductService
}

func NewProductController(svc service.ProductService) *ProductController {
	return &ProductController{svc: svc}
}

func (c *ProductController) CreateProduct(ctx *gin.Context) {
	if role := ctx.GetHeader("X-Role"); role != "ADMIN" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden: only admins can create products"})
		return
	}

	var input model.CreateProductInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: name, price (>0), stock (>=0) required"})
		return
	}

	user := ctx.GetHeader("X-User")
	if user == "" {
		user = "admin"
	}

	prod, err := c.svc.CreateProduct(ctx.Request.Context(), input, user)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, prod)
}

func (c *ProductController) GetProduct(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	prod, err := c.svc.GetProductByID(ctx.Request.Context(), uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, prod)
}

func (c *ProductController) ListProducts(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	prods, err := c.svc.ListProducts(ctx.Request.Context(), limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, prods)
}