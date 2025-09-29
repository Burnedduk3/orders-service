package ports

import (
	"context"
	"orders-service/internal/domain/entities"
)

// OrderRepository defines the interface for order persistence operations
type OrderRepository interface {
	// Create creates a new order in the repository
	Create(ctx context.Context, order *entities.Order) (*entities.Order, error)

	// GetByID retrieves an order by its ID
	GetByID(ctx context.Context, id uint) (*entities.Order, error)

	// Update updates an existing order
	Update(ctx context.Context, order *entities.Order) (*entities.Order, error)

	// Delete soft deletes an order by ID
	Delete(ctx context.Context, id uint) error

	// List retrieves a paginated list of all orders
	List(ctx context.Context, limit, offset int) ([]*entities.Order, error)

	// GetByCustomerID retrieves all orders for a specific customer
	GetByCustomerID(ctx context.Context, customerID uint, limit, offset int) ([]*entities.Order, error)

	// GetByStatus retrieves orders by status
	GetByStatus(ctx context.Context, status entities.OrderStatus, limit, offset int) ([]*entities.Order, error)

	// Count returns the total number of orders
	Count(ctx context.Context) (int64, error)

	// CountByCustomerID returns the total number of orders for a customer
	CountByCustomerID(ctx context.Context, customerID uint) (int64, error)

	// CountByStatus returns the total number of orders with a specific status
	CountByStatus(ctx context.Context, status entities.OrderStatus) (int64, error)
}
