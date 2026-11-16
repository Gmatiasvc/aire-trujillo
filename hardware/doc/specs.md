# Especificaciones de Hardware — Nodo Ambiental `aire-trujillo` (`HW-001` v1.0)

> **Trazabilidad (Zero-Drift):** Documento técnico definitivo generado a partir de los esquemáticos jerárquicos KiCad 9 (`aire-trujillo.kicad_sch` y sub-hojas `power_solar`, `cellular_modem`, `sensors_adc`), la PCB de 4 capas (`aire-trujillo.kicad_pcb`) y el paquete de manufactura en `hardware/pcb/manufacturing/`, alineado estrictamente con [`00-hardware-guide.md`](../../docs/design/hardware/00-hardware-guide.md), [`01-pcb-architecture.md`](../../docs/design/hardware/01-pcb-architecture.md), [`HW-002-firmware-telemetria-lte-mqtt.md`](../../docs/design/hardware/HW-002-firmware-telemetria-lte-mqtt.md) y [`02-domain-and-data.md`](../../docs/design/02-domain-and-data.md).

---

## 1. Arquitectura General y Entorno de Operación

* **Controlador principal:** `Espressif ESP32-S3-WROOM-1-N16R8` (`U1`, 16 MB Flash, 8 MB PSRAM, antena PCB integrada).
* **Conectividad celular / GNSS:** `SIMCom SIM7080G` (`U8`, LTE-M Cat-M1 / NB-IoT + GNSS pasivo/activo), conectado mediante traductor bidireccional de nivel lógico 3.3 V ↔ 1.8 V `TI SN74AVC4T245PW` (`U9`) y etapa discreta `BSS138` (`Q3`, `Q4`) para `PWRKEY` y `NETLIGHT`.
* **Autonomía energética:** Cargador solar MPPT `CN3791` (`U3`) para batería 1S Li-Ion/LiFePO4 (`3.7 V` nominal), conmutación ideal sin pérdida (`AO3401A` `Q1`/`Q2` + diodos Schottky `SS34`), regulador Buck `TLV62569DBV` (`U4`, `3V3_SYS`), Boost `TPS61023DRLR` (`U5`, `5V_BOOST`) e interruptores de carga de estado sólido `TPS22919DCKR` (`U6`, `U7`) para apagado completo de periféricos en Deep-Sleep.
* **Ambiente de despliegue:** Trujillo y balneario de Huanchaco (alta humedad marina, rocío nocturno, salinidad costera y radiación UV):
  * PCB de **4 capas** con acabado **ENIG (Electroless Nickel Immersion Gold)** resistente a oxidación de contactos y pads expuestos.
  * Recinto **IP66** estanco con prensaestopas M16/M20, membrana **ePTFE** hidrofóbica para ventana barométrica/química/acústica y **tropicalizado selectivo (Conformal Coating)** acrílico/siliconado excluyendo elementos sensores y conectores mediante cajas `User.Comments`.

---

## 2. Asignación Definitiva de Pines ESP32-S3 (`U1`)

| Módulo / Pin `U1` | GPIO ESP32-S3 | Net PCB | Subsistema | Función /Notas Eléctricas |
| :---: | :---: | :--- | :--- | :--- |
| `4` | `GPIO4` | `I2S_SCK` | Audio (`MK1`) | Reloj de bit I2S (`ICS-43434`) |
| `5` | `GPIO5` | `I2S_WS` | Audio (`MK1`) | Word Select I2S (`LRCLK`, canal izquierdo) |
| `6` | `GPIO6` | `I2S_SD` | Audio (`MK1`) | Datos serie I2S con pull-down `R43` (`100 kΩ`) a `GND` |
| `7` | `GPIO7` | `MODEM_STATUS` | Módem (`U8`/`U9`) | Entrada de estado encendido del módem (traducida 1.8 V → 3.3 V vía `U9`) |
| `8` | `GPIO15` | `MODEM_DTR` | Módem (`U8`/`U9`) | Control wake/sleep `DTR` hacia `U8` (traducida vía `U9`) |
| `9` | `GPIO16` | `MODEM_PWRKEY` | Módem (`U8`/`Q3`) | Pulso de encendido/apagado (`BSS138` `Q3` + `R29`/`R30` hacia `SIM_PWRKEY`) |
| `10` | `GPIO17` | `ESP_MODEM_TXD` | Módem (`U8`/`U9`) | `UART1 TX` ESP32-S3 → `RXD` SIM7080G (vía `U9`) |
| `11` | `GPIO18` | `ESP_MODEM_RXD` | Módem (`U8`/`U9`) | `UART1 RX` ESP32-S3 ← `TXD` SIM7080G (vía `U9`) |
| `12` | `GPIO8` | `I2C_SDA` | Bus I2C Sensores | Datos I2C con pull-up `R40` (`4.7 kΩ`) a `3V3_SENS` |
| `17` | `GPIO9` | `I2C_SCL` | Bus I2C Sensores | Reloj I2C con pull-up `R41` (`4.7 kΩ`) a `3V3_SENS` |
| `13` | `GPIO19` | `USB_DM` | USB-C (`J1`/`U2`) | Par diferencial USB 90 Ω (`USB D-`), protegido por `USBLC6-2SC6` (`U2`) |
| `14` | `GPIO20` | `USB_DP` | USB-C (`J1`/`U2`) | Par diferencial USB 90 Ω (`USB D+`), protegido por `U2` |
| `18` | `GPIO10` | `SD_CS` | MicroSD (`J2`) | SPI Chip Select (`R6` `10 kΩ` pull-up a `3V3_SENS`) |
| `19` | `GPIO11` | `SD_MOSI` | MicroSD (`J2`) | SPI MOSI (`R7` `10 kΩ` pull-up a `3V3_SENS`) |
| `20` | `GPIO12` | `SD_CLK` | MicroSD (`J2`) | SPI Clock |
| `21` | `GPIO13` | `SD_MISO` | MicroSD (`J2`) | SPI MISO (`R8` `10 kΩ` pull-up a `3V3_SENS`) |
| `22` | `GPIO14` | `SD_CD` | MicroSD (`J2`) | Card Detect (`R9` `10 kΩ` pull-up a `3V3_SENS`) |
| `23` | `GPIO21` | `PWR_5V_EN` | Potencia (`U5`) | Habilita elevador Boost `5V_BOOST` (`TPS61023`, activo en alto) |
| `24` | `GPIO47` | `EN_SENS_5V` | Potencia (`U6`) | Habilita `5V_SENS` (`TPS22919` para `SPS30` y 4× `ZE03`) |
| `25` | `GPIO48` | `EN_SENS_3V3` | Potencia (`U7`) | Habilita `3V3_SENS` (`TPS22919` para `ADS1115`, I2C, `MK1` y MicroSD) |
| `27` | `GPIO0` | `ESP_BOOT` | Sistema (`SW2`) | Pulsador `BOOT` (`R2` `10 kΩ`, `C4` `100 nF`) |
| `3` | `EN` | `ESP_EN` | Sistema (`SW1`) | Reset / Enable (`R1` `10 kΩ`, `C3` `1 µF`) |
| `31` | `GPIO38` | `ADS_ALERT` | ADC (`U14`) | Interrupción `ALERT/RDY` de `ADS1115 #1` (`R44` `10 kΩ` pull-up) |
| `32` | `GPIO39` | `CHG_STAT` | MPPT (`U3`) | Estado de carga `CN3791` open-drain (`R14` `10 kΩ` pull-up a `3V3_SYS`) |
| `33` | `GPIO40` | `CHG_DONE` | MPPT (`U3`) | Fin de carga `CN3791` open-drain (`R15` `10 kΩ` pull-up a `3V3_SYS`) |
| `34` | `GPIO41` | `LED_SYS` | Indicador (`D1`) | LED verde de telemetría/estado vía `R3` (`1 kΩ`) |
| `36` | `RXD0` | `UART0_RXD` | Debug (`J3`) | Puerto serie de diagnóstico (cabezal 2.54 mm `J3` pin 2) |
| `37` | `TXD0` | `UART0_TXD` | Debug (`J3`) | Puerto serie de diagnóstico (cabezal 2.54 mm `J3` pin 1) |
| `38` | `GPIO2` | `VSOLAR_ADC` | Telemetría Solar | Divisor `R12` (`470 kΩ`) / `R13` (`47 kΩ`), razón **1/11** (`V_solar = V_adc × 11.0`) |
| `39` | `GPIO1` | `VBAT_ADC` | Telemetría Batería | Divisor `R19` (`100 kΩ`) / `R20` (`100 kΩ`), razón **1/2** (`V_bat = V_adc × 2.0`) |

---

## 3. Mapa de Bus I2C y Canales Analógicos (`ADS1115` / `Winsen ZE03`)

### 3.1. Dispositivos en el Bus I2C (`GPIO8` SDA / `GPIO9` SCL, alimentados por `3V3_SENS`)

| Dirección 7-bit | Dispositivo | Designador | Selección de Dirección Hardware | Variables Canónicas (`02-domain-and-data.md`) |
| :---: | :--- | :---: | :--- | :--- |
| **`0x76`** | Bosch `BME280` | `U11` | Pin 5 (`BME_SDO`) unido a `GND` vía `R53` (`0 Ω`) | `temperature_c` (°C), `humidity_pct` (%), `pressure_hpa` (hPa) |
| **`0x59`** | Sensirion `SGP40` | `U13` | Dirección fija de silicio (`VDD` filtrado por `R42` `22 Ω` + `C46`/`C47`) | `voc_index` (índice 1–500) |
| **`0x62`** | Sensirion `SCD41` | `U12` | Dirección fija de silicio | `co2` (ppm) |
| **`0x69`** | Sensirion `SPS30` | `J9` | Conector JST `ZHR-5` (`5V_SENS`), pin 4 (`SPS_SEL`) a `GND` vía `JP1` | `pm25` (µg/m³), `pm10` (µg/m³) |
| **`0x48`** | TI `ADS1115` #1 | `U14` | Pin 1 (`ADDR`) conectado a `GND` (`ALERT/RDY` → `GPIO38`) | `co` (mg/m³), `no2` (µg/m³) |
| **`0x49`** | TI `ADS1115` #2 | `U15` | Pin 1 (`ADDR`) conectado a `3V3_SENS` | `so2` (µg/m³), `h2s` (µg/m³) |

### 3.2. Frontera Analógica de 8 Canales (`Winsen ZE03`)

Cada conector JST `XH-4` (`J10`–`J13`, alimentado por `5V_SENS`) entrega señal primaria de voltaje analógico (`VO`) y señal de compensación térmica/cero (`CORR`). Cada canal atraviesa un filtro pasa-bajos anti-aliasing `1 kΩ` (`R45`–`R52`) + `100 nF` (`C54`–`C61`, $f_c \approx 1.59\text{ kHz}$) y doble diodo de sujeción Schottky `BAT54S` (`D7`–`D14`) hacia `3V3_SENS` y `GND`:

| Gas / Variable | Conector | Pin 3 / Señal | Filtro & Clamp | Canal ADC | Pin 4 / Compensación | Filtro & Clamp | Canal ADC | Conversión en Firmware (`HW-002`) |
| :--- | :---: | :--- | :--- | :---: | :--- | :--- | :---: | :--- |
| **CO** (`co`) | `J10` | `ZE03_CO_VO` | `R45`+`C54`+`D7` | `U14.AIN0` (`0x48`) | `ZE03_CO_CORR` | `R46`+`C55`+`D8` | `U14.AIN1` (`0x48`) | `mg/m³` |
| **NO₂** (`no2`) | `J11` | `ZE03_NO2_VO` | `R47`+`C56`+`D9` | `U14.AIN2` (`0x48`) | `ZE03_NO2_CORR` | `R48`+`C57`+`D10` | `U14.AIN3` (`0x48`) | `µg/m³` |
| **SO₂** (`so2`) | `J12` | `ZE03_SO2_VO` | `R49`+`C58`+`D11` | `U15.AIN0` (`0x49`) | `ZE03_SO2_CORR` | `R50`+`C59`+`D12` | `U15.AIN1` (`0x49`) | `µg/m³` |
| **H₂S** (`h2s`) | `J13` | `ZE03_H2S_VO` | `R51`+`C60`+`D13` | `U15.AIN2` (`0x49`) | `ZE03_H2S_CORR` | `R52`+`C61`+`D14` | `U15.AIN3` (`0x49`) | `µg/m³` |

### 3.3. Contrato de Salida JSON (`MQTT` QoS 1)

Publicación sobre `aire-trujillo/v1/nodes/{node_code}/telemetry` con las 12 claves exactas definidas en `02-domain-and-data.md`: `pm25`, `pm10`, `co`, `no2`, `so2`, `h2s`, `co2`, `voc_index`, `noise_dba`, `temperature_c`, `humidity_pct`, `pressure_hpa`. El firmware debe omitir cualquier clave cuyo sensor falle (`nullptr` / lectura inválida) en lugar de transmitir `0`.

---

## 4. Subsistema de Energía Solar, Rieles y Secuencia Deep-Sleep

| Riel PCB | Origen | Tensión Nominal | Capacidad / Límite | Notas de Diseño |
| :--- | :--- | :---: | :---: | :--- |
| `V_SOLAR_RAW` | Conector `J4` (2P) | `5 V – 18 V` | Protección inversa `D2` (`SS34`) | Entrada fotovoltaica bruta |
| `V_SOLAR` | Salida `D2` | `5 V – 18 V` | Divisor MPPT `R10` (`100 kΩ`) / `R11` (`16.5 kΩ`) | Seguimiento MPPT: $V_{\text{MPPT}} = 1.205\,(1 + 100/16.5) \approx 8.51\text{ V}$ |
| `VBAT` | Conector `J5` + `U3` | `3.0 V – 4.2 V` | Shunt `R18` (`0.12 Ω 1 W`) → $I_{\text{chg,max}} = 1.0\text{ A}$ | Carga solar MPPT CC/CV (`CN3791` `U3` + `L1` `22 µH`) |
| `VSYS` | ORing `Q1`/`Q2` (`AO3401A`) | `3.0 V – 5.0 V` | `2 A` pico LTE | Bus principal tras selección automática Solar / USB / Batería; alimenta directo a `SIM7080G` (`3.0–4.3 V` en modo batería) y desacoplo `C32` (`100 µF`) |
| `VBUS_USB` | USB-C `J1` (`R4`/`R5` `5.1 kΩ` CC) | `5.0 V` | ESD `U2`, diodo `D4` (`SS34`) | Programación, consola y carga auxiliar |
| `3V3_SYS` | Buck `TLV62569` (`U4`, `L2 1.5 µH`) | `3.318 V` | `R22` (`453 kΩ`) / `R23` (`100 kΩ`) | Riel continuo para `ESP32-S3`, `U9` lado A y pull-ups de carga |
| `3V3_SENS` | Load switch `TPS22919` (`U7`) | `3.318 V` | Habilitado por `GPIO48` (`EN_SENS_3V3`) | Corta consumo en sueño de `ADS1115`, `BME280`, `SCD41`, `SGP40`, `ICS-43434` y `MicroSD` |
| `5V_BOOST` | Boost `TPS61023` (`U5`, `L3 1.0 µH`) | `4.992 V` | `R25` (`732 kΩ`) / `R26` (`100 kΩ`), `EN = GPIO21` | Elevador síncrono activo únicamente durante muestreo de sensores de `5 V` |
| `5V_SENS` | Load switch `TPS22919` (`U6`) | `4.992 V` | Habilitado por `GPIO47` (`EN_SENS_5V`) | Alimentación controlada para `SPS30` y 4× `ZE03` |

### Secuencia de Muestreo y Potencia (`HW-002`)
1. **Despertar (`ESP32-S3`):** Activar `GPIO48` (`EN_SENS_3V3 = HIGH`) para energizar bus I2C, ADCs, micrófono I2S y MicroSD.
2. **Precalentamiento / Sensores 5 V:** Activar `GPIO21` (`PWR_5V_EN = HIGH`), esperar estabilización de `5V_BOOST` (~2 ms), luego habilitar `GPIO47` (`EN_SENS_5V = HIGH`) para arranque del ventilador `SPS30` y celdas `ZE03`.
3. **Adquisición:** Lectura `BME280`/`SCD41`/`SGP40` (temperatura/humedad previa compensación VOC en `SGP40` y lectura `ADS1115`), muestreo I2S `ICS-43434` (ponderación A) y estabilización de `SPS30`.
4. **Apagado de sensores antes de TX LTE:** Apagar `EN_SENS_5V`, `PWR_5V_EN` y (tras persistir en MicroSD si falla enlace) `EN_SENS_3V3` para evitar picos simultáneos con ráfagas LTE-M de `SIM7080G` (`VSYS`).
5. **Transmisión LTE-M / Deep-Sleep:** Activar `MODEM_PWRKEY` (`GPIO16`) o despertar vía `MODEM_DTR` (`GPIO15`), publicar MQTT QoS 1 (o encolar en MicroSD `store-and-forward`), llevar módem a PSM/power-off y entrar en Deep-Sleep (`~50–150 µA` de corriente quiescente combinada objetivo de `U4`, `U5`, `U6`, `U7` y `U8` apagado).

---

## 5. Arquitectura Física de PCB, Apilado (Stackup) y Reglas de Manufactura

* **Dimensiones:** `150.0 × 105.0 mm`, esquinas redondeadas `R = 3.0 mm`, 4 taladros de montaje M3 (`H1`–`H4`, Ø `3.2 mm` chapados GND).
* **Apilado de 4 capas (JLCPCB `JLC04161H-7628`, espesor total `1.60 mm`, acabado superficial `ENIG`):**
  * **Capa 1 (`F.Cu`, `0.035 mm` / 1 oz):** Señales críticas, par diferencial USB `90 Ω` (`0.22 mm` ancho / `0.23 mm` gap), enlaces RF CPWG `50 Ω` (`0.30 mm` sobre prepreg `0.2104 mm` $\varepsilon_r = 4.6$ con valla de vías GND), conversión de potencia local y plano coplanar GND.
  * **Capa 2 (`In1.Cu`, `0.0175 mm` / 0.5 oz):** **Plano sólido e ininterrumpido de `GND`** (retorno continuo de RF, USB y señales mixtas; únicamente excluido bajo antena `ESP32-S3` y puerto acústico `MK1`).
  * **Capa 3 (`In2.Cu`, `0.0175 mm` / 0.5 oz):** Distribución principal de rieles de alimentación (`VSYS`, `VBAT`, `V_SOLAR`, `3V3_SYS`, `3V3_SENS`, `5V_BOOST`, `5V_SENS`) y cruces de señal.
  * **Capa 4 (`B.Cu`, `0.035 mm` / 1 oz):** Enrutamiento ortogonal de señales y plano coplanar `GND`.
* **Zonas de Exclusión (Keepouts) y Aislamiento Térmico:**
  * **Antena `ESP32-S3` (`ESP32_ANT_KEEPOUT`):** `[10.5, 59.5] × [0.0, 7.5] mm` sin cobre en las 4 capas.
  * **Slot de Aislamiento Térmico (`Edge.Cuts`):** Ranura fresada vertical `[(122.5, 56.0)..(123.7, 74.0)] mm` que separa térmicamente `BME280` (`U11`) y `SCD41` (`U12`) en la península derecha respecto al regulador `SIM7080G`/fuentes para evitar deriva térmica en `temperature_c`, `humidity_pct` y `co2`.
  * **Puerto Acústico (`MK1_ACOUSTIC_PORT`):** Exclusión `[140.4, 142.6] × [85.7, 87.8] mm` en `In1.Cu`, `In2.Cu` y `B.Cu` alrededor del taladro NPTH inferior del micrófono `ICS-43434` (`141.50, 86.71 mm`).
  * **Exclusión de Recubrimiento Conformado (`User.Comments`):** Marcos de enmascarado `MASK / DO NOT CONFORMAL COAT INSIDE BOXES` alrededor de `BME280`/`SCD41`, `SGP40`, `MK1`, `USB-C` (`J1`), `MicroSD` (`J2`), `Nano-SIM` (`J6`), conectores `U.FL` (`J7`, `J8`) y regletas terminales (`J4`, `J5`, `J9`–`J13`).

---

## 6. Verificación Eléctrica, DRC y Entregables de Producción

| Verificación / Artefacto | Herramienta | Resultado Observado | Ruta |
| :--- | :--- | :--- | :--- |
| **ERC Esquemático (4 hojas)** | `kicad-cli sch erc --severity-all` | **0 violaciones (0 errores, 0 advertencias)**, 171 símbolos | [`aire-trujillo.kicad_sch`](../pcb/aire-trujillo.kicad_sch) |
| **DRC PCB (4 capas) + Paridad** | `kicad-cli pcb drc --severity-all --schematic-parity` | **0 violaciones (0 errores, 0 advertencias), 0 ítems sin conectar (95/95 redes), 0 errores de paridad** | [`aire-trujillo.kicad_pcb`](../pcb/aire-trujillo.kicad_pcb) |
| **Gerbers X2 + Taladros PTH/NPTH** | `kicad-cli pcb export gerbers` + `drill` | Paquete completo empaquetado para JLCPCB (`4L ENIG`) | [`aire-trujillo-gerbers.zip`](../pcb/manufacturing/aire-trujillo-gerbers.zip) |
| **Lista de Materiales (BOM) & CPL** | `kicad-cli sch export bom` + `pcb export pos` | 171 referencias agrupadas en 66 líneas únicas + centroides SMD Top/Bottom | [`aire-trujillo-bom.csv`](../pcb/manufacturing/aire-trujillo-bom.csv), [`aire-trujillo-pos.csv`](../pcb/manufacturing/aire-trujillo-pos.csv) |
| **Esquemático PDF, STEP 3D, Renders y SVGs** | `kicad-cli` | PDF 4 hojas, modelo sólido STEP 3D, renders Top/Bottom 2400×1680 y SVGs por capa | [`manufacturing/`](../pcb/manufacturing/) |
