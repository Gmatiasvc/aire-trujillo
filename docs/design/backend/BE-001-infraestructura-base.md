# BE-001 — Infraestructura Base y Despliegue

* **Estado:** Implementado
* **Módulo:** Backend / Infraestructura

## 1. Objetivo
Proveer un entorno reproducible mediante `docker-compose.yml` para desarrollo local y despliegue de la demostración en un servicio gratuito, integrando base de datos geoespacial, broker MQTT y servidor API en Go.

## 2. Servicios en `docker-compose.yml`
1. **`postgres` (PostgreSQL 16 + PostGIS):**
   * Almacena nodos virtuales/físicos, series temporales de telemetría, credenciales de administrador y metadatos.
   * Extensiones requeridas: `postgis`.
2. **`mosquitto` (Eclipse Mosquitto):**
   * Broker MQTT ligero para recibir telemetría de nodos (simulados o reales vía LTE-M).
   * Configuración mínima con persistencia opcional y credenciales/ACL para tópicos de ingesta.
3. **`api` (Backend Go):**
   * Binario único compilado (multi-stage `Dockerfile` con imagen base `alpine` o `distroless` para mínimo consumo de memoria en *free tiers*).
   * Ejecuta migraciones SQL al arranque o mediante comando dedicado.

## 3. Consideraciones para Servicio Gratuito (Free Tier)
* **Bajo consumo de RAM/CPU:** Limitar el pool de conexiones `pgxpool` y evitar estructuras en memoria innecesariamente grandes.
* **Configuración por variables de entorno:** `DATABASE_URL`, `MQTT_BROKER_URL`, `ADMIN_PASSPHRASE_HASH`, `ADMIN_PASSPHRASE_SALT`, `PORT`, `CORS_ALLOWED_ORIGINS`.
* **Healthchecks:** Endpoint `/health` en Go para monitoreo de plataforma y `pg_isready` en Postgres.
