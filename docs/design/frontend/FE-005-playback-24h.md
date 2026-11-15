# FE-005 — Control de Playback de las Últimas 24 Horas

* **Estado:** Implementado
* **Módulo:** Frontend / Análisis Temporal

## 1. Objetivo
Permitir reproducir visualmente la evolución del mapa de calor a lo largo de las últimas 24 horas en intervalos de 5 minutos.

## 2. Diseño Técnico
* **Controles UI:** Botón Play / Pause, selector de velocidad (ej. `1x`, `2x`, `4x`), barra deslizante (*scrubber*) con marcas horarias en formato 24h y botón para volver a **"En vivo"** (snapshot más reciente).
* **Sincronización:** Al avanzar el *playback*, solicita/renderiza el snapshot correspondiente al *timestamp* seleccionado y actualiza el indicador de fecha/hora visible en pantalla.
* **Optimización:** Precarga progresiva o caché ligera en el cliente de los snapshots ya consultados durante la sesión de reproducción.
