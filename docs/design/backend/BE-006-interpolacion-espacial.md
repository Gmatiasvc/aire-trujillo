# BE-006 — Interpolación Espacial Dinámica y Manejo de Ausencia de Datos

* **Estado:** Implementado
* **Módulo:** Backend / Cómputo Geoespacial

## 1. Objetivo
Ejecutar toda la interpolación espacial exclusivamente en el backend para generar superficies de calor y consultas puntuales, actualizando dinámicamente la cobertura según los nodos activos y marcando explícitamente las zonas sin datos.

## 2. Diseño Técnico
* **Método de interpolación:** IDW tipo Shepard con soporte compacto ($w_i = ((R_{\max}-d_i)/(R_{\max}\cdot d_i))^p$, exponente por defecto `p = 2`, configurable vía `INTERP_IDW_POWER`) ponderado por distancia geográfica real (Haversine o PostGIS) sobre los nodos activos dentro del *timestamp* consultado, eliminando saltos en el borde $R_{\max}$.
* **Resolución de grilla:** `50 m` por celda (`INTERP_CELL_SIZE_M=50`), acorde a una densidad urbana de hasta 2 nodos por cuadra (~100 m/cuadra en Trujillo); el frontend usa la grilla `50 m` para consultas puntuales O(1) y la renderiza como superficie raster continua.
* **Cobertura dinámica:** El *bounding box* de cálculo se ajusta automáticamente al conjunto de nodos activos registrados más un margen igual a $R_{\max}$.
* **Interioridad entre nodos y decaimiento exterior hacia "No hay datos" (zona gris):**
  * Radio máximo de influencia por defecto: $R_{\max} = 600\text{ m}$ (`INTERP_MAX_RADIUS_M=600`), dimensionado para cubrir espaciamientos de 2 a 4 cuadras (200–400 m) más una zona de transición suave en los bordes.
  * Con un solo cálculo $O(K)$ por muestra mediante el vector resultante direccional de nodos vecinos ($\vec{R}(p)$), las zonas flanqueadas entre nodos activos mantienen `confidence -> 1.0` sin franjas grises intermedias, mientras que hacia los bordes exteriores (`400 m` a `600 m`) `confidence` decae progresivamente hacia `0.0` fundiéndose con **gris (`#9CA3AF`)**.
  * Si la distancia supera $R_{\max}$ (o no hay nodos activos con datos en esa ventana), el backend marca explícitamente el estado como `has_data: false` (nunca devuelve `0` como valor válido).
* **Color continuo e índices oficiales:** Las categorías regulatorias siguen usando sus umbrales discretos exactos, mientras el campo `color` interpola linealmente entre los centros de banda de cada escala antes de mezclarse con el gris de borde.
* **Salida lista para renderizar:** El backend entrega la grilla/celdas o puntos ya calculados (valor, color/índice, `confidence`, `has_data`) para que el frontend solo renderice.
