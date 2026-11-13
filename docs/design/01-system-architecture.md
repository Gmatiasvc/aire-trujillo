# 01 — Arquitectura del Sistema (`aire-trujillo`)

> **Nota:** El nombre `aire-trujillo` es temporal.
> **Documentos relacionados:** [02-domain-and-data.md](./02-domain-and-data.md) (fuente de verdad de contratos), [03-quality-and-testing.md](./03-quality-and-testing.md) (criterios de QC y pruebas).

## 1. Visión General
Plataforma de monitoreo ambiental y calidad del aire granularizada a nivel urbano, inspirada en el concepto de *Breathe London* pero orientada a visualización mediante mapas de calor (*heatmaps*).

* **Cobertura geográfica inicial:** Trujillo / Huanchaco, Perú (cobertura dinámica según nodos activos).
* **Frecuencia de actualización:** Recopilación de datos cada 5 a 10 minutos (pasos exactos de 5 minutos en simulación, zona horaria `America/Lima` / `UTC-5`).
* **Modo de operación actual (Demostración):** Simulación de nodos virtuales, con interfaz de administrador protegida (`/sixseven`) para registrar nodos y cargar series de datos para la demo pública.
* **Despliegue:** Orquestado con `docker-compose.yml` y optimizado para despliegue en servicio gratuito.

## 2. Componentes Principales

### 2.1. Nodos de Hardware (`hardware/`)
* Sensores instalados en postes con PCB custom (`ESP32-S3-WROOM-1-N16R8`, `SIMCom SIM7080G`, `SPS30`, 4x `Winsen ZE03` + 2x `ADS1115`, `SCD41`, `SGP40`, `BME280`, `ICS-43434`, `CN3791`, `TPS61023` + `TPS22919`, `MicroSD`, `USB-C 16P`), protegidos contra corrosión marina (*conformal coating* + caja IP66).
* Envían mediciones periódicas (cada 5 – 10 min) hacia el servidor central vía **MQTT sobre LTE-M**, con respaldo local (*offline buffering*) en MicroSD.
* **Documentación del módulo:** [00-hardware-guide.md](./hardware/00-hardware-guide.md), [01-pcb-architecture.md](./hardware/01-pcb-architecture.md), [HW-002](./hardware/HW-002-firmware-telemetria-lte-mqtt.md) y [hardware/features.md](./hardware/features.md).
* *(Para la fase de demostración actual, los nodos son virtuales/simulados).*

### 2.2. Infraestructura y Broker MQTT
* **Base de datos:** PostgreSQL 16 + PostGIS.
* **Broker MQTT:** Eclipse Mosquitto.
* **Orquestación:** `docker-compose.yml` en la raíz del repositorio.

### 2.3. Backend (`backend/` — Go 1.25+)
* **Ingesta y Simulación:** Recibe lecturas vía Eclipse Mosquitto y gestiona los nodos virtuales y sus series simuladas (pasos de 5 minutos).
* **Seguridad de Administración:** Autenticación mediante *passphrase* (Argon2id con salt + cookie `HttpOnly` firmada + *rate limiting*) para proteger los endpoints de administración en la demo pública.
* **Cómputo geoespacial exclusivo en backend:**
  * Interpolación espacial IDW ($p=2$, grilla `50 m`, inicio de degradado `400 m`, radio máximo $R_{\max}=600\text{ m}$) y cálculo de índices (**ECA/INCA Perú** y **AQI EPA**).
  * La cobertura se actualiza dinámicamente según los nodos activos; al alejarse de los nodos hacia los bordes (`400–600 m`), el color decae progresivamente hacia **gris (`#9CA3AF`)**, y más allá de `600 m` se indica explícitamente `"No hay datos"` (nunca `0`).
  * Entrega al frontend únicamente datos listos para renderizar.
* **Almacenamiento y Caché:** Histórico en PostgreSQL 16 + PostGIS (al menos 24 horas para *playback* y consultas temporales) y caché en memoria (`sync.RWMutex`).
* **Documentación del módulo:** [00-backend-guide.md](./backend/00-backend-guide.md) y [backend/features.md](./backend/features.md).

### 2.4. Frontend (`frontend/` — Vue 3.5 + TypeScript + Vite 8)
* **Página Home (`/`):**
  * Explicación del proyecto.
  * Visualizador interactivo del PCB (exploración de componentes; no requiere obligatoriamente ser 3D).
  * Link a la página personal del autor: `https://gmatiasvc.vercel.app`.
  * Link al paper técnico (`[URL/ruta por definir]`).
* **Mapa Principal (`/map` — renderizado ligero):**
  * Visualización en capas de mapas de calor (*heatmaps*) con actualización dinámica según los nodos registrados.
  * Selector de capas: **ECA / INCA (Perú)** como capa principal, botón para cambiar a **AQI EPA**, y capas individuales para las 12 variables.
  * Consulta puntual con interpolación: gráfico de evolución temporal en un punto seleccionado.
  * *Playback* de las últimas 24 horas (pasos de 5 min).
  * En la zona gris / fuera de cobertura muestra explícitamente **"No hay datos"** (nunca `0`).
* **Acceso Discreto e Interfaz de Administrador (`/sixseven`):**
  * **Pantalla de acceso discreta:** Solo un campo de texto y un botón para enviar la *passphrase*, sin ningún texto o pista que indique qué hay detrás (y sin enlaces públicos hacia esta ruta).
  * **Registro de nodos virtuales:** Nombre, ubicación (`lat/lon` manual o selección directa en mapa) y estado.
  * **Carga de datos simulados:** Valores separados por comas (el primero es el más reciente y cada siguiente tiene desfase de 5 minutos hacia atrás), con hora de inicio en formato 24h (`America/Lima`) cada 5 minutos exactos (`00:00`, `00:05`, ...) seleccionada mediante *combobox*.
  * **Previsualización:** Gráficos individuales por cada variable/dato ingresado.
* **Documentación del módulo:** [00-frontend-guide.md](./frontend/00-frontend-guide.md) y [frontend/features.md](./frontend/features.md).
