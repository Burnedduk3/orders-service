package usecases

import (
	"context"
	"testing"
	"time"

	"orders-service/internal/application/dto"
	"orders-service/internal/domain/entities"
	domainErrors "orders-service/internal/domain/errors"
	"orders-service/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockOrderRepository implements the OrderRepository interface for testing
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *entities.Order) (*entities.Order, error) {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByID(ctx context.Context, id uint) (*entities.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Order), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, order *entities.Order) (*entities.Order, error) {
	args := m.Called(ctx, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Order), args.Error(1)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) List(ctx context.Context, limit, offset int) ([]*entities.Order, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByCustomerID(ctx context.Context, customerID uint, limit, offset int) ([]*entities.Order, error) {
	args := m.Called(ctx, customerID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Order), args.Error(1)
}

func (m *MockOrderRepository) GetByStatus(ctx context.Context, status entities.OrderStatus, limit, offset int) ([]*entities.Order, error) {
	args := m.Called(ctx, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Order), args.Error(1)
}

func (m *MockOrderRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) CountByCustomerID(ctx context.Context, customerID uint) (int64, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) CountByStatus(ctx context.Context, status entities.OrderStatus) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func setupTestOrderUseCases() (OrderUseCases, *MockOrderRepository) {
	mockRepo := new(MockOrderRepository)
	log := logger.New("test")
	useCases := NewOrderUseCases(mockRepo, log)
	return useCases, mockRepo
}

// CreateOrder Tests
func TestOrderUseCases_CreateOrder_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	request := &dto.CreateOrderRequestDTO{
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

	expectedCreatedOrder := &entities.Order{
		ID:         1,
		CustomerID: 123,
		Items: []entities.OrderItem{
			{
				ID:          1,
				ProductID:   1,
				ProductSKU:  "SKU-001",
				ProductName: "Product 1",
				Quantity:    2,
				UnitPrice:   10.50,
				TotalPrice:  21.00,
			},
		},
		TotalAmount: 21.00,
		Status:      entities.OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(order *entities.Order) bool {
		return order.CustomerID == 123 &&
			order.Status == entities.OrderStatusPending &&
			len(order.Items) == 1
	})).Return(expectedCreatedOrder, nil)

	// When
	result, err := useCases.CreateOrder(ctx, request)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, uint(123), result.CustomerID)
	assert.Equal(t, entities.OrderStatusPending, result.Status)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, 21.00, result.TotalAmount)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_CreateOrder_InvalidCustomerID(t *testing.T) {
	// Given
	useCases, _ := setupTestOrderUseCases()
	ctx := context.Background()

	request := &dto.CreateOrderRequestDTO{
		CustomerID: 0, // Invalid
		Items:      []dto.CreateOrderItemDTO{},
	}

	// When
	result, err := useCases.CreateOrder(ctx, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "customer ID is required")
}

func TestOrderUseCases_CreateOrder_RepositoryError(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	request := &dto.CreateOrderRequestDTO{
		CustomerID: 123,
		Items:      []dto.CreateOrderItemDTO{},
	}

	mockRepo.On("Create", ctx, mock.Anything).Return(nil, assert.AnError)

	// When
	result, err := useCases.CreateOrder(ctx, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrFailedToCreateOrder, err)

	mockRepo.AssertExpectations(t)
}

// GetOrder Tests
func TestOrderUseCases_GetOrder_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	expectedOrder := &entities.Order{
		ID:          1,
		CustomerID:  123,
		Items:       []entities.OrderItem{},
		TotalAmount: 100.00,
		Status:      entities.OrderStatusConfirmed,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(expectedOrder, nil)

	// When
	result, err := useCases.GetOrder(ctx, 1)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, uint(123), result.CustomerID)
	assert.Equal(t, entities.OrderStatusConfirmed, result.Status)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_GetOrder_NotFound(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, domainErrors.ErrOrderNotFound)

	// When
	result, err := useCases.GetOrder(ctx, 999)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrOrderNotFound, err)

	mockRepo.AssertExpectations(t)
}

// AddItemToOrder Tests
func TestOrderUseCases_AddItemToOrder_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder := &entities.Order{
		ID:          1,
		CustomerID:  123,
		Items:       []entities.OrderItem{},
		TotalAmount: 0.00,
		Status:      entities.OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	request := &dto.AddOrderItemRequestDTO{
		ProductID:   1,
		ProductSKU:  "SKU-001",
		ProductName: "Product 1",
		Quantity:    2,
		UnitPrice:   10.50,
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(order *entities.Order) bool {
		return order.ID == 1 && len(order.Items) == 1
	})).Return(existingOrder, nil)

	// When
	result, err := useCases.AddItemToOrder(ctx, 1, request)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_AddItemToOrder_OrderNotFound(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	request := &dto.AddOrderItemRequestDTO{
		ProductID:   1,
		ProductSKU:  "SKU-001",
		ProductName: "Product 1",
		Quantity:    1,
		UnitPrice:   10.00,
	}

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, domainErrors.ErrOrderNotFound)

	// When
	result, err := useCases.AddItemToOrder(ctx, 999, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrOrderNotFound, err)

	mockRepo.AssertExpectations(t)
}

// RemoveItemFromOrder Tests
func TestOrderUseCases_RemoveItemFromOrder_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1
	existingOrder.AddItem(1, "SKU-001", "Product 1", 2, 10.50)

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(order *entities.Order) bool {
		return order.ID == 1 && len(order.Items) == 0
	})).Return(existingOrder, nil)

	// When
	result, err := useCases.RemoveItemFromOrder(ctx, 1, 1)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

// UpdateItemQuantity Tests
func TestOrderUseCases_UpdateItemQuantity_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1
	existingOrder.AddItem(1, "SKU-001", "Product 1", 2, 10.50)

	request := &dto.UpdateOrderItemQuantityRequestDTO{
		Quantity: 5,
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)
	mockRepo.On("Update", ctx, mock.Anything).Return(existingOrder, nil)

	// When
	result, err := useCases.UpdateItemQuantity(ctx, 1, 1, request)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)

	mockRepo.AssertExpectations(t)
}

// ConfirmOrder Tests
func TestOrderUseCases_ConfirmOrder_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1
	existingOrder.AddItem(1, "SKU-001", "Product 1", 2, 10.50)

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(order *entities.Order) bool {
		return order.Status == entities.OrderStatusConfirmed
	})).Return(existingOrder, nil)

	// When
	result, err := useCases.ConfirmOrder(ctx, 1)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entities.OrderStatusConfirmed, result.Status)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_ConfirmOrder_EmptyOrder(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1
	// No items added

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)

	// When
	result, err := useCases.ConfirmOrder(ctx, 1)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot confirm empty order")

	mockRepo.AssertExpectations(t)
}

// CancelOrder Tests
func TestOrderUseCases_CancelOrder_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1
	existingOrder.Status = entities.OrderStatusPending

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(order *entities.Order) bool {
		return order.Status == entities.OrderStatusCancelled
	})).Return(existingOrder, nil)

	// When
	result, err := useCases.CancelOrder(ctx, 1)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entities.OrderStatusCancelled, result.Status)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_CancelOrder_CannotCancel(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1
	existingOrder.Status = entities.OrderStatusDelivered // Cannot cancel delivered orders

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)

	// When
	result, err := useCases.CancelOrder(ctx, 1)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "order cannot be cancelled")

	mockRepo.AssertExpectations(t)
}

// TransitionOrderStatus Tests
func TestOrderUseCases_TransitionOrderStatus_ToProcessing(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1
	existingOrder.Status = entities.OrderStatusConfirmed

	request := &dto.UpdateOrderStatusRequestDTO{
		Status: entities.OrderStatusProcessing,
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(order *entities.Order) bool {
		return order.Status == entities.OrderStatusProcessing
	})).Return(existingOrder, nil)

	// When
	result, err := useCases.TransitionOrderStatus(ctx, 1, request)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, entities.OrderStatusProcessing, result.Status)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_TransitionOrderStatus_InvalidStatus(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1

	request := &dto.UpdateOrderStatusRequestDTO{
		Status: "invalid_status",
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)

	// When
	result, err := useCases.TransitionOrderStatus(ctx, 1, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrInvalidOrderStatus, err)

	mockRepo.AssertExpectations(t)
}

// GetCustomerOrders Tests
func TestOrderUseCases_GetCustomerOrders_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	expectedOrders := []*entities.Order{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []entities.OrderItem{},
			TotalAmount: 100.00,
			Status:      entities.OrderStatusPending,
		},
		{
			ID:          2,
			CustomerID:  123,
			Items:       []entities.OrderItem{},
			TotalAmount: 200.00,
			Status:      entities.OrderStatusConfirmed,
		},
	}

	mockRepo.On("GetByCustomerID", ctx, uint(123), 10, 0).Return(expectedOrders, nil)
	mockRepo.On("CountByCustomerID", ctx, uint(123)).Return(int64(2), nil)

	// When
	result, err := useCases.GetCustomerOrders(ctx, 123, 0, 10)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Orders, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 0, result.Page)
	assert.Equal(t, 10, result.PageSize)

	mockRepo.AssertExpectations(t)
}

// GetOrdersByStatus Tests
func TestOrderUseCases_GetOrdersByStatus_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	expectedOrders := []*entities.Order{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []entities.OrderItem{},
			TotalAmount: 100.00,
			Status:      entities.OrderStatusPending,
		},
	}

	mockRepo.On("GetByStatus", ctx, entities.OrderStatusPending, 10, 0).Return(expectedOrders, nil)
	mockRepo.On("CountByStatus", ctx, entities.OrderStatusPending).Return(int64(1), nil)

	// When
	result, err := useCases.GetOrdersByStatus(ctx, entities.OrderStatusPending, 0, 10)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Orders, 1)
	assert.Equal(t, int64(1), result.Total)

	mockRepo.AssertExpectations(t)
}

// ListOrders Tests
func TestOrderUseCases_ListOrders_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	expectedOrders := []*entities.Order{
		{
			ID:          1,
			CustomerID:  123,
			Items:       []entities.OrderItem{},
			TotalAmount: 100.00,
			Status:      entities.OrderStatusPending,
		},
		{
			ID:          2,
			CustomerID:  456,
			Items:       []entities.OrderItem{},
			TotalAmount: 200.00,
			Status:      entities.OrderStatusConfirmed,
		},
	}

	mockRepo.On("List", ctx, 10, 0).Return(expectedOrders, nil)
	mockRepo.On("Count", ctx).Return(int64(50), nil)

	// When
	result, err := useCases.ListOrders(ctx, 0, 10)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Orders, 2)
	assert.Equal(t, int64(50), result.Total)
	assert.Equal(t, 0, result.Page)
	assert.Equal(t, 10, result.PageSize)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_ListOrders_InvalidPagination(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	mockRepo.On("List", ctx, 10, 0).Return([]*entities.Order{}, nil)
	mockRepo.On("Count", ctx).Return(int64(0), nil)

	// When - Pass invalid pagination parameters
	result, err := useCases.ListOrders(ctx, -1, 150)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.Page)      // Should default to 0
	assert.Equal(t, 10, result.PageSize) // Should default to 10

	mockRepo.AssertExpectations(t)
}

// DeleteOrder Tests
func TestOrderUseCases_DeleteOrder_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)
	mockRepo.On("Delete", ctx, uint(1)).Return(nil)

	// When
	err := useCases.DeleteOrder(ctx, 1)

	// Then
	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_DeleteOrder_NotFound(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, domainErrors.ErrOrderNotFound)

	// When
	err := useCases.DeleteOrder(ctx, 999)

	// Then
	assert.Error(t, err)
	assert.Equal(t, domainErrors.ErrOrderNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestOrderUseCases_DeleteOrder_RepositoryError(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestOrderUseCases()
	ctx := context.Background()

	existingOrder, _ := entities.NewOrder(123)
	existingOrder.ID = 1

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingOrder, nil)
	mockRepo.On("Delete", ctx, uint(1)).Return(assert.AnError)

	// When
	err := useCases.DeleteOrder(ctx, 1)

	// Then
	assert.Error(t, err)
	assert.Equal(t, domainErrors.ErrFailedToDeleteOrder, err)

	mockRepo.AssertExpectations(t)
}
