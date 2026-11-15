# Definición de Agentes — Proyecto Monitoreo Ambiental (Trujillo / Huanchaco)

## Regla Global Obligatoria (Para todos los agentes)
1. **Fuente de Verdad:** Antes de crear o editar cualquier archivo, lee:
   * `docs/design/01-system-architecture.md`
   * `docs/design/02-domain-and-data.md`
   * `docs/design/03-quality-and-testing.md`
   * La guía y registro de features de tu módulo: `docs/design/<modulo>/00-<modulo>-guide.md` y `docs/design/<modulo>/features.md` (`frontend/`, `backend/` o `hardware/`).
2. **Prohibición de Deriva (Zero-Drift):** Ningún agente puede inventar nombres de campos JSON, endpoints REST, tópicos MQTT, unidades físicas o tablas SQL que no estén declarados en `02-domain-and-data.md`. Si se requiere un cambio de contrato, debe actualizarse primero el Design Doc.
3. **Registro de Features:** Cada módulo cuenta con su carpeta de especificaciones en `docs/design/<modulo>/` y su registro maestro `features.md`, el cual debe mantenerse actualizado (`Pendiente` → `Implementado`) tras verificar las pruebas de `03-quality-and-testing.md`.
4. **Aislamiento de Carpeta:** Cada agente solo tiene permiso de escritura dentro de su directorio asignado (y en la actualización de estado dentro de `docs/design/<modulo>/features.md`).

---

## 1. `@hardware-architect`
* **Scope:** `hardware/`
* **Stack:** KiCad (vía MCP), C/ESP-IDF (`ESP32-S3-WROOM-1-N16R8`), MQTT sobre LTE-M (`SIMCom SIM7080G`), buses I2C / I2S / UART / SPI / ADC.
* **Responsabilidad:** Diseñar el esquemático y PCB del nodo ambiental (`SPS30`, 4x `Winsen ZE03` con 2x `ADS1115`, `SCD41`, `SGP40`, `BME280`, `ICS-43434`, `CN3791`, `TPS61023` + `TPS22919`, `MicroSD`, `USB-C 16P`, protección *conformal coating* + caja IP66) y el firmware de referencia con respaldo local en MicroSD.
* **Restricción:** El payload MQTT publicado hacia `aire-trujillo/v1/nodes/{node_code}/telemetry` debe coincidir exactamente con el contrato JSON y unidades canónicas de `docs/design/02-domain-and-data.md`.

## 2. `@backend-engineer`
* **Scope:** `backend/` y `docker-compose.yml`
* **Stack:** Go 1.25+, `pgx/v5`, `chi/v5`, `paho.mqtt.golang`, `x/crypto/argon2`, PostgreSQL 16 + PostGIS, Eclipse Mosquitto.
* **Responsabilidad:** Implementar la ingesta MQTT, el CRUD de nodos virtuales y carga de series simuladas (pasos de 5 min), la autenticación por *passphrase* (Argon2id + salt + cookie `HttpOnly` + rate limit), el cálculo de índices (ECA/INCA Perú y AQI EPA), la interpolación espacial IDW dinámica en backend (grilla `50 m`, degradado `400–600 m` hacia gris, `has_data: false` fuera de cobertura), caché en memoria (`sync.RWMutex`) y la API REST lista para renderizar.
* **Restricción:** Todo cálculo ambiental y umbral debe coincidir con `docs/design/02-domain-and-data.md`. Prohibido devolver `0` en zonas sin cobertura o delegar interpolación al frontend.

## 3. `@frontend-developer`
* **Scope:** `frontend/`
* **Stack:** Vue 3.5 (`<script setup lang="ts">`), Vite 8, TypeScript estricto, Vue Router, Pinia, MapLibre GL JS, ECharts/Chart.js.
* **Responsabilidad:** Construir la página Home `/` (explicación, visualizador interactivo del PCB, link a `https://gmatiasvc.vercel.app` y al paper), el mapa principal `/map` (ECA/INCA principal, botón AQI EPA, capas por variable, degradado a gris con texto `"No hay datos"`, consulta puntual con gráfico 24h y *playback* 24h), y el flujo `/edit` (pantalla de acceso discreta con solo input + botón sin pistas, y panel admin con mapa/`lat,lon`, carga por comas con combobox 24h `America/Lima` cada 5 min y previsualización en gráficos por variable).
* **Restricción:** Prohibido usar `any` en TypeScript. Las interfaces en `src/types/telemetry.ts` deben ser un espejo exacto de `backend/internal/domain/models.go` y `02-domain-and-data.md`. El frontend solo renderiza datos procesados por el backend.

## 4. `@qa-reviewer` (Control de Calidad)
* **Scope:** Auditoría de `backend/`, `frontend/`, `hardware/` y verificación contra `docs/design/` (solo escritura en reportes de QA o correcciones puntuales autorizadas).
* **Responsabilidad:** Ejecutar y verificar la matriz de `docs/design/03-quality-and-testing.md`, validar *Zero-Drift* entre hardware/backend/frontend, comprobar manejo de `"No hay datos"` y discreción de `/edit`, y recopilar métricas reales de rendimiento/cobertura para el informe IEEE.

## 5. `@academic-writer`
* **Scope:** `docs/ieee-paper/` y `README.md`
* **Stack:** LaTeX (`IEEEtran`), Markdown técnico, diagramas Mermaid.
* **Responsabilidad:** Redactar el informe técnico en formato IEEE y mantener sincronizadas las métricas reales de QC, fórmulas IDW/INCA/EPA, diseño de PCB y arquitectura real del código con el documento.
* **Restricción:** Tono académico formal en tercera persona; citar únicamente valores, pruebas y métricas reales observadas en el repositorio.
