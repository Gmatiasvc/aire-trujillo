# 01 — Arquitectura del Sistema (`aire-trujillo`)

> **Nota:** El nombre `aire-trujillo` es temporal.

## 1. Visión General
Plataforma de monitoreo ambiental y calidad del aire granularizada a nivel urbano, inspirada en el concepto de *Breathe London* pero orientada a visualización mediante mapas de calor (*heatmaps*).

* **Cobertura geográfica inicial:** Trujillo / Huanchaco, Perú.
* **Frecuencia de actualización:** Recopilación de datos cada 5 a 10 minutos (intervalos de 5 minutos en simulación).
* **Modo de operación actual (Demostración):** Simulación de nodos virtuales, con interfaz de administrador protegida para registrar nodos y cargar series de datos para la demo pública.
* **Despliegue:** Orquestado con `docker-compose.yml` y pensado para desplegarse en un servicio gratuito.

## 2. Componentes Principales

### 2.1. Nodos de Hardware (Campo)
* Sensores instalados en postes con PCB custom (`ESP32-S3-WROOM-1-N16R8`, `SIMCom SIM7080G`, `SPS30`, 4x `Winsen ZE03` + 2x `ADS1115`, `SCD41`, `SGP40`, `BME280`, `ICS-43434`, `CN3791`, `TPS61023` + `TPS22919`, `MicroSD`, `USB-C 16P`), protegidos contra corrosión marina (*conformal coating* + caja IP66). Ver [docs/design/hardware/01-pcb-architecture.md](./hardware/01-pcb-architecture.md).
* Envían mediciones periódicas (cada 5 – 10 min) hacia el servidor central.
* **Conectividad y protocolo:** MQTT sobre LTE-M.
* *(Para la fase de demostración actual, los nodos son simulados).*

### 2.2. Infraestructura y Broker MQTT
* **Base de datos:** PostgreSQL 16 + PostGIS.
* **Broker MQTT:** Eclipse Mosquitto.
* **Orquestación:** `docker-compose.yml`.

### 2.3. Backend (Go)
* **Ingesta y Simulación:** Recibe lecturas vía Eclipse Mosquitto y gestiona los nodos virtuales y sus series simuladas (pasos de 5 minutos).
* **Seguridad de Administración:** Autenticación mediante *passphrase* (almacenada/verificada con hash y salt) para proteger los endpoints de administración en la demo pública.
* **Cómputo geoespacial exclusivo en backend:**
  * Realiza toda la interpolación espacial y el cálculo de índices (**ECA/INCA Perú** y **AQI EPA**).
  * La cobertura del mapa se actualiza dinámicamente según los nodos registrados; al alejarse de los nodos hacia los bordes, la intensidad/color decae progresivamente hasta llegar a **gris**, zona en la que se indica explícitamente ausencia de datos (nunca `0`).
  * Entrega al frontend únicamente datos listos para renderizar.
* **Almacenamiento:** Histórico en PostgreSQL 16 + PostGIS (al menos las últimas 24 horas para el *playback* y consultas temporales).
* **API REST:** Endpoints públicos de consulta geoespacial/temporal y endpoints protegidos para administración de nodos virtuales.

### 2.4. Frontend (Vue 3 + TypeScript + Vite)
* **Página Home (Landing):**
  * Explicación del proyecto.
  * Visualizador interactivo del PCB (exploración de componentes; no requiere obligatoriamente ser un modelo 3D).
  * Link a la página personal del autor: `https://gmatiasvc.vercel.app`.
  * Link al paper técnico (`[URL/ruta por definir]`).
* **Mapa Principal (renderizado ligero):**
  * Visualización en capas de mapas de calor (*heatmaps*) con actualización dinámica según los nodos registrados.
  * Selector de capas: **ECA / INCA (Perú)** como capa principal, botón para cambiar a **AQI EPA**, y capas individuales de contaminantes/variables.
  * Consulta puntual con interpolación: gráfico de evolución temporal en un punto seleccionado.
  * *Playback* de las últimas 24 horas.
  * Al alejarse de los nodos, el color decae progresivamente hacia **gris**; en la zona gris se muestra explícitamente **"No hay datos"** (nunca `0`).
* **Interfaz de Administrador:**
  * **Pantalla de acceso discreta:** Solo un campo de texto y un botón para enviar la *passphrase*, sin ningún texto o pista que indique qué hay detrás.
  * **Registro de nodos virtuales:** Nombre, ubicación (`lat/lon` manual o selección directa en mapa) y estado.
  * **Carga de datos simulados:** Valores separados por comas (el primero es el más reciente y cada siguiente tiene desfase de 5 minutos hacia atrás), con hora de inicio en formato 24h cada 5 minutos exactos (`16:00`, `16:05`, etc.) seleccionada mediante *combobox*.
  * **Previsualización:** Gráficos individuales por cada variable/dato ingresado.
