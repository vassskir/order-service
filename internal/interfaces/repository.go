package interfaces

import "L0/internal/models"

type OrderRepository interface {
	SaveOrder(order *models.Order) error

	GetOrderByUID(orderUID string) (*models.Order, error)

	GetAllOrders() ([]*models.Order, error)

	Close() error
}
