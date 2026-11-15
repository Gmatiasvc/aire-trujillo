# FE-007 — Panel de Administrador (Nodos Virtuales y Series 5 min)

* **Estado:** Implementado
* **Módulo:** Frontend / Administración

## 1. Objetivo
Permitir al administrador registrar nodos virtuales para la demo y cargar series simuladas de forma rápida, sin errores de formato horario y con verificación visual inmediata.

## 2. Funcionalidades
* **Registro y edición de nodos virtuales:**
  * Formulario con `nombre`, `estado` (activo/inactivo) y ubicación `lat/lon`.
  * Selector dual de coordenadas: ingreso numérico manual de `lat/lon` **o** selección directa haciendo clic sobre un mapa interactivo.
* **Carga de series simuladas:**
  * **Hora de inicio:** *Combobox* en formato **24h** (hora local de Perú, `America/Lima` / `UTC-5`) con opciones cada **5 minutos exactos** (`00:00`, `00:05`, ..., `23:55`) para evitar errores de tipeo.
  * **Entrada de valores:** Campo de texto por variable aceptando valores **separados por comas**, donde el **primer valor es el más reciente** (correspondiente a la hora seleccionada) y cada valor siguiente retrocede **5 minutos**.
* **Previsualización visual:**
  * Antes y después de guardar, renderizar **un gráfico individual por cada dato/variable** mostrando la serie temporal reconstruida con sus marcas de hora exactas, facilitando la validación visual de los datos ingresados.
