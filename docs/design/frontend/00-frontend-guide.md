# Guía de Implementación — Frontend (`frontend/`)

> Documento autocontenido para el chat de implementación del Frontend. Antes de editar, leer también [01-system-architecture.md](../01-system-architecture.md), [02-domain-and-data.md](../02-domain-and-data.md) y [03-quality-and-testing.md](../03-quality-and-testing.md).

---

## 1. Estado Actual del Scaffold
* Vue `^3.5.44`, TypeScript `~6.0.3`, Vite `^8.3.6` (`<script setup lang="ts">`, strict TS).
* **Nota sobre diseño visual:** El usuario proveerá un MCP/skill de diseño antes o durante la implementación del frontend para lograr una estética moderna y creativa (no genérica) en el Home y la UI. Respetar los requisitos funcionales y estéticos de cada feature doc.

## 2. Dependencias Recomendadas
* `vue-router` (rutas `/`, `/map`, `/sixseven`)
* `pinia` (estado global: capa activa, estándar `inca` | `epa`, timestamp de playback, sesión admin)
* `maplibre-gl` (renderizado WebGL ligero del mapa y capas de calor con datos ya calculados por backend)
* `echarts` o `chart.js` (gráfico de evolución temporal en consulta puntual y gráficos por variable en el panel de administrador)
* `tailwindcss` (o sistema de estilos alineado con el skill de diseño que indique el usuario)

## 3. Regla Zero-Drift de Tipos (`src/types/telemetry.ts`)
* Crear `src/types/telemetry.ts` como espejo exacto de `backend/internal/domain/models.go` y [02-domain-and-data.md](../02-domain-and-data.md).
* **Prohibido usar `any`** en todo el proyecto TypeScript.
* **Prohibido calcular IDW o INCA/EPA en el cliente:** el frontend únicamente solicita los endpoints REST y renderiza las celdas, colores, categorías y series devueltas por el backend.

## 4. Mapa de Rutas y Vistas

| Ruta | Vista | Descripción |
| :--- | :--- | :--- |
| `/` | Home (`FE-001`) | Explicación del proyecto, visualizador interactivo del PCB (con hotspots/inspección de componentes), botón hacia `/map`, link a `https://gmatiasvc.vercel.app` y link al paper (`[URL por definir]`). |
| `/map` | Mapa Principal (`FE-002`–`FE-005`) | Mapa MapLibre GL JS con capa principal **ECA / INCA**, botón para cambiar a **AQI EPA**, selector de variables individuales, degradado a gris `#9CA3AF` con etiqueta explícita `"No hay datos"`, panel lateral/drawer de consulta puntual con gráfico 24h y barra de *playback* 24h (pasos de 5 min). |
| `/sixseven` | Acceso Discreto + Admin (`FE-006`, `FE-007`) | **No enlazar esta ruta en ningún menú o footer público.** Si no hay sesión válida (`GET /api/v1/access/session`), renderiza exclusivamente un input de contraseña y un botón de envío (sin títulos, logos ni pistas). Tras autenticarse (`POST /api/v1/access/verify`), monta el panel de administrador. |

## 5. Detalles Críticos de UX
* **Zona sin cobertura (`FE-003`):** En cualquier hover/clic sobre celdas con `has_data === false` o respuesta de punto con `has_data === false`, mostrar textualmente `"No hay datos"`. Nunca mostrar `0`.
* **Carga de series en Admin (`FE-007`):**
  * **Combobox hora de inicio:** Opciones fijas cada 5 minutos en formato 24h (`00:00` a `23:55`), interpretadas en zona horaria `America/Lima` (`-05:00`).
  * **Input por comas:** Parsear en tiempo real los valores ingresados; índice `0` = hora seleccionada, índice `i` = `hora - i * 5 min`.
  * **Previsualización:** Renderizar **un gráfico individual por cada variable** que tenga valores ingresados (tanto antes de enviar como al consultar datos guardados del nodo).
* Al finalizar cada feature, actualizar el estado en [features.md](./features.md).
