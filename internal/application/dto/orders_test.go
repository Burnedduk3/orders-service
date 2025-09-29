package dto

import (
	"encoding/json"
	"testing"
	"time"

	"orders-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrderRequestDTO_ToEntity(t *testing.T) {
	tests := []struct {
		name          string
		dto           CreateOrderRequestDTO
		expectError   bool
		errorContains string
	}{
		{
			name: "valid conversion without items",
			dto: CreateOrderRequestDTO{
				CustomerID: 123,
				Items:      []CreateOrderItemDTO{},
			},
			expectError: false,
		},
		{
			name: "valid conversion with items",
			dto: CreateOrderRequestDTO{
				CustomerID: 123,
				Items: []CreateOrderItemDTO{
					{
						ProductID:   1,
						ProductSKU:  "SKU-001",
						ProductName: "Product 1",
						Quantity:    2,
						UnitPrice:   10.50,
					},
					{
						ProductID:   2,
						ProductSKU:  "SKU-002",
						ProductName: "Product 2",
						Quantity:    1,
						UnitPrice:   25.00,
					},
				},
			},
			expectError: false,
		},
		{
			name: "invalid customer ID",
			dto: CreateOrderRequestDTO{
				CustomerID: 0,
				Items:      []CreateOrderItemDTO{},
			},
			expectError:   true,
			errorContains: "customer ID is required",
		},
		{
			name: "invalid item - missing product ID",
			dto: CreateOrderRequestDTO{
				CustomerID: 123,
				Items: []CreateOrderItemDTO{
					{
						ProductID:   0,
						ProductSKU:  "SKU-001",
						ProductName: "Product 1",
						Quantity:    1,
						UnitPrice:   10.0,
					},
				},
			},
			expectError:   true,
			errorContains: "product ID is required",
		},
		{
			name: "invalid item - negative quantity",
			dto: CreateOrderRequestDTO{
				CustomerID: 123,
				Items: []CreateOrderItemDTO{
					{
						ProductID:   1,
						ProductSKU:  "SKU-001",
						ProductName: "Product 1",
						Quantity:    -1,
						UnitPrice:   10.0,
					},
				},
			},
			expectError:   true,
			errorContains: "quantity must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := tt.dto.ToEntity()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, entity)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entity)
				assert.Equal(t, tt.dto.CustomerID, entity.CustomerID)
				assert.Equal(t, entities.OrderStatusPending, entity.Status)
				assert.Equal(t, len(tt.dto.Items), len(entity.Items))

				// Verify items were added correctly
				for i, itemDTO := range tt.dto.Items {
					assert.Equal(t, itemDTO.ProductID, entity.Items[i].ProductID)
					assert.Equal(t, itemDTO.Quantity, entity.Items[i].Quantity)
					assert.Equal(t, itemDTO.UnitPrice, entity.Items[i].UnitPrice)
				}
			}
		})
	}
}

func TestAddOrderItemRequestDTO_ToOrderItem(t *testing.T) {
	tests := []struct {
		name          string
		dto           AddOrderItemRequestDTO
		expectError   bool
		errorContains string
	}{
		{
			name: "valid conversion",
			dto: AddOrderItemRequestDTO{
				ProductID:   1,
				ProductSKU:  "SKU-001",
				ProductName: "Product 1",
				Quantity:    5,
				UnitPrice:   19.99,
			},
			expectError: false,
		},
		{
			name: "invalid - empty SKU",
			dto: AddOrderItemRequestDTO{
				ProductID:   1,
				ProductSKU:  "",
				ProductName: "Product 1",
				Quantity:    1,
				UnitPrice:   10.0,
			},
			expectError:   true,
			errorContains: "product SKU is required",
		},
		{
			name: "invalid - zero price",
			dto: AddOrderItemRequestDTO{
				ProductID:   1,
				ProductSKU:  "SKU-001",
				ProductName: "Product 1",
				Quantity:    1,
				UnitPrice:   0.0,
			},
			expectError:   true,
			errorContains: "unit price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := tt.dto.ToOrderItem()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
				assert.Equal(t, tt.dto.ProductID, item.ProductID)
				assert.Equal(t, tt.dto.ProductSKU, item.ProductSKU)
				assert.Equal(t, tt.dto.ProductName, item.ProductName)
				assert.Equal(t, tt.dto.Quantity, item.Quantity)
				assert.Equal(t, tt.dto.UnitPrice, item.UnitPrice)
				assert.Equal(t, float64(tt.dto.Quantity)*tt.dto.UnitPrice, item.TotalPrice)
			}
		})
	}
}

func TestOrderToResponseDTO(t *testing.T) {
	// Given
	now := time.Now()
	order := &entities.Order{
		ID:         1,
		CustomerID: 123,
		Items: []entities.OrderItem{
			{
				ID:          1,
				ProductID:   1,
				ProductSKU:  "SKU-001",
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   10.0,
				TotalPrice:  20.0,
			},
			{
				ID:          2,
				ProductID:   2,
				ProductSKU:  "SKU-002",
				ProductName: "Product 2",
				Quantity:    1,
				UnitPrice:   15.0,
				TotalPrice:  15.0,
			},
		},
		TotalAmount: 35.0,
		Status:      entities.OrderStatusConfirmed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// When
	dto := OrderToResponseDTO(order)

	// Then
	assert.NotNil(t, dto)
	assert.Equal(t, order.ID, dto.ID)
	assert.Equal(t, order.CustomerID, dto.CustomerID)
	assert.Len(t, dto.Items, 2)
	assert.Equal(t, 2, dto.ItemCount)
	assert.Equal(t, 3, dto.TotalItems) // 2 + 1
	assert.Equal(t, 35.0, dto.TotalAmount)
	assert.Equal(t, entities.OrderStatusConfirmed, dto.Status)
	assert.Equal(t, now, dto.CreatedAt)
	assert.Equal(t, now, dto.UpdatedAt)

	// Verify items
	assert.Equal(t, uint(1), dto.Items[0].ProductID)
	assert.Equal(t, "SKU-001", dto.Items[0].ProductSKU)
	assert.Equal(t, 2, dto.Items[0].Quantity)
	assert.Equal(t, 10.0, dto.Items[0].UnitPrice)
	assert.Equal(t, 20.0, dto.Items[0].TotalPrice)
}

func TestOrderToSummaryResponseDTO(t *testing.T) {
	// Given
	now := time.Now()
	order := &entities.Order{
		ID:         1,
		CustomerID: 123,
		Items: []entities.OrderItem{
			{Quantity: 2},
			{Quantity: 3},
		},
		TotalAmount: 99.99,
		Status:      entities.OrderStatusShipped,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// When
	dto := OrderToSummaryResponseDTO(order)

	// Then
	assert.NotNil(t, dto)
	assert.Equal(t, order.ID, dto.ID)
	assert.Equal(t, order.CustomerID, dto.CustomerID)
	assert.Equal(t, 2, dto.ItemCount)
	assert.Equal(t, 99.99, dto.TotalAmount)
	assert.Equal(t, entities.OrderStatusShipped, dto.Status)
	assert.Equal(t, now, dto.CreatedAt)
	assert.Equal(t, now, dto.UpdatedAt)
}

func TestOrderItemToResponseDTO(t *testing.T) {
	// Given
	item := entities.OrderItem{
		ID:          1,
		ProductID:   123,
		ProductSKU:  "SKU-ABC",
		ProductName: "Test Product",
		Quantity:    5,
		UnitPrice:   12.50,
		TotalPrice:  62.50,
	}

	// When
	dto := OrderItemToResponseDTO(item)

	// Then
	assert.Equal(t, item.ID, dto.ID)
	assert.Equal(t, item.ProductID, dto.ProductID)
	assert.Equal(t, item.ProductSKU, dto.ProductSKU)
	assert.Equal(t, item.ProductName, dto.ProductName)
	assert.Equal(t, item.Quantity, dto.Quantity)
	assert.Equal(t, item.UnitPrice, dto.UnitPrice)
	assert.Equal(t, item.TotalPrice, dto.TotalPrice)
}

func TestOrdersToResponseDTOs(t *testing.T) {
	// Given
	now := time.Now()
	orders := []*entities.Order{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []entities.OrderItem{{Quantity: 2}},
			TotalAmount: 20.0,
			Status:      entities.OrderStatusPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			CustomerID:  456,
			Items:       []entities.OrderItem{{Quantity: 1}},
			TotalAmount: 50.0,
			Status:      entities.OrderStatusConfirmed,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// When
	dtos := OrdersToResponseDTOs(orders)

	// Then
	assert.Len(t, dtos, 2)
	assert.Equal(t, orders[0].ID, dtos[0].ID)
	assert.Equal(t, orders[0].CustomerID, dtos[0].CustomerID)
	assert.Equal(t, orders[1].ID, dtos[1].ID)
	assert.Equal(t, orders[1].CustomerID, dtos[1].CustomerID)
}

func TestOrdersToSummaryResponseDTOs(t *testing.T) {
	// Given
	now := time.Now()
	orders := []*entities.Order{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []entities.OrderItem{{Quantity: 2}},
			TotalAmount: 20.0,
			Status:      entities.OrderStatusPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			CustomerID:  456,
			Items:       []entities.OrderItem{{Quantity: 3}},
			TotalAmount: 50.0,
			Status:      entities.OrderStatusDelivered,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// When
	dtos := OrdersToSummaryResponseDTOs(orders)

	// Then
	assert.Len(t, dtos, 2)
	assert.Equal(t, orders[0].ID, dtos[0].ID)
	assert.Equal(t, 1, dtos[0].ItemCount)
	assert.Equal(t, orders[1].ID, dtos[1].ID)
	assert.Equal(t, 1, dtos[1].ItemCount)
}

func TestCreateOrderRequestDTO_JSONSerialization(t *testing.T) {
	// Given
	dto := CreateOrderRequestDTO{
		CustomerID: 123,
		Items: []CreateOrderItemDTO{
			{
				ProductID:   1,
				ProductSKU:  "SKU-001",
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   10.99,
			},
		},
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize back
	var decodedDTO CreateOrderRequestDTO
	err = json.Unmarshal(jsonData, &decodedDTO)
	require.NoError(t, err)

	assert.Equal(t, dto.CustomerID, decodedDTO.CustomerID)
	assert.Len(t, decodedDTO.Items, 1)
	assert.Equal(t, dto.Items[0].ProductID, decodedDTO.Items[0].ProductID)
	assert.Equal(t, dto.Items[0].ProductSKU, decodedDTO.Items[0].ProductSKU)
	assert.Equal(t, dto.Items[0].Quantity, decodedDTO.Items[0].Quantity)
	assert.Equal(t, dto.Items[0].UnitPrice, decodedDTO.Items[0].UnitPrice)
}

func TestOrderResponseDTO_JSONSerialization(t *testing.T) {
	// Given
	now := time.Now()
	dto := OrderResponseDTO{
		ID:         1,
		CustomerID: 123,
		Items: []OrderItemResponseDTO{
			{
				ID:          1,
				ProductID:   1,
				ProductSKU:  "SKU-001",
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   10.0,
				TotalPrice:  20.0,
			},
		},
		ItemCount:   1,
		TotalItems:  2,
		TotalAmount: 20.0,
		Status:      entities.OrderStatusConfirmed,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize back
	var decodedDTO OrderResponseDTO
	err = json.Unmarshal(jsonData, &decodedDTO)
	require.NoError(t, err)

	assert.Equal(t, dto.ID, decodedDTO.ID)
	assert.Equal(t, dto.CustomerID, decodedDTO.CustomerID)
	assert.Len(t, decodedDTO.Items, 1)
	assert.Equal(t, dto.ItemCount, decodedDTO.ItemCount)
	assert.Equal(t, dto.TotalItems, decodedDTO.TotalItems)
	assert.Equal(t, dto.TotalAmount, decodedDTO.TotalAmount)
	assert.Equal(t, dto.Status, decodedDTO.Status)
}

func TestUpdateOrderItemQuantityRequestDTO_JSONSerialization(t *testing.T) {
	// Given
	dto := UpdateOrderItemQuantityRequestDTO{
		Quantity: 5,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize back
	var decodedDTO UpdateOrderItemQuantityRequestDTO
	err = json.Unmarshal(jsonData, &decodedDTO)
	require.NoError(t, err)

	assert.Equal(t, dto.Quantity, decodedDTO.Quantity)
}

func TestUpdateOrderStatusRequestDTO_JSONSerialization(t *testing.T) {
	// Given
	dto := UpdateOrderStatusRequestDTO{
		Status: entities.OrderStatusShipped,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize back
	var decodedDTO UpdateOrderStatusRequestDTO
	err = json.Unmarshal(jsonData, &decodedDTO)
	require.NoError(t, err)

	assert.Equal(t, dto.Status, decodedDTO.Status)
}

func TestOrderListResponseDTO_Structure(t *testing.T) {
	// Given
	now := time.Now()
	orders := []*OrderResponseDTO{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []OrderItemResponseDTO{},
			ItemCount:   0,
			TotalAmount: 100.0,
			Status:      entities.OrderStatusPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			CustomerID:  456,
			Items:       []OrderItemResponseDTO{},
			ItemCount:   0,
			TotalAmount: 200.0,
			Status:      entities.OrderStatusConfirmed,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	dto := OrderListResponseDTO{
		Orders:   orders,
		Total:    50,
		Page:     1,
		PageSize: 2,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize and verify structure
	var decoded OrderListResponseDTO
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Orders, 2)
	assert.Equal(t, int64(50), decoded.Total)
	assert.Equal(t, 1, decoded.Page)
	assert.Equal(t, 2, decoded.PageSize)
}

func TestOrderSummaryListResponseDTO_Structure(t *testing.T) {
	// Given
	now := time.Now()
	orders := []*OrderSummaryResponseDTO{
		{
			ID:          1,
			CustomerID:  123,
			ItemCount:   2,
			TotalAmount: 100.0,
			Status:      entities.OrderStatusPending,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			CustomerID:  456,
			ItemCount:   1,
			TotalAmount: 200.0,
			Status:      entities.OrderStatusShipped,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	dto := OrderSummaryListResponseDTO{
		Orders:   orders,
		Total:    100,
		Page:     2,
		PageSize: 2,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize and verify structure
	var decoded OrderSummaryListResponseDTO
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Orders, 2)
	assert.Equal(t, int64(100), decoded.Total)
	assert.Equal(t, 2, decoded.Page)
	assert.Equal(t, 2, decoded.PageSize)
}

func TestOrderItemsToResponseDTOs(t *testing.T) {
	// Given
	items := []entities.OrderItem{
		{
			ID:          1,
			ProductID:   1,
			ProductSKU:  "SKU-001",
			ProductName: "Product 1",
			Quantity:    2,
			UnitPrice:   10.0,
			TotalPrice:  20.0,
		},
		{
			ID:          2,
			ProductID:   2,
			ProductSKU:  "SKU-002",
			ProductName: "Product 2",
			Quantity:    1,
			UnitPrice:   30.0,
			TotalPrice:  30.0,
		},
	}

	// When
	dtos := OrderItemsToResponseDTOs(items)

	// Then
	assert.Len(t, dtos, 2)
	assert.Equal(t, items[0].ID, dtos[0].ID)
	assert.Equal(t, items[0].ProductSKU, dtos[0].ProductSKU)
	assert.Equal(t, items[1].ID, dtos[1].ID)
	assert.Equal(t, items[1].ProductSKU, dtos[1].ProductSKU)
}

func TestCreateOrderRequestDTO_EmptyItems(t *testing.T) {
	// Given - Order without items
	dto := CreateOrderRequestDTO{
		CustomerID: 123,
		Items:      []CreateOrderItemDTO{},
	}

	// When
	entity, err := dto.ToEntity()

	// Then - Should create valid empty order
	assert.NoError(t, err)
	assert.NotNil(t, entity)
	assert.Equal(t, uint(123), entity.CustomerID)
	assert.Empty(t, entity.Items)
	assert.Equal(t, 0.0, entity.TotalAmount)
	assert.True(t, entity.IsEmpty())
}
