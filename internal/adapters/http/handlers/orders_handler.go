package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"orders-service/internal/application/dto"
	"orders-service/internal/application/usecases"
	"orders-service/internal/domain/entities"
	domainErrors "orders-service/internal/domain/errors"
	"orders-service/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type OrderHandler struct {
	orderUseCases usecases.OrderUseCases
	validator     *validator.Validate
	logger        logger.Logger
}

func NewOrderHandler(orderUseCases usecases.OrderUseCases, log logger.Logger) *OrderHandler {
	return &OrderHandler{
		orderUseCases: orderUseCases,
		validator:     validator.New(),
		logger:        log.With("component", "order_handler"),
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// CreateOrder handles POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Info("Create order request received",
		"request_id", requestID,
		"remote_ip", c.RealIP(),
		"user_agent", c.Request().UserAgent())

	// Parse request body
	var request dto.CreateOrderRequestDTO
	if err := c.Bind(&request); err != nil {
		h.logger.Warn("Failed to bind request body",
			"request_id", requestID,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body format",
		})
	}

	// Validate request
	if err := h.validator.Struct(request); err != nil {
		h.logger.Warn("Request validation failed",
			"request_id", requestID,
			"error", err)

		details := make(map[string]interface{})
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				details[fieldError.Field()] = getValidationErrorMessage(fieldError)
			}
		}

		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: details,
		})
	}

	// Execute use case
	response, err := h.orderUseCases.CreateOrder(c.Request().Context(), &request)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to create order")
	}

	h.logger.Info("Order created successfully",
		"request_id", requestID,
		"order_id", response.ID,
		"customer_id", response.CustomerID)

	return c.JSON(http.StatusCreated, response)
}

// GetOrder handles GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse order ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid order ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid order ID format",
		})
	}

	h.logger.Info("Get order request received",
		"request_id", requestID,
		"order_id", id,
		"remote_ip", c.RealIP())

	// Execute use case
	response, err := h.orderUseCases.GetOrder(c.Request().Context(), uint(id))
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to get order")
	}

	h.logger.Info("Order retrieved successfully",
		"request_id", requestID,
		"order_id", response.ID)

	return c.JSON(http.StatusOK, response)
}

// AddItemToOrder handles POST /api/v1/orders/:id/items
func (h *OrderHandler) AddItemToOrder(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse order ID
	idParam := c.Param("id")
	orderID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid order ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid order ID format",
		})
	}

	h.logger.Info("Add item to order request received",
		"request_id", requestID,
		"order_id", orderID,
		"remote_ip", c.RealIP())

	// Parse request body
	var request dto.AddOrderItemRequestDTO
	if err := c.Bind(&request); err != nil {
		h.logger.Warn("Failed to bind request body",
			"request_id", requestID,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body format",
		})
	}

	// Validate request
	if err := h.validator.Struct(request); err != nil {
		return h.handleValidationError(c, err, requestID)
	}

	// Execute use case
	response, err := h.orderUseCases.AddItemToOrder(c.Request().Context(), uint(orderID), &request)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to add item to order")
	}

	h.logger.Info("Item added to order successfully",
		"request_id", requestID,
		"order_id", response.ID,
		"product_id", request.ProductID)

	return c.JSON(http.StatusOK, response)
}

// RemoveItemFromOrder handles DELETE /api/v1/orders/:id/items/:product_id
func (h *OrderHandler) RemoveItemFromOrder(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse order ID and product ID
	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid order ID format",
		})
	}

	productID, err := parseUintParam(c, "product_id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	h.logger.Info("Remove item from order request received",
		"request_id", requestID,
		"order_id", orderID,
		"product_id", productID)

	// Execute use case
	response, err := h.orderUseCases.RemoveItemFromOrder(c.Request().Context(), orderID, productID)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to remove item from order")
	}

	h.logger.Info("Item removed from order successfully",
		"request_id", requestID,
		"order_id", orderID,
		"product_id", productID)

	return c.JSON(http.StatusOK, response)
}

// UpdateItemQuantity handles PUT /api/v1/orders/:id/items/:product_id
func (h *OrderHandler) UpdateItemQuantity(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse IDs
	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid order ID format",
		})
	}

	productID, err := parseUintParam(c, "product_id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	h.logger.Info("Update item quantity request received",
		"request_id", requestID,
		"order_id", orderID,
		"product_id", productID)

	// Parse request body
	var request dto.UpdateOrderItemQuantityRequestDTO
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body format",
		})
	}

	// Validate request
	if err := h.validator.Struct(request); err != nil {
		return h.handleValidationError(c, err, requestID)
	}

	// Execute use case
	response, err := h.orderUseCases.UpdateItemQuantity(c.Request().Context(), orderID, productID, &request)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to update item quantity")
	}

	h.logger.Info("Item quantity updated successfully",
		"request_id", requestID,
		"order_id", orderID,
		"product_id", productID,
		"quantity", request.Quantity)

	return c.JSON(http.StatusOK, response)
}

// ConfirmOrder handles POST /api/v1/orders/:id/confirm
func (h *OrderHandler) ConfirmOrder(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid order ID format",
		})
	}

	h.logger.Info("Confirm order request received",
		"request_id", requestID,
		"order_id", orderID)

	// Execute use case
	response, err := h.orderUseCases.ConfirmOrder(c.Request().Context(), orderID)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to confirm order")
	}

	h.logger.Info("Order confirmed successfully",
		"request_id", requestID,
		"order_id", orderID)

	return c.JSON(http.StatusOK, response)
}

// CancelOrder handles POST /api/v1/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid order ID format",
		})
	}

	h.logger.Info("Cancel order request received",
		"request_id", requestID,
		"order_id", orderID)

	// Execute use case
	response, err := h.orderUseCases.CancelOrder(c.Request().Context(), orderID)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to cancel order")
	}

	h.logger.Info("Order cancelled successfully",
		"request_id", requestID,
		"order_id", orderID)

	return c.JSON(http.StatusOK, response)
}

// UpdateOrderStatus handles PUT /api/v1/orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid order ID format",
		})
	}

	// Parse request body
	var request dto.UpdateOrderStatusRequestDTO
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body format",
		})
	}

	// Validate request
	if err := h.validator.Struct(request); err != nil {
		return h.handleValidationError(c, err, requestID)
	}

	h.logger.Info("Update order status request received",
		"request_id", requestID,
		"order_id", orderID,
		"new_status", request.Status)

	// Execute use case
	response, err := h.orderUseCases.TransitionOrderStatus(c.Request().Context(), orderID, &request)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to update order status")
	}

	h.logger.Info("Order status updated successfully",
		"request_id", requestID,
		"order_id", orderID,
		"new_status", request.Status)

	return c.JSON(http.StatusOK, response)
}

// ListOrders handles GET /api/v1/orders
func (h *OrderHandler) ListOrders(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Info("List orders request received",
		"request_id", requestID,
		"remote_ip", c.RealIP())

	// Parse query parameters
	page, pageSize := parsePaginationParams(c)

	h.logger.Info("List orders parameters",
		"request_id", requestID,
		"page", page,
		"page_size", pageSize)

	// Execute use case
	response, err := h.orderUseCases.ListOrders(c.Request().Context(), page, pageSize)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to list orders")
	}

	h.logger.Info("Orders listed successfully",
		"request_id", requestID,
		"count", len(response.Orders),
		"page", page)

	return c.JSON(http.StatusOK, response)
}

// GetCustomerOrders handles GET /api/v1/customers/:customer_id/orders
func (h *OrderHandler) GetCustomerOrders(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	customerID, err := parseUintParam(c, "customer_id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid customer ID format",
		})
	}

	// Parse query parameters
	page, pageSize := parsePaginationParams(c)

	h.logger.Info("Get customer orders request received",
		"request_id", requestID,
		"customer_id", customerID,
		"page", page,
		"page_size", pageSize)

	// Execute use case
	response, err := h.orderUseCases.GetCustomerOrders(c.Request().Context(), customerID, page, pageSize)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to get customer orders")
	}

	h.logger.Info("Customer orders retrieved successfully",
		"request_id", requestID,
		"customer_id", customerID,
		"count", len(response.Orders))

	return c.JSON(http.StatusOK, response)
}

// GetOrdersByStatus handles GET /api/v1/orders/status/:status
func (h *OrderHandler) GetOrdersByStatus(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	statusParam := c.Param("status")
	if statusParam == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_STATUS",
			Message: "Status parameter is required",
		})
	}

	status := entities.OrderStatus(statusParam)

	// Parse query parameters
	page, pageSize := parsePaginationParams(c)

	h.logger.Info("Get orders by status request received",
		"request_id", requestID,
		"status", status,
		"page", page,
		"page_size", pageSize)

	// Execute use case
	response, err := h.orderUseCases.GetOrdersByStatus(c.Request().Context(), status, page, pageSize)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to get orders by status")
	}

	h.logger.Info("Orders by status retrieved successfully",
		"request_id", requestID,
		"status", status,
		"count", len(response.Orders))

	return c.JSON(http.StatusOK, response)
}

// DeleteOrder handles DELETE /api/v1/orders/:id
func (h *OrderHandler) DeleteOrder(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	orderID, err := parseUintParam(c, "id")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid order ID format",
		})
	}

	h.logger.Info("Delete order request received",
		"request_id", requestID,
		"order_id", orderID)

	// Execute use case
	err = h.orderUseCases.DeleteOrder(c.Request().Context(), orderID)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to delete order")
	}

	h.logger.Info("Order deleted successfully",
		"request_id", requestID,
		"order_id", orderID)

	return c.JSON(http.StatusNoContent, nil)
}

// Helper functions

func (h *OrderHandler) handleError(c echo.Context, err error, requestID, logMessage string) error {
	h.logger.Error(logMessage,
		"request_id", requestID,
		"error", err)

	// Handle domain errors
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case domainErrors.ErrOrderNotFound.Code:
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		case domainErrors.ErrOrderAlreadyExists.Code:
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		case domainErrors.ErrInvalidCustomerID.Code,
			domainErrors.ErrInvalidOrderStatus.Code,
			domainErrors.ErrInvalidStatusTransition.Code,
			domainErrors.ErrOrderAlreadyConfirmed.Code,
			domainErrors.ErrOrderCannotBeCancelled.Code,
			domainErrors.ErrEmptyOrder.Code,
			domainErrors.ErrOrderItemNotFound.Code:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		default:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		}
	}

	// Handle generic errors
	return c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "INTERNAL_ERROR",
		Message: "An internal error occurred",
	})
}

func (h *OrderHandler) handleValidationError(c echo.Context, err error, requestID string) error {
	h.logger.Warn("Request validation failed",
		"request_id", requestID,
		"error", err)

	details := make(map[string]interface{})
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			details[fieldError.Field()] = getValidationErrorMessage(fieldError)
		}
	}

	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Error:   "VALIDATION_ERROR",
		Message: "Request validation failed",
		Details: details,
	})
}

func parseUintParam(c echo.Context, paramName string) (uint, error) {
	param := c.Param(paramName)
	id, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func parsePaginationParams(c echo.Context) (int, int) {
	page := 0
	pageSize := 10

	if pageParam := c.QueryParam("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p >= 0 {
			page = p
		}
	}

	if sizeParam := c.QueryParam("page_size"); sizeParam != "" {
		if ps, err := strconv.Atoi(sizeParam); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	return page, pageSize
}

func getValidationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Minimum value is " + fieldError.Param()
	case "max":
		return "Maximum value is " + fieldError.Param()
	case "gt":
		return "Value must be greater than " + fieldError.Param()
	case "oneof":
		return "Value must be one of: " + fieldError.Param()
	default:
		return "Invalid value"
	}
}
