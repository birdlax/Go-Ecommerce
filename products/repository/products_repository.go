package repository

import (
	"backend/products/domain"
	"gorm.io/gorm"
)

type ProductRepository interface {
	CreateProduct(product *domain.Product) error
	CreateImages(images []domain.ProductImage) error
	CreateCategory(category *domain.Category) error
}

// ... UploadRepository Interface ...

// struct ที่จะทำงานจริง
type productRepository struct {
	db *gorm.DB
}

// Constructor สำหรับ ProductRepository
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) CreateProduct(product *domain.Product) error {
	if err := r.db.Create(product).Error; err != nil {
		return err
	}
	return r.db.Preload("Category").First(product, product.ID).Error
}

func (r *productRepository) CreateImages(images []domain.ProductImage) error {
	if len(images) == 0 {
		return nil
	}
	return r.db.Create(&images).Error
}

func (r *productRepository) CreateCategory(category *domain.Category) error {
	result := r.db.Create(category)
	return result.Error
}
