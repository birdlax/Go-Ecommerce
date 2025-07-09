package repository

import (
	"errors"
	// [สำคัญ] แก้ไข "backend" เป็นชื่อ Module ใน go.mod ของคุณ
	"backend/domain"
	"backend/users/dto"

	"gorm.io/gorm"
)

// ErrNotFound คือ Custom Error ของ Repository ที่จะถูกส่งออกไป
var ErrNotFound = errors.New("record not found")

// ===================================================================
// UserRepository: จัดการเฉพาะข้อมูลในตาราง 'users'
// ===================================================================

type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id uint) (*domain.User, error)
	FindAll(params dto.UserQueryParams) ([]domain.User, error)
	Count() (int64, error)
	Update(user *domain.User) error
	Delete(id uint) error
	FindByRefreshToken(hashedToken string) (*domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &user, err
}

func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
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
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	result := r.db.Delete(&domain.User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepository) FindByRefreshToken(hashedToken string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("refresh_token = ?", hashedToken).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &user, err
}

// ===================================================================
// AddressRepository: จัดการเฉพาะข้อมูลในตาราง 'addresses'
// ===================================================================

type AddressRepository interface {
	Create(address *domain.Address) error
	FindByUserID(userID uint) ([]domain.Address, error)
	FindByID(addressID uint) (*domain.Address, error)
	Update(address *domain.Address) error
	Delete(addressID uint) error
	ClearDefault(userID uint) error
}

type addressRepository struct {
	db *gorm.DB
}

func NewAddressRepository(db *gorm.DB) AddressRepository {
	return &addressRepository{db: db}
}

func (r *addressRepository) Create(address *domain.Address) error {
	return r.db.Create(address).Error
}

func (r *addressRepository) FindByUserID(userID uint) ([]domain.Address, error) {
	var addresses []domain.Address
	err := r.db.Where("user_id = ?", userID).Find(&addresses).Error
	return addresses, err
}

func (r *addressRepository) FindByID(addressID uint) (*domain.Address, error) {
	var address domain.Address
	err := r.db.First(&address, addressID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &address, err
}

func (r *addressRepository) Update(address *domain.Address) error {
	return r.db.Save(address).Error
}

func (r *addressRepository) Delete(addressID uint) error {
	result := r.db.Delete(&domain.Address{}, addressID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *addressRepository) ClearDefault(userID uint) error {
	return r.db.Model(&domain.Address{}).Where("user_id = ? AND is_default = ?", userID, true).Update("is_default", false).Error
}
