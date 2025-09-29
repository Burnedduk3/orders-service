package dto

import (
	"orders-service/internal/domain/entities"
	"time"
)

// CreateOrderRequestDTO for order creation
type CreateOrderRequestDTO struct {
	CustomerID uint                 `json:"customer_id" validate:"required,min=1"`
	Items      []CreateOrderItemDTO `json:"items" validate:"omitempty,dive"`
}

// CreateOrderItemDTO for adding items when creating an order
type CreateOrderItemDTO struct {
	ProductID   uint    `json:"product_id" validate:"required,min=1"`
	ProductSKU  string  `json:"product_sku" validate:"required,min=1,max=100"`
	ProductName string  `json:"product_name" validate:"required,min=1,max=255"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

// AddOrderItemRequestDTO for adding a single item to an existing order
type AddOrderItemRequestDTO struct {
	ProductID   uint    `json:"product_id" validate:"required,min=1"`
	ProductSKU  string  `json:"product_sku" validate:"required,min=1,max=100"`
	ProductName string  `json:"product_name" validate:"required,min=1,max=255"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

// UpdateOrderItemQuantityRequestDTO for updating item quantity
type UpdateOrderItemQuantityRequestDTO struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}

// UpdateOrderStatusRequestDTO for updating order status
type UpdateOrderStatusRequestDTO struct {
	Status entities.OrderStatus `json:"status" validate:"required,oneof=pending confirmed processing shipped delivered cancelled refunded"`
}

// OrderItemResponseDTO for order item responses
type OrderItemResponseDTO struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	ProductSKU  string  `json:"product_sku"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

// OrderResponseDTO for order responses
type OrderResponseDTO struct {
	ID          uint                   `json:"id"`
	CustomerID  uint                   `json:"customer_id"`
	Items       []OrderItemResponseDTO `json:"items"`
	ItemCount   int                    `json:"item_count"`
	TotalItems  int                    `json:"total_items"`
	TotalAmount float64                `json:"total_amount"`
	Status      entities.OrderStatus   `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// OrderSummaryResponseDTO for lightweight order list responses
type OrderSummaryResponseDTO struct {
	ID          uint                 `json:"id"`
	CustomerID  uint                 `json:"customer_id"`
	ItemCount   int                  `json:"item_count"`
	TotalAmount float64              `json:"total_amount"`
	Status      entities.OrderStatus `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// OrderListResponseDTO for paginated order lists
type OrderListResponseDTO struct {
	Orders   []*OrderResponseDTO `json:"orders"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// OrderSummaryListResponseDTO for lightweight paginated order lists
type OrderSummaryListResponseDTO struct {
	Orders   []*OrderSummaryResponseDTO `json:"orders"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"page_size"`
}

// Conversion methods - Request DTOs to Domain Entities

func (dto *CreateOrderRequestDTO) ToEntity() (*entities.Order, error) {
	order, err := entities.NewOrder(dto.CustomerID)
	if err != nil {
		return nil, err
	}

	// Add items if provided
	for _, item := range dto.Items {
		err := order.AddItem(
			item.ProductID,
			item.ProductSKU,
			item.ProductName,
			item.Quantity,
			item.UnitPrice,
		)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
}

func (dto *AddOrderItemRequestDTO) ToOrderItem() (*entities.OrderItem, error) {
	return entities.NewOrderItem(
		dto.ProductID,
		dto.ProductSKU,
		dto.ProductName,
		dto.Quantity,
		dto.UnitPrice,
	)
}

// Conversion methods - Domain Entities to Response DTOs

func OrderToResponseDTO(order *entities.Order) *OrderResponseDTO {
	return &OrderResponseDTO{
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		Items:       OrderItemsToResponseDTOs(order.Items),
		ItemCount:   order.GetItemCount(),
		TotalItems:  order.GetTotalQuantity(),
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

func OrderToSummaryResponseDTO(order *entities.Order) *OrderSummaryResponseDTO {
	return &OrderSummaryResponseDTO{
		ID:          order.ID,
		CustomerID:  order.CustomerID,
		ItemCount:   order.GetItemCount(),
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		CreatedAt:   order.CreatedAt,
		UpdatedAt:   order.UpdatedAt,
	}
}

func OrderItemToResponseDTO(item entities.OrderItem) OrderItemResponseDTO {
	return OrderItemResponseDTO{
		ID:          item.ID,
		ProductID:   item.ProductID,
		ProductSKU:  item.ProductSKU,
		ProductName: item.ProductName,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		TotalPrice:  item.TotalPrice,
	}
}

func OrderItemsToResponseDTOs(items []entities.OrderItem) []OrderItemResponseDTO {
	dtos := make([]OrderItemResponseDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, OrderItemToResponseDTO(item))
	}
	return dtos
}

func OrdersToResponseDTOs(orders []*entities.Order) []*OrderResponseDTO {
	dtos := make([]*OrderResponseDTO, 0, len(orders))
	for _, order := range orders {
		dtos = append(dtos, OrderToResponseDTO(order))
	}
	return dtos
}

func OrdersToSummaryResponseDTOs(orders []*entities.Order) []*OrderSummaryResponseDTO {
	dtos := make([]*OrderSummaryResponseDTO, 0, len(orders))
	for _, order := range orders {
		dtos = append(dtos, OrderToSummaryResponseDTO(order))
	}
	return dtos
}
