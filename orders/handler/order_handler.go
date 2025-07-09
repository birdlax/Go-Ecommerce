package handler

import (
	"backend/domain"
	"backend/middleware"
	"backend/orders/dto"
	"backend/orders/service"
	"fmt"

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

func (h *OrderHandler) HandleGetMyOrders(c *fiber.Ctx) error {
	// ดึงข้อมูล user ที่ login อยู่จาก token
	claims := c.Locals("user").(*middleware.JwtClaims)

	// เรียก service เพื่อดึงข้อมูล order ทั้งหมด
	orders, err := h.orderSvc.GetMyOrders(claims.UserID)
	if err != nil {
		return err // ส่งต่อให้ Error Middleware จัดการ
	}

	return c.Status(fiber.StatusOK).JSON(orders)
}

// HandleGetOrderByID สำหรับให้ User ดูรายละเอียด Order ของตัวเองทีละรายการ
func (h *OrderHandler) HandleGetOrderByID(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)

	orderID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid order ID")
	}

	order, err := h.orderSvc.GetOrderByID(claims.UserID, uint(orderID))
	if err != nil {
		return err // ส่งต่อให้ Error Middleware จัดการ (ซึ่งจะคืน 404 ถ้าหาไม่เจอ)
	}

	return c.Status(fiber.StatusOK).JSON(order)
}

func (h *OrderHandler) HandleConfirmPayment(c *fiber.Ctx) error {
	orderID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid order ID")
	}

	if err := h.orderSvc.ConfirmPayment(uint(orderID)); err != nil {
		return err // ส่งต่อให้ Error Middleware จัดการ
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("Order %d has been marked as completed.", orderID),
	})
}

func (h *OrderHandler) HandlePaymentWebhook(c *fiber.Ctx) error {
	var req dto.PaymentWebhookRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid webhook payload")
	}

	// ในระบบจริง จะต้องมีการตรวจสอบ Signature ของ Webhook ก่อนเสมอ
	// เพื่อให้แน่ใจว่า Request นี้มาจาก Payment Gateway จริงๆ

	var newStatus domain.OrderStatus
	if req.Status == "success" {
		newStatus = domain.StatusProcessing // หรือ StatusProcessing ตาม Flow ของคุณ
	} else {
		newStatus = domain.StatusCancelled
	}

	if err := h.orderSvc.UpdateOrderStatus(req.OrderID, newStatus, req.PaymentMethod); err != nil {
		// ถ้าเกิด error ให้ตอบกลับ 500 เพื่อให้ Payment Gateway รู้ว่ามีปัญหา
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update order status")
	}

	return c.SendStatus(fiber.StatusOK) // ตอบกลับ 200 OK เพื่อบอกว่ารับทราบแล้ว
}

func (h *OrderHandler) HandleShipOrder(c *fiber.Ctx) error {
	orderID, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid order ID")
	}

	var req dto.ShipOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.orderSvc.ShipOrder(uint(orderID), req.TrackingNumber); err != nil {
		return err // ส่งต่อให้ Error Middleware จัดการ
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("Order %d has been shipped with tracking number %s", orderID, req.TrackingNumber),
	})
}
