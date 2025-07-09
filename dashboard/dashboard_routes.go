package dashboard

import (
	"backend/config"
	"backend/dashboard/handler"
	"backend/dashboard/service"
	"backend/internal/datastore"
	"backend/middleware"
	"github.com/gofiber/fiber/v2"
	"log"
)

func RegisterModule(api fiber.Router, uow datastore.UnitOfWork, cfg *config.Config) {
	dashboardSvc := service.NewDashboardService(uow)
	dashboardHdl := handler.NewDashboardHandler(dashboardSvc)

	adminAPI := api.Group("/admin", middleware.Protected(), middleware.AdminRequired())
	adminAPI.Get("/dashboard/stats", dashboardHdl.HandleGetDashboardStats)

	log.Println("✅ Dashboard module registered successfully.")
}
