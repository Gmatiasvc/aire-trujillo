# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- **Evaluadores académicos, comunidad de ingeniería y ciudadanía de Trujillo / Huanchaco (Público principal en `/` y `/map`):** Jurados, revisores técnicos (IEEE), investigadores y residentes locales que buscan comprender la arquitectura física del nodo de monitoreo ambiental en `/` y consultar en `/map` el estado actual y la evolución de las últimas 24 horas de la calidad del aire y el microclima urbano en su zona.
- **Administrador de demostración (`/edit`):** Operador autorizado que gestiona los nodos virtuales sobre el mapa de Trujillo / Huanchaco e inyecta series temporales simuladas en intervalos de 5 minutos para alimentar y validar la demostración pública.

## Product Purpose

Plataforma web de monitoreo ambiental y calidad del aire hiperlocal a escala urbana para Trujillo y Huanchaco (Perú), inspirada en *Breathe London* y centrada en la visualización mediante mapas de calor interpolados dinámicamente.

Su objetivo es hacer visible el comportamiento espacial y temporal de 12 variables ambientales urbanas con rigor ingenieril y transparencia cartográfica: desde el diseño físico del nodo autónomo solar con conectividad LTE-M/MQTT hasta la representación geoespacial fiel de la cobertura real de los sensores.

## Positioning

A diferencia de los portales ambientales que extrapolan una estación aislada a toda una ciudad o rellenan con `0` las zonas sin sensores, **Aire Trujillo** ejecuta en backend una interpolación espacial IDW ($p=2$, celdas de `50 m × 50 m`) con decaimiento explícito de confianza entre `400 m` y `600 m` hacia gris (`#9CA3AF`), declarando textualmente `"No hay datos"` más allá del radio máximo de cobertura ($R_{\max} = 600\text{ m}$).

Combina el estándar regulatorio nacional **ECA / INCA (MINAM — Perú)** como vista predeterminada con conmutación inmediata hacia **AQI (US EPA)**, capas individuales para 12 variables físicas, reproducción histórica de 24 horas en pasos de 5 minutos y trazabilidad *Zero-Drift* entre el hardware custom (`ESP32-S3` + `SIM7080G`), el backend geoespacial (`Go` + `PostGIS`) y el cliente web (`Vue 3.5` + `MapLibre GL JS`).

## Operating Context

- **Entorno geográfico y físico:** Zona urbana y costera de Trujillo y Huanchaco, Perú (zona horaria `America/Lima`, `UTC-5`). Los nodos físicos están diseñados para instalación en postes bajo condiciones de brisa y corrosión marina (*conformal coating* + gabinete estanco `IP66` + alimentación solar con respaldo en batería de litio y tarjeta MicroSD).
- **Superficies y flujos del producto:**
  - **Home (`/` — `FE-001`):** Explicación del contexto urbano/costero, arquitectura solar + LTE-M/MQTT e interpolación espacial; visualizador interactivo del PCB custom para inspeccionar cada bloque de hardware y las variables que mide; acceso directo a `/map`; enlace al sitio personal del autor (`https://gmatiasvc.vercel.app`) y enlace al paper técnico IEEE.
  - **Mapa Principal (`/map` — `FE-002` a `FE-005`):** Mapa WebGL (`MapLibre GL JS`) con capa principal **ECA / INCA (Perú)**, botón conmutable a **AQI (EPA)** y selector de las 12 variables individuales; leyenda dinámica con referencia explícita del gris `#9CA3AF` como `"Sin cobertura / No hay datos"`; consulta puntual (`lat, lon`) en panel/drawer con gráfico de evolución interpolada de 24 horas; y barra de *playback* de las últimas 24 horas (intervalos de 5 min, Play/Pause, velocidades `1x`/`2x`/`4x`, *scrubber* 24h y botón `"En vivo"`).
  - **Acceso Discreto y Panel de Administración (`/edit` — `FE-006` y `FE-007`):** Ruta no enlazada en ningún menú ni pie de página público. Sin sesión activa (`GET /api/v1/access/session`), renderiza únicamente un input `type="password"` y un botón de envío sin títulos, logos, placeholders descriptivos ni pistas ante errores. Tras autenticarse (`POST /api/v1/access/verify`), despliega el CRUD de nodos virtuales (ingreso manual de `lat/lon` o clic en mapa) y la carga por lotes de valores separados por comas con *combobox* horario en formato 24h (`00:00` a `23:55`, cada 5 min en `America/Lima`) y gráficos individuales de previsualización por variable.

## Capabilities and Constraints

- **Stack tecnológico:** Frontend en `Vue 3.5` (`<script setup lang="ts">`), `TypeScript ~6.0` estricto (prohibido usar `any`), `Vite 8`, `Vue Router`, `Pinia`, `MapLibre GL JS` y `ECharts` / `Chart.js`. Backend implementado en `Go 1.25+`, `PostgreSQL 16 + PostGIS` y `Eclipse Mosquitto` (`docker-compose.yml`).
- **Contrato Zero-Drift (`docs/design/02-domain-and-data.md`):** `frontend/src/types/telemetry.ts` debe ser un espejo exacto de `backend/internal/domain/models.go`. Ningún componente puede inventar nombres de campos JSON, unidades físicas o rutas REST fuera del documento de dominio.
- **Cómputo geoespacial exclusivo en backend:** El frontend jamás calcula interpolación IDW ni subíndices INCA/EPA; únicamente solicita `/api/v1/nodes`, `/api/v1/map/heatmap`, `/api/v1/map/point` y `/api/v1/map/playback/timestamps` y renderiza los datos recibidos.
- **Regla estricta de zonas sin datos (`FE-003`):** En cualquier celda o punto con `has_data === false`, el valor es `null`, el color es `#9CA3AF` y la interfaz debe mostrar textualmente `"No hay datos"` (nunca `0`, `null` ni `NaN`).
- **Variables canónicas (12):** `pm25` (`µg/m³`), `pm10` (`µg/m³`), `co` (`mg/m³`), `no2` (`µg/m³`), `so2` (`µg/m³`), `h2s` (`µg/m³`), `co2` (`ppm`), `voc_index` (`index` 1–500), `noise_dba` (`dBA`), `temperature_c` (`°C`), `humidity_pct` (`%`) y `pressure_hpa` (`hPa`).
- **BOM del PCB custom para el visor interactivo (`/`):** `ESP32-S3-WROOM-1-N16R8` (MCU), `SIMCom SIM7080G` (módem LTE-M MQTT), `Sensirion SPS30` (PM2.5/PM10 vía `Header JST-ZHR-5`), 4x `Winsen ZE03` (CO, NO2, SO2, H2S vía `Headers JST-XH-4`) leídos con 2x `TI ADS1115` (8 canales), `Sensirion SCD41` (CO2), `Sensirion SGP40` (VOC), `Bosch BME280` (temperatura, humedad, presión), `InvenSense ICS-43434` (micrófono digital I2S para ruido dBA), `CN3791` (cargador solar Li-Ion), `TPS61023` + `TPS22919` (regulador boost + *load switch*), `Socket MicroSD` (respaldo local offline) y `Conector USB-C 16P`.
- **Decisiones abiertas registradas:**
  - El nombre **Aire Trujillo** se mantiene como nombre del producto por ahora, sujeto a cambio futuro si se define una marca definitiva.
  - La URL o ruta final hacia el paper técnico IEEE en `/` permanece como pendiente / próximamente hasta que el documento tenga enlace definitivo.

## Brand Commitments

- **Nombre actual:** **Aire Trujillo** (`aire-trujillo`).
- **Idioma:** Español en toda la interfaz, etiquetas, leyendas y estados (respetando las denominaciones normativas `ECA / INCA (Perú)`, `AQI (EPA)` y el texto exacto `"No hay datos"`).
- **Enlaces externos requeridos en `/`:**
  - Página personal del autor: `https://gmatiasvc.vercel.app`
  - Paper técnico IEEE: enlace/acción en estado pendiente o ruta local por definir (`[URL/ruta por definir]`).

## Evidence on Hand

- **Documentación de arquitectura y contratos:** [01-system-architecture.md](docs/design/01-system-architecture.md), [02-domain-and-data.md](docs/design/02-domain-and-data.md), [03-quality-and-testing.md](docs/design/03-quality-and-testing.md), [00-frontend-guide.md](docs/design/frontend/00-frontend-guide.md) (`FE-001` a `FE-007`) y [01-pcb-architecture.md](docs/design/hardware/01-pcb-architecture.md).
- **Backend implementado (`backend/`):** Servidor Go con ingesta MQTT, cálculo de índices ECA/INCA y EPA AQI, interpolación IDW 50 m con degradado a gris, autenticación Argon2id por *passphrase* y endpoints de nodos virtuales y series por lotes (`BE-001` a `BE-007`).
- **Límites de evidencia (no fabricar):** En la fase actual de demostración los nodos activos operan en modo virtual/simulado. No se deben inventar métricas de estaciones físicas ya desplegadas, convenios municipales/gubernamentales inexistentes ni testimonios de terceros.

## Product Principles

1. **Honestidad cartográfica ante la incertidumbre:** La plataforma distingue con precisión qué zonas tienen respaldo de medición, cuáles están en decaimiento de confianza (`400–600 m`) y cuáles carecen de cobertura (`> 600 m`, `#9CA3AF`, `"No hay datos"`), evitando inducir a error con falsos ceros.
2. **Fidelidad contractual de extremo a extremo (*Zero-Drift*):** Cada lectura, unidad física, color de categoría e índice normativo (ECA/INCA o EPA AQI) proviene de una única fuente de verdad compartida entre firmware, backend y frontend.
3. **Ingeniería abierta y verificable:** La interfaz conecta los datos ambientales del mapa con la física real del instrumento, permitiendo inspeccionar la arquitectura electrónica del nodo solar/LTE-M en el propio portal.
4. **Simulación ágil tras acceso silencioso:** La administración de la demo permite reconstruir y validar visualmente escenarios de 24 horas en segundos mediante series por comas y gráficos por variable, resguardada tras un punto de acceso (`/edit`) libre de pistas públicas.
