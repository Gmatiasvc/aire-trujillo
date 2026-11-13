# 02 — Dominio, Contratos y Datos (`aire-trujillo`)

> **Fuente de Verdad (Zero-Drift):** Todo nombre de campo JSON, unidad física, tabla SQL, tópico MQTT y endpoint REST implementado en `hardware/`, `backend/` y `frontend/` debe coincidir exactamente con este documento.

---

## 1. Variables Medidas y Unidades Canónicas

Todas las lecturas almacenadas en base de datos y transportadas por MQTT/REST usan las siguientes claves (`snake_case`) y unidades físicas:

| Clave JSON / SQL | Variable | Sensor | Unidad Canónica | Rango Plausible |
| :--- | :--- | :--- | :--- | :--- |
| `pm25` | Material particulado PM2.5 | `Sensirion SPS30` | `µg/m³` | `0 – 1000` |
| `pm10` | Material particulado PM10 | `Sensirion SPS30` | `µg/m³` | `0 – 1000` |
| `co` | Monóxido de carbono | `Winsen ZE03-CO` (`ADS1115`) | `mg/m³` | `0 – 100` |
| `no2` | Dióxido de nitrógeno | `Winsen ZE03-NO2` (`ADS1115`) | `µg/m³` | `0 – 2000` |
| `so2` | Dióxido de azufre | `Winsen ZE03-SO2` (`ADS1115`) | `µg/m³` | `0 – 2000` |
| `h2s` | Sulfuro de hidrógeno | `Winsen ZE03-H2S` (`ADS1115`) | `µg/m³` | `0 – 2000` |
| `co2` | Dióxido de carbono | `Sensirion SCD41` | `ppm` | `400 – 10000` |
| `voc_index` | Índice VOC | `Sensirion SGP40` | `index` (1–500) | `1 – 500` |
| `noise_dba` | Nivel de ruido | `InvenSense ICS-43434` | `dBA` | `20 – 130` |
| `temperature_c` | Temperatura ambiente | `Bosch BME280` | `°C` | `-10 – 60` |
| `humidity_pct` | Humedad relativa | `Bosch BME280` | `%` | `0 – 100` |
| `pressure_hpa` | Presión atmosférica | `Bosch BME280` | `hPa` | `800 – 1100` |

---

## 2. Estándares de Calidad de Aire y Estados Visuales

### 2.1. Zona Sin Cobertura (Obligatorio)
* Cuando un punto o celda supera el radio máximo de interpolación ($R_{\max}$) o no existen nodos activos con datos:
  * `has_data`: `false`
  * `color`: `#9CA3AF` (gris)
  * `label`: `"No hay datos"`
  * **Prohibido** devolver `0` como sustituto de ausencia de datos.

### 2.2. Estándar Principal: ECA / INCA (Perú — MINAM)
El índice compuesto toma el máximo de los subíndices de contaminantes regulados disponibles (`pm25`, `pm10`, `co`, `no2`, `so2`, `h2s`).

| Rango INCA | Categoría (`category`) | Color Hex (`color`) |
| :--- | :--- | :--- |
| `0 – 50` | `Buena` | `#22C55E` |
| `51 – 100` | `Moderada` | `#EAB308` |
| `101 – 150` | `Mala` | `#F97316` |
| `> 150` | `Peligrosa` | `#EF4444` |

### 2.3. Estándar Secundario (Conmutable): AQI (EPA)
Calculado según los tramos oficiales US EPA para los contaminantes compatibles (`pm25`, `pm10`, `co`, `no2`, `so2`).

| Rango AQI | Categoría (`category`) | Color Hex (`color`) |
| :--- | :--- | :--- |
| `0 – 50` | `Good` | `#00E400` |
| `51 – 100` | `Moderate` | `#FFFF00` |
| `101 – 150` | `Unhealthy for Sensitive Groups` | `#FF7E00` |
| `151 – 200` | `Unhealthy` | `#FF0000` |
| `201 – 300` | `Very Unhealthy` | `#8F3F97` |
| `301 – 500` | `Hazardous` | `#7E0023` |

---

## 3. Contrato MQTT (Hardware / Simulador → Backend)

* **Broker:** Eclipse Mosquitto
* **Tópico de publicación:** `aire-trujillo/v1/nodes/{node_code}/telemetry`
* **QoS:** `1`

```json
{
  "node_code": "TRJ-HUANCHACO-01",
  "recorded_at": "2026-11-13T16:00:00Z",
  "metrics": {
    "pm25": 14.2,
    "pm10": 28.5,
    "co": 1.1,
    "no2": 32.0,
    "so2": 12.4,
    "h2s": 4.1,
    "co2": 435.0,
    "voc_index": 110,
    "noise_dba": 58.3,
    "temperature_c": 21.4,
    "humidity_pct": 78.2,
    "pressure_hpa": 1012.8
  }
}
```

---

## 4. Esquema de Base de Datos (PostgreSQL 16 + PostGIS)

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE nodes (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code        VARCHAR(64) NOT NULL UNIQUE,
  name        VARCHAR(120) NOT NULL,
  status      VARCHAR(16) NOT NULL CHECK (status IN ('active', 'inactive')),
  is_virtual  BOOLEAN NOT NULL DEFAULT TRUE,
  location    GEOGRAPHY(Point, 4326) NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_nodes_location ON nodes USING GIST (location);
CREATE INDEX idx_nodes_status ON nodes (status);

CREATE TABLE telemetry_readings (
  id            BIGSERIAL PRIMARY KEY,
  node_id       UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  recorded_at   TIMESTAMPTZ NOT NULL,
  pm25          DOUBLE PRECISION,
  pm10          DOUBLE PRECISION,
  co            DOUBLE PRECISION,
  no2           DOUBLE PRECISION,
  so2           DOUBLE PRECISION,
  h2s           DOUBLE PRECISION,
  co2           DOUBLE PRECISION,
  voc_index     DOUBLE PRECISION,
  noise_dba     DOUBLE PRECISION,
  temperature_c DOUBLE PRECISION,
  humidity_pct  DOUBLE PRECISION,
  pressure_hpa  DOUBLE PRECISION,
  source        VARCHAR(16) NOT NULL CHECK (source IN ('mqtt', 'admin_sim')),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (node_id, recorded_at)
);

CREATE INDEX idx_telemetry_recorded_at ON telemetry_readings (recorded_at DESC);
CREATE INDEX idx_telemetry_node_time ON telemetry_readings (node_id, recorded_at DESC);
```

---

## 5. Contratos API REST (Backend → Frontend)

### 5.1. Parámetros Comunes
* `layer`: `air_quality` (default), `pm25`, `pm10`, `co`, `no2`, `so2`, `h2s`, `co2`, `voc_index`, `noise_dba`, `temperature_c`, `humidity_pct`, `pressure_hpa`.
* `standard`: `inca` (default) | `epa` (aplica cuando `layer=air_quality`).

### 5.2. Endpoints Públicos

#### `GET /api/v1/health`
```json
{ "status": "ok" }
```

#### `GET /api/v1/nodes`
Devuelve los nodos activos para marcadores en el mapa.
```json
{
  "nodes": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "code": "TRJ-01",
      "name": "Plaza de Armas",
      "lat": -8.1116,
      "lon": -79.0288,
      "status": "active"
    }
  ]
}
```

#### `GET /api/v1/map/heatmap?layer=air_quality&standard=inca&at=2026-11-13T16:00:00Z`
Si se omite `at`, usa el snapshot más reciente.
```json
{
  "recorded_at": "2026-11-13T16:00:00Z",
  "layer": "air_quality",
  "standard": "inca",
  "bounds": {
    "min_lat": -8.15,
    "min_lon": -79.13,
    "max_lat": -8.06,
    "max_lon": -79.00
  },
  "cells": [
    {
      "lat": -8.1116,
      "lon": -79.0288,
      "value": 42.0,
      "category": "Buena",
      "color": "#22C55E",
      "confidence": 0.95,
      "has_data": true
    },
    {
      "lat": -8.1490,
      "lon": -79.1290,
      "value": null,
      "category": "No hay datos",
      "color": "#9CA3AF",
      "confidence": 0.0,
      "has_data": false
    }
  ]
}
```

#### `GET /api/v1/map/point?lat=-8.1116&lon=-79.0288&layer=air_quality&standard=inca`
* **Fuera de cobertura (zona gris):**
```json
{
  "lat": -8.1490,
  "lon": -79.1290,
  "has_data": false,
  "label": "No hay datos",
  "series": []
}
```
* **Dentro de cobertura:**
```json
{
  "lat": -8.1116,
  "lon": -79.0288,
  "has_data": true,
  "label": "Buena",
  "current": {
    "recorded_at": "2026-11-13T16:00:00Z",
    "value": 42.0,
    "category": "Buena",
    "color": "#22C55E"
  },
  "series": [
    { "recorded_at": "2026-11-13T15:55:00Z", "value": 41.2 },
    { "recorded_at": "2026-11-13T16:00:00Z", "value": 42.0 }
  ]
}
```

#### `GET /api/v1/map/playback/timestamps`
Devuelve los *timestamps* disponibles en las últimas 24 horas (pasos de 5 min, orden ascendente).
```json
{
  "timestamps": [
    "2026-11-12T16:05:00Z",
    "2026-11-12T16:10:00Z"
  ]
}
```

---

### 5.3. Endpoints de Acceso y Administración

#### `POST /api/v1/access/verify`
Valida la *passphrase* contra el hash + salt configurado. Emite cookie `HttpOnly` de sesión.
```json
{ "phrase": "..." }
```
*Respuesta válida (`200 OK`):* `{ "ok": true }`
*Respuesta inválida (`401 Unauthorized`):* `{ "ok": false }`

#### Gestión de Nodos Virtuales (requieren sesión válida)
* `GET /api/v1/admin/nodes`
* `POST /api/v1/admin/nodes` → `{ "code": "TRJ-02", "name": "Óvalo Papal", "lat": -8.118, "lon": -79.035, "status": "active" }`
* `PATCH /api/v1/admin/nodes/{id}` → actualiza `name`, `lat`, `lon`, `status`
* `DELETE /api/v1/admin/nodes/{id}`

#### `POST /api/v1/admin/nodes/{id}/readings/batch`
Carga de series simuladas. `start_time` debe estar alineado a múltiplos exactos de 5 minutos (`minuto % 5 == 0`) en hora de Perú (`America/Lima`, `-05:00`). En cada arreglo de `metrics`, el índice `0` corresponde a `start_time` (el más reciente) y cada índice `i` retrocede `i * 5 minutos`.
```json
{
  "start_time": "2026-11-13T16:00:00-05:00",
  "metrics": {
    "pm25": [18.4, 17.9, 16.5],
    "pm10": [34.0, 33.1, 30.8],
    "temperature_c": [21.5, 21.6, 21.8]
  }
}
```
