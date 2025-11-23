package interfaces

import "L0/internal/models"

type OrderCache interface {
	Set(order *models.Order)

	Get(orderUID string) (*models.Order, bool)

	GetAll() map[string]*models.Order

	LoadFromRepository(getAllOrders func() ([]*models.Order, error)) error

	Count() int
}
