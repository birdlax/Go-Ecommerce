package repository

import (
	"backend/domain"
	"errors"

	"gorm.io/gorm"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrNotFound = errors.New("record not found")

type OrderRepository interface {
	Create(order *domain.Order) error
	FindByID(orderID uint) (*domain.Order, error)
	FindAllByUserID(userID uint) ([]domain.Order, error)
	Update(order *domain.Order) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create จะสร้าง Order และ OrderItems พร้อมกันใน Transaction เดียว
func (r *orderRepository) Create(order *domain.Order) error {
	return r.db.Create(order).Error
}

// FindByID ค้นหา Order ตาม ID พร้อมข้อมูลสินค้า
func (r *orderRepository) FindByID(orderID uint) (*domain.Order, error) {
	var order domain.Order
	err := r.db.Preload("OrderItems.Product").Preload("ShippingAddress").First(&order, orderID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &order, nil
}

// FindAllByUserID ค้นหาทุก Order ของ User คนนั้น
func (r *orderRepository) FindAllByUserID(userID uint) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Preload("OrderItems.Product").Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) Update(order *domain.Order) error {
	// Save จะทำการอัปเดตทุกฟิลด์ของ object ที่มี Primary Key อยู่แล้ว
	return r.db.Save(order).Error
}

func (r *orderRepository) UpdateTrackingInfo(orderID uint, trackingNumber string) error {
	var order domain.Order
	err := r.db.First(&order, orderID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrOrderNotFound
		}
		return err
	}

	order.TrackingNumber = &trackingNumber
	return r.db.Save(&order).Error
}
