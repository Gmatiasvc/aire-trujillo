# FE-003 — Degradado hacia Gris y Manejo de "No hay datos"

* **Estado:** Implementado
* **Módulo:** Frontend / Visualización Geoespacial

## 1. Objetivo
Evitar interpretaciones erróneas en las zonas sin cobertura real de sensores, dejando visual y textualmente explícito cuando no existe información válida.

## 2. Comportamiento Visual y de Interacción
* **Transición en bordes:** A medida que un área se aleja de los nodos activos, la escala cromática del mapa de calor decae suavemente (ej. rojo → amarillo → verde) hasta fundirse en **gris** en el límite de cobertura (según los datos/confianza entregados por `BE-006`).
* **Indicador explícito:**
  * Al hacer hover o clic sobre una zona gris o fuera de cobertura, el tooltip/panel debe mostrar exactamente **"No hay datos"** (nunca `0`, `null` ni `NaN`).
  * En la leyenda del mapa incluir la referencia visual del color gris como **"Sin cobertura / No hay datos"**.
