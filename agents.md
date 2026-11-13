# Definición de Agentes — Proyecto Monitoreo Ambiental (Trujillo / Huanchaco)

## Regla Global Obligatoria (Para todos los agentes)
1. **Fuente de Verdad:** Antes de crear o editar cualquier archivo, lee `docs/design/01-system-architecture.md`, `docs/design/02-domain-and-data.md` y el registro de features de tu módulo en `docs/design/<modulo>/features.md` (`frontend/`, `backend/` o `hardware/`).
2. **Prohibición de Deriva (Zero-Drift):** Ningún agente puede inventar nombres de campos JSON, endpoints REST, unidades físicas o tablas SQL que no estén declarados en `02-domain-and-data.md`. Si se requiere un cambio de contrato, debe actualizarse primero el Design Doc.
3. **Registro de Features:** Cada módulo cuenta con su carpeta de especificaciones en `docs/design/<modulo>/` y su registro maestro `features.md`, el cual debe mantenerse actualizado conforme se implementen las features.
4. **Aislamiento de Carpeta:** Cada agente solo tiene permiso de escritura dentro de su directorio asignado.

---

## 1. `@hardware-architect`
* **Scope:** `hardware/`
* **Stack:** KiCad (vía MCP), C/ESP-IDF (`ESP32-S3-WROOM-1-N16R8`), MQTT sobre LTE-M (`SIMCom SIM7080G`), buses I2C / I2S / UART / ADC.
* **Responsabilidad:** Diseñar el esquemático y PCB del nodo ambiental (`SPS30`, 4x `Winsen ZE03` con 2x `ADS1115`, `SCD41`, `SGP40`, `BME280`, `ICS-43434`, `CN3791`, `TPS61023` + `TPS22919`, `MicroSD`, `USB-C 16P`, protección *conformal coating* + caja IP66) y el firmware de referencia con respaldo local en MicroSD.
* **Restricción:** El payload MQTT publicado hacia el backend debe coincidir exactamente con el contrato JSON definido en `docs/design/02-domain-and-data.md`.

## 2. `@backend-engineer`
* **Scope:** `backend/` y `docker-compose.yml`
* **Stack:** Go 1.23+, `pgx/v5`, `chi/v5`, PostgreSQL 16 + PostGIS, Eclipse Mosquitto (MQTT).
* **Responsabilidad:** Implementar la ingesta MQTT, el CRUD de nodos virtuales y carga de series simuladas (intervalos de 5 min), la autenticación por *passphrase* (hash + salt), el cálculo de índices (ECA/INCA Perú y AQI EPA), la interpolación espacial dinámica en backend (indicando ausencia de datos fuera de cobertura) y la API REST lista para renderizar.
* **Restricción:** Todo cálculo ambiental debe respetar los umbrales ECA/INCA Perú y AQI EPA documentados en `docs/design/`. Ninguna lógica de interpolación debe delegarse al frontend.

## 3. `@frontend-developer`
* **Scope:** `frontend/`
* **Stack:** Vue 3 (`<script setup lang="ts">`), Vite, TypeScript estricto (librerías de mapas, gráficos y 3D por definir en Design Docs).
* **Responsabilidad:** Construir la página Home (explicación, links y visor 3D interactivo del PCB), el mapa principal con capas de calor (ECA/INCA, botón AQI EPA, variables), consulta puntual con gráfico, *playback* 24h (mostrando explícitamente "No hay datos" fuera de cobertura), y la interfaz de administrador (acceso discreto por *passphrase*, registro de nodos en mapa/`lat,lon`, carga de valores por comas con combobox 24h cada 5 min y previsualización en gráficos por variable).
* **Restricción:** No usar `any` en TypeScript. Las interfaces en `src/types/telemetry.ts` deben ser un espejo exacto de `backend/internal/domain/models.go`. El frontend solo renderiza datos procesados por el backend.

## 4. `@academic-writer`
* **Scope:** `docs/ieee-paper/` y `README.md`
* **Stack:** LaTeX (`IEEEtran`), Markdown técnico, diagramas Mermaid.
* **Responsabilidad:** Redactar el informe técnico en formato IEEE, documentar la metodología de gobernanza de IA (`03-ai-governance.md`) y mantener sincronizadas las métricas, fórmulas IDW y arquitectura real del código con el documento.
* **Restricción:** Tono académico formal en tercera persona; citar siempre las referencias en `references.bib` y los valores reales implementados en el repositorio.