package handler

import (
	"backend/coupons/dto"
	"backend/coupons/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type CouponHandler struct {
	couponSvc service.CouponService
}

func NewCouponHandler(couponSvc service.CouponService) *CouponHandler {
	return &CouponHandler{couponSvc: couponSvc}
}

func (h *CouponHandler) HandleCreateCoupon(c *fiber.Ctx) error {
	var req dto.CouponRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if err := validator.New().Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	res, err := h.couponSvc.Create(req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

func (h *CouponHandler) HandleGetAllCoupons(c *fiber.Ctx) error {
	res, err := h.couponSvc.GetAll()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *CouponHandler) HandleGetCouponByID(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid coupon ID")
	}
	res, err := h.couponSvc.GetByID(uint(id))
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *CouponHandler) HandleUpdateCoupon(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid coupon ID")
	}
	var req dto.CouponRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}
	if err := validator.New().Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed: "+err.Error())
	}
	res, err := h.couponSvc.Update(uint(id), req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(res)
}

func (h *CouponHandler) HandleDeleteCoupon(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid coupon ID")
	}
	if err := h.couponSvc.Delete(uint(id)); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
