package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
)

// EventRepository implementa ports.EventRepository usando PostgreSQL
type EventRepository struct {
    pool *pgxpool.Pool
}

// NewEventRepository crea una nueva instancia
func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
    return &EventRepository{pool: pool}
}

// Save guarda un evento en la base de datos
func (r *EventRepository) Save(ctx context.Context, event *domain.OrderEvent) error {
    dataJSON, err := json.Marshal(event.Data)
    if err != nil {
        return fmt.Errorf("failed to marshal event data: %w", err)
    }

    query := `
        INSERT INTO order_events (id, order_id, type, data, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `

    _, err = r.pool.Exec(ctx, query,
        event.ID,
        event.OrderID,
        event.Type,
        dataJSON,
        event.CreatedAt,
    )
    if err != nil {
        return fmt.Errorf("failed to save event: %w", err)
    }

    return nil
}

// GetByOrderID obtiene todos los eventos de una orden
func (r *EventRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderEvent, error) {
    query := `
        SELECT id, order_id, type, data, created_at
        FROM order_events
        WHERE order_id = $1
        ORDER BY created_at ASC
    `

    rows, err := r.pool.Query(ctx, query, orderID)
    if err != nil {
        return nil, fmt.Errorf("failed to query events: %w", err)
    }
    defer rows.Close()

    var events []*domain.OrderEvent

    for rows.Next() {
        var e domain.OrderEvent
        var dataJSON []byte
        var eventType string

        if err := rows.Scan(&e.ID, &e.OrderID, &eventType, &dataJSON, &e.CreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan event: %w", err)
        }

        if err := json.Unmarshal(dataJSON, &e.Data); err != nil {
            return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
        }

        e.Type = domain.EventType(eventType)
        events = append(events, &e)
    }

    return events, rows.Err()
}

// GetByType obtiene eventos por tipo
func (r *EventRepository) GetByType(ctx context.Context, eventType domain.EventType, limit int) ([]*domain.OrderEvent, error) {
    query := `
        SELECT id, order_id, type, data, created_at
        FROM order_events
        WHERE type = $1
        ORDER BY created_at DESC
        LIMIT $2
    `

    rows, err := r.pool.Query(ctx, query, eventType, limit)
    if err != nil {
        return nil, fmt.Errorf("failed to query events by type: %w", err)
    }
    defer rows.Close()

    var events []*domain.OrderEvent

    for rows.Next() {
        var e domain.OrderEvent
        var dataJSON []byte
        var evType string

        if err := rows.Scan(&e.ID, &e.OrderID, &evType, &dataJSON, &e.CreatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan event: %w", err)
        }

        if err := json.Unmarshal(dataJSON, &e.Data); err != nil {
            return nil, fmt.Errorf("failed to unmarshal event data: %w", err)
        }

        e.Type = domain.EventType(evType)
        events = append(events, &e)
    }

    return events, rows.Err()
}