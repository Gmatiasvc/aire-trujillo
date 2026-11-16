# HW-002 — Firmware de Telemetría (ESP32-S3 + SIM7080G + MicroSD)

* **Estado:** Implementado en código fuente y verificado en host (`ctest` C11 + validación JSON contra backend Go); pendiente compilación cruzada con toolchain `ESP-IDF v5.x` (`idf.py build` no instalado en el entorno actual).
* **Módulo:** Hardware / Firmware (`C11` sobre `ESP-IDF v5.x`)
* **Código Fuente:** [`hardware/firmware/`](../../../hardware/firmware/) · [`README.md`](../../../hardware/firmware/README.md)

## 1. Decisión de Arquitectura y Lenguaje (`C11 + ESP-IDF` vs. Ensamblador)

El firmware está diseñado e implementado en **C11 estándar sobre ESP-IDF v5.x (FreeRTOS)** para el módulo `ESP32-S3-WROOM-1-N16R8` (`16 MB` QIO Flash, `8 MB` Octal PSRAM), y no en ensamblador Xtensa LX7 directo:

* **Por qué C / ESP-IDF y no ASM:** `ESP32-S3` requiere pilas completas para `I2C` maestro, `I2S` DMA (`48 kHz / 32-bit`), `SDSPI + VFS FATFS` para respaldo `store-and-forward` en MicroSD, calibración ADC por eFuse (`adc_cali` *curve fitting*), gestión de modos `light-sleep`/`deep-sleep` con retención de pads (`gpio_hold_en`) y temporización AT por UART. C11 con `aire_core` desacoplado permite compilar exactamente el mismo motor de cálculo, colas, parseo AT y serialización JSON tanto en host Linux (`gcc -std=c11 -Wall -Wextra -Werror -Wpedantic -Wconversion -Wshadow`) como en `xtensa-esp32s3-elf-gcc`.

## 2. Arquitectura Modular Implementada

1. **Núcleo Portable C11 ([`components/aire_core`](../../../hardware/firmware/components/aire_core/)):**
   * [`aire_board.h`](../../../hardware/firmware/components/aire_core/include/aire_board.h): mapa de pines (`GPIO1..GPIO48`), direcciones I2C (`0x48`, `0x49`, `0x59`, `0x62`, `0x69`, `0x76`) y divisores resistivos (`VBAT ×2.0`, `VSOLAR ×11.0`) sincronizados exactamente con [`hardware/doc/specs.md`](../../../hardware/doc/specs.md).
   * [`aire_telemetry.c`](../../../hardware/firmware/components/aire_core/src/aire_telemetry.c): serializador JSON **Zero-Drift** para las 12 métricas canónicas de [`02-domain-and-data.md`](../02-domain-and-data.md) (`pm25`, `pm10`, `co`, `no2`, `so2`, `h2s`, `co2`, `voc_index`, `noise_dba`, `temperature_c`, `humidity_pct`, `pressure_hpa`), omitiendo lecturas ausentes, `NaN` o fuera de rango físico, y construyendo el tópico `aire-trujillo/v1/nodes/{node_code}/telemetry` rechazando comodines (`+`, `#`, `/`).
   * [`aire_calibration.c`](../../../hardware/firmware/components/aire_core/src/aire_calibration.c):
     * CRC-8 Sensirion (`0x31`, init `0xFF`).
     * Compensación entera/flotante de fábrica `BME280` (registros `0x88..0xA1` y `0xE1..0xE7`).
     * Decodificación `SCD41`, ticks T/RH e índice `VOC` EWMA persistido en memoria `RTC_DATA_ATTR` para `SGP40`, y tramas IEEE-754 big-endian de 60 bytes para `SPS30`.
     * Conversión `ADS1115` (`PGA ±4.096 V` -> `0.125 mV/LSB`) y compensación diferencial (`VO` − deriva `CORR` + deriva térmica ref. `25 °C`) para los 4 sensores `Winsen ZE03` (`CO` en `mg/m³`; `NO2`, `SO2`, `H2S` en `µg/m³`).
     * DSP acústico `ICS-43434` @ `48 kHz`: bloqueador DC + cascada de 3 secciones biquad bilineales exactas **IEC 61672 Ponderación A** (`0.000 dB @ 1 kHz`, `-19.17 dB @ 100 Hz`) y cálculo RMS `dBA`.
     * Planificador energético Li-Ion 1S / Solar (`VBAT < 3250 mV` -> deep-sleep protector `4×`; noche con batería media-baja `< 3500 mV` -> intervalo `2×`).
   * [`aire_queue.c`](../../../hardware/firmware/components/aire_core/src/aire_queue.c): cola FIFO persistente sobre `/sdcard/queue.log` + `/sdcard/queue.idx` con registro tabular `|CRC32|topic\tjson\n`, salto automático de líneas truncadas por corte de luz, actualización atómica de cursor (`.tmp` + `rename`) y compactación acotada por `max_bytes`.
   * [`aire_modem_at.c`](../../../hardware/firmware/components/aire_core/src/aire_modem_at.c): máquina de estados AT para `SIMCom SIM7080G` (`PWRKEY` vía `Q3`, `DTR`, `CNMP=38`, `CMNB=1` LTE-M, `CGATT?`, `CNACT=0,1`), sincronización horaria UTC vía GNSS (`+CGNSINF`) con fallback NITZ (`+CCLK` convirtiendo huso `-20` UTC-5 Perú a `Z` incluso en cruce de medianoche/año bisiesto), y publicación `AT+SMCONF`/`AT+SMCONN`/`AT+SMPUB` QoS 1.
2. **HAL ESP-IDF v5.x y Ciclo Principal ([`main/`](../../../hardware/firmware/main/)):**
   * [`bus_hal.c`](../../../hardware/firmware/main/drivers/bus_hal.c) y [`sensors_hal.c`](../../../hardware/firmware/main/drivers/sensors_hal.c): aislamiento energético entre etapas (habilitación `3V3_SENS`, calentamiento `5V_SENS` en `esp_light_sleep_start()`, corte estricto de rieles de 5 V antes de ráfaga SD/LTE-M y retención de apagado `gpio_hold_en` antes de `esp_deep_sleep_start()`).
   * [`app_main.c`](../../../hardware/firmware/main/app_main.c) y [`Kconfig.projbuild`](../../../hardware/firmware/main/Kconfig.projbuild): configuración de `node_code`, intervalo nominal (default `300 s`), warmup `5V`, APN LTE-M, host/puerto/credenciales Mosquitto y offset acústico.

## 3. Verificación Actual y Pendiente

* **Pruebas ejecutadas después de la última edición:**
  * Suite unitaria host C11 (`cmake` + `ctest` en [`tests/test_firmware_core.c`](../../../hardware/firmware/tests/test_firmware_core.c)) compilada con `-std=c11 -Wall -Wextra -Werror -Wpedantic -Wconversion -Wshadow`: **100 % superada (`0` fallos)**.
  * Validación cruzada de payload JSON emitido por el binario C contra decodificador Go estricto (`DisallowUnknownFields` + `RFC3339`) y ejecución de `go test ./internal/domain ./internal/mqtt` en `backend/`: **superadas**.
  * Verificación sintáctica de archivos ESP-IDF (`app_main.c`, `bus_hal.c`, `sensors_hal.c`) mediante *stubs* C11/GNU11: **0 errores y 0 advertencias**.
* **Pendiente en banco / máquina con SDK Espressif:**
  * Ejecutar `idf.py set-target esp32s3 && idf.py build` tras instalar ESP-IDF v5.2+ y flashear prototipo físico rev. 1.0 para calibración de campo `ZE03` / micrófono `ICS-43434`.
