package repository

import (
	"order/model"

	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *model.Order) error
	GetByID(id uint) (*model.Order, error)
	GetAll() ([]model.Order, error)
	UpdateStatus(id uint, status string) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetAll() ([]model.Order, error) {
	var orders []model.Order
	if err := r.db.Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *orderRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}
