package service

import (
	"product/model"
	"product/repository"
)

type ProductService struct {
	repo *repository.ProductRepository
}

func NewProductService(r *repository.ProductRepository) *ProductService {
	return &ProductService{repo: r}
}

func (s *ProductService) GetAllProducts() ([]model.Product, error) {
	return s.repo.GetAll()
}

func (s *ProductService) GetProductByID(id uint) (*model.Product, error) {
	return s.repo.GetByID(id)
}

func (s *ProductService) CreateProduct(p *model.Product) (*model.Product, error) {
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) UpdateProduct(id uint, p *model.Product) (*model.Product, error) {
	p.ID = id
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *ProductService) DeleteProduct(id uint) error {
	return s.repo.Delete(id)
}
