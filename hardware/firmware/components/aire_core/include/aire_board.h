#ifndef AIRE_BOARD_H
#define AIRE_BOARD_H

/*
 * aire-trujillo — Mapa Definitivo de Pines, Buses y Reglas de Hardware (HW-001 v1.0)
 * MCU: ESP32-S3-WROOM-1-N16R8 (16 MB QIO Flash / 8 MB Octal PSRAM)
 * Sincronizado exactamente con hardware/doc/specs.md y aire-trujillo.kicad_pcb.
 */

/* Bus I2S — InvenSense ICS-43434 (MK1, Canal Izquierdo LR=GND) */
#define AIRE_GPIO_I2S_SCK       4
#define AIRE_GPIO_I2S_WS        5
#define AIRE_GPIO_I2S_SD        6

/* Módem Celular / GNSS — SIMCom SIM7080G (U8) vía SN74AVC4T245 (U9) y BSS138 (Q3) */
#define AIRE_GPIO_MODEM_STATUS  7
#define AIRE_GPIO_MODEM_DTR     15
#define AIRE_GPIO_MODEM_PWRKEY  16
#define AIRE_GPIO_MODEM_TXD     17   /* ESP32 TX -> U8 RXD */
#define AIRE_GPIO_MODEM_RXD     18   /* ESP32 RX <- U8 TXD */

/* Bus I2C Principal (3.3 V conmutado 3V3_SENS, R40/R41 4.7k pull-ups) */
#define AIRE_GPIO_I2C_SDA       8
#define AIRE_GPIO_I2C_SCL       9
#define AIRE_I2C_FREQ_HZ        100000U

/* Direcciones I2C de 7 bits verificadas en esquemático */
#define AIRE_I2C_ADDR_ADS1115_1 0x48U  /* U14: ADDR=GND -> CO (AIN0/1), NO2 (AIN2/3) */
#define AIRE_I2C_ADDR_ADS1115_2 0x49U  /* U15: ADDR=3V3_SENS -> SO2 (AIN0/1), H2S (AIN2/3) */
#define AIRE_I2C_ADDR_SGP40     0x59U  /* U13: VOC index */
#define AIRE_I2C_ADDR_SCD41     0x62U  /* U12: CO2 NDIR */
#define AIRE_I2C_ADDR_SPS30     0x69U  /* J9 : SPS_SEL=GND (JP1) */
#define AIRE_I2C_ADDR_BME280    0x76U  /* U11: BME_SDO=GND (R53 0R) */

/* Bus SPI — Socket MicroSD Hirose DM3AT-SF-PEJM5 (J2) */
#define AIRE_GPIO_SD_CS         10
#define AIRE_GPIO_SD_MOSI       11
#define AIRE_GPIO_SD_CLK        12
#define AIRE_GPIO_SD_MISO       13
#define AIRE_GPIO_SD_CD         14

/* Control de Potencia y Load Switches (TPS61023 U5 + 2x TPS22919 U6/U7) */
#define AIRE_GPIO_PWR_5V_EN     21   /* Habilita Boost 5V_BOOST (U5 TPS61023) */
#define AIRE_GPIO_EN_SENS_5V    47   /* Habilita 5V_SENS (U6 para SPS30 + 4x ZE03) */
#define AIRE_GPIO_EN_SENS_3V3   48   /* Habilita 3V3_SENS (U7 para I2C, ADCs, MK1, SD) */

/* Interrupción ADC y Telemetría de Carga MPPT CN3791 (U3) / LED */
#define AIRE_GPIO_ADS_ALERT     38   /* U14 ALERT/RDY open-drain (R44 10k pull-up) */
#define AIRE_GPIO_CHG_STAT      39   /* U3 CHG_STAT open-drain (R14 10k pull-up) */
#define AIRE_GPIO_CHG_DONE      40   /* U3 CHG_DONE open-drain (R15 10k pull-up) */
#define AIRE_GPIO_LED_SYS       41   /* LED verde D1 vía R3 1k */

/* Consola Serie Debug UART0 (J3) */
#define AIRE_GPIO_UART0_RXD     44
#define AIRE_GPIO_UART0_TXD     43

/* Telemetría Analógica de Energía (ADC1 ESP32-S3) */
#define AIRE_GPIO_VBAT_ADC      1    /* ADC1_CH0: R19 100k / R20 100k (x2.0) */
#define AIRE_GPIO_VSOLAR_ADC    2    /* ADC1_CH1: R12 470k / R13 47k  (x11.0) */
#define AIRE_VBAT_DIV_NUM       2.0f
#define AIRE_VSOLAR_DIV_NUM     11.0f

/* Temporizaciones Hardware Seguras (ms) */
#define AIRE_SETTLE_3V3_SENS_MS 10U
#define AIRE_SETTLE_5V_BOOST_MS 5U
#define AIRE_SETTLE_5V_SENS_MS  20U
#define AIRE_MODEM_PWRKEY_MS    1250U

#endif /* AIRE_BOARD_H */
