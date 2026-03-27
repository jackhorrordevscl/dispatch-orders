package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
)

// OrderRepository define las operaciones de persistencia de órdenes
type OrderRepository interface {
    // Create guarda una nueva orden en la base de datos
    Create(ctx context.Context, order *domain.Order) error
    
    // GetByID obtiene una orden por su ID
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
    
    // Update actualiza una orden existente
    Update(ctx context.Context, order *domain.Order) error
    
    // Delete elimina una orden (soft delete recomendado)
    Delete(ctx context.Context, id uuid.UUID) error
    
    // List obtiene órdenes con filtros opcionales
    List(ctx context.Context, filters OrderFilters) ([]*domain.Order, error)
}

// OrderFilters representa filtros para búsqueda de órdenes
type OrderFilters struct {
    CustomerID *string
    Status     *domain.OrderStatus
    Limit      int
    Offset     int
}