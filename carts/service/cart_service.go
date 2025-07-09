package service

import (
	"backend/carts/dto"
	"backend/carts/repository"
	"backend/domain"
	"backend/internal/datastore"
	"errors"
)

var ErrProductNotFound = errors.New("product not found")
var ErrNotEnoughStock = errors.New("not enough stock")
var ErrItemNotInCart = errors.New("item not in user's cart")

type CartService interface {
	AddItemToCart(userID uint, req dto.AddItemRequest) (*dto.CartResponse, error)
	GetCart(userID uint) (*dto.CartResponse, error)
	UpdateCartItem(userID, cartItemID uint, quantity uint) (*dto.CartResponse, error)
	RemoveCartItem(userID, cartItemID uint) (*dto.CartResponse, error)
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
			emptyCart := &dto.CartResponse{UserID: userID, Items: []dto.CartItemResponse{}, TotalPrice: 0}
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

// mapCartToCartResponse คือ helper function สำหรับแปลงข้อมูล
func (s *cartService) mapCartToCartResponse(cart *domain.Cart) *dto.CartResponse {
	var totalPrice float64
	itemResponses := make([]dto.CartItemResponse, 0, len(cart.Items))

	for _, item := range cart.Items {
		var imageURL string
		if len(item.Product.Images) > 0 {
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
		totalPrice += item.Product.Price * float64(item.Quantity)
	}

	return &dto.CartResponse{
		ID:         cart.ID,
		UserID:     cart.UserID,
		Items:      itemResponses,
		TotalPrice: totalPrice,
	}
}
