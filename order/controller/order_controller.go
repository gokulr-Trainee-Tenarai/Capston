package controller

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"order/model"
	"order/service"
)

type OrderController struct {
	orderService *service.OrderService
}

func NewOrderController(svc *service.OrderService) *OrderController {
	return &OrderController{orderService: svc}
}

type CreateOrderInput struct {
	CustomerID uint `json:"customerId" binding:"required"`
	ProductID  uint `json:"productId" binding:"required"`
	Quantity   int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateStatusInput struct {
	Status string `json:"status" binding:"required"`
}

func (ctrl *OrderController) CreateOrder(c *gin.Context) {
	var input CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	traceID := c.GetHeader("X-Request-ID")
	if traceID == "" {
		traceID = c.GetHeader("X-Trace-ID")
	}
	ctx := context.WithValue(c.Request.Context(), "TraceID", traceID)

	order := &model.Order{
		CustomerID: input.CustomerID,
		ProductID:  input.ProductID,
		Quantity:   input.Quantity,
		Status:     "CREATED",
	}

	createdOrder, err := ctrl.orderService.CreateOrder(ctx, order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdOrder)
}

func (ctrl *OrderController) GetOrderByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := ctrl.orderService.GetOrderByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (ctrl *OrderController) ListOrders(c *gin.Context) {
	orders, err := ctrl.orderService.ListOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func (ctrl *OrderController) UpdateOrderStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var input UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status payload"})
		return
	}

	validStatuses := map[string]bool{
		"CREATED":   true,
		"CONFIRMED": true,
		"CANCELLED": true,
	}

	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Allowed values: CREATED, CONFIRMED, CANCELLED",
		})
		return
	}

	err = ctrl.orderService.UpdateStatus(uint(id), input.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"status":  input.Status,
	})
}
