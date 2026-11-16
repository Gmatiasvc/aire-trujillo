# Registro Maestro de Features — Hardware

| ID | Feature | Estado | Design Doc |
| :--- | :--- | :--- | :--- |
| HW-001 | Arquitectura base de PCB custom (ESP32-S3, SIM7080G, sensores, solar/batería) | Implementado | [01-pcb-architecture.md](./01-pcb-architecture.md) · [specs.md](../../../hardware/doc/specs.md) |
| HW-002 | Firmware de telemetría MQTT sobre LTE-M (intervalo 5–10 min, respaldo MicroSD) | Implementado (verificado en host; pendiente idf.py build) | [HW-002-firmware-telemetria-lte-mqtt.md](./HW-002-firmware-telemetria-lte-mqtt.md) · [firmware/](../../../hardware/firmware/) |
