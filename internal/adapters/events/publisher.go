package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
)

// InMemoryPublisher implementa EventPublisher usando logs estructurados
// En producción se reemplazaría por Kafka, RabbitMQ, etc.
type InMemoryPublisher struct {
    logger *slog.Logger
}

// NewInMemoryPublisher crea un publisher que loguea los eventos
func NewInMemoryPublisher(logger *slog.Logger) *InMemoryPublisher {
    return &InMemoryPublisher{logger: logger}
}

// Publish publica un evento (lo loguea como JSON estructurado)
func (p *InMemoryPublisher) Publish(ctx context.Context, event *domain.OrderEvent) error {
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    p.logger.InfoContext(ctx, "event published",
        "event_id", event.ID,
        "event_type", event.Type,
        "order_id", event.OrderID,
        "payload", string(data),
    )

    return nil
}

// PublishBatch publica múltiples eventos
func (p *InMemoryPublisher) PublishBatch(ctx context.Context, events []*domain.OrderEvent) error {
    for _, event := range events {
        if err := p.Publish(ctx, event); err != nil {
            return err
        }
    }
    return nil
}