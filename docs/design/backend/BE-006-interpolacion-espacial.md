# BE-006 — Interpolación Espacial Dinámica y Manejo de Ausencia de Datos

* **Estado:** Implementado
* **Módulo:** Backend / Cómputo Geoespacial

## 1. Objetivo
Ejecutar toda la interpolación espacial exclusivamente en el backend para generar superficies de calor y consultas puntuales, actualizando dinámicamente la cobertura según los nodos activos y marcando explícitamente las zonas sin datos.

## 2. Diseño Técnico
* **Método de interpolación:** IDW (*Inverse Distance Weighting*, exponente por defecto `p = 2`, configurable vía `INTERP_IDW_POWER`) ponderado por distancia geográfica real (Haversine o PostGIS) sobre los nodos activos dentro del *timestamp* consultado.
* **Resolución de grilla:** `50 m` por celda (`INTERP_CELL_SIZE_M=50`), acorde a una densidad urbana de hasta 2 nodos por cuadra (~100 m/cuadra en Trujillo).
* **Cobertura dinámica:** El *bounding box* de cálculo se ajusta automáticamente al conjunto de nodos activos registrados más un margen igual a $R_{\max}$.
* **Decaimiento hacia "No hay datos" (zona gris):**
  * Radio máximo de influencia por defecto: $R_{\max} = 600\text{ m}$ (`INTERP_MAX_RADIUS_M=600`), dimensionado para cubrir espaciamientos de 2 a 4 cuadras (200–400 m) más una zona de transición suave en los bordes.
  * Conforme la distancia al nodo más cercano se acerca a $R_{\max}$ (ej. entre `400 m` y `600 m`), `confidence` decae progresivamente hacia `0.0` (permitiendo al frontend degradar el color hacia **gris**).
  * Si la distancia supera $R_{\max}$ (o no hay nodos activos con datos en esa ventana), el backend marca explícitamente el estado como `has_data: false` (nunca devuelve `0` como valor válido).
* **Salida lista para renderizar:** El backend entrega la grilla/celdas o puntos ya calculados (valor, color/índice, `confidence`, `has_data`) para que el frontend solo renderice.
