package handler

import (
	"backend/middleware"
	"backend/orders/dto"
	"backend/orders/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type OrderHandler struct {
	orderSvc service.OrderService
}

func NewOrderHandler(orderSvc service.OrderService) *OrderHandler {
	return &OrderHandler{orderSvc: orderSvc}
}

func (h *OrderHandler) HandleCreateOrder(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)

	var req dto.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if err := validator.New().Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	orderResponse, err := h.orderSvc.CreateOrderFromCart(c.Context(), claims.UserID, req)
	if err != nil {
		// สามารถเช็ค error ประเภทต่างๆ แล้วส่ง status code ที่เหมาะสมได้
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(orderResponse)
}
