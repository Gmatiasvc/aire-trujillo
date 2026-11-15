# Guía de Implementación — Backend (`backend/`)

> Documento autocontenido para el chat de implementación del Backend. Antes de editar, leer también [01-system-architecture.md](../01-system-architecture.md), [02-domain-and-data.md](../02-domain-and-data.md) y [03-quality-and-testing.md](../03-quality-and-testing.md).

---

## 1. Estado Actual del Scaffold
El directorio `backend/` ya cuenta con:
* `go.mod`: módulo `github.com/gmatiasvc/air-quality-trujillo/backend`, Go `1.25.0`, con `chi/v5`, `go-chi/cors` y `pgx/v5`.
* Carpetas y archivos vacíos (0 bytes):
  * `cmd/server/main.go`
  * `internal/config/config.go`
  * `internal/domain/models.go`
  * `internal/store/postgres.go`
  * `internal/spatial/idw.go`
  * `internal/cache/memory.go`
  * `internal/api/router.go` y `internal/api/handlers.go`
  * `internal/simulator/engine.go` y `internal/simulator/stations.go`
  * `migrations/001_init_postgis.sql` y `migrations/002_seed_stations.sql`
  * `Dockerfile`

## 2. Organización Recomendada de Paquetes
Aprovechar los paquetes existentes y añadir los módulos faltantes de forma limpia:
* `internal/domain/models.go` — Estructuras espejo exacto de [02-domain-and-data.md](../02-domain-and-data.md).
* `internal/aqi/` — Cálculo de subíndices y categorías **ECA/INCA (Perú)** y **AQI (EPA)** (`BE-005`).
* `internal/spatial/idw.go` — Interpolación IDW (`p=2`), grilla `50 m`, `R_fade=400 m`, `R_max=600 m`, mezcla RGB hacia `#9CA3AF` y cálculo puntual (`BE-006`).
* `internal/auth/` — Verificación de *passphrase* con Argon2id (`golang.org/x/crypto/argon2`) + salt, emisión/validación de cookie `HttpOnly` firmada (HMAC-SHA256) y *rate limiter* por IP (`BE-002`).
* `internal/mqtt/` — Cliente Eclipse Mosquitto (`github.com/eclipse/paho.mqtt.golang`) con reconexión automática e ingesta hacia Postgres + invalidación de caché (`BE-004`).
* `internal/store/postgres.go` — Repositorio `pgxpool` para `nodes` y `telemetry_readings` (upsert en lotes, consultas de ventana 24h y snapshot).
* `internal/simulator/` — Servicio de carga y validación de series simuladas (`BE-003`) y utilidades opcionales de seed/demo (sin usar `math/rand` para datos falsos no configurados).
* `internal/cache/memory.go` — Caché concurrente (`sync.RWMutex`) para grillas de heatmap y timestamps de playback, invalidada al insertar lecturas o modificar nodos.
* `internal/api/` — Router `chi/v5`, middlewares (CORS, logging, auth admin) y handlers REST (`BE-007`).

## 3. Variables de Entorno Esperadas
| Variable | Default / Ejemplo | Uso |
| :--- | :--- | :--- |
| `PORT` | `8080` | Puerto HTTP del servidor Go |
| `DATABASE_URL` | `postgres://...` | Cadena de conexión PostgreSQL 16 + PostGIS |
| `MQTT_BROKER_URL` | `tcp://mosquitto:1883` | URL de Eclipse Mosquitto |
| `MQTT_ENABLED` | `true` | Permite arrancar en free tier sin broker si se usa solo HTTP/admin |
| `ADMIN_PASSPHRASE_HASH` | hex/base64 | Hash Argon2id de la passphrase |
| `ADMIN_PASSPHRASE_SALT` | hex/base64 | Salt para Argon2id |
| `SESSION_SECRET` | string aleatorio | Clave HMAC para firmar cookie `aire_admin_session` |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | Orígenes permitidos |
| `INTERP_CELL_SIZE_M` | `50` | Tamaño de celda de grilla (m) |
| `INTERP_IDW_POWER` | `2` | Exponente IDW |
| `INTERP_FADE_START_M` | `400` | Inicio de degradado a gris (m) |
| `INTERP_MAX_RADIUS_M` | `600` | Radio máximo de cobertura (m) |

## 4. Notas Importantes para Free Tier y Demo
* Incluir en `docker-compose.yml` (en la raíz del repo): `postgres` (`postgis/postgis:16-3.4-alpine`), `mosquitto` (`eclipse-mosquitto:2`), y `api` (`./backend`).
* Incluir una herramienta CLI pequeña o subcomando (ej. `go run ./cmd/hashphrase`) para que el usuario genere fácilmente su `ADMIN_PASSPHRASE_HASH` + `ADMIN_PASSPHRASE_SALT`.
* Mantener `migrations/002_seed_stations.sql` idempotente (opcionalmente vacío o con nodos de prueba solo si se habilita seed de desarrollo, respetando que en producción/demo el administrador pueda gestionar todo desde `/edit`).
* Al finalizar cada feature, actualizar el estado en [features.md](./features.md).
