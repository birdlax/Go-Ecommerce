package repository

import (
	// [สำคัญ] แก้ไข "backend" เป็นชื่อ Module ใน go.mod ของคุณ
	"backend/domain"
	"errors"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

// CartRepository คือ Interface สำหรับจัดการข้อมูล Cart ทั้งหมด
type CartRepository interface {
	GetOrCreateCart(userID uint) (*domain.Cart, error)
	AddItem(cartID, productID uint, quantity uint) (*domain.CartItem, error)
	GetCartByUserID(userID uint) (*domain.Cart, error)
	UpdateItemQuantity(cartItemID uint, quantity uint) error
	RemoveItem(cartItemID uint) error
	ClearCart(cartID uint) error
	FindItemByCartIDAndProductID(cartID, productID uint) (*domain.CartItem, error)
}

type cartRepository struct {
	db *gorm.DB
}

// NewCartRepository คือ Constructor
func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) GetOrCreateCart(userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	// .FirstOrCreate จะหาแถวแรกที่ UserID ตรงกัน ถ้าไม่เจอก็จะสร้างใหม่ให้เลย
	err := r.db.Where(domain.Cart{UserID: userID}).FirstOrCreate(&cart).Error
	return &cart, err
}

func (r *cartRepository) AddItem(cartID, productID uint, quantity uint) (*domain.CartItem, error) {
	// ตรวจสอบก่อนว่าสินค้านี้มีในตะกร้าแล้วหรือยัง
	cartItem, err := r.FindItemByCartIDAndProductID(cartID, productID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err // ถ้าเกิด error อื่นที่ไม่ใช่ "หาไม่เจอ"
	}

	if cartItem.ID != 0 {
		// ถ้ามีอยู่แล้ว ให้อัปเดตจำนวน (บวกเพิ่ม)
		cartItem.Quantity += quantity
		err = r.db.Save(cartItem).Error
	} else {
		// ถ้ายังไม่มี ให้สร้างรายการใหม่
		cartItem = &domain.CartItem{
			CartID:    cartID,
			ProductID: productID,
			Quantity:  quantity,
		}
		err = r.db.Create(cartItem).Error
	}
	return cartItem, err
}

func (r *cartRepository) GetCartByUserID(userID uint) (*domain.Cart, error) {
	var cart domain.Cart
	err := r.db.
		Preload("Items.Product.Category").
		Preload("Items.Product.Images"). // Preload รูปภาพของ Product ในตะกร้าด้วย
		First(&cart, "user_id = ?", userID).Error

	// --- จุดที่แก้ไข ---
	// ตรวจสอบ error ที่นี่
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ถ้าหาไม่เจอ ให้ "แปล" เป็น error ของเราเอง
			return nil, ErrNotFound
		}
		// ถ้าเป็น error อื่น ก็ส่งออกไปตามเดิม
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) UpdateItemQuantity(cartItemID uint, quantity uint) error {
	return r.db.Model(&domain.CartItem{}).Where("id = ?", cartItemID).Update("quantity", quantity).Error
}

func (r *cartRepository) RemoveItem(cartItemID uint) error {
	return r.db.Delete(&domain.CartItem{}, cartItemID).Error
}

func (r *cartRepository) ClearCart(cartID uint) error {
	return r.db.Where("cart_id = ?", cartID).Delete(&domain.CartItem{}).Error
}

func (r *cartRepository) FindItemByCartIDAndProductID(cartID, productID uint) (*domain.CartItem, error) {
	var cartItem domain.CartItem
	err := r.db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&cartItem).Error
	return &cartItem, err
}
