package order_repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"orders-service/internal/application/ports"
	"orders-service/internal/domain/entities"
	domainErrors "orders-service/internal/domain/errors"

	"gorm.io/gorm"
)

// OrderModel represents the database model for orders
type OrderModel struct {
	ID          uint             `gorm:"primarykey"`
	CustomerID  uint             `gorm:"not null;index"`
	Items       []OrderItemModel `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
	TotalAmount float64          `gorm:"type:decimal(10,2);not null;default:0"`
	Status      string           `gorm:"not null;default:'pending';index"`
	CreatedAt   time.Time        `gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time        `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt   `gorm:"index"` // For soft deletes
}

// OrderItemModel represents the database model for order items
type OrderItemModel struct {
	ID          uint      `gorm:"primarykey"`
	OrderID     uint      `gorm:"not null;index"`
	ProductID   uint      `gorm:"not null;index"`
	ProductSKU  string    `gorm:"not null;index"`
	ProductName string    `gorm:"not null"`
	Quantity    int       `gorm:"not null"`
	UnitPrice   float64   `gorm:"type:decimal(10,2);not null"`
	TotalPrice  float64   `gorm:"type:decimal(10,2);not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (OrderModel) TableName() string {
	return "orders"
}

// TableName specifies the table name for GORM
func (OrderItemModel) TableName() string {
	return "order_items"
}

// GormOrderRepository implements the OrderRepository interface using GORM
type GormOrderRepository struct {
	db *gorm.DB
}

// NewGormOrderRepository creates a new GORM order repository
func NewGormOrderRepository(db *gorm.DB) ports.OrderRepository {
	return &GormOrderRepository{db: db}
}

// Create implements ports.OrderRepository
func (r *GormOrderRepository) Create(ctx context.Context, order *entities.Order) (*entities.Order, error) {
	gormModel := r.toModel(order)

	// Create order with items in a transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(gormModel).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, r.handleError(err)
	}

	// Reload to get all generated IDs
	return r.GetByID(ctx, gormModel.ID)
}

// GetByID implements ports.OrderRepository
func (r *GormOrderRepository) GetByID(ctx context.Context, id uint) (*entities.Order, error) {
	var model OrderModel

	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntity(&model), nil
}

// Update implements ports.OrderRepository
func (r *GormOrderRepository) Update(ctx context.Context, order *entities.Order) (*entities.Order, error) {
	gormModel := r.toModel(order)

	// Update order and items in a transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update order fields
		if err := tx.Model(&OrderModel{}).
			Where("id = ?", gormModel.ID).
			Updates(map[string]interface{}{
				"customer_id":  gormModel.CustomerID,
				"total_amount": gormModel.TotalAmount,
				"status":       gormModel.Status,
				"updated_at":   time.Now(),
			}).Error; err != nil {
			return err
		}

		// Delete existing items
		if err := tx.Where("order_id = ?", gormModel.ID).Delete(&OrderItemModel{}).Error; err != nil {
			return err
		}

		// Insert updated items
		if len(gormModel.Items) > 0 {
			for i := range gormModel.Items {
				gormModel.Items[i].OrderID = gormModel.ID
			}
			if err := tx.Create(&gormModel.Items).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.GetByID(ctx, order.ID)
}

// Delete implements ports.OrderRepository
func (r *GormOrderRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&OrderModel{}, id)
	if result.Error != nil {
		return r.handleError(result.Error)
	}

	if result.RowsAffected == 0 {
		return domainErrors.ErrOrderNotFound
	}

	return nil
}

// List implements ports.OrderRepository
func (r *GormOrderRepository) List(ctx context.Context, limit, offset int) ([]*entities.Order, error) {
	var models []OrderModel

	err := r.db.WithContext(ctx).
		Preload("Items").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// GetByCustomerID implements ports.OrderRepository
func (r *GormOrderRepository) GetByCustomerID(ctx context.Context, customerID uint, limit, offset int) ([]*entities.Order, error) {
	var models []OrderModel

	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("customer_id = ?", customerID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// GetByStatus implements ports.OrderRepository
func (r *GormOrderRepository) GetByStatus(ctx context.Context, status entities.OrderStatus, limit, offset int) ([]*entities.Order, error) {
	var models []OrderModel

	err := r.db.WithContext(ctx).
		Preload("Items").
		Where("status = ?", string(status)).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// Count implements ports.OrderRepository
func (r *GormOrderRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&OrderModel{}).Count(&count).Error
	if err != nil {
		return 0, r.handleError(err)
	}
	return count, nil
}

// CountByCustomerID implements ports.OrderRepository
func (r *GormOrderRepository) CountByCustomerID(ctx context.Context, customerID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&OrderModel{}).
		Where("customer_id = ?", customerID).
		Count(&count).Error
	if err != nil {
		return 0, r.handleError(err)
	}
	return count, nil
}

// CountByStatus implements ports.OrderRepository
func (r *GormOrderRepository) CountByStatus(ctx context.Context, status entities.OrderStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&OrderModel{}).
		Where("status = ?", string(status)).
		Count(&count).Error
	if err != nil {
		return 0, r.handleError(err)
	}
	return count, nil
}

// Helper functions for conversion between domain entities and GORM models

func (r *GormOrderRepository) toModel(order *entities.Order) *OrderModel {
	model := &OrderModel{
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		TotalAmount: order.TotalAmount,
		Status:      string(order.Status),
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}

	// Convert items
	if len(order.Items) > 0 {
		model.Items = make([]OrderItemModel, 0, len(order.Items))
		for _, item := range order.Items {
			model.Items = append(model.Items, OrderItemModel{
				ID:          item.ID,
				OrderID:     order.ID,
				ProductID:   item.ProductID,
				ProductSKU:  item.ProductSKU,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				TotalPrice:  item.TotalPrice,
			})
		}
	}

	return model
}

func (r *GormOrderRepository) toEntity(model *OrderModel) *entities.Order {
	order := &entities.Order{
		ID:          model.ID,
		CustomerID:  model.CustomerID,
		TotalAmount: model.TotalAmount,
		Status:      entities.OrderStatus(model.Status),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	// Convert items
	if len(model.Items) > 0 {
		order.Items = make([]entities.OrderItem, 0, len(model.Items))
		for _, item := range model.Items {
			order.Items = append(order.Items, entities.OrderItem{
				ID:          item.ID,
				ProductID:   item.ProductID,
				ProductSKU:  item.ProductSKU,
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				TotalPrice:  item.TotalPrice,
			})
		}
	} else {
		order.Items = make([]entities.OrderItem, 0)
	}

	return order
}

func (r *GormOrderRepository) toEntities(models []OrderModel) []*entities.Order {
	orders := make([]*entities.Order, 0, len(models))
	for _, model := range models {
		orders = append(orders, r.toEntity(&model))
	}
	return orders
}

// Helper to convert GORM errors to domain errors
func (r *GormOrderRepository) handleError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainErrors.ErrOrderNotFound
	}

	// Handle foreign key constraint violations
	if strings.Contains(err.Error(), "foreign key constraint") ||
		strings.Contains(err.Error(), "FOREIGN KEY constraint") {
		return domainErrors.NewOrderValidationError("customer_id", "invalid customer ID")
	}

	// Handle unique constraint violations
	if errors.Is(err, gorm.ErrDuplicatedKey) ||
		(err.Error() != "" && (strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint"))) {
		return domainErrors.ErrOrderAlreadyExists
	}

	// Return wrapped error for other cases
	return err
}
