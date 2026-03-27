package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventType representa el tipo de evento
type EventType string

const (
    EventOrderCreated        EventType = "order.created"
    EventOrderStatusChanged  EventType = "order.status.changed"
    EventOrderCancelled      EventType = "order.cancelled"
)

// OrderEvent representa un evento de dominio
type OrderEvent struct {
    ID        uuid.UUID              `json:"id"`
    OrderID   uuid.UUID              `json:"order_id"`
    Type      EventType              `json:"type"`
    Data      map[string]interface{} `json:"data"`
    CreatedAt time.Time              `json:"created_at"`
}

// NewOrderCreatedEvent crea un evento de orden creada
func NewOrderCreatedEvent(order *Order) *OrderEvent {
    return &OrderEvent{
        ID:      uuid.New(),
        OrderID: order.ID,
        Type:    EventOrderCreated,
        Data: map[string]interface{}{
            "customer_id": order.CustomerID,
            "status":      order.Status,
            "items_count": len(order.Items),
        },
        CreatedAt: time.Now(),
    }
}

// NewOrderStatusChangedEvent crea un evento de cambio de estado
func NewOrderStatusChangedEvent(order *Order, oldStatus, newStatus OrderStatus) *OrderEvent {
    return &OrderEvent{
        ID:      uuid.New(),
        OrderID: order.ID,
        Type:    EventOrderStatusChanged,
        Data: map[string]interface{}{
            "old_status": oldStatus,
            "new_status": newStatus,
        },
        CreatedAt: time.Now(),
    }
}

// NewOrderCancelledEvent crea un evento de orden cancelada
func NewOrderCancelledEvent(order *Order, reason string) *OrderEvent {
    return &OrderEvent{
        ID:      uuid.New(),
        OrderID: order.ID,
        Type:    EventOrderCancelled,
        Data: map[string]interface{}{
            "reason": reason,
        },
        CreatedAt: time.Now(),
    }
}