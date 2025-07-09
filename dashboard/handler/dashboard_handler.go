package handler

import (
	"backend/dashboard/service"
	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	dashboardSvc service.DashboardService
}

func NewDashboardHandler(dashboardSvc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardSvc: dashboardSvc}
}

func (h *DashboardHandler) HandleGetDashboardStats(c *fiber.Ctx) error {
	stats, err := h.dashboardSvc.GetDashboardStats()
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(stats)
}
