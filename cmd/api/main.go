package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackhorrordevscl/dispatch-orders/internal/adapters/events"
	httpAdapter "github.com/jackhorrordevscl/dispatch-orders/internal/adapters/http"
	"github.com/jackhorrordevscl/dispatch-orders/internal/adapters/postgres"
	"github.com/jackhorrordevscl/dispatch-orders/internal/application"
	"github.com/jackhorrordevscl/dispatch-orders/internal/observability"
	"github.com/joho/godotenv"
)

func main() {
    // 1. Cargar variables de entorno
    if err := godotenv.Load(); err != nil {
        slog.Warn("no .env file found, using environment variables")
    }

    // 2. Logger estructurado
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    // 3. Inicializar OpenTelemetry
    ctx := context.Background()
    tp, err := observability.InitTracer(ctx, "dispatch-orders")
    if err != nil {
        logger.Error("failed to initialize tracer", "error", err)
        os.Exit(1)
    }
    defer func() {
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        if err := tp.Shutdown(shutdownCtx); err != nil {
            logger.Error("tracer shutdown error", "error", err)
        }
    }()

    logger.Info("OpenTelemetry initialized", "service", "dispatch-orders")

    // 4. Conectar a PostgreSQL
    dbConfig, err := postgres.LoadConfig()
    if err != nil {
        logger.Error("failed to load db config", "error", err)
        os.Exit(1)
    }

    connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    pool, err := postgres.NewPool(connCtx, dbConfig)
    if err != nil {
        logger.Error("failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer pool.Close()

    logger.Info("database connected", "host", dbConfig.Host, "port", dbConfig.Port)

    // 5. Instanciar repositorios y servicios
    orderRepo    := postgres.NewOrderRepository(pool)
    eventRepo    := postgres.NewEventRepository(pool)
    publisher    := events.NewInMemoryPublisher(logger)
    orderService := application.NewOrderService(orderRepo, eventRepo, publisher)

    // 6. Configurar HTTP
    handler := httpAdapter.NewHandler(orderService, logger)
    router  := httpAdapter.NewRouter(handler)

    port := getEnv("SERVER_PORT", "8080")
    server := &http.Server{
        Addr:         fmt.Sprintf(":%s", port),
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // 7. Arrancar con graceful shutdown
    go func() {
        logger.Info("server starting", "port", port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error("server error", "error", err)
            os.Exit(1)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("shutting down server...")

    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := server.Shutdown(shutdownCtx); err != nil {
        logger.Error("forced shutdown", "error", err)
    }

    logger.Info("server stopped")
}

func getEnv(key, defaultValue string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultValue
}