package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
	"github.com/jackhorrordevscl/dispatch-orders/internal/ports"
)

// OrderService contiene los casos de uso de órdenes
type OrderService struct {
    orderRepo      ports.OrderRepository
    eventRepo      ports.EventRepository
    eventPublisher ports.EventPublisher
}

// NewOrderService crea una nueva instancia del servicio
func NewOrderService(
    orderRepo ports.OrderRepository,
    eventRepo ports.EventRepository,
    eventPublisher ports.EventPublisher,
) *OrderService {
    return &OrderService{
        orderRepo:      orderRepo,
        eventRepo:      eventRepo,
        eventPublisher: eventPublisher,
    }
}

// CreateOrderInput representa los datos de entrada para crear una orden
type CreateOrderInput struct {
    CustomerID string
    Items      []domain.Item
    Metadata   map[string]interface{}
}

// UpdateStatusInput representa los datos para cambiar estado
type UpdateStatusInput struct {
    OrderID   uuid.UUID
    NewStatus domain.OrderStatus
}

// CreateOrder crea una nueva orden de despacho
func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
    // 1. Crear entidad de dominio (con validaciones)
    order, err := domain.NewOrder(input.CustomerID, input.Items, input.Metadata)
    if err != nil {
        return nil, fmt.Errorf("invalid order data: %w", err)
    }

    // 2. Persistir en base de datos
    if err := s.orderRepo.Create(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to save order: %w", err)
    }

    // 3. Crear y persistir evento
    event := domain.NewOrderCreatedEvent(order)
    if err := s.eventRepo.Save(ctx, event); err != nil {
        // Log pero no falla (el pedido ya fue creado)
        fmt.Printf("warning: failed to save event: %v\n", err)
    }

    // 4. Publicar evento
    if err := s.eventPublisher.Publish(ctx, event); err != nil {
        fmt.Printf("warning: failed to publish event: %v\n", err)
    }

    return order, nil
}

// GetOrder obtiene una orden por ID
func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
    order, err := s.orderRepo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    return order, nil
}

// ListOrders lista órdenes con filtros opcionales
func (s *OrderService) ListOrders(ctx context.Context, filters ports.OrderFilters) ([]*domain.Order, error) {
    return s.orderRepo.List(ctx, filters)
}

// UpdateOrderStatus cambia el estado de una orden
func (s *OrderService) UpdateOrderStatus(ctx context.Context, input UpdateStatusInput) (*domain.Order, error) {
    // 1. Obtener orden actual
    order, err := s.orderRepo.GetByID(ctx, input.OrderID)
    if err != nil {
        return nil, err
    }

    // 2. Guardar estado anterior para el evento
    oldStatus := order.Status

    // 3. Actualizar estado (con validación de transición)
    if err := order.UpdateStatus(input.NewStatus); err != nil {
        return nil, err
    }

    // 4. Persistir cambio
    if err := s.orderRepo.Update(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to update order: %w", err)
    }

    // 5. Emitir evento
    event := domain.NewOrderStatusChangedEvent(order, oldStatus, input.NewStatus)
    if err := s.eventRepo.Save(ctx, event); err != nil {
        fmt.Printf("warning: failed to save event: %v\n", err)
    }
    if err := s.eventPublisher.Publish(ctx, event); err != nil {
        fmt.Printf("warning: failed to publish event: %v\n", err)
    }

    return order, nil
}

// DeleteOrder elimina una orden
func (s *OrderService) DeleteOrder(ctx context.Context, id uuid.UUID) error {
    return s.orderRepo.Delete(ctx, id)
}

// GetOrderEvents obtiene el historial de eventos de una orden
func (s *OrderService) GetOrderEvents(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderEvent, error) {
    return s.eventRepo.GetByOrderID(ctx, orderID)
}