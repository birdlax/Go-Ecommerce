package repository

import (
	"backend/domain"
	"errors"

	"gorm.io/gorm"
)

// ถ้ายังไม่มี ให้สร้าง error กลางของ repository ขึ้นมา
var ErrNotFound = errors.New("record not found")

type CategoryRepository interface {
	Create(category *domain.Category) error
	FindAll() ([]domain.Category, error)
	FindByID(id uint) (*domain.Category, error)
	Update(category *domain.Category) error
	Delete(id uint) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(category *domain.Category) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) FindAll() ([]domain.Category, error) {
	var categories []domain.Category
	err := r.db.Order("id asc").Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) FindByID(id uint) (*domain.Category, error) {
	var category domain.Category
	err := r.db.First(&category, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &category, err
}

func (r *categoryRepository) Update(category *domain.Category) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
	result := r.db.Delete(&domain.Category{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
