package service

import (
	"backend/domain"
	"backend/internal/datastore"
	"backend/orders/dto"
	"context"
	"errors"
	"fmt"
)

var (
	ErrCartIsEmpty       = errors.New("cannot create order from an empty cart")
	ErrProductOutOfStock = errors.New("a product in the cart is out of stock")
	ErrOrderAccessDenied = errors.New("you do not have permission to view this order")
)

type OrderService interface {
	CreateOrderFromCart(ctx context.Context, userID uint, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	GetMyOrders(userID uint) ([]dto.OrderResponse, error)
	GetOrderByID(userID, orderID uint) (*dto.OrderResponse, error)
}

type orderService struct {
	uow datastore.UnitOfWork
}

func NewOrderService(uow datastore.UnitOfWork) OrderService {
	return &orderService{uow: uow}
}

func (s *orderService) CreateOrderFromCart(ctx context.Context, userID uint, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	var createdOrder *domain.Order

	err := s.uow.Execute(func(repos *repository.Repositories) error {
		cart, err := repos.Cart.GetCartByUserID(userID)
		if err != nil || len(cart.Items) == 0 {
			return ErrCartIsEmpty
		}

		orderItems := make([]domain.OrderItem, 0)
		var totalPrice float64

		for _, cartItem := range cart.Items {
			product, err := repos.Product.FindProductByID(cartItem.ProductID)
			if err != nil {
				return fmt.Errorf("product with id %d not found", cartItem.ProductID)
			}
			if product.Quantity < int(cartItem.Quantity) {
				return fmt.Errorf("%w: %s", ErrProductOutOfStock, product.Name)
			}

			product.Quantity -= int(cartItem.Quantity)
			if err := repos.Product.Update(product); err != nil {
				return fmt.Errorf("failed to update stock for product %d: %w", product.ID, err)
			}

			orderItems = append(orderItems, domain.OrderItem{
				ProductID: product.ID,
				Quantity:  cartItem.Quantity,
				Price:     product.Price,
			})
			totalPrice += product.Price * float64(cartItem.Quantity)
		}

		order := &domain.Order{
			UserID:            userID,
			OrderItems:        orderItems,
			TotalPrice:        totalPrice,
			ShippingAddressID: req.ShippingAddressID,
			Status:            domain.StatusProcessing,
		}

		if err := repos.Order.Create(order); err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}
		createdOrder = order

		if err := repos.Cart.ClearCart(cart.ID); err != nil {
			return fmt.Errorf("failed to clear cart: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return mapOrderToOrderResponse(createdOrder), nil
}

func (s *orderService) GetMyOrders(userID uint) ([]dto.OrderResponse, error) {
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
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrOrderAccessDenied
	}
	return mapOrderToOrderResponse(order), nil
}

// Helper function
func mapOrderToOrderResponse(order *domain.Order) *dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0, len(order.OrderItems))
	for _, item := range order.OrderItems {
		items = append(items, dto.OrderItemResponse{
			ProductID: item.ProductID,
			Name:      item.Product.Name,
			Sku:       item.Product.Sku,
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
		Items:             items,
	}
}
