package domain

import "errors"

// OrderStatus representa los estados posibles de una orden
type OrderStatus string

const (
    OrderStatusCreated    OrderStatus = "created"
    OrderStatusProcessing OrderStatus = "processing"
    OrderStatusShipped    OrderStatus = "shipped"
    OrderStatusDelivered  OrderStatus = "delivered"
    OrderStatusCancelled  OrderStatus = "cancelled"
)

// Errores de dominio
var (
    ErrInvalidOrderStatus    = errors.New("invalid order status")
    ErrInvalidCustomerID     = errors.New("customer ID cannot be empty")
    ErrInvalidItems          = errors.New("order must have at least one item")
    ErrInvalidSKU            = errors.New("SKU cannot be empty")
    ErrInvalidQuantity       = errors.New("quantity must be greater than zero")
    ErrOrderNotFound         = errors.New("order not found")
    ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// IsValid verifica si el status es válido
func (s OrderStatus) IsValid() bool {
    switch s {
    case OrderStatusCreated, OrderStatusProcessing, 
         OrderStatusShipped, OrderStatusDelivered, 
         OrderStatusCancelled:
        return true
    }
    return false
}

// CanTransitionTo verifica si se puede cambiar de un estado a otro
func (s OrderStatus) CanTransitionTo(newStatus OrderStatus) bool {
    transitions := map[OrderStatus][]OrderStatus{
        OrderStatusCreated: {
            OrderStatusProcessing,
            OrderStatusCancelled,
        },
        OrderStatusProcessing: {
            OrderStatusShipped,
            OrderStatusCancelled,
        },
        OrderStatusShipped: {
            OrderStatusDelivered,
        },
        OrderStatusDelivered: {},
        OrderStatusCancelled: {},
    }
    
    allowedStatuses, exists := transitions[s]
    if !exists {
        return false
    }
    
    for _, allowed := range allowedStatuses {
        if allowed == newStatus {
            return true
        }
    }
    
    return false
}