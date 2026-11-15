# BE-002 — Autenticación de Administrador por Passphrase

* **Estado:** Implementado
* **Módulo:** Backend / Seguridad

## 1. Objetivo
Proteger los endpoints de administración en la demostración pública para evitar vandalismo, sin exponer información innecesaria sobre la existencia del panel.

## 2. Flujo de Autenticación
1. El cliente envía la *passphrase* mediante `POST` al endpoint de autenticación (`[ruta exacta por definir en 02-domain-and-data.md]`).
2. El backend aplica la función de derivación/hash (ej. **Argon2id** o **scrypt** / **bcrypt**) usando el **salt** configurado y compara en tiempo constante (`subtle.ConstantTimeCompare`) contra el hash almacenado.
3. Si es válida, emite un token de sesión firmado (JWT HS256 o token opaco en cookie `HttpOnly`, `Secure`, `SameSite=Strict`) con expiración acotada.
4. Si es inválida, responde con código genérico (`401 Unauthorized` o `404` según política definida) sin mensajes descriptivos ni pistas.

## 3. Medidas Anti-Abuso
* **Rate limiting** por IP en el endpoint de verificación para mitigar ataques de fuerza bruta.
* **Middleware de autorización** obligatorio en todas las rutas de creación/edición/eliminación de nodos virtuales y carga de datos simulados.
