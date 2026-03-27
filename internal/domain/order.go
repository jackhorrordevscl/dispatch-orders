package domain

import (
	"time"

	"github.com/google/uuid"
)

// Order representa una orden de despacho
type Order struct {
    ID         uuid.UUID              `json:"id"`
    CustomerID string                 `json:"customer_id"`
    Status     OrderStatus            `json:"status"`
    Metadata   map[string]interface{} `json:"metadata"`
    Items      []Item                 `json:"items"`
    CreatedAt  time.Time              `json:"created_at"`
    UpdatedAt  time.Time              `json:"updated_at"`
}

// NewOrder crea una nueva orden con validaciones
func NewOrder(customerID string, items []Item, metadata map[string]interface{}) (*Order, error) {
    if customerID == "" {
        return nil, ErrInvalidCustomerID
    }
    
    if len(items) == 0 {
        return nil, ErrInvalidItems
    }
    
    // Validar cada item
    for _, item := range items {
        if err := item.Validate(); err != nil {
            return nil, err
        }
    }
    
    // Si metadata es nil, inicializar como mapa vacío
    if metadata == nil {
        metadata = make(map[string]interface{})
    }
    
    now := time.Now()
    
    return &Order{
        ID:         uuid.New(),
        CustomerID: customerID,
        Status:     OrderStatusCreated,
        Metadata:   metadata,
        Items:      items,
        CreatedAt:  now,
        UpdatedAt:  now,
    }, nil
}

// UpdateStatus actualiza el estado de la orden con validaciones
func (o *Order) UpdateStatus(newStatus OrderStatus) error {
    if !newStatus.IsValid() {
        return ErrInvalidOrderStatus
    }
    
    if !o.Status.CanTransitionTo(newStatus) {
        return ErrInvalidStatusTransition
    }
    
    o.Status = newStatus
    o.UpdatedAt = time.Now()
    
    return nil
}

// Validate verifica que la orden sea válida
func (o *Order) Validate() error {
    if o.CustomerID == "" {
        return ErrInvalidCustomerID
    }
    
    if !o.Status.IsValid() {
        return ErrInvalidOrderStatus
    }
    
    if len(o.Items) == 0 {
        return ErrInvalidItems
    }
    
    for _, item := range o.Items {
        if err := item.Validate(); err != nil {
            return err
        }
    }
    
    return nil
}