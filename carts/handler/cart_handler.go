package handler

import (
	"backend/carts/dto"
	"backend/carts/service"
	"backend/middleware"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type CartHandler struct {
	cartSvc service.CartService
}

func NewCartHandler(cartSvc service.CartService) *CartHandler {
	return &CartHandler{cartSvc: cartSvc}
}

func (h *CartHandler) HandleGetCart(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)
	cart, err := h.cartSvc.GetCart(claims.UserID)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(cart)
}

func (h *CartHandler) HandleAddItemToCart(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)
	var req dto.AddItemRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if err := validator.New().Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	cart, err := h.cartSvc.AddItemToCart(claims.UserID, req)
	if err != nil {
		return err // ส่งให้ Error Middleware จัดการ
	}
	return c.Status(fiber.StatusCreated).JSON(cart)
}

func (h *CartHandler) HandleUpdateCartItem(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)

	itemId, err := c.ParamsInt("itemId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid item ID")
	}

	var req dto.UpdateItemRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if err := validator.New().Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}

	// ถ้า quantity เป็น 0, service จะทำการลบ item นั้นให้โดยอัตโนมัติ
	updatedCart, err := h.cartSvc.UpdateCartItem(claims.UserID, uint(itemId), req.Quantity)
	if err != nil {
		return err // ส่งต่อให้ Error Middleware จัดการ
	}

	return c.Status(fiber.StatusOK).JSON(updatedCart)
}

func (h *CartHandler) HandleRemoveCartItem(c *fiber.Ctx) error {
	claims := c.Locals("user").(*middleware.JwtClaims)

	itemId, err := c.ParamsInt("itemId")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid item ID")
	}

	updatedCart, err := h.cartSvc.RemoveCartItem(claims.UserID, uint(itemId))
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(updatedCart)
}
