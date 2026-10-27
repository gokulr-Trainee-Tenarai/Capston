package repository

import (
	"context"

	"order/model"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, o *model.Order) error
	GetByID(ctx context.Context, id uint) (*model.Order, error)
	List(ctx context.Context, limit, offset int) ([]model.Order, error)
	UpdateStatus(ctx context.Context, id uint, status model.OrderStatus) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, o *model.Order) error {
	return r.db.WithContext(ctx).Create(o).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id uint) (*model.Order, error) {
	var o model.Order
	err := r.db.WithContext(ctx).First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *orderRepository) List(ctx context.Context, limit, offset int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("id DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint, status model.OrderStatus) error {
	res := r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", id).Update("status", status)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}