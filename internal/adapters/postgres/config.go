package postgres

import (
	"fmt"
	"os"
	"strconv"
)

// Config contiene la configuración de PostgreSQL
type Config struct {
    Host     string
    Port     int
    User     string
    Password string
    Database string
    SSLMode  string
}

// LoadConfig carga la configuración desde variables de entorno
func LoadConfig() (*Config, error) {
    port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
    if err != nil {
        return nil, fmt.Errorf("invalid DB_PORT: %w", err)
    }
    
    return &Config{
        Host:     getEnv("DB_HOST", "localhost"),
        Port:     port,
        User:     getEnv("DB_USER", "postgres"),
        Password: getEnv("DB_PASSWORD", "postgres"),
        Database: getEnv("DB_NAME", "dispatch_orders"),
        SSLMode:  getEnv("DB_SSLMODE", "disable"),
    }, nil
}

// DSN genera el string de conexión
func (c *Config) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
    )
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}