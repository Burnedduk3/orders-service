package errors

import "fmt"

type DomainError struct {
	Code    string
	Message string
	Field   string
}

func (e *DomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Order-specific domain errors
var (
	ErrOrderNotFound = &DomainError{
		Code:    "ORDER_NOT_FOUND",
		Message: "Order not found",
	}

	ErrOrderAlreadyExists = &DomainError{
		Code:    "ORDER_ALREADY_EXISTS",
		Message: "Order with this ID already exists",
		Field:   "id",
	}

	ErrInvalidCustomerID = &DomainError{
		Code:    "INVALID_CUSTOMER_ID",
		Message: "Customer ID is required",
		Field:   "customer_id",
	}

	ErrInvalidOrderStatus = &DomainError{
		Code:    "INVALID_ORDER_STATUS",
		Message: "Invalid order status",
		Field:   "status",
	}

	ErrInvalidStatusTransition = &DomainError{
		Code:    "INVALID_STATUS_TRANSITION",
		Message: "Invalid status transition",
		Field:   "status",
	}

	ErrOrderAlreadyConfirmed = &DomainError{
		Code:    "ORDER_ALREADY_CONFIRMED",
		Message: "Order is already confirmed and cannot be modified",
	}

	ErrOrderAlreadyCancelled = &DomainError{
		Code:    "ORDER_ALREADY_CANCELLED",
		Message: "Order is already cancelled",
	}

	ErrOrderCannotBeCancelled = &DomainError{
		Code:    "ORDER_CANNOT_BE_CANCELLED",
		Message: "Order cannot be cancelled in current status",
	}

	ErrEmptyOrder = &DomainError{
		Code:    "EMPTY_ORDER",
		Message: "Order must have at least one item",
	}

	ErrInvalidTotalAmount = &DomainError{
		Code:    "INVALID_TOTAL_AMOUNT",
		Message: "Total amount must be positive",
		Field:   "total_amount",
	}

	// Order Item errors
	ErrOrderItemNotFound = &DomainError{
		Code:    "ORDER_ITEM_NOT_FOUND",
		Message: "Order item not found",
	}

	ErrInvalidProductID = &DomainError{
		Code:    "INVALID_PRODUCT_ID",
		Message: "Product ID is required",
		Field:   "product_id",
	}

	ErrInvalidProductSKU = &DomainError{
		Code:    "INVALID_PRODUCT_SKU",
		Message: "Product SKU is required",
		Field:   "product_sku",
	}

	ErrInvalidProductName = &DomainError{
		Code:    "INVALID_PRODUCT_NAME",
		Message: "Product name is required",
		Field:   "product_name",
	}

	ErrInvalidQuantity = &DomainError{
		Code:    "INVALID_QUANTITY",
		Message: "Quantity must be positive",
		Field:   "quantity",
	}

	ErrInvalidUnitPrice = &DomainError{
		Code:    "INVALID_UNIT_PRICE",
		Message: "Unit price must be positive",
		Field:   "unit_price",
	}

	ErrDuplicateOrderItem = &DomainError{
		Code:    "DUPLICATE_ORDER_ITEM",
		Message: "Product already exists in order",
		Field:   "product_id",
	}

	// Repository errors
	ErrFailedToCreateOrder = &DomainError{
		Code:    "FAILED_TO_CREATE_ORDER",
		Message: "Failed to create order",
	}

	ErrFailedToUpdateOrder = &DomainError{
		Code:    "FAILED_TO_UPDATE_ORDER",
		Message: "Failed to update order",
	}

	ErrFailedToDeleteOrder = &DomainError{
		Code:    "FAILED_TO_DELETE_ORDER",
		Message: "Failed to delete order",
	}

	ErrFailedToListOrders = &DomainError{
		Code:    "FAILED_TO_LIST_ORDERS",
		Message: "Failed to list orders",
	}
)

// Helper functions to create specific errors
func NewOrderValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:    "ORDER_VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

func NewOrderItemValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:    "ORDER_ITEM_VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

func NewInvalidStatusTransitionError(from, to string) *DomainError {
	return &DomainError{
		Code:    "INVALID_STATUS_TRANSITION",
		Message: fmt.Sprintf("Cannot transition from %s to %s", from, to),
		Field:   "status",
	}
}
