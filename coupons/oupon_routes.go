package coupons

import (
	"backend/config"
	"backend/coupons/handler"
	"backend/coupons/service"
	"backend/internal/datastore"
	"backend/middleware"
	"github.com/gofiber/fiber/v2"
	"log"
)

func RegisterModule(api fiber.Router, uow datastore.UnitOfWork, cfg *config.Config) {
	couponSvc := service.NewCouponService(uow)
	couponHdl := handler.NewCouponHandler(couponSvc)

	adminAPI := api.Group("/admin/coupons", middleware.Protected(), middleware.AdminRequired())

	adminAPI.Post("/", couponHdl.HandleCreateCoupon)
	adminAPI.Get("/", couponHdl.HandleGetAllCoupons)
	adminAPI.Get("/:id", couponHdl.HandleGetCouponByID)
	adminAPI.Patch("/:id", couponHdl.HandleUpdateCoupon)
	adminAPI.Delete("/:id", couponHdl.HandleDeleteCoupon)

	log.Println("✅ Coupon module registered successfully.")
}
