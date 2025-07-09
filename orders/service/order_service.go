package service

import (
	"backend/domain"
	"backend/internal/datastore"
	"backend/orders/dto"
	"backend/orders/repository"
	"context"
	"errors"
	"fmt"
)

var (
	ErrCartIsEmpty        = errors.New("cannot create order from an empty cart")
	ErrProductOutOfStock  = errors.New("a product in the cart is out of stock")
	ErrOrderAccessDenied  = errors.New("you do not have permission to view this order")
	ErrInvalidOrderStatus = errors.New("order status is not valid for this operation")
)

type OrderService interface {
	CreateOrderFromCart(ctx context.Context, userID uint, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetMyOrders(userID uint) ([]dto.OrderResponse, error)
	GetOrderByID(userID, orderID uint) (*dto.OrderResponse, error)
	ConfirmPayment(orderID uint) error
	UpdateOrderStatus(orderID uint, status domain.OrderStatus, paymentMethod string) error
	ShipOrder(orderID uint, trackingNumber string) error
}

type orderService struct {
	uow datastore.UnitOfWork
}

func NewOrderService(uow datastore.UnitOfWork) OrderService {
	return &orderService{uow: uow}
}

func (s *orderService) CreateOrderFromCart(ctx context.Context, userID uint, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	var createdOrder *domain.Order

	// ทำทุกอย่างใน Transaction เดียวผ่าน Unit of Work
	err := s.uow.Execute(func(repos *datastore.Repositories) error {
		// 1. ดึงข้อมูลตะกร้าล่าสุดของผู้ใช้
		cart, err := repos.Cart.GetCartByUserID(userID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrCartIsEmpty
			}
			return fmt.Errorf("could not get user cart: %w", err)
		}
		if len(cart.Items) == 0 {
			return ErrCartIsEmpty
		}

		// 2. เตรียมข้อมูล Order และคำนวณราคารวม
		orderItems := make([]domain.OrderItem, 0)
		var totalPrice float64

		for _, cartItem := range cart.Items {
			product, err := repos.Product.FindProductByID(cartItem.ProductID)
			if err != nil {
				return fmt.Errorf("product with id %d not found: %w", cartItem.ProductID, err)
			}

			if product.Quantity < int(cartItem.Quantity) {
				return fmt.Errorf("%w: %s has only %d in stock", ErrProductOutOfStock, product.Name, product.Quantity)
			}

			// [แก้ไข] ลดสต็อกสินค้า
			newQuantity := product.Quantity - int(cartItem.Quantity)
			// [แก้ไข] เรียกใช้เมธอด Update ที่ถูกต้อง
			if err := repos.Product.Update(product.ID, map[string]interface{}{"quantity": newQuantity}); err != nil {
				return fmt.Errorf("failed to update stock for product %d: %w", product.ID, err)
			}

			orderItems = append(orderItems, domain.OrderItem{
				ProductID: product.ID,
				Quantity:  cartItem.Quantity,
				Price:     product.Price,
			})
			totalPrice += product.Price * float64(cartItem.Quantity)
		}

		// 3. สร้าง Order หลัก
		order := &domain.Order{
			UserID:            userID,
			OrderItems:        orderItems,
			TotalPrice:        totalPrice,
			ShippingAddressID: req.ShippingAddressID,
			Status:            domain.StatusPending,
		}

		if err := repos.Order.Create(order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
		createdOrder = order

		// 4. ล้างตะกร้าสินค้า
		if err := repos.Cart.ClearCart(cart.ID); err != nil {
			return fmt.Errorf("failed to clear cart: %w", err)
		}
		return nil // Commit Transaction
	})

	if err != nil {
		return nil, err
	}

	// 5. แปลงข้อมูลเป็น DTO เพื่อส่งกลับ
	return mapOrderToOrderResponse(createdOrder), nil
}

func (s *orderService) GetMyOrders(userID uint) ([]dto.OrderResponse, error) {
	// การอ่านข้อมูลอย่างเดียว สามารถเรียก Repo จาก UoW ได้โดยตรง
	orders, err := s.uow.OrderRepository().FindAllByUserID(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.OrderResponse, 0, len(orders))
	for _, order := range orders {
		responses = append(responses, *mapOrderToOrderResponse(&order))
	}
	return responses, nil
}

func (s *orderService) GetOrderByID(userID, orderID uint) (*dto.OrderResponse, error) {
	order, err := s.uow.OrderRepository().FindByID(orderID)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, err // ส่งต่อ error ที่มีความหมายไปให้ handler
		}
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrOrderAccessDenied
	}
	return mapOrderToOrderResponse(order), nil
}

// Helper function สำหรับแปลงข้อมูล
func mapOrderToOrderResponse(order *domain.Order) *dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(order.OrderItems))
	for _, item := range order.OrderItems {
		items = append(items, dto.OrderItemResponse{
			ProductID: item.ProductID,
			Name:      item.Product.Name,
			Sku:       item.Product.SKU,
			Price:     item.Price,
			Quantity:  item.Quantity,
		})
	}
	return &dto.OrderResponse{
		ID:                order.ID,
		UserID:            order.UserID,
		TotalPrice:        order.TotalPrice,
		Status:            order.Status,
		ShippingAddressID: order.ShippingAddressID,
		CreatedAt:         order.CreatedAt,
		ShippingAddress:   mapAddressToResponse(&order.ShippingAddress),
		PaymentMethod:     order.PaymentMethod,
		Items:             items,
	}
}

func (s *orderService) ConfirmPayment(orderID uint) error {
	return s.uow.Execute(func(repos *datastore.Repositories) error {
		// 1. ค้นหา Order ที่ต้องการ
		order, err := repos.Order.FindByID(orderID)
		if err != nil {
			return err // จะคืนค่า ErrOrderNotFound ถ้าไม่มี
		}

		if order.Status != domain.StatusShipped {
			return fmt.Errorf("%w: current status is '%s'", ErrInvalidOrderStatus, order.Status)
		}

		// 3. อัปเดตสถานะใหม่
		order.Status = domain.StatusCompleted

		// 4. บันทึกการเปลี่ยนแปลงลง Database
		return repos.Order.Update(order)
	})
}

func (s *orderService) UpdateOrderStatus(orderID uint, status domain.OrderStatus, paymentMethod string) error {
	return s.uow.Execute(func(repos *datastore.Repositories) error {
		order, err := repos.Order.FindByID(orderID)
		if err != nil {
			return err
		}

		if status == domain.StatusProcessing {
			order.Status = domain.StatusProcessing
		}

		order.Status = status
		order.PaymentMethod = &paymentMethod
		return repos.Order.Update(order)
	})
}

func (s *orderService) ShipOrder(orderID uint, trackingNumber string) error {
	return s.uow.Execute(func(repos *datastore.Repositories) error {
		order, err := repos.Order.FindByID(orderID)
		if err != nil {
			return err
		}

		if order.Status != domain.StatusProcessing {
			return fmt.Errorf("%w: cannot ship order in status '%s'", ErrInvalidOrderStatus, order.Status)
		}

		order.Status = domain.StatusShipped
		order.TrackingNumber = &trackingNumber

		return repos.Order.Update(order)
	})
}
func mapAddressToResponse(address *domain.Address) *dto.AddressResponse {
	// ป้องกันกรณีที่ Address เป็น nil
	if address == nil || address.ID == 0 {
		return nil
	}

	return &dto.AddressResponse{
		ID:           address.ID,
		AddressLine1: address.AddressLine1,
		AddressLine2: address.AddressLine2,
		City:         address.City,
		State:        address.State,
		PostalCode:   address.PostalCode,
		Country:      address.Country,
		IsDefault:    address.IsDefault,
	}
}
