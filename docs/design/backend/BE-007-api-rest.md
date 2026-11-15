# BE-007 — API REST (Heatmap, Consulta Puntual y Playback 24h)

* **Estado:** Implementado
* **Módulo:** Backend / API

## 1. Objetivo
Exponer endpoints REST ligeros y cacheados con `chi/v5` para alimentar el mapa principal, las consultas puntuales y la reproducción de las últimas 24 horas.

## 2. Grupos de Endpoints (Contratos exactos en `02-domain-and-data.md`)
* **Heatmap por capa y timestamp:** Devuelve los datos espaciales precalculados para la capa solicitada (`inca`, `aqi_epa`, o variable individual) en el instante indicado (por defecto el más reciente).
* **Consulta puntual interpolada:** Recibe `(lat, lon)` y capa/estándar; devuelve el estado de cobertura (`has_data`) y, si está dentro de cobertura, la serie temporal interpolada de las últimas 24h para graficar. Si está en zona gris/fuera de cobertura, indica ausencia de datos.
* **Playback 24h:** Lista de *timestamps* disponibles en las últimas 24 horas (pasos de 5 min) y entrega eficiente de *snapshots* por instante.
* **Administración (protegidos por `BE-002`):** CRUD de nodos virtuales y carga de series simuladas.

## 3. Rendimiento
* **Caché en memoria (`sync.RWMutex`):** Cachear grillas interpoladas recientes y snapshots de playback, invalidando cuando ingresan nuevos datos o se editan nodos virtuales.
* **CORS:** Configurado explícitamente para los orígenes permitidos del frontend.
