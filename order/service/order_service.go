package service

import (
	"context"
	"errors"
	"order/model"
	"order/repository"
)

type OrderService struct {
	orderRepo     *repository.OrderRepository
	productClient *ProductClient
}

func NewOrderService(repo *repository.OrderRepository, prodClient *ProductClient) *OrderService {
	return &OrderService{
		orderRepo:     repo,
		productClient: prodClient,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	price, err := s.productClient.GetProductPrice(ctx, order.ProductID)
	if err != nil {
		return nil, errors.New("failed to retrieve product price: " + err.Error())
	}

	order.TotalPrice = price * float64(order.Quantity)
	order.Status = "CREATED"

	if err := s.orderRepo.Create(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetOrderByID(id uint) (*model.Order, error) {
	return s.orderRepo.GetByID(id)
}

func (s *OrderService) ListOrders() ([]model.Order, error) {
	return s.orderRepo.GetAll()
}

func (s *OrderService) UpdateStatus(id uint, status string) error {
	return s.orderRepo.UpdateStatus(id, status)
}
