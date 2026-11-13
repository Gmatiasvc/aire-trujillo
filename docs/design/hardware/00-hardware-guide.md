# Guía de Implementación — Hardware (`hardware/`)

> Documento autocontenido para el chat de implementación de Hardware (KiCad + Firmware ESP-IDF). Antes de editar, leer también [01-pcb-architecture.md](./01-pcb-architecture.md), [HW-002-firmware-telemetria-lte-mqtt.md](./HW-002-firmware-telemetria-lte-mqtt.md), [02-domain-and-data.md](../02-domain-and-data.md) y [03-quality-and-testing.md](../03-quality-and-testing.md).

---

## 1. Estructura de Carpetas
* `hardware/pcb/` — Proyecto KiCad (esquemático `.kicad_sch`, layout `.kicad_pcb`, librerías locales si aplican, archivos de fabricación Gerber/BOM).
* `hardware/firmware/` — Proyecto ESP-IDF en C para `ESP32-S3-WROOM-1-N16R8`.
* `hardware/doc/specs.md` — Documentación de pinout, asignación de buses, presupuesto de potencia solar/batería y notas de ensamblaje.

## 2. Asignación de Buses y Subsistemas del PCB (`HW-001`)
* **MCU:** `ESP32-S3-WROOM-1-N16R8`.
* **Bus I2C principal (3.3 V):**
  * `Bosch BME280` (temperatura, humedad, presión)
  * `Sensirion SCD41` (CO2)
  * `Sensirion SGP40` (VOC index)
  * `TI ADS1115` #1 y #2 (direcciones I2C distintas vía pin `ADDR`, ej. `0x48` y `0x49`)
  * `Sensirion SPS30` vía `Header JST-ZHR-5` (I2C o UART; verificar pin `SEL` si se usa I2C y pull-ups adecuados / level shifting si requiere 5 V de alimentación)
* **Entradas Analógicas (2x `TI ADS1115` + 4x `Headers JST-XH-4`):**
  * 4 sensores `Winsen ZE03` (`CO`, `NO2`, `SO2`, `H2S`): cada sensor ocupa 1 canal de señal principal + 1 canal libre de corrección/referencia (total 8 canales = 2x `ADS1115` de 4 canales).
* **Bus I2S:** `InvenSense ICS-43434` (micrófono MEMS digital para `noise_dba`).
* **Bus UART (LTE-M):** `SIMCom SIM7080G` (con antena LTE-M, ranura SIM/eSIM y alimentación dedicada capaz de picos de corriente TX).
* **Bus SPI:** `Socket MicroSD` para respaldo local (*offline buffering*).
* **USB:** `Conector USB-C 16P` conectado al USB nativo del `ESP32-S3` (y protección ESD).
* **Energía Solar + Litio:**
  * Cargador solar MPPT `CN3791` para celda de litio.
  * Regulador boost `TPS61023` (para riel de sensores/sistema de 3.3 V o 5 V según diseño) + *load switch* `TPS22919` para cortar alimentación de sensores en *deep sleep*.
* **Protección Marina:** Acabado resistente a corrosión (recomendado ENIG), especificaciones para *conformal coating* (excluyendo áreas activas de sensores de gas/presión/micrófono) y montaje en caja **IP66** con membranas hidrofóbicas/respiraderos adecuados para muestreo de aire.

## 3. Reglas del Firmware (`HW-002`)
* El payload JSON publicado por MQTT al tópico `aire-trujillo/v1/nodes/{node_code}/telemetry` debe coincidir exactamente con [02-domain-and-data.md](../02-domain-and-data.md) (claves `snake_case`, unidades canónicas, `recorded_at` ISO-8601 UTC).
* La compensación/calibración de los `Winsen ZE03` con los canales adicionales de `ADS1115` debe convertir la lectura a las unidades canónicas (`mg/m³` para `co`; `µg/m³` para `no2`, `so2`, `h2s`) antes de serializar.
* Implementar cola FIFO en MicroSD cuando falle LTE-M / MQTT ACK para reenviar sin pérdida de datos al reconectar.
* Actualizar `hardware/doc/specs.md` con el pinout definitivo y marcar el avance en [features.md](./features.md).
