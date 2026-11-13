# 02 — Dominio, Contratos y Datos (`aire-trujillo`)

> **Fuente de Verdad (Zero-Drift):** Todo nombre de campo JSON, unidad física, tabla SQL, tópico MQTT y endpoint REST implementado en `hardware/`, `backend/` y `frontend/` debe coincidir exactamente con este documento.

---

## 1. Variables Medidas y Unidades Canónicas

Todas las lecturas almacenadas en base de datos y transportadas por MQTT/REST usan las siguientes claves (`snake_case`) y unidades físicas:

| Clave JSON / SQL | Variable | Sensor | Unidad Canónica | Rango Plausible | Escala Visual Individual (Capas) |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `pm25` | Material particulado PM2.5 | `Sensirion SPS30` | `µg/m³` | `0 – 1000` | `0` (verde) → `25` → `50` → `100+` (rojo/morado) |
| `pm10` | Material particulado PM10 | `Sensirion SPS30` | `µg/m³` | `0 – 1000` | `0` (verde) → `50` → `100` → `200+` (rojo/morado) |
| `co` | Monóxido de carbono | `Winsen ZE03-CO` (`ADS1115`) | `mg/m³` | `0 – 100` | `0` → `10` → `30` (ECA 1h = 30 mg/m³) |
| `no2` | Dióxido de nitrógeno | `Winsen ZE03-NO2` (`ADS1115`) | `µg/m³` | `0 – 2000` | `0` → `100` → `200` → `400+` |
| `so2` | Dióxido de azufre | `Winsen ZE03-SO2` (`ADS1115`) | `µg/m³` | `0 – 2000` | `0` → `125` → `250` → `500+` |
| `h2s` | Sulfuro de hidrógeno | `Winsen ZE03-H2S` (`ADS1115`) | `µg/m³` | `0 – 2000` | `0` → `75` → `150` → `300+` |
| `co2` | Dióxido de carbono | `Sensirion SCD41` | `ppm` | `400 – 10000` | `400` (verde) → `1000` (amarillo) → `2000+` (rojo) |
| `voc_index` | Índice VOC | `Sensirion SGP40` | `index` (1–500) | `1 – 500` | `1–100` (bueno) → `200` → `400+` |
| `noise_dba` | Nivel de ruido | `InvenSense ICS-43434` | `dBA` | `20 – 130` | `30` (verde) → `60` (amarillo) → `80+` (rojo) |
| `temperature_c` | Temperatura ambiente | `Bosch BME280` | `°C` | `-10 – 60` | `12` (azul) → `22` (verde) → `32+` (naranja/rojo) |
| `humidity_pct` | Humedad relativa | `Bosch BME280` | `%` | `0 – 100` | `20` → `60` → `95` |
| `pressure_hpa` | Presión atmosférica | `Bosch BME280` | `hPa` | `800 – 1100` | `990` → `1013` → `1030` |

*Nota sobre `Winsen ZE03`:* Los sensores electroquímicos entregan señal analógica/voltaje compensado por temperatura/cero usando los canales libres de los 2x `TI ADS1115`; el firmware convierte la lectura física calibrada a las unidades canónicas de esta tabla antes de publicar por MQTT.

---

## 2. Estándares de Calidad de Aire, Fórmulas y Zona Sin Datos

### 2.1. Zona Sin Cobertura (Obligatorio)
* Cuando un punto o celda supera el radio máximo de interpolación ($R_{\max} = 600\text{ m}$) o no existen nodos activos con datos recientes en la ventana temporal:
  * `has_data`: `false`
  * `value`: `null`
  * `color`: `#9CA3AF` (gris)
  * `confidence`: `0.0`
  * `category` / `label`: `"No hay datos"`
* **Regla estricta:** **Prohibido** devolver `0` como sustituto de ausencia de datos.

### 2.2. Interpolación Espacial IDW y Decaimiento Hacia Gris (`BE-006`)
* **Parámetros configurables por entorno:**
  * `INTERP_CELL_SIZE_M=50` (resolución de grilla: 50 m × 50 m)
  * `INTERP_IDW_POWER=2` ($p = 2$)
  * `INTERP_FADE_START_M=400` ($R_{\text{fade}} = 400\text{ m}$, inicio de degradado hacia gris)
  * `INTERP_MAX_RADIUS_M=600` ($R_{\max} = 600\text{ m}$, límite absoluto de cobertura)
* **Distancia geodésica:** Haversine en metros entre el centro de la celda / punto consultado $(lat, lon)$ y cada nodo activo $k$ con lectura válida.
* **Ventana de snapshot:** Para un instante $t$ (por defecto el último múltiplo o timestamp disponible), se toma la lectura más reciente de cada nodo activo dentro de $[t - 15\text{ min}, t]$.
* **Cálculo IDW:** Solo participan los nodos activos a distancia $d_k \le R_{\max}$:
  * Si existe un nodo con $d_k < 1\text{ m}$, el valor interpolado es exactamente el valor de ese nodo.
  * En otro caso:
    $$\hat{v}(x) = \frac{\sum_{k: d_k \le R_{\max}} w_k \, v_k}{\sum_{k: d_k \le R_{\max}} w_k}, \quad w_k = \frac{1}{d_k^p}$$
* **Confianza y transición cromática al gris:** Sea $d_{\min} = \min_k d_k$:
  * Si $d_{\min} \le R_{\text{fade}}$: `confidence = 1.0`, `has_data = true`.
  * Si $R_{\text{fade}} < d_{\min} \le R_{\max}$:
    $$\text{confidence} = 1.0 - \frac{d_{\min} - R_{\text{fade}}}{R_{\max} - R_{\text{fade}}} \in (0, 1), \quad \text{has\_data} = \text{true}$$
  * El color final de la celda se interpola linealmente en RGB entre el color de la categoría (`color_base`) y `#9CA3AF` usando `confidence`:
    $$\text{RGB}_{\text{final}} = \text{confidence} \cdot \text{RGB}_{\text{base}} + (1 - \text{confidence}) \cdot \text{RGB}_{\text{gris}}$$
  * Si $d_{\min} > R_{\max}$ o no hay nodos: `confidence = 0.0`, `has_data = false`, `color = "#9CA3AF"`, `category = "No hay datos"`.

### 2.3. Fórmula de Subíndices (Piecewise Linear Interpolation)
Tanto para **ECA/INCA (Perú)** como para **AQI (EPA)**, el subíndice $I_p$ de cada contaminante con concentración $C_p$ se calcula por interpolación lineal en el tramo $[C_{\text{lo}}, C_{\text{hi}}] \to [I_{\text{lo}}, I_{\text{hi}}]$:
$$I_p = \frac{I_{\text{hi}} - I_{\text{lo}}}{C_{\text{hi}} - C_{\text{lo}}} \cdot (C_p - C_{\text{lo}}) + I_{\text{lo}}$$
El índice compuesto en un punto/nodo es $\max_p I_p$ sobre los contaminantes disponibles. Si una concentración supera el límite superior del último tramo, se mantiene en el nivel máximo correspondiente (`> 150` en INCA / `500` en EPA).

### 2.4. Estándar Principal: ECA / INCA (Perú — MINAM)
Tramos de referencia alineados con los Estándares de Calidad Ambiental (D.S. N° 003-2017-MINAM) e Índice de Calidad del Aire (INCA):

| Contaminante | Unidad | Buena (`0–50`) | Moderada (`51–100`) | Mala (`101–150`) | Peligrosa / VUEC (`151–300`) |
| :--- | :--- | :---: | :---: | :---: | :---: |
| `pm25` | `µg/m³` | `0.0 – 25.0` | `25.1 – 50.0` | `50.1 – 125.0` | `125.1 – 250.0` |
| `pm10` | `µg/m³` | `0.0 – 50.0` | `50.1 – 100.0` | `100.1 – 250.0` | `250.1 – 420.0` |
| `no2` | `µg/m³` | `0.0 – 100.0` | `100.1 – 200.0` | `200.1 – 400.0` | `400.1 – 600.0` |
| `so2` | `µg/m³` | `0.0 – 125.0` | `125.1 – 250.0` | `250.1 – 500.0` | `500.1 – 800.0` |
| `co` | `mg/m³` | `0.0 – 10.0` | `10.1 – 30.0` | `30.1 – 45.0` | `45.1 – 60.0` |
| `h2s` | `µg/m³` | `0.0 – 75.0` | `75.1 – 150.0` | `150.1 – 300.0` | `300.1 – 500.0` |

| Rango INCA | Categoría (`category`) | Color Hex (`color`) |
| :--- | :--- | :--- |
| `0 – 50` | `Buena` | `#22C55E` |
| `51 – 100` | `Moderada` | `#EAB308` |
| `101 – 150` | `Mala` | `#F97316` |
| `> 150` | `Peligrosa` | `#EF4444` |

### 2.5. Estándar Secundario (Conmutable): AQI (US EPA)
Nota de conversión interna en backend para `co`, `no2` y `so2` (almacenados en unidades másicas canónicas a 25 °C, 1 atm):
* `co`: $1\text{ mg/m}^3 \approx 0.873\text{ ppm}$
* `no2`: $1\text{ µg/m}^3 \approx 0.532\text{ ppb}$
* `so2`: $1\text{ µg/m}^3 \approx 0.382\text{ ppb}$

Tramos US EPA (PM2.5 actualizado 2024 y gases estándar):

| Contaminante (unidad EPA) | Good (`0–50`) | Moderate (`51–100`) | USG (`101–150`) | Unhealthy (`151–200`) | Very Unhealthy (`201–300`) | Hazardous (`301–500`) |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| `pm25` (`µg/m³`) | `0.0–9.0` | `9.1–35.4` | `35.5–55.4` | `55.5–125.4` | `125.5–225.4` | `225.5–325.4` |
| `pm10` (`µg/m³`) | `0–54` | `55–154` | `155–254` | `255–354` | `355–424` | `425–604` |
| `co` (`ppm`) | `0.0–4.4` | `4.5–9.4` | `9.5–12.4` | `12.5–15.4` | `15.5–30.4` | `30.5–50.4` |
| `no2` (`ppb`) | `0–53` | `54–100` | `101–360` | `361–649` | `650–1249` | `1250–2049` |
| `so2` (`ppb`) | `0–35` | `36–75` | `76–185` | `186–304` | `305–604` | `605–1004` |

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
*Nota:* Si `node_code` no existe en la tabla `nodes` o el nodo está `inactive`, el consumidor MQTT registra advertencia y descarta/no expone en mapa público. Todos los campos dentro de `metrics` son opcionales (`*float64` en Go) para tolerar desconexión puntual de un sensor sin perder el resto de la trama.

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

### 5.1. Parámetros Comunes y Respuesta de Error Estándar
* `layer`: `air_quality` (default), `pm25`, `pm10`, `co`, `no2`, `so2`, `h2s`, `co2`, `voc_index`, `noise_dba`, `temperature_c`, `humidity_pct`, `pressure_hpa`.
* `standard`: `inca` (default) | `epa` (aplica cuando `layer=air_quality`).
* **Formato de error JSON (`400`, `401`, `404`, `429`, `500`):**
```json
{
  "error": "invalid_request",
  "message": "Descripción breve del error"
}
```
*(En `POST /api/v1/access/verify` y rutas `/admin/*` sin sesión, devolver `401` genérico sin pistas sobre el panel).*

### 5.2. Endpoints Públicos

#### `GET /api/v1/health`
```json
{ "status": "ok" }
```

#### `GET /api/v1/nodes`
Devuelve los nodos con `status = 'active'` para renderizar marcadores en el mapa.
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
Si se omite `at`, usa el snapshot más reciente. Si no hay nodos activos o lecturas, devuelve `cells: []` y `bounds: null`.
```json
{
  "recorded_at": "2026-11-13T16:00:00Z",
  "layer": "air_quality",
  "standard": "inca",
  "unit": "INCA",
  "cell_size_m": 50,
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
  "layer": "air_quality",
  "standard": "inca",
  "unit": "INCA",
  "has_data": false,
  "label": "No hay datos",
  "current": null,
  "series": []
}
```
* **Dentro de cobertura:**
```json
{
  "lat": -8.1116,
  "lon": -79.0288,
  "layer": "air_quality",
  "standard": "inca",
  "unit": "INCA",
  "has_data": true,
  "label": "Buena",
  "current": {
    "recorded_at": "2026-11-13T16:00:00Z",
    "value": 42.0,
    "category": "Buena",
    "color": "#22C55E",
    "confidence": 0.92
  },
  "series": [
    { "recorded_at": "2026-11-13T15:55:00Z", "value": 41.2, "has_data": true },
    { "recorded_at": "2026-11-13T16:00:00Z", "value": 42.0, "has_data": true }
  ]
}
```

#### `GET /api/v1/map/playback/timestamps`
Devuelve los *timestamps* disponibles en las últimas 24 horas (pasos de 5 min, orden ascendente). Si no hay datos, devuelve `{ "timestamps": [] }`.
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
Valida la *passphrase* contra `ADMIN_PASSPHRASE_HASH` + `ADMIN_PASSPHRASE_SALT` (Argon2id). Emite cookie `HttpOnly` de sesión (`aire_admin_session`, `SameSite=Lax` o `Strict`, `Secure` en producción).
```json
{ "phrase": "..." }
```
*Respuesta válida (`200 OK`):* `{ "ok": true }`
*Respuesta inválida (`401 Unauthorized`):* `{ "ok": false }`

#### `GET /api/v1/access/session`
Verifica si el cliente tiene sesión válida activa (útil al recargar `/sixseven` sin exponer nada en HTML público).
*Respuesta válida (`200 OK`):* `{ "ok": true }`
*Sin sesión (`401 Unauthorized`):* `{ "ok": false }`

#### `POST /api/v1/access/logout`
Invalida/limpia la cookie de sesión.

#### Gestión de Nodos Virtuales (requieren sesión válida)
* `GET /api/v1/admin/nodes` → `{ "nodes": [ ... ] }` (incluye nodos `active` e `inactive`, `is_virtual`, `created_at`, `updated_at`)
* `POST /api/v1/admin/nodes` → `{ "code": "TRJ-02", "name": "Óvalo Papal", "lat": -8.118, "lon": -79.035, "status": "active" }` → `201 Created` con el nodo creado
* `PATCH /api/v1/admin/nodes/{id}` → actualiza parcialmente `name`, `lat`, `lon`, `status` → `200 OK`
* `DELETE /api/v1/admin/nodes/{id}` → `204 No Content`

#### `GET /api/v1/admin/nodes/{id}/readings?hours=24`
Devuelve las últimas lecturas almacenadas del nodo (para previsualizar/inspeccionar en el panel admin).
```json
{
  "node_id": "550e8400-e29b-41d4-a716-446655440000",
  "readings": [
    {
      "recorded_at": "2026-11-13T16:00:00-05:00",
      "pm25": 18.4,
      "pm10": 34.0,
      "temperature_c": 21.5
    }
  ]
}
```

#### `POST /api/v1/admin/nodes/{id}/readings/batch`
Carga de series simuladas.
* `start_time` debe incluir offset `-05:00` (`America/Lima`) y minutos múltiplos exactos de 5 (`minuto % 5 == 0`).
* En cada arreglo de `metrics`, el índice `0` corresponde a `start_time` (el más reciente) y cada índice `i` retrocede `i * 5 minutos`.
* Se aceptan subconjuntos de variables (no es obligatorio enviar las 12 a la vez).
* Máximo recomendado por lote: `288` puntos (24 horas × 12 puntos/hora).
* Persistencia mediante `INSERT ... ON CONFLICT (node_id, recorded_at) DO UPDATE`, actualizando solo los campos enviados (`COALESCE(EXCLUDED.col, telemetry_readings.col)`).

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
*Respuesta (`200 OK`):*
```json
{
  "inserted_or_updated": 3,
  "from": "2026-11-13T15:50:00-05:00",
  "to": "2026-11-13T16:00:00-05:00"
}
```
