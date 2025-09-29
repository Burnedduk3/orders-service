package usecases

import (
	"context"
	"errors"

	"orders-service/internal/application/dto"
	"orders-service/internal/application/ports"
	"orders-service/internal/domain/entities"
	domainErrors "orders-service/internal/domain/errors"
	"orders-service/pkg/logger"
)

// OrderUseCases defines the interface for order business operations
type OrderUseCases interface {
	CreateOrder(ctx context.Context, request *dto.CreateOrderRequestDTO) (*dto.OrderResponseDTO, error)
	GetOrder(ctx context.Context, id uint) (*dto.OrderResponseDTO, error)
	AddItemToOrder(ctx context.Context, orderID uint, request *dto.AddOrderItemRequestDTO) (*dto.OrderResponseDTO, error)
	RemoveItemFromOrder(ctx context.Context, orderID, productID uint) (*dto.OrderResponseDTO, error)
	UpdateItemQuantity(ctx context.Context, orderID, productID uint, request *dto.UpdateOrderItemQuantityRequestDTO) (*dto.OrderResponseDTO, error)
	ConfirmOrder(ctx context.Context, orderID uint) (*dto.OrderResponseDTO, error)
	CancelOrder(ctx context.Context, orderID uint) (*dto.OrderResponseDTO, error)
	TransitionOrderStatus(ctx context.Context, orderID uint, request *dto.UpdateOrderStatusRequestDTO) (*dto.OrderResponseDTO, error)
	GetCustomerOrders(ctx context.Context, customerID uint, page, pageSize int) (*dto.OrderListResponseDTO, error)
	GetOrdersByStatus(ctx context.Context, status entities.OrderStatus, page, pageSize int) (*dto.OrderListResponseDTO, error)
	ListOrders(ctx context.Context, page, pageSize int) (*dto.OrderListResponseDTO, error)
	DeleteOrder(ctx context.Context, orderID uint) error
}

// orderUseCasesImpl implements OrderUseCases interface
type orderUseCasesImpl struct {
	orderRepo ports.OrderRepository
	logger    logger.Logger
}

// NewOrderUseCases creates a new instance of order use cases
func NewOrderUseCases(orderRepo ports.OrderRepository, log logger.Logger) OrderUseCases {
	return &orderUseCasesImpl{
		orderRepo: orderRepo,
		logger:    log.With("component", "order_usecases"),
	}
}

// CreateOrder creates a new order
func (uc *orderUseCasesImpl) CreateOrder(ctx context.Context, request *dto.CreateOrderRequestDTO) (*dto.OrderResponseDTO, error) {
	uc.logger.Info("CreateOrder use case called", "customer_id", request.CustomerID)

	// Convert DTO to domain entity
	domainEntity, err := request.ToEntity()
	if err != nil {
		uc.logger.Error("Failed to convert DTO to entity", "error", err)
		return nil, err
	}

	// Create order in repository
	createdOrder, err := uc.orderRepo.Create(ctx, domainEntity)
	if err != nil {
		uc.logger.Error("Failed to create order", "error", err)
		return nil, domainErrors.ErrFailedToCreateOrder
	}

	uc.logger.Info("CreateOrder success", "order_id", createdOrder.ID, "customer_id", request.CustomerID)
	return dto.OrderToResponseDTO(createdOrder), nil
}

// GetOrder retrieves an order by ID
func (uc *orderUseCasesImpl) GetOrder(ctx context.Context, id uint) (*dto.OrderResponseDTO, error) {
	uc.logger.Info("GetOrder use case called", "order_id", id)

	order, err := uc.orderRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", id, "error", err)
		return nil, err
	}

	uc.logger.Info("GetOrder success", "order_id", id)
	return dto.OrderToResponseDTO(order), nil
}

// AddItemToOrder adds an item to an existing order
func (uc *orderUseCasesImpl) AddItemToOrder(ctx context.Context, orderID uint, request *dto.AddOrderItemRequestDTO) (*dto.OrderResponseDTO, error) {
	uc.logger.Info("AddItemToOrder use case called", "order_id", orderID, "product_id", request.ProductID)

	// Get existing order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Add item to order
	err = order.AddItem(
		request.ProductID,
		request.ProductSKU,
		request.ProductName,
		request.Quantity,
		request.UnitPrice,
	)
	if err != nil {
		uc.logger.Error("Failed to add item to order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Update order in repository
	updatedOrder, err := uc.orderRepo.Update(ctx, order)
	if err != nil {
		uc.logger.Error("Failed to update order", "order_id", orderID, "error", err)
		return nil, domainErrors.ErrFailedToUpdateOrder
	}

	uc.logger.Info("AddItemToOrder success", "order_id", orderID, "product_id", request.ProductID)
	return dto.OrderToResponseDTO(updatedOrder), nil
}

// RemoveItemFromOrder removes an item from an order
func (uc *orderUseCasesImpl) RemoveItemFromOrder(ctx context.Context, orderID, productID uint) (*dto.OrderResponseDTO, error) {
	uc.logger.Info("RemoveItemFromOrder use case called", "order_id", orderID, "product_id", productID)

	// Get existing order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Remove item from order
	err = order.RemoveItem(productID)
	if err != nil {
		uc.logger.Error("Failed to remove item from order", "order_id", orderID, "product_id", productID, "error", err)
		return nil, err
	}

	// Update order in repository
	updatedOrder, err := uc.orderRepo.Update(ctx, order)
	if err != nil {
		uc.logger.Error("Failed to update order", "order_id", orderID, "error", err)
		return nil, domainErrors.ErrFailedToUpdateOrder
	}

	uc.logger.Info("RemoveItemFromOrder success", "order_id", orderID, "product_id", productID)
	return dto.OrderToResponseDTO(updatedOrder), nil
}

// UpdateItemQuantity updates the quantity of an item in an order
func (uc *orderUseCasesImpl) UpdateItemQuantity(ctx context.Context, orderID, productID uint, request *dto.UpdateOrderItemQuantityRequestDTO) (*dto.OrderResponseDTO, error) {
	uc.logger.Info("UpdateItemQuantity use case called", "order_id", orderID, "product_id", productID, "quantity", request.Quantity)

	// Get existing order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Update item quantity
	err = order.UpdateItemQuantity(productID, request.Quantity)
	if err != nil {
		uc.logger.Error("Failed to update item quantity", "order_id", orderID, "product_id", productID, "error", err)
		return nil, err
	}

	// Update order in repository
	updatedOrder, err := uc.orderRepo.Update(ctx, order)
	if err != nil {
		uc.logger.Error("Failed to update order", "order_id", orderID, "error", err)
		return nil, domainErrors.ErrFailedToUpdateOrder
	}

	uc.logger.Info("UpdateItemQuantity success", "order_id", orderID, "product_id", productID)
	return dto.OrderToResponseDTO(updatedOrder), nil
}

// ConfirmOrder confirms a pending order
func (uc *orderUseCasesImpl) ConfirmOrder(ctx context.Context, orderID uint) (*dto.OrderResponseDTO, error) {
	uc.logger.Info("ConfirmOrder use case called", "order_id", orderID)

	// Get existing order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Confirm order
	err = order.ConfirmOrder()
	if err != nil {
		uc.logger.Error("Failed to confirm order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Update order in repository
	updatedOrder, err := uc.orderRepo.Update(ctx, order)
	if err != nil {
		uc.logger.Error("Failed to update order", "order_id", orderID, "error", err)
		return nil, domainErrors.ErrFailedToUpdateOrder
	}

	uc.logger.Info("ConfirmOrder success", "order_id", orderID)
	return dto.OrderToResponseDTO(updatedOrder), nil
}

// CancelOrder cancels an order
func (uc *orderUseCasesImpl) CancelOrder(ctx context.Context, orderID uint) (*dto.OrderResponseDTO, error) {
	uc.logger.Info("CancelOrder use case called", "order_id", orderID)

	// Get existing order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Cancel order
	err = order.CancelOrder()
	if err != nil {
		uc.logger.Error("Failed to cancel order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Update order in repository
	updatedOrder, err := uc.orderRepo.Update(ctx, order)
	if err != nil {
		uc.logger.Error("Failed to update order", "order_id", orderID, "error", err)
		return nil, domainErrors.ErrFailedToUpdateOrder
	}

	uc.logger.Info("CancelOrder success", "order_id", orderID)
	return dto.OrderToResponseDTO(updatedOrder), nil
}

// TransitionOrderStatus transitions an order to a new status
func (uc *orderUseCasesImpl) TransitionOrderStatus(ctx context.Context, orderID uint, request *dto.UpdateOrderStatusRequestDTO) (*dto.OrderResponseDTO, error) {
	uc.logger.Info("TransitionOrderStatus use case called", "order_id", orderID, "new_status", request.Status)

	// Get existing order
	order, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", orderID, "error", err)
		return nil, err
	}

	// Validate status
	if err := entities.ValidateOrderStatus(request.Status); err != nil {
		uc.logger.Error("Invalid order status", "status", request.Status, "error", err)
		return nil, domainErrors.ErrInvalidOrderStatus
	}

	// Transition based on target status
	switch request.Status {
	case entities.OrderStatusConfirmed:
		err = order.ConfirmOrder()
	case entities.OrderStatusProcessing:
		err = order.TransitionToProcessing()
	case entities.OrderStatusShipped:
		err = order.TransitionToShipped()
	case entities.OrderStatusDelivered:
		err = order.TransitionToDelivered()
	case entities.OrderStatusCancelled:
		err = order.CancelOrder()
	case entities.OrderStatusRefunded:
		err = order.TransitionToRefunded()
	default:
		err = errors.New("unsupported status transition")
	}

	if err != nil {
		uc.logger.Error("Failed to transition order status", "order_id", orderID, "error", err)
		return nil, err
	}

	// Update order in repository
	updatedOrder, err := uc.orderRepo.Update(ctx, order)
	if err != nil {
		uc.logger.Error("Failed to update order", "order_id", orderID, "error", err)
		return nil, domainErrors.ErrFailedToUpdateOrder
	}

	uc.logger.Info("TransitionOrderStatus success", "order_id", orderID, "new_status", request.Status)
	return dto.OrderToResponseDTO(updatedOrder), nil
}

// GetCustomerOrders retrieves all orders for a specific customer
func (uc *orderUseCasesImpl) GetCustomerOrders(ctx context.Context, customerID uint, page, pageSize int) (*dto.OrderListResponseDTO, error) {
	uc.logger.Info("GetCustomerOrders use case called", "customer_id", customerID, "page", page, "page_size", pageSize)

	// Validate and normalize pagination
	page, pageSize = normalizePagination(page, pageSize)

	// Get orders from repository
	orders, err := uc.orderRepo.GetByCustomerID(ctx, customerID, pageSize, page)
	if err != nil {
		uc.logger.Error("Failed to get customer orders", "customer_id", customerID, "error", err)
		return nil, domainErrors.ErrFailedToListOrders
	}

	// Get total count
	total, err := uc.orderRepo.CountByCustomerID(ctx, customerID)
	if err != nil {
		uc.logger.Error("Failed to count customer orders", "customer_id", customerID, "error", err)
		total = int64(len(orders))
	}

	uc.logger.Info("GetCustomerOrders success", "customer_id", customerID, "count", len(orders))
	return &dto.OrderListResponseDTO{
		Orders:   dto.OrdersToResponseDTOs(orders),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetOrdersByStatus retrieves orders by status
func (uc *orderUseCasesImpl) GetOrdersByStatus(ctx context.Context, status entities.OrderStatus, page, pageSize int) (*dto.OrderListResponseDTO, error) {
	uc.logger.Info("GetOrdersByStatus use case called", "status", status, "page", page, "page_size", pageSize)

	// Validate status
	if err := entities.ValidateOrderStatus(status); err != nil {
		uc.logger.Error("Invalid order status", "status", status, "error", err)
		return nil, domainErrors.ErrInvalidOrderStatus
	}

	// Validate and normalize pagination
	page, pageSize = normalizePagination(page, pageSize)

	// Get orders from repository
	orders, err := uc.orderRepo.GetByStatus(ctx, status, pageSize, page)
	if err != nil {
		uc.logger.Error("Failed to get orders by status", "status", status, "error", err)
		return nil, domainErrors.ErrFailedToListOrders
	}

	// Get total count
	total, err := uc.orderRepo.CountByStatus(ctx, status)
	if err != nil {
		uc.logger.Error("Failed to count orders by status", "status", status, "error", err)
		total = int64(len(orders))
	}

	uc.logger.Info("GetOrdersByStatus success", "status", status, "count", len(orders))
	return &dto.OrderListResponseDTO{
		Orders:   dto.OrdersToResponseDTOs(orders),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ListOrders retrieves a paginated list of all orders
func (uc *orderUseCasesImpl) ListOrders(ctx context.Context, page, pageSize int) (*dto.OrderListResponseDTO, error) {
	uc.logger.Info("ListOrders use case called", "page", page, "page_size", pageSize)

	// Validate and normalize pagination
	page, pageSize = normalizePagination(page, pageSize)

	// Get orders from repository
	orders, err := uc.orderRepo.List(ctx, pageSize, page)
	if err != nil {
		uc.logger.Error("Failed to list orders", "error", err)
		return nil, domainErrors.ErrFailedToListOrders
	}

	// Get total count
	total, err := uc.orderRepo.Count(ctx)
	if err != nil {
		uc.logger.Error("Failed to count orders", "error", err)
		total = int64(len(orders))
	}

	uc.logger.Info("ListOrders success", "count", len(orders))
	return &dto.OrderListResponseDTO{
		Orders:   dto.OrdersToResponseDTOs(orders),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// DeleteOrder soft deletes an order
func (uc *orderUseCasesImpl) DeleteOrder(ctx context.Context, orderID uint) error {
	uc.logger.Info("DeleteOrder use case called", "order_id", orderID)

	// Check if order exists
	_, err := uc.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to get order", "order_id", orderID, "error", err)
		return err
	}

	// Delete order
	err = uc.orderRepo.Delete(ctx, orderID)
	if err != nil {
		uc.logger.Error("Failed to delete order", "order_id", orderID, "error", err)
		return domainErrors.ErrFailedToDeleteOrder
	}

	uc.logger.Info("DeleteOrder success", "order_id", orderID)
	return nil
}

// Helper function to normalize pagination parameters
func normalizePagination(page, pageSize int) (int, int) {
	if page < 0 {
		page = 0
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return page, pageSize
}
