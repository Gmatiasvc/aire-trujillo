# FE-002 — Mapa Principal con Capas de Calor

* **Estado:** Pendiente
* **Módulo:** Frontend / Mapa Geoespacial

## 1. Objetivo
Renderizar de forma fluida y ligera los mapas de calor generados por el backend sobre Trujillo / Huanchaco, permitiendo alternar entre estándares de calidad de aire y variables individuales.

## 2. Diseño Técnico
* **Motor de mapa:** MapLibre GL JS (WebGL) consumiendo datos ya interpolados por el backend (sin cálculos espaciales en el cliente).
* **Selector de capas:**
  * **Capa principal por defecto:** Calidad de aire bajo estándar **ECA / INCA (Perú)**.
  * **Botón conmutable:** Cambio inmediato entre **ECA / INCA (Perú)** y **AQI (EPA)**.
  * **Capas individuales:** PM2.5, PM10, CO, NO2, SO2, H2S, CO2, VOC, Ruido, Temperatura, Humedad y Presión.
* **Leyenda dinámica:** Muestra la escala de colores, rangos y categorías correspondientes a la capa y estándar activo.
* **Marcadores de nodos:** Visualización opcional/interactiva de la ubicación y estado de los nodos activos.
