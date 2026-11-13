# BE-003 — Gestión de Nodos Virtuales y Simulación (5 min)

* **Estado:** Pendiente
* **Módulo:** Backend / Administración y Simulación

## 1. Objetivo
Permitir que el administrador registre nodos virtuales para la demostración y cargue series temporales de variables ambientales con resolución de 5 minutos.

## 2. Gestión de Nodos Virtuales
* **Campos base:** `nombre`, ubicación (`lat`, `lon`, almacenada como punto geográfico en PostGIS) y `estado` (activo / inactivo).
* Operaciones requeridas (protegidas por `BE-002`): listar, crear, actualizar, cambiar estado y eliminar nodos virtuales.

## 3. Carga de Series Simuladas
* **Formato de entrada:** Cadena de valores numéricos separados por comas por cada variable, junto con la **hora de inicio** en formato 24h alineada a múltiplos exactos de 5 minutos (ej. `16:00`, `16:05`).
* **Regla temporal:**
  * El **primer valor** de la lista corresponde a la hora de inicio seleccionada (el dato más reciente).
  * Cada valor subsiguiente $i$ (índice base 0) se asigna al timestamp `hora_inicio - (i * 5 minutos)`.
* **Validación:** Verificar formato numérico, rangos físicos plausibles por sensor/variable, y alineación exacta de minutos (`minuto % 5 == 0`).
* **Persistencia / Publicación:** Guardar las lecturas generadas en PostGIS (y/o publicarlas al flujo MQTT interno si se requiere trazabilidad uniforme con nodos reales) e invalidar/actualizar la caché espacial.
