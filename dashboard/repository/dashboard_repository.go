package repository

import (
	"backend/dashboard/dto" // [สำคัญ] แก้ไขชื่อ Module ให้ถูกต้อง
	"backend/domain"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository interface {
	GetSalesSummary() (*dto.SalesSummaryResponse, error)
	GetCountSummary() (*dto.CountSummaryResponse, error)
	GetRecentOrders(limit int) ([]domain.Order, error)
	GetLowStockProducts(threshold, limit int) ([]domain.Product, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

func (r *dashboardRepository) GetSalesSummary() (*dto.SalesSummaryResponse, error) {
	summary := &dto.SalesSummaryResponse{}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	sevenDaysAgo := todayStart.AddDate(0, 0, -7)
	thirtyDaysAgo := todayStart.AddDate(0, 0, -30)

	completedStatus := string(domain.StatusCompleted)

	r.db.Model(&domain.Order{}).Where("created_at >= ? AND status = ?", todayStart, completedStatus).Select("COALESCE(SUM(total_price), 0)").Row().Scan(&summary.Today)
	r.db.Model(&domain.Order{}).Where("created_at >= ? AND status = ?", sevenDaysAgo, completedStatus).Select("COALESCE(SUM(total_price), 0)").Row().Scan(&summary.Last7Days)
	r.db.Model(&domain.Order{}).Where("created_at >= ? AND status = ?", thirtyDaysAgo, completedStatus).Select("COALESCE(SUM(total_price), 0)").Row().Scan(&summary.Last30Days)

	return summary, nil
}

func (r *dashboardRepository) GetCountSummary() (*dto.CountSummaryResponse, error) {
	summary := &dto.CountSummaryResponse{}
	r.db.Model(&domain.User{}).Count(&summary.TotalUsers)
	r.db.Model(&domain.Product{}).Count(&summary.TotalProducts)
	r.db.Model(&domain.Order{}).Count(&summary.TotalOrders)
	return summary, nil
}

func (r *dashboardRepository) GetRecentOrders(limit int) ([]domain.Order, error) {
	var orders []domain.Order
	err := r.db.Preload("User").Order("created_at desc").Limit(limit).Find(&orders).Error
	return orders, err
}

func (r *dashboardRepository) GetLowStockProducts(threshold, limit int) ([]domain.Product, error) {
	var products []domain.Product
	err := r.db.Where("quantity < ?", threshold).Order("quantity asc").Limit(limit).Find(&products).Error
	return products, err
}
