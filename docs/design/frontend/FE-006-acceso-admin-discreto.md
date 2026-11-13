# FE-006 — Pantalla de Acceso Discreta a Administración

* **Estado:** Pendiente
* **Módulo:** Frontend / Seguridad Admin

## 1. Objetivo
Proveer una puerta de entrada minimalista para el administrador en la demo pública sin revelar en absoluto qué funcionalidad existe detrás.

## 2. Reglas de Diseño e Interfaz
* **Contenido estricto:** Únicamente un campo de entrada de texto (`type="password"`) y un botón para enviar la *passphrase*.
* **Prohibiciones:**
  * Sin títulos, subtítulos, logos de "Admin", placeholders descriptivos, enlaces desde la navegación pública ni mensajes de ayuda.
  * Ante una frase incorrecta, no dar pistas ni mensajes detallados; mantener la pantalla neutra.
* **Acceso:** Ruta dedicada `/sixseven`, no enlazada en el menú público; tras autenticación exitosa contra `BE-002`, redirige/monta el panel de administración (`FE-007`).
