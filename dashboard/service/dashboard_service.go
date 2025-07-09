package service

import (
	"backend/dashboard/dto"
	"backend/domain"
	"backend/internal/datastore"
)

type DashboardService interface {
	GetDashboardStats() (*dto.DashboardStatsResponse, error)
}

type dashboardService struct {
	uow datastore.UnitOfWork
}

func NewDashboardService(uow datastore.UnitOfWork) DashboardService {
	return &dashboardService{uow: uow}
}

func (s *dashboardService) GetDashboardStats() (*dto.DashboardStatsResponse, error) {
	var response dto.DashboardStatsResponse
	var err error

	err = s.uow.Execute(func(repos *datastore.Repositories) error {
		salesSummary, err := repos.Dashboard.GetSalesSummary()
		if err != nil {
			return err
		}

		countSummary, err := repos.Dashboard.GetCountSummary()
		if err != nil {
			return err
		}

		recentOrders, err := repos.Dashboard.GetRecentOrders(5)
		if err != nil {
			return err
		}

		lowStock, err := repos.Dashboard.GetLowStockProducts(10, 5)
		if err != nil {
			return err
		}

		response.SalesSummary = *salesSummary
		response.CountSummary = *countSummary
		response.RecentOrders = mapRecentOrdersToResponse(recentOrders)
		response.LowStockProducts = mapLowStockProductsToResponse(lowStock)

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &response, nil
}

func mapRecentOrdersToResponse(orders []domain.Order) []dto.RecentOrderResponse {
	responses := make([]dto.RecentOrderResponse, 0, len(orders))
	for _, order := range orders {
		responses = append(responses, dto.RecentOrderResponse{
			OrderID:      order.ID,
			CustomerName: order.User.FirstName + " " + order.User.LastName,
			TotalPrice:   order.TotalPrice,
			Status:       order.Status,
			CreatedAt:    order.CreatedAt,
		})
	}
	return responses
}

func mapLowStockProductsToResponse(products []domain.Product) []dto.LowStockProductResponse {
	responses := make([]dto.LowStockProductResponse, 0, len(products))
	for _, product := range products {
		responses = append(responses, dto.LowStockProductResponse{
			ProductID:   product.ID,
			ProductName: product.Name,
			SKU:         product.SKU,
			Quantity:    product.Quantity,
		})
	}
	return responses
}
