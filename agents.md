# Definición de Agentes — Proyecto Monitoreo Ambiental (Trujillo / Huanchaco)

## Regla Global Obligatoria (Para todos los agentes)
1. **Fuente de Verdad:** Antes de crear o editar cualquier archivo, lee `docs/design/01-system-architecture.md`, `docs/design/02-domain-and-data.md` y el registro de features de tu módulo en `docs/design/<modulo>/features.md` (`frontend/`, `backend/` o `hardware/`).
2. **Prohibición de Deriva (Zero-Drift):** Ningún agente puede inventar nombres de campos JSON, endpoints REST, unidades físicas o tablas SQL que no estén declarados en `02-domain-and-data.md`. Si se requiere un cambio de contrato, debe actualizarse primero el Design Doc.
3. **Registro de Features:** Cada módulo cuenta con su carpeta de especificaciones en `docs/design/<modulo>/` y su registro maestro `features.md`, el cual debe mantenerse actualizado conforme se implementen las features.
4. **Aislamiento de Carpeta:** Cada agente solo tiene permiso de escritura dentro de su directorio asignado.

---

## 1. `@hardware-architect`
* **Scope:** `hardware/`
* **Stack:** KiCad (vía MCP), C/ESP-IDF, buses I2C / I2S / UART.
* **Responsabilidad:** Diseñar el esquemático y PCB del nodo ambiental (`SPS30`, `SCD41`, `BME280`, `ICS-43434`, `ADS1115`) y el firmware de referencia.
* **Restricción:** El payload serializado en `telemetry_http.c` debe coincidir byte a byte con el contrato JSON definido en `docs/design/02-domain-and-data.md`.

## 2. `@backend-engineer`
* **Scope:** `backend/` y `docker-compose.yml`
* **Stack:** Go 1.23+, `pgx/v5`, `chi/v5`, PostgreSQL 16 + PostGIS.
* **Responsabilidad:** Implementar el motor de simulación física (perfiles Trujillo vs. Huanchaco), la interpolación espacial IDW, la caché en memoria (`sync.RWMutex`), la persistencia en PostGIS y la API REST con CORS.
* **Restricción:** Prohibido usar `math/rand` puro sin semillas o funciones de tendencia temporal/eólica. Todo cálculo ambiental debe respetar los umbrales ECA Perú documentados.

## 3. `@frontend-developer`
* **Scope:** `frontend/`
* **Stack:** Vue 3 (`<script setup lang="ts">`), Vite, TypeScript estricto, Pinia, MapLibre GL JS, Tailwind CSS.
* **Responsabilidad:** Construir el dashboard geoespacial, las capas WebGL de mapas de calor (ICA, PM2.5, NO2, Ruido dBA, Microclima) y el control de *time-travel* entre snapshots.
* **Restricción:** No usar `any` en TypeScript. Las interfaces en `src/types/telemetry.ts` deben ser un espejo exacto de `backend/internal/domain/models.go`.

## 4. `@academic-writer`
* **Scope:** `docs/ieee-paper/` y `README.md`
* **Stack:** LaTeX (`IEEEtran`), Markdown técnico, diagramas Mermaid.
* **Responsabilidad:** Redactar el informe técnico en formato IEEE, documentar la metodología de gobernanza de IA (`03-ai-governance.md`) y mantener sincronizadas las métricas, fórmulas IDW y arquitectura real del código con el documento.
* **Restricción:** Tono académico formal en tercera persona; citar siempre las referencias en `references.bib` y los valores reales implementados en el repositorio.