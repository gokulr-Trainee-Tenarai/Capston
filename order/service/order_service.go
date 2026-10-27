package service

import (
	"context"
	"errors"

	"order/model"
	"order/repository"
)

var ErrInvalidStatus = errors.New("invalid status value")

type OrderService interface {
	CreateOrder(ctx context.Context, input model.CreateOrderInput) (*model.Order, error)
	GetOrderByID(ctx context.Context, id uint) (*model.Order, error)
	ListOrders(ctx context.Context, limit, offset int) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, id uint, status model.OrderStatus) error
}

type orderService struct {
	repo          repository.OrderRepository
	productClient ProductClient
}

func NewOrderService(repo repository.OrderRepository, prodClient ProductClient) OrderService {
	return &orderService{repo: repo, productClient: prodClient}
}

func (s *orderService) CreateOrder(ctx context.Context, input model.CreateOrderInput) (*model.Order, error) {
	prod, err := s.productClient.GetProductWithResilience(ctx, input.ProductID)
	unitPrice := 10.0
	if err == nil && prod != nil {
		unitPrice = prod.Price
	}

	order := &model.Order{
		CustomerID:  input.CustomerID,
		ProductID:   input.ProductID,
		Quantity:    input.Quantity,
		UnitPrice:   unitPrice,
		TotalAmount: unitPrice * float64(input.Quantity),
		Status:      model.StatusCreated,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *orderService) GetOrderByID(ctx context.Context, id uint) (*model.Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *orderService) ListOrders(ctx context.Context, limit, offset int) ([]model.Order, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id uint, status model.OrderStatus) error {
	if !status.IsValid() {
		return ErrInvalidStatus
	}
	return s.repo.UpdateStatus(ctx, id, status)
}
