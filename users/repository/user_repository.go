package repository

import (
	"backend/domain"
	"backend/users/dto"
	"gorm.io/gorm"
)

// UserRepository Interface
type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uint) (*domain.User, error)
	FindAll(params dto.UserQueryParams) ([]domain.User, error)
	Count() (int64, error)
	Update(user *domain.User) error
	Delete(id uint) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository Constructor
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error // GORM จะจัดการเรื่อง ErrRecordNotFound ให้
	return &user, err
}

func (r *userRepository) FindAll(params dto.UserQueryParams) ([]domain.User, error) {
	var users []domain.User
	offset := (params.Page - 1) * params.Limit
	orderBy := params.SortBy + " " + params.Order
	err := r.db.Offset(offset).Limit(params.Limit).Order(orderBy).Find(&users).Error
	return users, err
}

func (r *userRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Count(&count).Error
	return count, err
}

func (r *userRepository) Update(user *domain.User) error {
	// .Save จะทำการอัปเดตทุกฟิลด์ถ้ามี Primary Key อยู่
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	// GORM จะทำ Soft Delete เพราะ User model มี gorm.Model
	result := r.db.Delete(&domain.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // คืนค่า error ถ้าไม่เจอ ID ที่จะลบ
	}
	return nil
}
