package tests

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackhorrordevscl/dispatch-orders/internal/application"
	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
	"github.com/jackhorrordevscl/dispatch-orders/internal/ports"
)

// ── Mocks ─────────────────────────────────────────────────────────────────────

type mockOrderRepo struct {
    orders map[uuid.UUID]*domain.Order
}

func newMockOrderRepo() *mockOrderRepo {
    return &mockOrderRepo{orders: make(map[uuid.UUID]*domain.Order)}
}

func (m *mockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
    m.orders[order.ID] = order
    return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
    o, ok := m.orders[id]
    if !ok {
        return nil, domain.ErrOrderNotFound
    }
    return o, nil
}

func (m *mockOrderRepo) Update(ctx context.Context, order *domain.Order) error {
    if _, ok := m.orders[order.ID]; !ok {
        return domain.ErrOrderNotFound
    }
    m.orders[order.ID] = order
    return nil
}

func (m *mockOrderRepo) Delete(ctx context.Context, id uuid.UUID) error {
    if _, ok := m.orders[id]; !ok {
        return domain.ErrOrderNotFound
    }
    delete(m.orders, id)
    return nil
}

func (m *mockOrderRepo) List(ctx context.Context, filters ports.OrderFilters) ([]*domain.Order, error) {
    var result []*domain.Order
    for _, o := range m.orders {
        result = append(result, o)
    }
    return result, nil
}

type mockEventRepo struct {
    events []*domain.OrderEvent
}

func (m *mockEventRepo) Save(ctx context.Context, e *domain.OrderEvent) error {
    m.events = append(m.events, e)
    return nil
}

func (m *mockEventRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderEvent, error) {
    var result []*domain.OrderEvent
    for _, e := range m.events {
        if e.OrderID == orderID {
            result = append(result, e)
        }
    }
    return result, nil
}

func (m *mockEventRepo) GetByType(ctx context.Context, t domain.EventType, limit int) ([]*domain.OrderEvent, error) {
    return nil, nil
}

type mockPublisher struct {
    published []*domain.OrderEvent
}

func (m *mockPublisher) Publish(ctx context.Context, e *domain.OrderEvent) error {
    m.published = append(m.published, e)
    return nil
}

func (m *mockPublisher) PublishBatch(ctx context.Context, events []*domain.OrderEvent) error {
    m.published = append(m.published, events...)
    return nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestService() (*application.OrderService, *mockOrderRepo, *mockEventRepo, *mockPublisher) {
    repo      := newMockOrderRepo()
    eventRepo := &mockEventRepo{}
    publisher := &mockPublisher{}
    svc       := application.NewOrderService(repo, eventRepo, publisher)
    return svc, repo, eventRepo, publisher
}

func sampleInput() application.CreateOrderInput {
    return application.CreateOrderInput{
        CustomerID: "customer-123",
        Items: []domain.Item{
            {SKU: "ABC123", Quantity: 2},
        },
        Metadata: map[string]interface{}{
            "warehouse": "SCL-01",
            "priority":  "high",
        },
    }
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreateOrder_Success(t *testing.T) {
    svc, _, eventRepo, publisher := newTestService()

    order, err := svc.CreateOrder(context.Background(), sampleInput())

    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if order.CustomerID != "customer-123" {
        t.Errorf("expected customer-123, got %s", order.CustomerID)
    }
    if order.Status != domain.OrderStatusCreated {
        t.Errorf("expected status 'created', got %s", order.Status)
    }
    if len(eventRepo.events) != 1 {
        t.Errorf("expected 1 event saved, got %d", len(eventRepo.events))
    }
    if len(publisher.published) != 1 {
        t.Errorf("expected 1 event published, got %d", len(publisher.published))
    }
}

func TestCreateOrder_MissingCustomerID(t *testing.T) {
    svc, _, _, _ := newTestService()

    input := sampleInput()
    input.CustomerID = ""

    _, err := svc.CreateOrder(context.Background(), input)

    if err == nil {
        t.Fatal("expected error for empty customer ID")
    }
}

func TestCreateOrder_NoItems(t *testing.T) {
    svc, _, _, _ := newTestService()

    input := sampleInput()
    input.Items = []domain.Item{}

    _, err := svc.CreateOrder(context.Background(), input)

    if err == nil {
        t.Fatal("expected error for empty items")
    }
}

func TestGetOrder_NotFound(t *testing.T) {
    svc, _, _, _ := newTestService()

    _, err := svc.GetOrder(context.Background(), uuid.New())

    if err != domain.ErrOrderNotFound {
        t.Errorf("expected ErrOrderNotFound, got %v", err)
    }
}

func TestUpdateOrderStatus_ValidTransition(t *testing.T) {
    svc, _, _, _ := newTestService()

    order, _ := svc.CreateOrder(context.Background(), sampleInput())

    updated, err := svc.UpdateOrderStatus(context.Background(), application.UpdateStatusInput{
        OrderID:   order.ID,
        NewStatus: domain.OrderStatusProcessing,
    })

    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if updated.Status != domain.OrderStatusProcessing {
        t.Errorf("expected 'processing', got %s", updated.Status)
    }
}

func TestUpdateOrderStatus_InvalidTransition(t *testing.T) {
    svc, _, _, _ := newTestService()

    order, _ := svc.CreateOrder(context.Background(), sampleInput())

    // created → delivered es una transición inválida
    _, err := svc.UpdateOrderStatus(context.Background(), application.UpdateStatusInput{
        OrderID:   order.ID,
        NewStatus: domain.OrderStatusDelivered,
    })

    if err != domain.ErrInvalidStatusTransition {
        t.Errorf("expected ErrInvalidStatusTransition, got %v", err)
    }
}

func TestDeleteOrder_Success(t *testing.T) {
    svc, _, _, _ := newTestService()

    order, _ := svc.CreateOrder(context.Background(), sampleInput())

    if err := svc.DeleteOrder(context.Background(), order.ID); err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }

    _, err := svc.GetOrder(context.Background(), order.ID)
    if err != domain.ErrOrderNotFound {
        t.Error("expected order to be deleted")
    }
}

func TestGetOrderEvents(t *testing.T) {
    svc, _, _, _ := newTestService()

    order, _ := svc.CreateOrder(context.Background(), sampleInput())
    svc.UpdateOrderStatus(context.Background(), application.UpdateStatusInput{
        OrderID:   order.ID,
        NewStatus: domain.OrderStatusProcessing,
    })

    evts, err := svc.GetOrderEvents(context.Background(), order.ID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if len(evts) != 2 {
        t.Errorf("expected 2 events, got %d", len(evts))
    }
}