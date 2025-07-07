package datastore

import (
	cartRepo "backend/carts/repository"
	orderRepo "backend/orders/repository"
	productRepo "backend/products/repository"
	userRepo "backend/users/repository"
	"gorm.io/gorm"
)

// Repositories คือ struct ที่รวบรวม "Interface" ของ Repository ทั้งหมด
// ที่ต้องการทำงานภายใต้ Transaction เดียวกัน
type Repositories struct {
	User    userRepo.UserRepository
	Product productRepo.ProductRepository
	Cart    cartRepo.CartRepository
	Order   orderRepo.OrderRepository
}

// UnitOfWork คือ Interface หลักที่ Service จะเรียกใช้
type UnitOfWork interface {
	// Execute จะรับฟังก์ชันเข้ามาทำงานภายใน Database Transaction
	Execute(fn func(repos *Repositories) error) error

	// เรามี Getter สำหรับ Repository ที่ไม่เกี่ยวกับ DB Transaction ด้วย (เช่น Azure)
	ProductRepository() productRepo.ProductRepository
	UserRepository() userRepo.UserRepository
	CartRepository() cartRepo.CartRepository
	UploadRepository() UploadRepository
}

// unitOfWork คือ struct ที่ทำงานจริง
type unitOfWork struct {
	db          *gorm.DB
	userRepo    userRepo.UserRepository
	productRepo productRepo.ProductRepository
	cartRepo    cartRepo.CartRepository
	orderRepo   orderRepo.OrderRepository
	uploadRepo  UploadRepository
}

func NewUnitOfWork(db *gorm.DB, uploadRepo UploadRepository) UnitOfWork {
	return &unitOfWork{
		db: db,
		// --- [เพิ่ม] สร้าง repo ทั้งหมดตอนเริ่มต้น และเก็บไว้ ---
		userRepo:    userRepo.NewUserRepository(db),
		productRepo: productRepo.NewProductRepository(db),
		cartRepo:    cartRepo.NewCartRepository(db),
		orderRepo:   orderRepo.NewOrderRepository(db),
		uploadRepo:  uploadRepo,
	}
}

func (u *unitOfWork) Execute(fn func(repos *Repositories) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repos := &Repositories{
			User:    userRepo.NewUserRepository(tx),
			Product: productRepo.NewProductRepository(tx),
			Cart:    cartRepo.NewCartRepository(tx),
			Order:   orderRepo.NewOrderRepository(tx),
		}
		return fn(repos)
	})
}

func (u *unitOfWork) ProductRepository() productRepo.ProductRepository {
	return u.productRepo
}

func (u *unitOfWork) UserRepository() userRepo.UserRepository {
	return u.userRepo
}

func (u *unitOfWork) CartRepository() cartRepo.CartRepository {
	return u.cartRepo
}

func (u *unitOfWork) UploadRepository() UploadRepository {
	return u.uploadRepo
}
