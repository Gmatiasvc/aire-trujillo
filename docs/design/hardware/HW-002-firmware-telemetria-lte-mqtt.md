# HW-002 — Firmware de Telemetría (ESP32-S3 + SIM7080G + MicroSD)

* **Estado:** Pendiente
* **Módulo:** Hardware / Firmware (ESP-IDF)

## 1. Objetivo
Implementar el firmware de referencia en C/ESP-IDF para `ESP32-S3-WROOM-1-N16R8` encargado de muestrear todos los sensores cada 5 – 10 minutos, compensar las lecturas electroquímicas, respaldar localmente en MicroSD y publicar por MQTT sobre LTE-M.

## 2. Arquitectura del Ciclo de Operación
1. **Despertar y control de potencia:** Activar rieles de sensores mediante *load switch* (`TPS22919`) y estabilizar tiempo de calentamiento requerido (especialmente `SPS30` y `SGP40`).
2. **Adquisición por buses:**
   * **I2C:** `BME280` (T, HR, P), `SCD41` (CO2), `SGP40` (VOC), y 2x `TI ADS1115` (4 canales de señal + 4 canales libres de corrección para los 4 `Winsen ZE03`: CO, NO2, SO2, H2S).
   * **UART / I2C (`JST-ZHR-5`):** `Sensirion SPS30` (PM2.5, PM10).
   * **I2S:** `InvenSense ICS-43434` (cálculo de nivel de ruido).
3. **Serialización y Respaldo (`Socket MicroSD`):**
   * Construir el payload JSON exactamente según el contrato de `02-domain-and-data.md`.
   * Si el enlace LTE-M (`SIMCom SIM7080G`) no está disponible o falla el ACK MQTT, encolar la lectura con timestamp en la MicroSD y reintentar en el siguiente ciclo (*store-and-forward*).
4. **Transmisión y Sueño Profundo:**
   * Publicar vía MQTT sobre LTE-M usando comandos AT / PPP del `SIM7080G`.
   * Apagar periféricos no esenciales y entrar en *deep sleep* hasta el siguiente intervalo (5 – 10 min).
