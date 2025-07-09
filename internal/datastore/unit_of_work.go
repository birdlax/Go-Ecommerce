package datastore

import (
	cartRepo "backend/carts/repository"
	categoryRepo "backend/categories/repository"
	cuponRepo "backend/coupons/repository"
	dashboardRepo "backend/dashboard/repository"
	orderRepo "backend/orders/repository"
	productRepo "backend/products/repository"
	userRepo "backend/users/repository"

	"gorm.io/gorm"
)

// Repositories คือ struct ที่รวบรวม "Interface" ของ Repository ทั้งหมด
// ที่ต้องการทำงานภายใต้ Transaction เดียวกัน
type Repositories struct {
	User      userRepo.UserRepository
	Address   userRepo.AddressRepository
	Product   productRepo.ProductRepository
	Category  categoryRepo.CategoryRepository
	Cart      cartRepo.CartRepository
	Order     orderRepo.OrderRepository
	Coupon    cuponRepo.CouponRepository
	Dashboard dashboardRepo.DashboardRepository
}

// UnitOfWork คือ Interface หลักที่ Service จะเรียกใช้
type UnitOfWork interface {
	// Execute จะรับฟังก์ชันเข้ามาทำงานภายใน Database Transaction
	Execute(fn func(repos *Repositories) error) error

	// เรามี Getter สำหรับ Repository ที่ไม่เกี่ยวกับ DB Transaction ด้วย (เช่น Azure)
	ProductRepository() productRepo.ProductRepository
	CategoryRepository() categoryRepo.CategoryRepository
	UserRepository() userRepo.UserRepository
	CouponRepository() cuponRepo.CouponRepository
	AddressRepository() userRepo.AddressRepository
	CartRepository() cartRepo.CartRepository
	DashboardRepository() dashboardRepo.DashboardRepository
	OrderRepository() orderRepo.OrderRepository
	UploadRepository() UploadRepository
}

// unitOfWork คือ struct ที่ทำงานจริง
type unitOfWork struct {
	db            *gorm.DB
	userRepo      userRepo.UserRepository
	addressRepo   userRepo.AddressRepository
	categoryRepo  categoryRepo.CategoryRepository
	productRepo   productRepo.ProductRepository
	couponRepo    cuponRepo.CouponRepository
	dashboardRepo dashboardRepo.DashboardRepository
	cartRepo      cartRepo.CartRepository
	orderRepo     orderRepo.OrderRepository
	uploadRepo    UploadRepository
}

func NewUnitOfWork(db *gorm.DB, uploadRepo UploadRepository) UnitOfWork {
	return &unitOfWork{
		db: db,
		// --- [เพิ่ม] สร้าง repo ทั้งหมดตอนเริ่มต้น และเก็บไว้ ---
		userRepo:      userRepo.NewUserRepository(db),
		addressRepo:   userRepo.NewAddressRepository(db),
		productRepo:   productRepo.NewProductRepository(db),
		categoryRepo:  categoryRepo.NewCategoryRepository(db),
		cartRepo:      cartRepo.NewCartRepository(db),
		couponRepo:    cuponRepo.NewCouponRepository(db),
		dashboardRepo: dashboardRepo.NewDashboardRepository(db),
		orderRepo:     orderRepo.NewOrderRepository(db),
		uploadRepo:    uploadRepo,
	}
}

func (u *unitOfWork) Execute(fn func(repos *Repositories) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		repos := &Repositories{
			User:      userRepo.NewUserRepository(tx),
			Address:   userRepo.NewAddressRepository(tx),
			Product:   productRepo.NewProductRepository(tx),
			Category:  categoryRepo.NewCategoryRepository(tx),
			Cart:      cartRepo.NewCartRepository(tx),
			Order:     orderRepo.NewOrderRepository(tx),
			Dashboard: dashboardRepo.NewDashboardRepository(tx),
			Coupon:    cuponRepo.NewCouponRepository(tx),
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

func (u *unitOfWork) AddressRepository() userRepo.AddressRepository {
	return u.addressRepo
}

func (u *unitOfWork) CartRepository() cartRepo.CartRepository {
	return u.cartRepo
}

func (u *unitOfWork) UploadRepository() UploadRepository {
	return u.uploadRepo
}
func (u *unitOfWork) OrderRepository() orderRepo.OrderRepository {
	return u.orderRepo
}
func (u *unitOfWork) CategoryRepository() categoryRepo.CategoryRepository {
	return u.categoryRepo
}
func (u *unitOfWork) CouponRepository() cuponRepo.CouponRepository {
	return u.couponRepo
}

func (u *unitOfWork) DashboardRepository() dashboardRepo.DashboardRepository {
	return u.dashboardRepo
}
