# 03 — Control de Calidad, Pruebas y Evidencia Técnica (`aire-trujillo`)

> Este documento define los criterios de aceptación, pruebas obligatorias y reglas de auditoría (*Quality Control*) que deben cumplir todos los módulos antes de considerarse terminados, además de la evidencia técnica trazable para el informe IEEE.

---

## 1. Reglas Globales de Calidad (Zero-Drift)
1. **Sincronía de contratos:** Cualquier campo JSON, unidad física, tópico MQTT, tabla SQL o ruta REST implementada debe coincidir carácter por carácter con [02-domain-and-data.md](./02-domain-and-data.md).
2. **Estado de features:** Cada implementación debe actualizar el estado en su registro maestro (`docs/design/<modulo>/features.md`) de `Pendiente` a `Implementado` (solo tras pasar pruebas).
3. **Ausencia de datos explícita:** Jamás representar una zona fuera de cobertura ($d > R_{\max}$) con `0`; siempre debe producirse `has_data: false`, color `#9CA3AF` y etiqueta `"No hay datos"`.
4. **Cómputo exclusivo en backend:** El frontend no puede ejecutar fórmulas IDW ni cálculo de subíndices INCA/EPA; toda la lógica ambiental reside y se prueba en Go.

---

## 2. Matriz de Pruebas por Módulo

### 2.1. Backend (`backend/`)
* **Pruebas unitarias obligatorias (`go test ./...`):**
  * **Índices (`BE-005`):** Casos de borde en tramos INCA Perú y US EPA (valores límite, conversión de unidades `mg/m³ ↔ ppm` y `µg/m³ ↔ ppb`, contaminante crítico, valores por encima del último tramo).
  * **Interpolación IDW (`BE-006`):** Coincidencia exacta sobre un nodo ($d < 1\text{ m}$), simetría entre nodos equidistantes, decaimiento lineal de `confidence` entre `400 m` y `600 m`, y `has_data: false` cuando $d > 600\text{ m}$ o sin nodos activos.
  * **Series simuladas (`BE-003`):** Validación de múltiplos de 5 min, asignación temporal regresiva (`i * 5 min`), rechazo de horas no alineadas y de valores fuera de rango o `NaN`/`Inf`.
  * **Autenticación (`BE-002`):** Verificación Argon2id con salt, comparación en tiempo constante, rechazo sin pistas y *rate limiting*.
* **Compilación y análisis estático:** `go vet ./...` y `go build ./cmd/server` sin errores ni advertencias.
* **Integración (`docker-compose.yml`):** Arranque limpio de `postgres` (PostGIS), `mosquitto` y `api`; verificación de migraciones SQL, ingesta MQTT y endpoints REST.

### 2.2. Frontend (`frontend/`)
* **Tipado estricto (`vue-tsc -b`):** Cero usos de `any`; `src/types/telemetry.ts` debe reflejar exactamente `backend/internal/domain/models.go` y `02-domain-and-data.md`.
* **Build de producción:** `npm run build` (`vue-tsc -b && vite build`) sin errores.
* **Verificación funcional de UI:**
  * Ruta `/edit`: únicamente 1 campo de texto y 1 botón, sin textos descriptivos ni enlaces desde la UI pública.
  * Zona gris en mapa y consulta puntual: muestra `"No hay datos"`, nunca `0`.
  * Admin: combobox en formato 24h cada 5 min (`00:00`–`23:55`), entrada por comas (primero el más reciente) y gráfico individual por cada variable ingresada.

### 2.3. Hardware (`hardware/`)
* **KiCad (`hardware/pcb/`):**
  * ERC (*Electrical Rules Check*) sin errores en esquemático.
  * DRC (*Design Rules Check*) sin errores ni pistas sin conectar en PCB.
  * Verificación de BOM contra [01-pcb-architecture.md](./hardware/01-pcb-architecture.md) (incluyendo 2x `ADS1115`, 4x `Winsen ZE03` con canales de corrección, `CN3791`, `TPS61023` + `TPS22919`).
* **Firmware (`hardware/firmware/`):**
  * Compilación limpia con ESP-IDF (`idf.py build`) para target `esp32s3`.
  * Verificación unitaria/serial del payload JSON generado contra el contrato MQTT de [02-domain-and-data.md](./02-domain-and-data.md).

---

## 3. Evidencia para el Informe IEEE (`docs/ieee-paper/`)
Durante la implementación y el Control de Calidad se deben conservar datos reales y reproducibles (sin inventar métricas):
* Latencias medidas de cálculo IDW sobre la grilla de `50 m` y tamaño de payload JSON.
* Consumo de memoria del servicio Go en ejecución (relevante para despliegue en *free tier*).
* Cobertura real de pruebas (`go test -cover ./...`) y resultados de ERC/DRC.
