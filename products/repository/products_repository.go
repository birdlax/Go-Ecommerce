package repository

import (
	"backend/domain"
	"backend/products/dto"
	"errors"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

type ProductRepository interface {
	CreateProduct(product *domain.Product) error
	CreateImages(images []domain.ProductImage) error

	FindProductByID(id uint) (*domain.Product, error)
	FindAll(params dto.QueryParams) ([]domain.Product, error)
	Count(params dto.QueryParams) (int64, error)
	Delete(id uint) error
	Update(id uint, updates map[string]interface{}) error
	// Method สำหรับจัดการ ProductImage
	FindImageByID(id uint) (*domain.ProductImage, error)
	FindImagesByIDs(ids []uint) ([]domain.ProductImage, error)
	DeleteImagesByIDs(ids []uint) error
	FindByID(id uint) (*domain.Product, error)
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

func (r *productRepository) FindProductByID(id uint) (*domain.Product, error) {
	var product domain.Product

	// เราใช้ Preload เพื่อสั่งให้ GORM ดึงข้อมูลจากตาราง 'Category' และ 'Images'
	// ที่มีความสัมพันธ์กับ Product ชิ้นนี้มาด้วยในคราวเดียว
	err := r.db.Preload("Category").Preload("Images").First(&product, id).Error

	// ถ้า GORM หาข้อมูลไม่เจอ มันจะคืนค่า error gorm.ErrRecordNotFound
	// เราจะส่ง error นี้กลับไปให้ Service Layer จัดการต่อ
	return &product, err
}

func (r *productRepository) FindAll(params dto.QueryParams) ([]domain.Product, error) {
	var products []domain.Product
	offset := (params.Page - 1) * params.Limit
	orderBy := params.SortBy + " " + params.Order

	// เริ่มสร้าง Query
	query := r.db.Model(&domain.Product{}).Preload("Category").Preload("Images")

	// --- เพิ่ม Logic การ Filter แบบไดนามิก ---
	if params.Search != "" {
		query = query.Where("name LIKE ?", "%"+params.Search+"%")
	}
	if params.CategoryID != 0 {
		query = query.Where("category_id = ?", params.CategoryID)
	}
	if params.MinPrice > 0 {
		query = query.Where("price >= ?", params.MinPrice)
	}
	if params.MaxPrice > 0 {
		query = query.Where("price <= ?", params.MaxPrice)
	}

	// ทำ Pagination และ Sorting ต่อท้าย
	err := query.Offset(offset).Limit(params.Limit).Order(orderBy).Find(&products).Error
	return products, err
}

func (r *productRepository) Count(params dto.QueryParams) (int64, error) {
	var count int64
	query := r.db.Model(&domain.Product{})

	// --- เพิ่ม Logic การ Filter ให้ตรงกับ FindAll ---
	if params.Search != "" {
		query = query.Where("name LIKE ?", "%"+params.Search+"%")
	}
	// ... เพิ่มเงื่อนไข filter อื่นๆ ให้ครบ ...

	err := query.Count(&count).Error
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

func (r *productRepository) FindImageByID(id uint) (*domain.ProductImage, error) {
	var image domain.ProductImage
	err := r.db.First(&image, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &image, nil
}

func (r *productRepository) FindImagesByIDs(ids []uint) ([]domain.ProductImage, error) {
	var images []domain.ProductImage
	err := r.db.Where("id IN ?", ids).Find(&images).Error
	return images, err
}

// DeleteImagesByIDs ลบรูปภาพหลายใบจาก ID ที่ระบุ (Soft Delete)
func (r *productRepository) DeleteImagesByIDs(ids []uint) error {
	// ใช้ gorm.Model จะทำการ Soft Delete โดยอัตโนมัติ
	return r.db.Delete(&domain.ProductImage{}, ids).Error
}

func (r *productRepository) FindByID(id uint) (*domain.Product, error) {
	var product domain.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &product, nil
}
