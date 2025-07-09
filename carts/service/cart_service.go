package service

import (
	"backend/carts/dto"
	"backend/carts/repository"
	"backend/domain"
	"backend/internal/datastore"
	"errors"
	"time"
)

var ErrProductNotFound = errors.New("product not found")
var ErrNotEnoughStock = errors.New("not enough stock")
var ErrItemNotInCart = errors.New("item not in user's cart")
var (
	ErrCouponNotFound   = errors.New("coupon not found or is invalid")
	ErrCouponExpired    = errors.New("coupon has expired")
	ErrCouponUsageLimit = errors.New("coupon has reached its usage limit")
)

type CartService interface {
	AddItemToCart(userID uint, req dto.AddItemRequest) (*dto.CartResponse, error)
	GetCart(userID uint) (*dto.CartResponse, error)
	UpdateCartItem(userID, cartItemID uint, quantity uint) (*dto.CartResponse, error)
	RemoveCartItem(userID, cartItemID uint) (*dto.CartResponse, error)
	ApplyCoupon(userID uint, couponCode string) (*dto.CartResponse, error)
	RemoveCoupon(userID uint) (*dto.CartResponse, error)
}

type cartService struct {
	uow          datastore.UnitOfWork
	imageBaseURL string
}

func NewCartService(uow datastore.UnitOfWork, imageBaseURL string) CartService {
	return &cartService{
		uow:          uow,
		imageBaseURL: imageBaseURL,
	}
}

// AddItemToCart เพิ่มสินค้าลงตะกร้า
func (s *cartService) AddItemToCart(userID uint, req dto.AddItemRequest) (*dto.CartResponse, error) {
	// 1. ตรวจสอบว่าสินค้ามีอยู่จริงและมีสต็อกเพียงพอหรือไม่
	product, err := s.uow.ProductRepository().FindProductByID(req.ProductID)
	if err != nil {
		return nil, ErrProductNotFound
	}
	if product.Quantity < int(req.Quantity) {
		return nil, ErrNotEnoughStock
	}

	// 2. หาหรือสร้างตะกร้าสำหรับ User คนนี้
	cart, err := s.uow.CartRepository().GetOrCreateCart(userID)

	if err != nil {
		return nil, err
	}

	// 3. เพิ่ม Item ลงในตะกร้า (Repository จะจัดการเรื่องบวกจำนวนเอง)
	if _, err := s.uow.CartRepository().AddItem(cart.ID, req.ProductID, req.Quantity); err != nil {
		return nil, err
	}

	// 4. ดึงข้อมูลตะกร้าล่าสุดแล้วส่งกลับไป
	return s.GetCart(userID)
}

// GetCart ดึงข้อมูลตะกร้าทั้งหมด
func (s *cartService) GetCart(userID uint) (*dto.CartResponse, error) {
	cart, err := s.uow.CartRepository().GetCartByUserID(userID)
	if err != nil {
		// ถ้าหาไม่เจอ (เช่น user ใหม่) ให้สร้างตะกร้าเปล่าๆ คืนไป
		if errors.Is(err, repository.ErrNotFound) {
			emptyCart := &dto.CartResponse{UserID: userID, Items: []dto.CartItemResponse{}}
			// เราอาจจะสร้าง cart จริงๆ ใน db ไปเลยก็ได้
			newCart, dbErr := s.uow.CartRepository().GetOrCreateCart(userID)
			if dbErr != nil {
				return nil, dbErr
			}
			emptyCart.ID = newCart.ID
			return emptyCart, nil
		}
		return nil, err
	}

	// แปลง Domain Model เป็น DTO
	return s.mapCartToCartResponse(cart), nil
}

// UpdateCartItem อัปเดตจำนวนสินค้า
func (s *cartService) UpdateCartItem(userID, cartItemID uint, quantity uint) (*dto.CartResponse, error) {
	// Logic การตรวจสอบความเป็นเจ้าของควรจะทำที่นี่
	// (เช็คว่า cartItemID นี้อยู่ใน cart ของ userID จริงๆ)
	// ... (ละไว้เพื่อให้โค้ดกระชับ) ...

	if quantity == 0 {
		// ถ้าจำนวนเป็น 0 ให้ลบ Item นั้นทิ้ง
		return s.RemoveCartItem(userID, cartItemID)
	}

	if err := s.uow.CartRepository().UpdateItemQuantity(cartItemID, quantity); err != nil {
		return nil, err
	}

	return s.GetCart(userID)
}

// RemoveCartItem ลบสินค้าออกจากตะกร้า
func (s *cartService) RemoveCartItem(userID, cartItemID uint) (*dto.CartResponse, error) {
	// Logic การตรวจสอบความเป็นเจ้าของควรจะทำที่นี่
	// ...

	if err := s.uow.CartRepository().RemoveItem(cartItemID); err != nil {
		return nil, err
	}
	return s.GetCart(userID)
}

func (s *cartService) mapCartToCartResponse(cart *domain.Cart) *dto.CartResponse {
	var subtotal float64
	var discount float64

	itemResponses := make([]dto.CartItemResponse, 0, len(cart.Items))

	for _, item := range cart.Items {
		var imageURL string
		// ตรวจสอบให้แน่ใจว่า Product และ Images ถูก Preload มาด้วย
		if item.Product.ID != 0 && len(item.Product.Images) > 0 {
			imageURL = s.imageBaseURL + "/" + item.Product.Images[0].Path
		}

		itemResponses = append(itemResponses, dto.CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      item.Product.Name,
			Price:     item.Product.Price,
			Quantity:  item.Quantity,
			ImageURL:  imageURL,
		})
		// คำนวณราคารวมก่อนหักส่วนลด
		subtotal += item.Product.Price * float64(item.Quantity)
	}

	// คำนวณส่วนลดถ้ามีคูปองผูกอยู่
	if cart.Coupon != nil && cart.Coupon.ID != 0 {
		switch cart.Coupon.DiscountType {
		case domain.DiscountTypeFixed:
			discount = cart.Coupon.DiscountValue
		case domain.DiscountTypePercentage:
			discount = subtotal * (cart.Coupon.DiscountValue / 100)
		}
	}

	grandTotal := subtotal - discount
	if grandTotal < 0 {
		grandTotal = 0 // ราคาสุดท้ายต้องไม่ติดลบ
	}

	// สร้าง Response DTO
	response := &dto.CartResponse{
		ID:         cart.ID,
		UserID:     cart.UserID,
		Items:      itemResponses,
		Subtotal:   subtotal,
		Discount:   discount,
		GrandTotal: grandTotal,
	}

	// เพิ่มโค้ดคูปองเข้าไปใน Response ถ้ามี
	if cart.Coupon != nil && cart.Coupon.ID != 0 {
		response.AppliedCoupon = &cart.Coupon.Code
	}

	return response
}

func (s *cartService) ApplyCoupon(userID uint, couponCode string) (*dto.CartResponse, error) {
	var cart *domain.Cart
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		// 1. หา Coupon
		coupon, err := repos.Coupon.FindByCode(couponCode)
		if err != nil {
			return ErrCouponNotFound
		}

		// 2. ตรวจสอบเงื่อนไข Coupon
		if !coupon.IsActive || time.Now().After(coupon.ExpiryDate) {
			return ErrCouponExpired
		}
		if coupon.UsageCount >= coupon.UsageLimit {
			return ErrCouponUsageLimit
		}

		// 3. หาตะกร้าของผู้ใช้
		cart, err = repos.Cart.GetOrCreateCart(userID)
		if err != nil {
			return err
		}

		// 4. ผูก Coupon กับ Cart
		cart.CouponID = &coupon.ID
		return repos.Cart.Update(cart) // ต้องเพิ่ม Update ใน Repo
	})

	if err != nil {
		return nil, err
	}
	return s.GetCart(userID)
}

func (s *cartService) RemoveCoupon(userID uint) (*dto.CartResponse, error) {
	// var cart *domain.Cart // <-- ลบบรรทัดนี้ทิ้ง

	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		cart, err := repos.Cart.GetCartByUserID(userID)
		if err != nil {
			return err
		}

		cart.CouponID = nil
		cart.Coupon = nil // อาจจะต้องเคลียร์ relation ด้วย
		return repos.Cart.Update(cart)
	})

	if err != nil {
		return nil, err
	}
	return s.GetCart(userID)
}
