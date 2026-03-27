package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
)

// EventRepository define operaciones para persistir eventos
type EventRepository interface {
    // Save guarda un evento en la base de datos
    Save(ctx context.Context, event *domain.OrderEvent) error
    
    // GetByOrderID obtiene todos los eventos de una orden
    GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderEvent, error)
    
    // GetByType obtiene eventos por tipo
    GetByType(ctx context.Context, eventType domain.EventType, limit int) ([]*domain.OrderEvent, error)
}