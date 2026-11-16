# Firmware Aire-Trujillo (`HW-002` — ESP32-S3 + SIM7080G + MicroSD)

Firmware en **C11 sobre ESP-IDF v5.x** para el nodo ambiental urbano/costero de **aire-trujillo** (`ESP32-S3-WROOM-1-N16R8`), sincronizado pin a pin con [specs.md](../doc/specs.md) y el contrato JSON Zero-Drift de [02-domain-and-data.md](../../docs/design/02-domain-and-data.md).

## Decisión de Lenguaje: C11 + ESP-IDF v5.x (por qué no ASM puro)

- El proyecto exige FreeRTOS, pila USB/UART/SPI/I2C/I2S, sistema de archivos VFS FATFS para cola `store-and-forward` en MicroSD y calibración por eFuse (`adc_cali`), además de portabilidad entre pruebas unitarias en host (`gcc -std=c11`) y compilación cruzada Xtensa LX7 (`xtensa-esp32s3-elf-gcc`).
- Toda la lógica crítica determinista (CRC-8/CRC-32, compensación BME280/ZE03, filtros bilineales IIR de ponderación A IEC 61672 a 48 kHz, serialización JSON estricta y máquina de estados AT LTE-M/MQTT) vive en el componente portable [`components/aire_core`](components/aire_core/), libre de dependencias de SDK.

## Estructura del Proyecto

```text
hardware/firmware/
├── CMakeLists.txt
├── partitions.csv                  # Tabla OTA 16 MB Flash + NVS + coredump
├── sdkconfig.defaults              # Target esp32s3, QIO 16 MB, Octal PSRAM 8 MB, FATFS LFN
├── components/
│   └── aire_core/                  # Núcleo C11 portable (100% testeable en host)
│       ├── include/
│       │   ├── aire_board.h        # Mapa GPIO0..48, direcciones I2C y divisores ADC
│       │   ├── aire_telemetry.h    # Contrato Zero-Drift (12 métricas) + tópico MQTT
│       │   ├── aire_calibration.h  # CRC8, BME280, SCD41, SGP40, SPS30, ZE03, DSP dBA, energía
│       │   ├── aire_queue.h        # Cola FIFO persistente en SD con CRC32 e índice atómico
│       │   └── aire_modem_at.h     # Máquina de estados AT SIMCom SIM7080G + parseo GNSS/CCLK
│       └── src/
│           └── *.c
├── main/
│   ├── Kconfig.projbuild           # Parámetros de despliegue (node_code, APN, broker MQTT)
│   ├── CMakeLists.txt
│   ├── app_main.c                  # Máquina de ciclo deep-sleep -> sensores -> SD -> LTE-M
│   └── drivers/
│       ├── bus_hal.h / bus_hal.c   # HAL ESP-IDF v5 (power gating TPS22919/TPS61023, I2C, I2S, SD, UART)
│       └── sensors_hal.h / .c      # Orquestación de BME280, SCD41, SGP40, SPS30, ADS1115 y ICS-43434
└── tests/                          # Suite unitaria host C11 (CMake + CTest)
    ├── CMakeLists.txt
    └── test_firmware_core.c
```

## Secuencia Energética de Ciclo (`app_main.c`)

1. **Chequeo de Batería (`ADC1` + `CN3791`):** Lee `VBAT` (`GPIO1`, divisor `×2.0`) y `VSOLAR` (`GPIO2`, divisor `×11.0`). Si `VBAT < 3250 mV`, omite riel `5V_SENS` y LTE-M y duerme `4×` el intervalo base para evitar *brownout* o sobredescarga Li-Ion.
2. **Adquisición conmutada (`3V3_SENS` -> `5V_SENS`):**
   - Habilita `GPIO48` (`EN_SENS_3V3`), inicia I2C principal (`GPIO8/9`, 100 kHz).
   - Habilita `GPIO21` (`PWR_5V_EN`) + `GPIO47` (`EN_SENS_5V`) e inicia `SPS30` (`0x69`) y 4× `Winsen ZE03` mientras lee `BME280` (`0x76`), `SCD41` (`0x62` compensado por presión), `SGP40` (`0x59` compensado por T/RH) e `ICS-43434` (`I2S`, 1.5 s @ 48 kHz -> filtro A-weighting IEC 61672).
   - Completa la ventana de estabilización (`CONFIG_AIRE_WARMUP_5V_MS`) en `esp_light_sleep_start()` reteniendo rieles de sensores.
   - Lee `SPS30` (`PM2.5`/`PM10`) y 2× `ADS1115` (`0x48`/`0x49`, pares `VO`/`CORR` de `CO`, `NO2`, `SO2`, `H2S`), detiene el ventilador `SPS30` y apaga `5V_SENS` antes del ráfaga LTE/SD.
3. **Store-and-Forward + LTE-M / MQTT QoS 1:**
   - Monta MicroSD (`/sdcard`), enciende `SIM7080G` (`GPIO16` vía `Q3`, despierta `DTR GPIO15`), registra `LTE-M` (`AT+CNMP=38`, `AT+CMNB=1`), sincroniza reloj UTC vía GNSS (`+CGNSINF`) o NITZ (`+CCLK` convirtiendo huso `-20` UTC-5 Perú a `Z`).
   - Serializa payload JSON omitiendo métricas ausentes/inválidas (sin ceros falsos) en `/sdcard/queue.log` con integridad `CRC32` y drena hasta 32 tramas FIFO por ciclo vía `AT+SMPUB` QoS 1.
   - Retiene apagados los load switches (`gpio_hold_en`) y entra a `deep-sleep`.

## Pruebas Unitarias en Host (C11 estricto)

```bash
cmake -S hardware/firmware/tests -B /tmp/aire-fw-tests
cmake --build /tmp/aire-fw-tests
ctest --test-dir /tmp/aire-fw-tests --output-on-failure
```

## Compilación de Producción (ESP-IDF v5.x)

Requiere entorno **ESP-IDF v5.2+** activo (`export.sh`):

```bash
cd hardware/firmware
idf.py set-target esp32s3
idf.py menuconfig   # Configurar AIRE_NODE_CODE, APN LTE-M y AIRE_MQTT_BROKER_HOST
idf.py build
idf.py -p /dev/ttyACM0 flash monitor
```
