package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/jackhorrordevscl/dispatch-orders/internal/application"
	"github.com/jackhorrordevscl/dispatch-orders/internal/domain"
	"github.com/jackhorrordevscl/dispatch-orders/internal/ports"
)

const tracerName = "dispatch-orders/http"

// Handler contiene todos los handlers HTTP
type Handler struct {
    service *application.OrderService
    logger  *slog.Logger
}

// NewHandler crea una nueva instancia del handler
func NewHandler(service *application.OrderService, logger *slog.Logger) *Handler {
    return &Handler{service: service, logger: logger}
}

// NewRouter crea y configura el router con todas las rutas
func NewRouter(h *Handler) http.Handler {
    r := chi.NewRouter()

    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Recoverer)

    r.Get("/health", h.healthCheck)

    r.Route("/orders", func(r chi.Router) {
        r.Post("/", h.createOrder)
        r.Get("/", h.listOrders)

        r.Route("/{id}", func(r chi.Router) {
            r.Get("/", h.getOrder)
            r.Patch("/status", h.updateStatus)
            r.Delete("/", h.deleteOrder)
            r.Get("/events", h.getOrderEvents)
        })
    })

    return r
}

// ── Requests & Responses ──────────────────────────────────────────────────────

type createOrderRequest struct {
    CustomerID string                 `json:"customer_id"`
    Items      []domain.Item          `json:"items"`
    Metadata   map[string]interface{} `json:"metadata"`
}

type updateStatusRequest struct {
    Status string `json:"status"`
}

type errorResponse struct {
    Error string `json:"error"`
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
    writeJSON(w, status, errorResponse{Error: msg})
}

func parseUUID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        writeError(w, http.StatusBadRequest, "invalid UUID format")
        return uuid.UUID{}, false
    }
    return id, true
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer(tracerName).Start(r.Context(), "http.createOrder")
    defer span.End()

    var req createOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        span.SetStatus(codes.Error, "invalid body")
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    span.SetAttributes(attribute.String("customer_id", req.CustomerID))

    order, err := h.service.CreateOrder(ctx, application.CreateOrderInput{
        CustomerID: req.CustomerID,
        Items:      req.Items,
        Metadata:   req.Metadata,
    })
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        h.handleDomainError(w, err)
        return
    }

    span.SetAttributes(attribute.String("order_id", order.ID.String()))
    span.SetStatus(codes.Ok, "order created")
    writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer(tracerName).Start(r.Context(), "http.getOrder")
    defer span.End()

    id, ok := parseUUID(w, r)
    if !ok {
        span.SetStatus(codes.Error, "invalid uuid")
        return
    }

    span.SetAttributes(attribute.String("order_id", id.String()))

    order, err := h.service.GetOrder(ctx, id)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        h.handleDomainError(w, err)
        return
    }

    span.SetStatus(codes.Ok, "")
    writeJSON(w, http.StatusOK, order)
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer(tracerName).Start(r.Context(), "http.listOrders")
    defer span.End()

    filters := ports.OrderFilters{Limit: 20}

    if cid := r.URL.Query().Get("customer_id"); cid != "" {
        filters.CustomerID = &cid
        span.SetAttributes(attribute.String("filter.customer_id", cid))
    }
    if st := r.URL.Query().Get("status"); st != "" {
        status := domain.OrderStatus(st)
        filters.Status = &status
        span.SetAttributes(attribute.String("filter.status", st))
    }

    orders, err := h.service.ListOrders(ctx, filters)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        writeError(w, http.StatusInternalServerError, "failed to list orders")
        return
    }

    if orders == nil {
        orders = []*domain.Order{}
    }

    span.SetAttributes(attribute.Int("result.count", len(orders)))
    span.SetStatus(codes.Ok, "")
    writeJSON(w, http.StatusOK, orders)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer(tracerName).Start(r.Context(), "http.updateStatus")
    defer span.End()

    id, ok := parseUUID(w, r)
    if !ok {
        span.SetStatus(codes.Error, "invalid uuid")
        return
    }

    var req updateStatusRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        span.SetStatus(codes.Error, "invalid body")
        writeError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    span.SetAttributes(
        attribute.String("order_id", id.String()),
        attribute.String("new_status", req.Status),
    )

    order, err := h.service.UpdateOrderStatus(ctx, application.UpdateStatusInput{
        OrderID:   id,
        NewStatus: domain.OrderStatus(req.Status),
    })
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        h.handleDomainError(w, err)
        return
    }

    span.SetStatus(codes.Ok, "status updated")
    writeJSON(w, http.StatusOK, order)
}

func (h *Handler) deleteOrder(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer(tracerName).Start(r.Context(), "http.deleteOrder")
    defer span.End()

    id, ok := parseUUID(w, r)
    if !ok {
        span.SetStatus(codes.Error, "invalid uuid")
        return
    }

    span.SetAttributes(attribute.String("order_id", id.String()))

    if err := h.service.DeleteOrder(ctx, id); err != nil {
        span.SetStatus(codes.Error, err.Error())
        h.handleDomainError(w, err)
        return
    }

    span.SetStatus(codes.Ok, "")
    w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getOrderEvents(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer(tracerName).Start(r.Context(), "http.getOrderEvents")
    defer span.End()

    id, ok := parseUUID(w, r)
    if !ok {
        span.SetStatus(codes.Error, "invalid uuid")
        return
    }

    span.SetAttributes(attribute.String("order_id", id.String()))

    events, err := h.service.GetOrderEvents(ctx, id)
    if err != nil {
        span.SetStatus(codes.Error, err.Error())
        h.handleDomainError(w, err)
        return
    }

    if events == nil {
        events = []*domain.OrderEvent{}
    }

    span.SetAttributes(attribute.Int("result.count", len(events)))
    span.SetStatus(codes.Ok, "")
    writeJSON(w, http.StatusOK, events)
}

func (h *Handler) handleDomainError(w http.ResponseWriter, err error) {
    switch {
    case errors.Is(err, domain.ErrOrderNotFound):
        writeError(w, http.StatusNotFound, err.Error())
    case errors.Is(err, domain.ErrInvalidStatusTransition):
        writeError(w, http.StatusUnprocessableEntity, err.Error())
    case errors.Is(err, domain.ErrInvalidOrderStatus),
        errors.Is(err, domain.ErrInvalidCustomerID),
        errors.Is(err, domain.ErrInvalidItems),
        errors.Is(err, domain.ErrInvalidSKU),
        errors.Is(err, domain.ErrInvalidQuantity):
        writeError(w, http.StatusBadRequest, err.Error())
    default:
        h.logger.Error("internal error", "error", err)
        writeError(w, http.StatusInternalServerError, "internal server error")
    }
}