package ports

import (
	"context"

	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
)

// EventPublisher define cómo publicar eventos de dominio
type EventPublisher interface {
    // Publish publica un evento al sistema de mensajería
    Publish(ctx context.Context, event *domain.OrderEvent) error
    
    // PublishBatch publica múltiples eventos de forma atómica
    PublishBatch(ctx context.Context, events []*domain.OrderEvent) error
}