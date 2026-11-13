# FE-004 — Consulta Puntual en Mapa y Gráfico de Evolución

* **Estado:** Pendiente
* **Módulo:** Frontend / Análisis Puntual

## 1. Objetivo
Permitir al usuario hacer clic en cualquier punto del mapa para inspeccionar cómo ha cambiado la calidad del aire o una variable ambiental en ese lugar específico mediante interpolación del backend.

## 2. Flujo de Interacción
1. El usuario selecciona un punto `(lat, lon)` en el mapa.
2. El frontend solicita al backend la serie temporal interpolada para ese punto y la capa/estándar activo.
3. **Si está dentro de cobertura:** Renderiza un panel/drawer con el valor actual, categoría y un **gráfico temporal** (últimas 24 horas en pasos de 5 min) con tooltip por instante.
4. **Si está fuera de cobertura (zona gris):** Muestra el estado explícito **"No hay datos"** en lugar de un gráfico con ceros.
