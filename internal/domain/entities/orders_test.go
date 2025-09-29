package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOrder(t *testing.T) {
	tests := []struct {
		name          string
		customerID    uint
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid order creation",
			customerID:  123,
			expectError: false,
		},
		{
			name:          "zero customer ID",
			customerID:    0,
			expectError:   true,
			errorContains: "customer ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(tt.customerID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.customerID, order.CustomerID)
				assert.Equal(t, OrderStatusPending, order.Status)
				assert.Equal(t, 0.0, order.TotalAmount)
				assert.Empty(t, order.Items)
				assert.WithinDuration(t, time.Now(), order.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), order.UpdatedAt, time.Second)
			}
		})
	}
}

func TestNewOrderItem(t *testing.T) {
	tests := []struct {
		name          string
		productID     uint
		productSKU    string
		productName   string
		quantity      int
		unitPrice     float64
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid order item creation",
			productID:   1,
			productSKU:  "SKU-001",
			productName: "Test Product",
			quantity:    2,
			unitPrice:   10.50,
			expectError: false,
		},
		{
			name:          "zero product ID",
			productID:     0,
			productSKU:    "SKU-001",
			productName:   "Test Product",
			quantity:      1,
			unitPrice:     10.0,
			expectError:   true,
			errorContains: "product ID is required",
		},
		{
			name:          "empty product SKU",
			productID:     1,
			productSKU:    "",
			productName:   "Test Product",
			quantity:      1,
			unitPrice:     10.0,
			expectError:   true,
			errorContains: "product SKU is required",
		},
		{
			name:          "empty product name",
			productID:     1,
			productSKU:    "SKU-001",
			productName:   "",
			quantity:      1,
			unitPrice:     10.0,
			expectError:   true,
			errorContains: "product name is required",
		},
		{
			name:          "zero quantity",
			productID:     1,
			productSKU:    "SKU-001",
			productName:   "Test Product",
			quantity:      0,
			unitPrice:     10.0,
			expectError:   true,
			errorContains: "quantity must be positive",
		},
		{
			name:          "negative quantity",
			productID:     1,
			productSKU:    "SKU-001",
			productName:   "Test Product",
			quantity:      -1,
			unitPrice:     10.0,
			expectError:   true,
			errorContains: "quantity must be positive",
		},
		{
			name:          "zero unit price",
			productID:     1,
			productSKU:    "SKU-001",
			productName:   "Test Product",
			quantity:      1,
			unitPrice:     0.0,
			expectError:   true,
			errorContains: "unit price must be positive",
		},
		{
			name:          "negative unit price",
			productID:     1,
			productSKU:    "SKU-001",
			productName:   "Test Product",
			quantity:      1,
			unitPrice:     -10.0,
			expectError:   true,
			errorContains: "unit price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewOrderItem(tt.productID, tt.productSKU, tt.productName, tt.quantity, tt.unitPrice)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, item)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, item)
				assert.Equal(t, tt.productID, item.ProductID)
				assert.Equal(t, tt.productSKU, item.ProductSKU)
				assert.Equal(t, tt.productName, item.ProductName)
				assert.Equal(t, tt.quantity, item.Quantity)
				assert.Equal(t, tt.unitPrice, item.UnitPrice)
				assert.Equal(t, float64(tt.quantity)*tt.unitPrice, item.TotalPrice)
			}
		})
	}
}

func TestOrder_AddItem(t *testing.T) {
	tests := []struct {
		name          string
		orderStatus   OrderStatus
		productID     uint
		productSKU    string
		productName   string
		quantity      int
		unitPrice     float64
		expectError   bool
		errorContains string
	}{
		{
			name:        "add item to pending order",
			orderStatus: OrderStatusPending,
			productID:   1,
			productSKU:  "SKU-001",
			productName: "Test Product",
			quantity:    2,
			unitPrice:   10.50,
			expectError: false,
		},
		{
			name:          "add item to cancelled order",
			orderStatus:   OrderStatusCancelled,
			productID:     1,
			productSKU:    "SKU-001",
			productName:   "Test Product",
			quantity:      1,
			unitPrice:     10.0,
			expectError:   true,
			errorContains: "order cannot be modified",
		},
		{
			name:          "add item with invalid data",
			orderStatus:   OrderStatusPending,
			productID:     0,
			productSKU:    "SKU-001",
			productName:   "Test Product",
			quantity:      1,
			unitPrice:     10.0,
			expectError:   true,
			errorContains: "product ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, _ := NewOrder(123)
			order.Status = tt.orderStatus

			err := order.AddItem(tt.productID, tt.productSKU, tt.productName, tt.quantity, tt.unitPrice)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Len(t, order.Items, 1)
				assert.Equal(t, tt.productID, order.Items[0].ProductID)
				assert.Equal(t, tt.quantity, order.Items[0].Quantity)
				assert.Equal(t, float64(tt.quantity)*tt.unitPrice, order.TotalAmount)
			}
		})
	}
}

func TestOrder_AddItem_UpdateExisting(t *testing.T) {
	order, _ := NewOrder(123)

	// Add initial item
	err := order.AddItem(1, "SKU-001", "Test Product", 2, 10.0)
	assert.NoError(t, err)
	assert.Len(t, order.Items, 1)
	assert.Equal(t, 2, order.Items[0].Quantity)
	assert.Equal(t, 20.0, order.TotalAmount)

	// Add same product again (should update quantity)
	err = order.AddItem(1, "SKU-001", "Test Product", 3, 10.0)
	assert.NoError(t, err)
	assert.Len(t, order.Items, 1)
	assert.Equal(t, 5, order.Items[0].Quantity) // 2 + 3
	assert.Equal(t, 50.0, order.TotalAmount)
}

func TestOrder_RemoveItem(t *testing.T) {
	order, _ := NewOrder(123)
	order.AddItem(1, "SKU-001", "Product 1", 2, 10.0)
	order.AddItem(2, "SKU-002", "Product 2", 1, 15.0)

	tests := []struct {
		name          string
		productID     uint
		expectError   bool
		errorContains string
		expectedItems int
	}{
		{
			name:          "remove existing item",
			productID:     1,
			expectError:   false,
			expectedItems: 1,
		},
		{
			name:          "remove non-existing item",
			productID:     999,
			expectError:   true,
			errorContains: "item not found",
			expectedItems: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialItems := len(order.Items)
			err := order.RemoveItem(tt.productID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Len(t, order.Items, initialItems)
			} else {
				assert.NoError(t, err)
				assert.Len(t, order.Items, tt.expectedItems)
			}
		})
	}
}

func TestOrder_UpdateItemQuantity(t *testing.T) {
	order, _ := NewOrder(123)
	order.AddItem(1, "SKU-001", "Product 1", 2, 10.0)

	tests := []struct {
		name          string
		productID     uint
		quantity      int
		expectError   bool
		errorContains string
	}{
		{
			name:        "update existing item quantity",
			productID:   1,
			quantity:    5,
			expectError: false,
		},
		{
			name:          "update with zero quantity",
			productID:     1,
			quantity:      0,
			expectError:   true,
			errorContains: "quantity must be positive",
		},
		{
			name:          "update non-existing item",
			productID:     999,
			quantity:      3,
			expectError:   true,
			errorContains: "item not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := order.UpdateItemQuantity(tt.productID, tt.quantity)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				item, _ := order.GetItem(tt.productID)
				assert.Equal(t, tt.quantity, item.Quantity)
				assert.Equal(t, float64(tt.quantity)*item.UnitPrice, item.TotalPrice)
			}
		})
	}
}

func TestOrder_CalculateTotal(t *testing.T) {
	order, _ := NewOrder(123)
	order.AddItem(1, "SKU-001", "Product 1", 2, 10.0) // 20.0
	order.AddItem(2, "SKU-002", "Product 2", 3, 15.0) // 45.0

	total := order.CalculateTotal()

	assert.Equal(t, 65.0, total)
	assert.Equal(t, 65.0, order.TotalAmount)
}

func TestOrder_ConfirmOrder(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus OrderStatus
		hasItems      bool
		expectError   bool
		errorContains string
	}{
		{
			name:          "confirm pending order with items",
			initialStatus: OrderStatusPending,
			hasItems:      true,
			expectError:   false,
		},
		{
			name:          "confirm empty order",
			initialStatus: OrderStatusPending,
			hasItems:      false,
			expectError:   true,
			errorContains: "cannot confirm empty order",
		},
		{
			name:          "confirm already confirmed order",
			initialStatus: OrderStatusConfirmed,
			hasItems:      true,
			expectError:   true,
			errorContains: "only pending orders can be confirmed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, _ := NewOrder(123)
			order.Status = tt.initialStatus

			if tt.hasItems {
				order.AddItem(1, "SKU-001", "Product", 1, 10.0)
			}

			err := order.ConfirmOrder()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, OrderStatusConfirmed, order.Status)
			}
		})
	}
}

func TestOrder_CancelOrder(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus OrderStatus
		expectError   bool
		errorContains string
	}{
		{
			name:          "cancel pending order",
			initialStatus: OrderStatusPending,
			expectError:   false,
		},
		{
			name:          "cancel confirmed order",
			initialStatus: OrderStatusConfirmed,
			expectError:   false,
		},
		{
			name:          "cancel processing order",
			initialStatus: OrderStatusProcessing,
			expectError:   false,
		},
		{
			name:          "cancel shipped order",
			initialStatus: OrderStatusShipped,
			expectError:   true,
			errorContains: "order cannot be cancelled",
		},
		{
			name:          "cancel delivered order",
			initialStatus: OrderStatusDelivered,
			expectError:   true,
			errorContains: "order cannot be cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, _ := NewOrder(123)
			order.Status = tt.initialStatus

			err := order.CancelOrder()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, OrderStatusCancelled, order.Status)
			}
		})
	}
}

func TestOrder_StatusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		method        func(*Order) error
		fromStatus    OrderStatus
		toStatus      OrderStatus
		expectError   bool
		errorContains string
	}{
		{
			name:        "transition to processing",
			method:      (*Order).TransitionToProcessing,
			fromStatus:  OrderStatusConfirmed,
			toStatus:    OrderStatusProcessing,
			expectError: false,
		},
		{
			name:          "invalid transition to processing",
			method:        (*Order).TransitionToProcessing,
			fromStatus:    OrderStatusPending,
			toStatus:      OrderStatusPending,
			expectError:   true,
			errorContains: "only confirmed orders can be moved to processing",
		},
		{
			name:        "transition to shipped",
			method:      (*Order).TransitionToShipped,
			fromStatus:  OrderStatusProcessing,
			toStatus:    OrderStatusShipped,
			expectError: false,
		},
		{
			name:        "transition to delivered",
			method:      (*Order).TransitionToDelivered,
			fromStatus:  OrderStatusShipped,
			toStatus:    OrderStatusDelivered,
			expectError: false,
		},
		{
			name:        "transition to refunded",
			method:      (*Order).TransitionToRefunded,
			fromStatus:  OrderStatusDelivered,
			toStatus:    OrderStatusRefunded,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, _ := NewOrder(123)
			order.Status = tt.fromStatus

			err := tt.method(order)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.toStatus, order.Status)
			}
		})
	}
}

func TestOrder_BusinessRules(t *testing.T) {
	order, _ := NewOrder(123)

	// Test CanBeCancelled
	assert.True(t, order.CanBeCancelled()) // Pending

	order.Status = OrderStatusConfirmed
	assert.True(t, order.CanBeCancelled())

	order.Status = OrderStatusShipped
	assert.False(t, order.CanBeCancelled())

	// Test IsEmpty
	assert.True(t, order.IsEmpty())
	order.AddItem(1, "SKU-001", "Product", 1, 10.0)
	assert.False(t, order.IsEmpty())

	// Test status checks
	order.Status = OrderStatusPending
	assert.True(t, order.IsPending())
	assert.False(t, order.IsConfirmed())

	order.Status = OrderStatusConfirmed
	assert.False(t, order.IsPending())
	assert.True(t, order.IsConfirmed())

	order.Status = OrderStatusCancelled
	assert.True(t, order.IsCancelled())

	order.Status = OrderStatusDelivered
	assert.True(t, order.IsDelivered())
}

func TestOrder_GetItem(t *testing.T) {
	order, _ := NewOrder(123)
	order.AddItem(1, "SKU-001", "Product 1", 2, 10.0)

	// Test existing item
	item, err := order.GetItem(1)
	assert.NoError(t, err)
	assert.NotNil(t, item)
	assert.Equal(t, uint(1), item.ProductID)

	// Test non-existing item
	item, err = order.GetItem(999)
	assert.Error(t, err)
	assert.Nil(t, item)
	assert.Contains(t, err.Error(), "item not found")
}

func TestOrder_GetCounts(t *testing.T) {
	order, _ := NewOrder(123)
	order.AddItem(1, "SKU-001", "Product 1", 2, 10.0)
	order.AddItem(2, "SKU-002", "Product 2", 3, 15.0)

	assert.Equal(t, 2, order.GetItemCount())
	assert.Equal(t, 5, order.GetTotalQuantity()) // 2 + 3
}

func TestValidateOrderStatus(t *testing.T) {
	validStatuses := []OrderStatus{
		OrderStatusPending, OrderStatusConfirmed, OrderStatusProcessing,
		OrderStatusShipped, OrderStatusDelivered, OrderStatusCancelled, OrderStatusRefunded,
	}

	for _, status := range validStatuses {
		assert.NoError(t, ValidateOrderStatus(status))
	}

	assert.Error(t, ValidateOrderStatus("invalid_status"))
}
