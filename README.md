# Dispatch Orders — Microservicio de Órdenes de Despacho

Microservicio en Go para gestión de órdenes de despacho, construido con arquitectura hexagonal. Permite crear órdenes, consultar su estado, realizar transiciones de estado y observar el historial de eventos generados.

---

## Tecnologías

| Capa | Tecnología |
|---|---|
| Lenguaje | Go 1.23 |
| Base de datos | PostgreSQL 16 (Alpine) |
| Driver DB | pgx/v5 + pgxpool |
| Router HTTP | chi v5 |
| Contenedores | Docker + Docker Compose |
| Logs | slog (JSON estructurado) |
| CI | GitHub Actions |

---

## Arquitectura

El proyecto sigue **arquitectura hexagonal** (Ports & Adapters):

```
dispatch-orders/
├── cmd/api/              # Punto de entrada (main.go)
├── internal/
│   ├── domain/           # Entidades, reglas de negocio, errores
│   ├── ports/            # Interfaces (contratos)
│   ├── adapters/
│   │   ├── http/         # Handler HTTP + router chi
│   │   ├── postgres/     # Repositorios PostgreSQL
│   │   └── events/       # Publisher de eventos (in-memory)
│   └── application/      # Casos de uso (orquestación)
├── migrations/           # SQL de creación de tablas
├── tests/                # Tests unitarios con mocks
├── .github/workflows/    # CI pipeline
├── docker-compose.yml
├── Makefile
└── .env
```

**Flujo de una petición:**

```
HTTP Request → Handler → OrderService → OrderRepository (PostgreSQL)
                                      → EventRepository (PostgreSQL)
                                      → EventPublisher  (logs JSON)
```

---

## Requisitos previos

- [Go 1.23+](https://go.dev/dl/)
- [Docker + Docker Compose](https://docs.docker.com/get-docker/)
- `make` (opcional, pero recomendado)

---

## Configuración

### 1. Clonar el repositorio

```bash
git clone https://github.com/jackhorrordevscl/dispatch-orders.git
cd dispatch-orders
```

### 2. Configurar variables de entorno

Edita el archivo `.env` en la raíz del proyecto:

```env
# Base de datos
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=dispatch_orders
DB_SSLMODE=disable

# Servidor
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Entorno
ENV=development
```

> **Nota:** El puerto externo de PostgreSQL es `5433` para evitar conflictos con instalaciones locales. Internamente el contenedor sigue usando `5432`.

### 3. Levantar la base de datos

```bash
make db-up
```

Las migraciones se ejecutan automáticamente al iniciar el contenedor (via `docker-entrypoint-initdb.d`).

### 4. Iniciar el servidor

```bash
make run
```

El servidor estará disponible en `http://localhost:8080`.

---

## Comandos disponibles

```bash
make build          # Compilar binario en bin/dispatch-orders
make run            # Ejecutar sin compilar
make dev            # Levantar DB + servidor en un solo comando
make test           # Ejecutar todos los tests
make test-coverage  # Tests con reporte HTML de cobertura
make lint           # Ejecutar linter (instala golangci-lint si falta)
make db-up          # Levantar PostgreSQL
make db-down        # Detener contenedores
make db-shell       # Abrir psql interactivo
make tidy           # go mod tidy
make clean          # Eliminar binarios
```

---

## API Reference

### Health Check

```
GET /health
```

```json
{ "status": "ok" }
```

---

### Crear orden

```
POST /orders
Content-Type: application/json
```

**Body:**
```json
{
  "customer_id": "customer-123",
  "items": [
    { "sku": "ABC123", "quantity": 2 },
    { "sku": "XYZ789", "quantity": 1 }
  ],
  "metadata": {
    "warehouse": "SCL-01",
    "priority": "high",
    "tags": ["fragile", "international"]
  }
}
```

**Response `201 Created`:**
```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "customer_id": "customer-123",
  "status": "created",
  "metadata": {
    "warehouse": "SCL-01",
    "priority": "high",
    "tags": ["fragile", "international"]
  },
  "items": [
    { "sku": "ABC123", "quantity": 2 },
    { "sku": "XYZ789", "quantity": 1 }
  ],
  "created_at": "2026-03-26T21:30:00Z",
  "updated_at": "2026-03-26T21:30:00Z"
}
```

---

### Obtener orden por ID

```
GET /orders/{id}
```

**Response `200 OK`:** misma estructura que la creación.

**Response `404 Not Found`:**
```json
{ "error": "order not found" }
```

---

### Listar órdenes

```
GET /orders
GET /orders?customer_id=customer-123
GET /orders?status=processing
```

**Response `200 OK`:**
```json
[
  { ... },
  { ... }
]
```

---

### Cambiar estado de una orden

```
PATCH /orders/{id}/status
Content-Type: application/json
```

**Body:**
```json
{ "status": "processing" }
```

**Response `200 OK`:** orden actualizada.

**Response `422 Unprocessable Entity`** (transición inválida):
```json
{ "error": "invalid status transition" }
```

---

### Eliminar orden

```
DELETE /orders/{id}
```

**Response `204 No Content`**

---

### Historial de eventos de una orden

```
GET /orders/{id}/events
```

**Response `200 OK`:**
```json
[
  {
    "id": "a1b2c3d4-...",
    "order_id": "f47ac10b-...",
    "type": "order.created",
    "data": {
      "customer_id": "customer-123",
      "status": "created",
      "items_count": 2
    },
    "created_at": "2026-03-26T21:30:00Z"
  },
  {
    "id": "b2c3d4e5-...",
    "order_id": "f47ac10b-...",
    "type": "order.status.changed",
    "data": {
      "old_status": "created",
      "new_status": "processing"
    },
    "created_at": "2026-03-26T21:35:00Z"
  }
]
```

---

## Transiciones de estado válidas

```
created ──→ processing ──→ shipped ──→ delivered
   │              │
   └──────────────┴──→ cancelled
```

Cualquier transición fuera de este diagrama retorna `422 Unprocessable Entity`.

---

## Ejemplos con curl

```bash
# Crear orden
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "items": [{"sku": "ABC123", "quantity": 2}],
    "metadata": {"warehouse": "SCL-01", "priority": "high"}
  }' | jq .

# Obtener orden (reemplaza el UUID)
curl -s http://localhost:8080/orders/f47ac10b-58cc-4372-a567-0e02b2c3d479 | jq .

# Listar órdenes
curl -s http://localhost:8080/orders | jq .

# Cambiar estado
curl -s -X PATCH http://localhost:8080/orders/f47ac10b-58cc-4372-a567-0e02b2c3d479/status \
  -H "Content-Type: application/json" \
  -d '{"status": "processing"}' | jq .

# Ver eventos
curl -s http://localhost:8080/orders/f47ac10b-58cc-4372-a567-0e02b2c3d479/events | jq .

# Eliminar orden
curl -s -X DELETE http://localhost:8080/orders/f47ac10b-58cc-4372-a567-0e02b2c3d479
```

---

## Tests

```bash
# Ejecutar tests
make test

# Con cobertura
make test-coverage
```

Los tests usan **mocks en memoria** para repositorios y publisher, sin necesidad de una base de datos real. Cubren: creación de órdenes, validaciones de dominio, transiciones de estado válidas e inválidas, eliminación y consulta de eventos.

---

## Modelo de datos

### Tabla `orders`

| Columna | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador único |
| customer_id | VARCHAR | ID del cliente |
| status | ENUM | Estado actual de la orden |
| metadata | JSONB | Datos flexibles (warehouse, priority, tags) |
| items | JSONB | Lista de items con SKU y cantidad |
| created_at | TIMESTAMPTZ | Fecha de creación |
| updated_at | TIMESTAMPTZ | Última modificación (auto-update via trigger) |

### Tabla `order_events`

| Columna | Tipo | Descripción |
|---|---|---|
| id | UUID | Identificador del evento |
| order_id | UUID | Referencia a la orden |
| type | VARCHAR | Tipo de evento |
| data | JSONB | Payload del evento |
| created_at | TIMESTAMPTZ | Fecha del evento |

---

## CI/CD

El pipeline de GitHub Actions (`.github/workflows/ci.yml`) se ejecuta en cada push a `main` o `develop` y en pull requests. Levanta un contenedor de PostgreSQL, compila el proyecto y ejecuta todos los tests.