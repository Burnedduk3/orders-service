package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"orders-service/internal/application/dto"
	"orders-service/internal/domain/entities"
	domainErrors "orders-service/internal/domain/errors"
	"orders-service/pkg/logger"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockOrderUseCases implements the OrderUseCases interface for testing
type MockOrderUseCases struct {
	mock.Mock
}

func (m *MockOrderUseCases) CreateOrder(ctx context.Context, request *dto.CreateOrderRequestDTO) (*dto.OrderResponseDTO, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) GetOrder(ctx context.Context, id uint) (*dto.OrderResponseDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) AddItemToOrder(ctx context.Context, orderID uint, request *dto.AddOrderItemRequestDTO) (*dto.OrderResponseDTO, error) {
	args := m.Called(ctx, orderID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) RemoveItemFromOrder(ctx context.Context, orderID, productID uint) (*dto.OrderResponseDTO, error) {
	args := m.Called(ctx, orderID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) UpdateItemQuantity(ctx context.Context, orderID, productID uint, request *dto.UpdateOrderItemQuantityRequestDTO) (*dto.OrderResponseDTO, error) {
	args := m.Called(ctx, orderID, productID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) ConfirmOrder(ctx context.Context, orderID uint) (*dto.OrderResponseDTO, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) CancelOrder(ctx context.Context, orderID uint) (*dto.OrderResponseDTO, error) {
	args := m.Called(ctx, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) TransitionOrderStatus(ctx context.Context, orderID uint, request *dto.UpdateOrderStatusRequestDTO) (*dto.OrderResponseDTO, error) {
	args := m.Called(ctx, orderID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) GetCustomerOrders(ctx context.Context, customerID uint, page, pageSize int) (*dto.OrderListResponseDTO, error) {
	args := m.Called(ctx, customerID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderListResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) GetOrdersByStatus(ctx context.Context, status entities.OrderStatus, page, pageSize int) (*dto.OrderListResponseDTO, error) {
	args := m.Called(ctx, status, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderListResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) ListOrders(ctx context.Context, page, pageSize int) (*dto.OrderListResponseDTO, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.OrderListResponseDTO), args.Error(1)
}

func (m *MockOrderUseCases) DeleteOrder(ctx context.Context, orderID uint) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func setupTestOrderHandler() (*OrderHandler, *MockOrderUseCases) {
	mockUseCases := new(MockOrderUseCases)
	log := logger.New("test")
	handler := NewOrderHandler(mockUseCases, log)
	return handler, mockUseCases
}

// CreateOrder Tests
func TestOrderHandler_CreateOrder_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	requestBody := dto.CreateOrderRequestDTO{
		CustomerID: 123,
		Items: []dto.CreateOrderItemDTO{
			{
				ProductID:   1,
				ProductSKU:  "SKU-001",
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   10.50,
			},
		},
	}

	expectedResponse := &dto.OrderResponseDTO{
		ID:          1,
		CustomerID:  123,
		Items:       []dto.OrderItemResponseDTO{},
		ItemCount:   1,
		TotalItems:  2,
		TotalAmount: 21.00,
		Status:      entities.OrderStatusPending,
	}

	mockUseCases.On("CreateOrder", mock.Anything, &requestBody).Return(expectedResponse, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.CreateOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response dto.OrderResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.Equal(t, expectedResponse.CustomerID, response.CustomerID)
	assert.Equal(t, expectedResponse.Status, response.Status)

	mockUseCases.AssertExpectations(t)
}

func TestOrderHandler_CreateOrder_ValidationError(t *testing.T) {
	// Setup
	handler, _ := setupTestOrderHandler()

	requestBody := dto.CreateOrderRequestDTO{
		CustomerID: 0, // Invalid - required
		Items:      []dto.CreateOrderItemDTO{},
	}

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.CreateOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "VALIDATION_ERROR", response.Error)
	assert.NotNil(t, response.Details)
}

// GetOrder Tests
func TestOrderHandler_GetOrder_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	expectedResponse := &dto.OrderResponseDTO{
		ID:          1,
		CustomerID:  123,
		Items:       []dto.OrderItemResponseDTO{},
		TotalAmount: 100.00,
		Status:      entities.OrderStatusConfirmed,
	}

	mockUseCases.On("GetOrder", mock.Anything, uint(1)).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/1", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.GetOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.OrderResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.Equal(t, expectedResponse.CustomerID, response.CustomerID)

	mockUseCases.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_NotFound(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	mockUseCases.On("GetOrder", mock.Anything, uint(999)).Return(nil, domainErrors.ErrOrderNotFound)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/999", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")

	// Execute
	err := handler.GetOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ORDER_NOT_FOUND", response.Error)
	mockUseCases.AssertExpectations(t)
}

func TestOrderHandler_GetOrder_InvalidID(t *testing.T) {
	// Setup
	handler, _ := setupTestOrderHandler()

	// Create request with invalid ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/invalid", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	// Execute
	err := handler.GetOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "INVALID_ID", response.Error)
}

// AddItemToOrder Tests
func TestOrderHandler_AddItemToOrder_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	requestBody := dto.AddOrderItemRequestDTO{
		ProductID:   1,
		ProductSKU:  "SKU-001",
		ProductName: "Product 1",
		Quantity:    2,
		UnitPrice:   10.50,
	}

	expectedResponse := &dto.OrderResponseDTO{
		ID:          1,
		CustomerID:  123,
		Items:       []dto.OrderItemResponseDTO{},
		TotalAmount: 21.00,
		Status:      entities.OrderStatusPending,
	}

	mockUseCases.On("AddItemToOrder", mock.Anything, uint(1), &requestBody).Return(expectedResponse, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/1/items", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.AddItemToOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockUseCases.AssertExpectations(t)
}

// RemoveItemFromOrder Tests
func TestOrderHandler_RemoveItemFromOrder_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	expectedResponse := &dto.OrderResponseDTO{
		ID:          1,
		CustomerID:  123,
		Items:       []dto.OrderItemResponseDTO{},
		TotalAmount: 0.00,
		Status:      entities.OrderStatusPending,
	}

	mockUseCases.On("RemoveItemFromOrder", mock.Anything, uint(1), uint(1)).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/orders/1/items/1", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id", "product_id")
	c.SetParamValues("1", "1")

	// Execute
	err := handler.RemoveItemFromOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockUseCases.AssertExpectations(t)
}

// UpdateItemQuantity Tests
func TestOrderHandler_UpdateItemQuantity_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	requestBody := dto.UpdateOrderItemQuantityRequestDTO{
		Quantity: 5,
	}

	expectedResponse := &dto.OrderResponseDTO{
		ID:          1,
		CustomerID:  123,
		Items:       []dto.OrderItemResponseDTO{},
		TotalAmount: 52.50,
		Status:      entities.OrderStatusPending,
	}

	mockUseCases.On("UpdateItemQuantity", mock.Anything, uint(1), uint(1), &requestBody).Return(expectedResponse, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/orders/1/items/1", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id", "product_id")
	c.SetParamValues("1", "1")

	// Execute
	err := handler.UpdateItemQuantity(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockUseCases.AssertExpectations(t)
}

// ConfirmOrder Tests
func TestOrderHandler_ConfirmOrder_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	expectedResponse := &dto.OrderResponseDTO{
		ID:          1,
		CustomerID:  123,
		Items:       []dto.OrderItemResponseDTO{},
		TotalAmount: 100.00,
		Status:      entities.OrderStatusConfirmed,
	}

	mockUseCases.On("ConfirmOrder", mock.Anything, uint(1)).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/1/confirm", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.ConfirmOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.OrderResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, entities.OrderStatusConfirmed, response.Status)

	mockUseCases.AssertExpectations(t)
}

func TestOrderHandler_ConfirmOrder_EmptyOrder(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	mockUseCases.On("ConfirmOrder", mock.Anything, uint(1)).Return(nil, domainErrors.ErrEmptyOrder)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/1/confirm", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.ConfirmOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "EMPTY_ORDER", response.Error)

	mockUseCases.AssertExpectations(t)
}

// CancelOrder Tests
func TestOrderHandler_CancelOrder_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	expectedResponse := &dto.OrderResponseDTO{
		ID:          1,
		CustomerID:  123,
		Items:       []dto.OrderItemResponseDTO{},
		TotalAmount: 100.00,
		Status:      entities.OrderStatusCancelled,
	}

	mockUseCases.On("CancelOrder", mock.Anything, uint(1)).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders/1/cancel", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.CancelOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.OrderResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, entities.OrderStatusCancelled, response.Status)

	mockUseCases.AssertExpectations(t)
}

// UpdateOrderStatus Tests
func TestOrderHandler_UpdateOrderStatus_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	requestBody := dto.UpdateOrderStatusRequestDTO{
		Status: entities.OrderStatusProcessing,
	}

	expectedResponse := &dto.OrderResponseDTO{
		ID:          1,
		CustomerID:  123,
		Items:       []dto.OrderItemResponseDTO{},
		TotalAmount: 100.00,
		Status:      entities.OrderStatusProcessing,
	}

	mockUseCases.On("TransitionOrderStatus", mock.Anything, uint(1), &requestBody).Return(expectedResponse, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/orders/1/status", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.UpdateOrderStatus(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.OrderResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, entities.OrderStatusProcessing, response.Status)

	mockUseCases.AssertExpectations(t)
}

// ListOrders Tests
func TestOrderHandler_ListOrders_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	expectedOrders := []*dto.OrderResponseDTO{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []dto.OrderItemResponseDTO{},
			TotalAmount: 100.00,
			Status:      entities.OrderStatusPending,
		},
		{
			ID:          2,
			CustomerID:  456,
			Items:       []dto.OrderItemResponseDTO{},
			TotalAmount: 200.00,
			Status:      entities.OrderStatusConfirmed,
		},
	}

	expectedResponse := &dto.OrderListResponseDTO{
		Orders:   expectedOrders,
		Total:    50,
		Page:     0,
		PageSize: 10,
	}

	mockUseCases.On("ListOrders", mock.Anything, 0, 10).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.ListOrders(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.OrderListResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Orders, 2)
	assert.Equal(t, int64(50), response.Total)

	mockUseCases.AssertExpectations(t)
}

func TestOrderHandler_ListOrders_WithPagination(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	expectedResponse := &dto.OrderListResponseDTO{
		Orders:   []*dto.OrderResponseDTO{},
		Total:    0,
		Page:     2,
		PageSize: 5,
	}

	mockUseCases.On("ListOrders", mock.Anything, 2, 5).Return(expectedResponse, nil)

	// Create request with pagination parameters
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders?page=2&page_size=5", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.ListOrders(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockUseCases.AssertExpectations(t)
}

// GetCustomerOrders Tests
func TestOrderHandler_GetCustomerOrders_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	expectedOrders := []*dto.OrderResponseDTO{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []dto.OrderItemResponseDTO{},
			TotalAmount: 100.00,
			Status:      entities.OrderStatusPending,
		},
	}

	expectedResponse := &dto.OrderListResponseDTO{
		Orders:   expectedOrders,
		Total:    1,
		Page:     0,
		PageSize: 10,
	}

	mockUseCases.On("GetCustomerOrders", mock.Anything, uint(123), 0, 10).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/customers/123/orders", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("customer_id")
	c.SetParamValues("123")

	// Execute
	err := handler.GetCustomerOrders(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.OrderListResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Orders, 1)
	assert.Equal(t, uint(123), response.Orders[0].CustomerID)

	mockUseCases.AssertExpectations(t)
}

// GetOrdersByStatus Tests
func TestOrderHandler_GetOrdersByStatus_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	expectedOrders := []*dto.OrderResponseDTO{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []dto.OrderItemResponseDTO{},
			TotalAmount: 100.00,
			Status:      entities.OrderStatusPending,
		},
	}

	expectedResponse := &dto.OrderListResponseDTO{
		Orders:   expectedOrders,
		Total:    1,
		Page:     0,
		PageSize: 10,
	}

	mockUseCases.On("GetOrdersByStatus", mock.Anything, entities.OrderStatusPending, 0, 10).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/status/pending", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("status")
	c.SetParamValues("pending")

	// Execute
	err := handler.GetOrdersByStatus(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.OrderListResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Orders, 1)
	assert.Equal(t, entities.OrderStatusPending, response.Orders[0].Status)

	mockUseCases.AssertExpectations(t)
}

// DeleteOrder Tests
func TestOrderHandler_DeleteOrder_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	mockUseCases.On("DeleteOrder", mock.Anything, uint(1)).Return(nil)

	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/orders/1", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.DeleteOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	mockUseCases.AssertExpectations(t)
}

func TestOrderHandler_DeleteOrder_NotFound(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestOrderHandler()

	mockUseCases.On("DeleteOrder", mock.Anything, uint(999)).Return(domainErrors.ErrOrderNotFound)

	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/orders/999", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")

	// Execute
	err := handler.DeleteOrder(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ORDER_NOT_FOUND", response.Error)

	mockUseCases.AssertExpectations(t)
}
