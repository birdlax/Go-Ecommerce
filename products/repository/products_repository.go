package repository

import (
	"backend/products/domain"
	"errors"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

type ProductRepository interface {
	CreateProduct(product *domain.Product) error
	CreateImages(images []domain.ProductImage) error
	CreateCategory(category *domain.Category) error
	FindAll(params domain.QueryParams) ([]domain.Product, error)
	FindProductByID(id uint) (*domain.Product, error)
	Count() (int64, error)
	Delete(id uint) error
	Update(id uint, updates map[string]interface{}) error
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

func (r *productRepository) FindProductByID(id uint) (*domain.Product, error) {
	var product domain.Product

	// เราใช้ Preload เพื่อสั่งให้ GORM ดึงข้อมูลจากตาราง 'Category' และ 'Images'
	// ที่มีความสัมพันธ์กับ Product ชิ้นนี้มาด้วยในคราวเดียว
	err := r.db.Preload("Category").Preload("Images").First(&product, id).Error

	// ถ้า GORM หาข้อมูลไม่เจอ มันจะคืนค่า error gorm.ErrRecordNotFound
	// เราจะส่ง error นี้กลับไปให้ Service Layer จัดการต่อ
	return &product, err
}

func (r *productRepository) FindAll(params domain.QueryParams) ([]domain.Product, error) {
	var products []domain.Product

	// คำนวณ offset สำหรับการแบ่งหน้า
	offset := (params.Page - 1) * params.Limit

	// สร้างสตริงสำหรับ Order By
	orderBy := params.SortBy + " " + params.Order

	// ใช้ .Offset(), .Limit(), และ .Order() ของ GORM
	err := r.db.
		Preload("Category").
		Preload("Images").
		Offset(offset).
		Limit(params.Limit).
		Order(orderBy).
		Find(&products).Error

	return products, err
}

func (r *productRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&domain.Product{}).Count(&count).Error
	return count, err
}

func (r *productRepository) Delete(id uint) error {
	// สร้าง transaction object จาก DB connection
	tx := r.db.Delete(&domain.Product{}, id)

	// ตรวจสอบ error ที่อาจเกิดขึ้นระหว่างการ query
	if tx.Error != nil {
		return tx.Error
	}

	// ตรวจสอบว่ามีแถวที่ถูกลบไปจริงๆ หรือไม่
	// ถ้าไม่มีแถวไหนถูกลบเลย แสดงว่าหา record นั้นไม่เจอ
	if tx.RowsAffected == 0 {
		return ErrNotFound // คืนค่าเป็น custom error ของเรา
	}

	return nil
}

func (r *productRepository) Update(id uint, updates map[string]interface{}) error {
	// ใช้ .Model() และ .Updates() เพื่ออัปเดตเฉพาะฟิลด์ที่ส่งมาใน map
	tx := r.db.Model(&domain.Product{}).Where("id = ?", id).Updates(updates)

	// ตรวจสอบ Error จากการ Query ก่อน
	if tx.Error != nil {
		return tx.Error
	}

	// ตรวจสอบว่ามีแถวข้อมูลที่ถูกอัปเดตจริงๆ หรือไม่
	// ถ้าไม่มีเลย (RowsAffected == 0) แสดงว่าหา Product ID นั้นไม่เจอ
	if tx.RowsAffected == 0 {
		return ErrNotFound // คืนค่า custom error ของเรากลับไป
	}

	return nil
}
