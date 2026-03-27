package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
	"github.com/jackhorrordevscl/dispatch-orders/internal/ports"
)

// OrderRepository implementa ports.OrderRepository usando PostgreSQL
type OrderRepository struct {
    pool *pgxpool.Pool
}

// NewOrderRepository crea una nueva instancia del repositorio
func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
    return &OrderRepository{
        pool: pool,
    }
}

// Create guarda una nueva orden en PostgreSQL
func (r *OrderRepository) Create(ctx context.Context, order *domain.Order) error {
    // Serializar metadata e items a JSONB
    metadataJSON, err := json.Marshal(order.Metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }
    
    itemsJSON, err := json.Marshal(order.Items)
    if err != nil {
        return fmt.Errorf("failed to marshal items: %w", err)
    }
    
    query := `
        INSERT INTO orders (id, customer_id, status, metadata, items, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
    
    _, err = r.pool.Exec(ctx, query,
        order.ID,
        order.CustomerID,
        order.Status,
        metadataJSON,
        itemsJSON,
        order.CreatedAt,
        order.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }
    
    return nil
}

// GetByID obtiene una orden por su ID
func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
    query := `
        SELECT id, customer_id, status, metadata, items, created_at, updated_at
        FROM orders
        WHERE id = $1
    `
    
    var order domain.Order
    var metadataJSON []byte
    var itemsJSON []byte
    var status string
    
    err := r.pool.QueryRow(ctx, query, id).Scan(
        &order.ID,
        &order.CustomerID,
        &status,
        &metadataJSON,
        &itemsJSON,
        &order.CreatedAt,
        &order.UpdatedAt,
    )
    
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, domain.ErrOrderNotFound
        }
        return nil, fmt.Errorf("failed to get order: %w", err)
    }
    
    // Deserializar JSONB
    if err := json.Unmarshal(metadataJSON, &order.Metadata); err != nil {
        return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
    }
    
    if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
        return nil, fmt.Errorf("failed to unmarshal items: %w", err)
    }
    
    order.Status = domain.OrderStatus(status)
    
    return &order, nil
}

// Update actualiza una orden existente
func (r *OrderRepository) Update(ctx context.Context, order *domain.Order) error {
    metadataJSON, err := json.Marshal(order.Metadata)
    if err != nil {
        return fmt.Errorf("failed to marshal metadata: %w", err)
    }
    
    itemsJSON, err := json.Marshal(order.Items)
    if err != nil {
        return fmt.Errorf("failed to marshal items: %w", err)
    }
    
    query := `
        UPDATE orders
        SET customer_id = $2, status = $3, metadata = $4, items = $5, updated_at = $6
        WHERE id = $1
    `
    
    result, err := r.pool.Exec(ctx, query,
        order.ID,
        order.CustomerID,
        order.Status,
        metadataJSON,
        itemsJSON,
        order.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    
    if result.RowsAffected() == 0 {
        return domain.ErrOrderNotFound
    }
    
    return nil
}

// Delete elimina una orden (implementación real - se puede cambiar a soft delete)
func (r *OrderRepository) Delete(ctx context.Context, id uuid.UUID) error {
    query := `DELETE FROM orders WHERE id = $1`
    
    result, err := r.pool.Exec(ctx, query, id)
    if err != nil {
        return fmt.Errorf("failed to delete order: %w", err)
    }
    
    if result.RowsAffected() == 0 {
        return domain.ErrOrderNotFound
    }
    
    return nil
}

// List obtiene órdenes con filtros
func (r *OrderRepository) List(ctx context.Context, filters ports.OrderFilters) ([]*domain.Order, error) {
    query := `
        SELECT id, customer_id, status, metadata, items, created_at, updated_at
        FROM orders
        WHERE 1=1
    `
    
    args := []interface{}{}
    argPos := 1
    
    // Filtro por customer_id
    if filters.CustomerID != nil {
        query += fmt.Sprintf(" AND customer_id = $%d", argPos)
        args = append(args, *filters.CustomerID)
        argPos++
    }
    
    // Filtro por status
    if filters.Status != nil {
        query += fmt.Sprintf(" AND status = $%d", argPos)
        args = append(args, *filters.Status)
        argPos++
    }
    
    // Ordenar por fecha de creación descendente
    query += " ORDER BY created_at DESC"
    
    // Límite y offset
    if filters.Limit > 0 {
        query += fmt.Sprintf(" LIMIT $%d", argPos)
        args = append(args, filters.Limit)
        argPos++
    }
    
    if filters.Offset > 0 {
        query += fmt.Sprintf(" OFFSET $%d", argPos)
        args = append(args, filters.Offset)
    }
    
    rows, err := r.pool.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to list orders: %w", err)
    }
    defer rows.Close()
    
    var orders []*domain.Order
    
    for rows.Next() {
        var order domain.Order
        var metadataJSON []byte
        var itemsJSON []byte
        var status string
        
        err := rows.Scan(
            &order.ID,
            &order.CustomerID,
            &status,
            &metadataJSON,
            &itemsJSON,
            &order.CreatedAt,
            &order.UpdatedAt,
        )
        
        if err != nil {
            return nil, fmt.Errorf("failed to scan order: %w", err)
        }
        
        if err := json.Unmarshal(metadataJSON, &order.Metadata); err != nil {
            return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
        }
        
        if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
            return nil, fmt.Errorf("failed to unmarshal items: %w", err)
        }
        
        order.Status = domain.OrderStatus(status)
        orders = append(orders, &order)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("rows iteration error: %w", err)
    }
    
    return orders, nil
}