package repository

import (
	"backend/domain"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("record not found")
)

type CouponRepository interface {
	Create(coupon *domain.Coupon) error
	FindAll() ([]domain.Coupon, error)
	FindByID(id uint) (*domain.Coupon, error)
	FindByCode(code string) (*domain.Coupon, error)
	Update(coupon *domain.Coupon) error
	Delete(id uint) error
}

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) Create(coupon *domain.Coupon) error {
	return r.db.Create(coupon).Error
}

func (r *couponRepository) FindAll() ([]domain.Coupon, error) {
	var coupons []domain.Coupon
	err := r.db.Find(&coupons).Error
	return coupons, err
}

func (r *couponRepository) FindByID(id uint) (*domain.Coupon, error) {
	var coupon domain.Coupon
	err := r.db.First(&coupon, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &coupon, err
}

func (r *couponRepository) FindByCode(code string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	err := r.db.Where("code = ?", code).First(&coupon).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &coupon, err
}

func (r *couponRepository) Update(coupon *domain.Coupon) error {
	return r.db.Save(coupon).Error
}

func (r *couponRepository) Delete(id uint) error {
	result := r.db.Delete(&domain.Coupon{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
