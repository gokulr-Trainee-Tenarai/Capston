package repository

import (
	"context"

	"product/model"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error
	GetByID(ctx context.Context, id uint) (*model.Product, error)
	List(ctx context.Context, limit, offset int) ([]model.Product, error)
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, p *model.Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *productRepository) GetByID(ctx context.Context, id uint) (*model.Product, error) {
	var p model.Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) List(ctx context.Context, limit, offset int) ([]model.Product, error) {
	var products []model.Product
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Order("id DESC").Find(&products).Error
	return products, err
}