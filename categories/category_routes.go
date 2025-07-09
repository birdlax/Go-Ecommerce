package categories

import (
	"backend/categories/handler"
	"backend/categories/service"
	"backend/config"
	"backend/internal/datastore"
	"backend/middleware"
	"github.com/gofiber/fiber/v2"
	"log"
)

func RegisterModule(api fiber.Router, uow datastore.UnitOfWork, cfg *config.Config) {
	categorySvc := service.NewCategoryService(uow)
	categoryHdl := handler.NewCategoryHandler(categorySvc)

	categoriesAPI := api.Group("/categories")

	// GET สามารถเป็น Public
	categoriesAPI.Get("/", categoryHdl.HandleGetAllCategories)
	categoriesAPI.Get("/:id", categoryHdl.HandleGetCategoryByID)

	// POST, PATCH, DELETE สำหรับ Admin
	adminAPI := categoriesAPI.Group("", middleware.Protected(), middleware.AdminRequired())
	adminAPI.Post("/", categoryHdl.HandleCreateCategory)
	adminAPI.Patch("/:id", categoryHdl.HandleUpdateCategory)
	adminAPI.Delete("/:id", categoryHdl.HandleDeleteCategory)

	log.Println("✅ Category module registered successfully.")
}
