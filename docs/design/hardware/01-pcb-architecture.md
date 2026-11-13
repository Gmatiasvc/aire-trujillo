# Hardware — Arquitectura de PCB Custom (`aire-trujillo`)

## 1. Requisitos Generales
* **Despliegue:** Nodos instalados en postes en Trujillo / Huanchaco, Perú.
* **Alimentación:** Sistema solar + batería de litio.
* **Protección ambiental:** Protección contra corrosión marina mediante *conformal coating* y caja **IP66**.

## 2. Lista de Componentes Principales (BOM Base)

| Subsistema | Componente | Cant. | Función / Notas |
| :--- | :--- | :---: | :--- |
| **MCU** | `ESP32-S3-WROOM-1-N16R8` | 1 | Microcontrolador principal |
| **Conectividad Celular** | `SIMCom SIM7080G` | 1 | Módem LTE-M para envío vía MQTT |
| **Material Particulado** | `Sensirion SPS30` | 1 | Conectado vía `Header JST-ZHR-5` (PM2.5, PM10) |
| **Gases Electroquímicos** | `Winsen ZE03` (CO, NO2, SO2, H2S) | 4 | Conectados vía `Headers JST-XH-4` |
| **Conversión Analógica** | `TI ADS1115` | 2 | 2 unidades (8 canales totales) para leer los 4 `Winsen ZE03` + 1 puerto libre de corrección por cada sensor |
| **CO2** | `Sensirion SCD41` | 1 | Medición de CO2 |
| **VOC** | `Sensirion SGP40` | 1 | Medición de compuestos orgánicos volátiles (VOC) |
| **Microclima** | `Bosch BME280` | 1 | Temperatura, humedad y presión |
| **Acústica (Ruido)** | `InvenSense ICS-43434` | 1 | Micrófono digital (I2S) |
| **Gestión Solar / Batería** | `CN3791` | 1 | Cargador solar para batería de litio |
| **Regulación y Control de Potencia** | `TPS61023` + `TPS22919` | 1 | Regulador boost + *load switch* |
| **Conectores de Expansión** | `Header JST-ZHR-5` | 1 | Conexión para `Sensirion SPS30` |
| **Conectores de Expansión** | `Headers JST-XH-4` | 4 | Conexión para sensores `Winsen ZE03` |
| **Interfaz / Depuración** | `Conector USB-C 16P` | 1 | Conexión USB-C |
| **Almacenamiento Local** | `Socket MicroSD` | 1 | Respaldo local (*offline buffering*) ante pérdida de conectividad |
