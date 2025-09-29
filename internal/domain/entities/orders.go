package entities

import (
	"errors"
	"strings"
	"time"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusConfirmed  OrderStatus = "confirmed"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type OrderItem struct {
	ID          uint    `json:"id"`
	ProductID   uint    `json:"product_id"`
	ProductSKU  string  `json:"product_sku"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

type Order struct {
	ID          uint        `json:"id"`
	CustomerID  uint        `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Domain methods for Order

// AddItem adds a new item to the order or updates quantity if product already exists
func (o *Order) AddItem(productID uint, productSKU, productName string, quantity int, unitPrice float64) error {
	if o.isImmutable() {
		return errors.New("order cannot be modified in current status")
	}

	if err := validateOrderItem(productID, productSKU, productName, quantity, unitPrice); err != nil {
		return err
	}

	// Check if item already exists
	for i := range o.Items {
		if o.Items[i].ProductID == productID {
			// Update existing item quantity
			o.Items[i].Quantity += quantity
			o.Items[i].TotalPrice = float64(o.Items[i].Quantity) * o.Items[i].UnitPrice
			o.CalculateTotal()
			o.UpdatedAt = time.Now()
			return nil
		}
	}

	// Add new item
	newItem := OrderItem{
		ProductID:   productID,
		ProductSKU:  strings.TrimSpace(productSKU),
		ProductName: strings.TrimSpace(productName),
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalPrice:  float64(quantity) * unitPrice,
	}

	o.Items = append(o.Items, newItem)
	o.CalculateTotal()
	o.UpdatedAt = time.Now()
	return nil
}

// RemoveItem removes an item from the order
func (o *Order) RemoveItem(productID uint) error {
	if o.isImmutable() {
		return errors.New("order cannot be modified in current status")
	}

	for i, item := range o.Items {
		if item.ProductID == productID {
			// Remove item by slicing
			o.Items = append(o.Items[:i], o.Items[i+1:]...)
			o.CalculateTotal()
			o.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("item not found in order")
}

// UpdateItemQuantity updates the quantity of an existing item
func (o *Order) UpdateItemQuantity(productID uint, quantity int) error {
	if o.isImmutable() {
		return errors.New("order cannot be modified in current status")
	}

	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	for i := range o.Items {
		if o.Items[i].ProductID == productID {
			o.Items[i].Quantity = quantity
			o.Items[i].TotalPrice = float64(quantity) * o.Items[i].UnitPrice
			o.CalculateTotal()
			o.UpdatedAt = time.Now()
			return nil
		}
	}

	return errors.New("item not found in order")
}

// CalculateTotal recalculates and updates the total amount
func (o *Order) CalculateTotal() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.TotalPrice
	}
	o.TotalAmount = total
	return total
}

// ConfirmOrder transitions the order from pending to confirmed
func (o *Order) ConfirmOrder() error {
	if o.Status != OrderStatusPending {
		return errors.New("only pending orders can be confirmed")
	}

	if len(o.Items) == 0 {
		return errors.New("cannot confirm empty order")
	}

	o.Status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()
	return nil
}

// CancelOrder cancels the order if cancellation is allowed
func (o *Order) CancelOrder() error {
	if !o.CanBeCancelled() {
		return errors.New("order cannot be cancelled in current status")
	}

	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}

// TransitionToProcessing moves order from confirmed to processing
func (o *Order) TransitionToProcessing() error {
	if o.Status != OrderStatusConfirmed {
		return errors.New("only confirmed orders can be moved to processing")
	}

	o.Status = OrderStatusProcessing
	o.UpdatedAt = time.Now()
	return nil
}

// TransitionToShipped moves order from processing to shipped
func (o *Order) TransitionToShipped() error {
	if o.Status != OrderStatusProcessing {
		return errors.New("only processing orders can be shipped")
	}

	o.Status = OrderStatusShipped
	o.UpdatedAt = time.Now()
	return nil
}

// TransitionToDelivered moves order from shipped to delivered
func (o *Order) TransitionToDelivered() error {
	if o.Status != OrderStatusShipped {
		return errors.New("only shipped orders can be delivered")
	}

	o.Status = OrderStatusDelivered
	o.UpdatedAt = time.Now()
	return nil
}

// TransitionToRefunded moves order from delivered to refunded
func (o *Order) TransitionToRefunded() error {
	if o.Status != OrderStatusDelivered {
		return errors.New("only delivered orders can be refunded")
	}

	o.Status = OrderStatusRefunded
	o.UpdatedAt = time.Now()
	return nil
}

// Business rule methods

// CanBeCancelled checks if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending ||
		o.Status == OrderStatusConfirmed ||
		o.Status == OrderStatusProcessing
}

// IsEmpty checks if the order has no items
func (o *Order) IsEmpty() bool {
	return len(o.Items) == 0
}

// IsPending checks if order is in pending status
func (o *Order) IsPending() bool {
	return o.Status == OrderStatusPending
}

// IsConfirmed checks if order is confirmed
func (o *Order) IsConfirmed() bool {
	return o.Status == OrderStatusConfirmed
}

// IsCancelled checks if order is cancelled
func (o *Order) IsCancelled() bool {
	return o.Status == OrderStatusCancelled
}

// IsDelivered checks if order is delivered
func (o *Order) IsDelivered() bool {
	return o.Status == OrderStatusDelivered
}

// GetItem returns an item by product ID
func (o *Order) GetItem(productID uint) (*OrderItem, error) {
	for i := range o.Items {
		if o.Items[i].ProductID == productID {
			return &o.Items[i], nil
		}
	}
	return nil, errors.New("item not found")
}

// GetItemCount returns the total number of items in the order
func (o *Order) GetItemCount() int {
	return len(o.Items)
}

// GetTotalQuantity returns the total quantity of all items
func (o *Order) GetTotalQuantity() int {
	total := 0
	for _, item := range o.Items {
		total += item.Quantity
	}
	return total
}

// isImmutable checks if the order can be modified
func (o *Order) isImmutable() bool {
	return o.Status == OrderStatusCancelled ||
		o.Status == OrderStatusDelivered ||
		o.Status == OrderStatusRefunded
}

// Factory function for creating new orders
func NewOrder(customerID uint) (*Order, error) {
	if customerID == 0 {
		return nil, errors.New("customer ID is required")
	}

	now := time.Now()

	return &Order{
		CustomerID:  customerID,
		Items:       make([]OrderItem, 0),
		TotalAmount: 0.0,
		Status:      OrderStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Factory function for creating new order items
func NewOrderItem(productID uint, productSKU, productName string, quantity int, unitPrice float64) (*OrderItem, error) {
	if err := validateOrderItem(productID, productSKU, productName, quantity, unitPrice); err != nil {
		return nil, err
	}

	return &OrderItem{
		ProductID:   productID,
		ProductSKU:  strings.TrimSpace(productSKU),
		ProductName: strings.TrimSpace(productName),
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TotalPrice:  float64(quantity) * unitPrice,
	}, nil
}

// Domain validation functions
func validateOrderItem(productID uint, productSKU, productName string, quantity int, unitPrice float64) error {
	if productID == 0 {
		return errors.New("product ID is required")
	}

	if strings.TrimSpace(productSKU) == "" {
		return errors.New("product SKU is required")
	}

	if strings.TrimSpace(productName) == "" {
		return errors.New("product name is required")
	}

	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}

	if unitPrice <= 0 {
		return errors.New("unit price must be positive")
	}

	return nil
}

func ValidateOrderStatus(status OrderStatus) error {
	switch status {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusProcessing,
		OrderStatusShipped, OrderStatusDelivered, OrderStatusCancelled, OrderStatusRefunded:
		return nil
	default:
		return errors.New("invalid order status")
	}
}
