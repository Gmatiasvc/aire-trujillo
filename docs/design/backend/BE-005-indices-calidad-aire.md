# BE-005 — Motor de Cálculo de Índices (ECA/INCA Perú y AQI EPA)

* **Estado:** Implementado
* **Módulo:** Backend / Dominio Ambiental

## 1. Objetivo
Calcular en el backend los índices de calidad de aire a partir de las concentraciones medidas, soportando el estándar **ECA / INCA (Perú)** como principal y **AQI (EPA)** como estándar secundario conmutable.

## 2. Diseño Técnico
* **Cálculo de subíndices:** Para cada contaminante regulado presente en el estándar seleccionado (ej. PM2.5, PM10, CO, NO2, SO2, H2S según corresponda a ECA/INCA o EPA), interpolar linealmente dentro de los tramos oficiales de concentración → índice.
* **Índice compuesto:** El valor general de calidad de aire del nodo/punto corresponde al máximo de los subíndices calculados (contaminante crítico), devolviendo además la categoría textual y el color asociado.
* **Separación estricta:** Todas las tablas de umbrales y funciones de conversión viven en un paquete dedicado del dominio en Go, referenciadas desde `02-domain-and-data.md`, sin lógica duplicada en el frontend.
