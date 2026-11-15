# BE-004 — Ingesta MQTT (Eclipse Mosquitto) y Persistencia PostGIS

* **Estado:** Implementado
* **Módulo:** Backend / Ingesta

## 1. Objetivo
Recibir las lecturas periódicas (cada 5 – 10 minutos) publicadas por los nodos hacia Eclipse Mosquitto y almacenarlas de forma confiable en PostgreSQL + PostGIS.

## 2. Diseño Técnico
* **Cliente MQTT en Go:** Conexión persistente al broker Mosquitto con reconexión automática (*exponential backoff*) y suscripción a los tópicos de telemetría de nodos.
* **Validación de Payload:** Deserialización estricta contra el contrato definido en `02-domain-and-data.md` (descartar o registrar errores de formato sin detener el consumidor).
* **Idempotencia:** Evitar duplicados si un nodo reintenta envíos tras recuperar conectividad (ej. respaldo desde MicroSD en campo) usando clave única `(node_id, recorded_at)`.
* **Almacenamiento PostGIS:** Tablas indexadas por tiempo (`recorded_at`) y geometría (`geom` con SRID 4326) para consultas rápidas de ventanas de 24h y cálculos espaciales.
