package service

import (
	"backend/coupons/dto"
	"backend/coupons/repository"
	"backend/domain"
	"backend/internal/datastore"
	"errors"
)

var (
	ErrCouponNotFound = errors.New("coupon not found")
	ErrCouponExists   = errors.New("coupon code already exists")
)

type CouponService interface {
	Create(req dto.CouponRequest) (*dto.CouponResponse, error)
	GetAll() ([]dto.CouponResponse, error)
	GetByID(id uint) (*dto.CouponResponse, error)
	Update(id uint, req dto.CouponRequest) (*dto.CouponResponse, error)
	Delete(id uint) error
}

type couponService struct {
	uow datastore.UnitOfWork
}

func NewCouponService(uow datastore.UnitOfWork) CouponService {
	return &couponService{uow: uow}
}

func (s *couponService) Create(req dto.CouponRequest) (*dto.CouponResponse, error) {
	// ตรวจสอบว่าโค้ดซ้ำหรือไม่
	_, err := s.uow.CouponRepository().FindByCode(req.Code)
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, ErrCouponExists
	}

	newCoupon := &domain.Coupon{
		Code:          req.Code,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		ExpiryDate:    req.ExpiryDate,
		UsageLimit:    req.UsageLimit,
		IsActive:      req.IsActive,
	}

	err = s.uow.Execute(func(repos *datastore.Repositories) error {
		return repos.Coupon.Create(newCoupon)
	})

	if err != nil {
		return nil, err
	}
	return mapCouponToResponse(newCoupon), nil
}

func (s *couponService) GetAll() ([]dto.CouponResponse, error) {
	coupons, err := s.uow.CouponRepository().FindAll()
	if err != nil {
		return nil, err
	}
	responses := make([]dto.CouponResponse, 0, len(coupons))
	for _, c := range coupons {
		responses = append(responses, *mapCouponToResponse(&c))
	}
	return responses, nil
}

func (s *couponService) GetByID(id uint) (*dto.CouponResponse, error) {
	coupon, err := s.uow.CouponRepository().FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	return mapCouponToResponse(coupon), nil
}

func (s *couponService) Update(id uint, req dto.CouponRequest) (*dto.CouponResponse, error) {
	var updatedCoupon *domain.Coupon
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		coupon, err := repos.Coupon.FindByID(id)
		if err != nil {
			return err
		}
		coupon.Code = req.Code
		coupon.DiscountType = req.DiscountType
		coupon.DiscountValue = req.DiscountValue
		coupon.ExpiryDate = req.ExpiryDate
		coupon.UsageLimit = req.UsageLimit
		coupon.IsActive = req.IsActive

		updatedCoupon = coupon
		return repos.Coupon.Update(coupon)
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	return mapCouponToResponse(updatedCoupon), nil
}

func (s *couponService) Delete(id uint) error {
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		return repos.Coupon.Delete(id)
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrCouponNotFound
		}
		return err
	}
	return nil
}

// Helper function
func mapCouponToResponse(coupon *domain.Coupon) *dto.CouponResponse {
	return &dto.CouponResponse{
		ID:            coupon.ID,
		Code:          coupon.Code,
		DiscountType:  coupon.DiscountType,
		DiscountValue: coupon.DiscountValue,
		ExpiryDate:    coupon.ExpiryDate,
		UsageLimit:    coupon.UsageLimit,
		UsageCount:    coupon.UsageCount,
		IsActive:      coupon.IsActive,
		CreatedAt:     coupon.CreatedAt,
	}
}
