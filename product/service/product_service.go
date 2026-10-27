package service

import (
	"context"

	"product/model"
	"product/repository"
)

type ProductService interface {
	CreateProduct(ctx context.Context, input model.CreateProductInput, user string) (*model.Product, error)
	GetProductByID(ctx context.Context, id uint) (*model.Product, error)
	ListProducts(ctx context.Context, limit, offset int) ([]model.Product, error)
}

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, input model.CreateProductInput, user string) (*model.Product, error) {
	p := &model.Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		InStock:     input.Stock > 0,
		CreatedBy:   user,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *productService) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *productService) ListProducts(ctx context.Context, limit, offset int) ([]model.Product, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}