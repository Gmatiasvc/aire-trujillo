#include "sensors_hal.h"
#include "bus_hal.h"
#include "aire_board.h"

#include <math.h>
#include <stdlib.h>
#include <string.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "driver/gpio.h"
#include "esp_log.h"
#include "esp_sleep.h"
#include "esp_timer.h"

static const char *TAG = "sensors_hal";

static uint32_t now_ms(void) {
    return (uint32_t)(esp_timer_get_time() / 1000ULL);
}

static void sleep_energy_aware_ms(uint32_t ms) {
    if (ms == 0U) {
        return;
    }
    if (ms < 50U) {
        vTaskDelay(pdMS_TO_TICKS(ms));
        return;
    }
    if (esp_sleep_enable_timer_wakeup((uint64_t)ms * 1000ULL) == ESP_OK) {
        gpio_hold_en(AIRE_GPIO_EN_SENS_3V3);
        gpio_hold_en(AIRE_GPIO_PWR_5V_EN);
        gpio_hold_en(AIRE_GPIO_EN_SENS_5V);
        esp_light_sleep_start();
        gpio_hold_dis(AIRE_GPIO_EN_SENS_5V);
        gpio_hold_dis(AIRE_GPIO_PWR_5V_EN);
        gpio_hold_dis(AIRE_GPIO_EN_SENS_3V3);
        esp_sleep_disable_wakeup_source(ESP_SLEEP_WAKEUP_TIMER);
    } else {
        vTaskDelay(pdMS_TO_TICKS(ms));
    }
}

/* ==================== 1. Bosch BME280 (0x76) ==================== */

static bool read_bme280(double *out_t, double *out_rh, double *out_p_hpa) {
    uint8_t reg = 0xD0U;
    uint8_t chip_id = 0U;
    if (bus_hal_i2c_write_read(AIRE_I2C_ADDR_BME280, &reg, 1U, &chip_id, 1U) != ESP_OK || chip_id != 0x60U) {
        ESP_LOGW(TAG, "BME280 ausente o ID inválido (0x%02X)", chip_id);
        return false;
    }

    uint8_t cal_a[26];
    uint8_t cal_b[7];
    reg = 0x88U;
    if (bus_hal_i2c_write_read(AIRE_I2C_ADDR_BME280, &reg, 1U, cal_a, sizeof(cal_a)) != ESP_OK) {
        return false;
    }
    reg = 0xE1U;
    if (bus_hal_i2c_write_read(AIRE_I2C_ADDR_BME280, &reg, 1U, cal_b, sizeof(cal_b)) != ESP_OK) {
        return false;
    }

    aire_bme280_calib_t calib;
    if (!aire_bme280_parse_calib(cal_a, cal_b, &calib)) {
        return false;
    }

    const uint8_t cfg_hum[2]  = { 0xF2U, 0x01U }; /* osrs_h = x1 */
    const uint8_t cfg_filt[2] = { 0xF5U, 0x00U }; /* Filtro IIR apagado en forced mode */
    const uint8_t cfg_meas[2] = { 0xF4U, 0x25U }; /* osrs_t=x1, osrs_p=x1, mode=forced */
    if (bus_hal_i2c_write(AIRE_I2C_ADDR_BME280, cfg_hum, 2U) != ESP_OK ||
        bus_hal_i2c_write(AIRE_I2C_ADDR_BME280, cfg_filt, 2U) != ESP_OK ||
        bus_hal_i2c_write(AIRE_I2C_ADDR_BME280, cfg_meas, 2U) != ESP_OK) {
        return false;
    }

    vTaskDelay(pdMS_TO_TICKS(15));
    for (int retry = 0; retry < 5; ++retry) {
        uint8_t st_reg = 0xF3U, st = 0U;
        if (bus_hal_i2c_write_read(AIRE_I2C_ADDR_BME280, &st_reg, 1U, &st, 1U) == ESP_OK && (st & 0x08U) == 0U) {
            break;
        }
        vTaskDelay(pdMS_TO_TICKS(5));
    }

    uint8_t raw[8];
    reg = 0xF7U;
    if (bus_hal_i2c_write_read(AIRE_I2C_ADDR_BME280, &reg, 1U, raw, sizeof(raw)) != ESP_OK) {
        return false;
    }

    int32_t adc_P = ((int32_t)raw[0] << 12) | ((int32_t)raw[1] << 4) | ((int32_t)raw[2] >> 4);
    int32_t adc_T = ((int32_t)raw[3] << 12) | ((int32_t)raw[4] << 4) | ((int32_t)raw[5] >> 4);
    int32_t adc_H = ((int32_t)raw[6] << 8)  | (int32_t)raw[7];
    if (adc_P == 0x80000 || adc_T == 0x80000 || adc_H == 0x8000) {
        return false;
    }

    return aire_bme280_compensate(&calib, adc_T, adc_P, adc_H, out_t, out_p_hpa, out_rh);
}

/* ==================== 2. Sensirion SCD41 (0x62) ==================== */

static bool read_scd41(bool has_press, double press_hpa,
                       double *out_co2, double *out_t, double *out_rh) {
    if (has_press && press_hpa >= 700.0 && press_hpa <= 1200.0) {
        uint16_t mbar = (uint16_t)llround(press_hpa);
        uint8_t p_cmd[5] = { 0xE0U, 0x00U, (uint8_t)(mbar >> 8), (uint8_t)(mbar & 0xFFU), 0U };
        p_cmd[4] = aire_crc8_sensirion(&p_cmd[2], 2U);
        (void)bus_hal_i2c_write(AIRE_I2C_ADDR_SCD41, p_cmd, sizeof(p_cmd));
        vTaskDelay(pdMS_TO_TICKS(2));
    }

    const uint8_t trig[2] = { 0x21U, 0x9DU }; /* measure_single_shot (~5000 ms) */
    if (bus_hal_i2c_write(AIRE_I2C_ADDR_SCD41, trig, sizeof(trig)) != ESP_OK) {
        ESP_LOGW(TAG, "SCD41 no respondió al disparo single-shot");
        return false;
    }

    vTaskDelay(pdMS_TO_TICKS(5050));

    bool ready = false;
    for (int retry = 0; retry < 15; ++retry) {
        const uint8_t drdy_cmd[2] = { 0xE4U, 0xB8U };
        uint8_t rx[3] = { 0U };
        if (bus_hal_i2c_write(AIRE_I2C_ADDR_SCD41, drdy_cmd, 2U) == ESP_OK) {
            vTaskDelay(pdMS_TO_TICKS(2));
            if (bus_hal_i2c_read(AIRE_I2C_ADDR_SCD41, rx, 3U) == ESP_OK &&
                aire_sensirion_check_words(rx, 1U)) {
                uint16_t st = ((uint16_t)rx[0] << 8) | rx[1];
                if ((st & 0x07FFU) != 0U) {
                    ready = true;
                    break;
                }
            }
        }
        vTaskDelay(pdMS_TO_TICKS(100));
    }
    if (!ready) {
        return false;
    }

    const uint8_t rd_cmd[2] = { 0xECU, 0x05U }; /* read_measurement */
    uint8_t data[9] = { 0U };
    if (bus_hal_i2c_write(AIRE_I2C_ADDR_SCD41, rd_cmd, 2U) != ESP_OK) {
        return false;
    }
    vTaskDelay(pdMS_TO_TICKS(2));
    if (bus_hal_i2c_read(AIRE_I2C_ADDR_SCD41, data, sizeof(data)) != ESP_OK ||
        !aire_sensirion_check_words(data, 3U)) {
        return false;
    }

    const uint16_t words[3] = {
        (uint16_t)(((uint16_t)data[0] << 8) | data[1]),
        (uint16_t)(((uint16_t)data[3] << 8) | data[4]),
        (uint16_t)(((uint16_t)data[6] << 8) | data[7]),
    };
    aire_scd41_decode(words, out_co2, out_t, out_rh);

    const uint8_t pdn[2] = { 0x36U, 0xE0U }; /* power_down */
    (void)bus_hal_i2c_write(AIRE_I2C_ADDR_SCD41, pdn, sizeof(pdn));
    return true;
}

/* ==================== 3. Sensirion SGP40 (0x59) ==================== */

static bool read_sgp40(aire_voc_state_t *rtc_voc, double temp_c, double rh_pct, double *out_voc) {
    uint8_t cmd[8];
    cmd[0] = 0x26U;
    cmd[1] = 0x0FU; /* measure_raw_signal */
    aire_sgp40_encode_ticks(temp_c, rh_pct, &cmd[2]);

    if (bus_hal_i2c_write(AIRE_I2C_ADDR_SGP40, cmd, sizeof(cmd)) != ESP_OK) {
        ESP_LOGW(TAG, "SGP40 no respondió");
        return false;
    }
    vTaskDelay(pdMS_TO_TICKS(35));

    uint8_t rx[3] = { 0U };
    if (bus_hal_i2c_read(AIRE_I2C_ADDR_SGP40, rx, sizeof(rx)) != ESP_OK ||
        !aire_sensirion_check_words(rx, 1U)) {
        return false;
    }
    uint16_t sraw = ((uint16_t)rx[0] << 8) | rx[1];
    *out_voc = aire_sgp40_calc_voc_index(rtc_voc, sraw);

    const uint8_t off[2] = { 0x36U, 0x15U }; /* turn_heater_off */
    (void)bus_hal_i2c_write(AIRE_I2C_ADDR_SGP40, off, sizeof(off));
    return true;
}

/* ==================== 4. Sensirion SPS30 (0x69 @ 5V_SENS) ==================== */

static bool sps30_start(void) {
    uint8_t cmd[5] = { 0x00U, 0x10U, 0x03U, 0x00U, 0U }; /* start_measurement float big-endian */
    cmd[4] = aire_crc8_sensirion(&cmd[2], 2U);
    return bus_hal_i2c_write(AIRE_I2C_ADDR_SPS30, cmd, sizeof(cmd)) == ESP_OK;
}

static bool sps30_read_and_stop(double *out_pm25, double *out_pm10) {
    bool ok = false;
    for (int retry = 0; retry < 20; ++retry) {
        const uint8_t drdy[2] = { 0x02U, 0x02U };
        uint8_t rx[3] = { 0U };
        if (bus_hal_i2c_write(AIRE_I2C_ADDR_SPS30, drdy, 2U) == ESP_OK) {
            vTaskDelay(pdMS_TO_TICKS(2));
            if (bus_hal_i2c_read(AIRE_I2C_ADDR_SPS30, rx, 3U) == ESP_OK &&
                aire_sensirion_check_words(rx, 1U) && rx[1] == 0x01U) {
                ok = true;
                break;
            }
        }
        vTaskDelay(pdMS_TO_TICKS(100));
    }

    bool decoded = false;
    if (ok) {
        const uint8_t rd[2] = { 0x03U, 0x00U }; /* read_measured_values (60 B) */
        uint8_t payload[60] = { 0U };
        if (bus_hal_i2c_write(AIRE_I2C_ADDR_SPS30, rd, 2U) == ESP_OK) {
            vTaskDelay(pdMS_TO_TICKS(2));
            if (bus_hal_i2c_read(AIRE_I2C_ADDR_SPS30, payload, sizeof(payload)) == ESP_OK) {
                decoded = aire_sps30_decode_pm(payload, out_pm25, out_pm10);
            }
        }
    }

    const uint8_t stop[2]  = { 0x01U, 0x04U };
    const uint8_t sleep[2] = { 0x10U, 0x01U };
    (void)bus_hal_i2c_write(AIRE_I2C_ADDR_SPS30, stop, 2U);
    vTaskDelay(pdMS_TO_TICKS(5));
    (void)bus_hal_i2c_write(AIRE_I2C_ADDR_SPS30, sleep, 2U);
    return decoded;
}

/* ==================== 5. TI ADS1115 + 4x Winsen ZE03 ==================== */

static bool ads1115_read_channel_mv(uint8_t addr7, uint8_t ch, double *out_mv) {
    if (ch > 3U || !out_mv) {
        return false;
    }
    double sum_mv = 0.0;
    const uint8_t hi = (uint8_t)(0xC3U | ((ch & 0x03U) << 4)); /* OS=1, MUX=4+ch, PGA=+-4.096V, MODE=single */
    const uint8_t cfg[3] = { 0x01U, hi, 0x83U };               /* DR=128 SPS, COMP_QUE=11 */

    for (int sample = 0; sample < 4; ++sample) {
        if (bus_hal_i2c_write(addr7, cfg, 3U) != ESP_OK) {
            return false;
        }
        vTaskDelay(pdMS_TO_TICKS(10));

        bool conv_done = false;
        uint8_t reg_cfg = 0x01U;
        uint8_t st[2] = { 0U };
        for (int w = 0; w < 5; ++w) {
            if (bus_hal_i2c_write_read(addr7, &reg_cfg, 1U, st, 2U) == ESP_OK && (st[0] & 0x80U) != 0U) {
                conv_done = true;
                break;
            }
            vTaskDelay(pdMS_TO_TICKS(4));
        }
        if (!conv_done) {
            return false;
        }

        uint8_t reg_conv = 0x00U;
        uint8_t raw[2] = { 0U };
        if (bus_hal_i2c_write_read(addr7, &reg_conv, 1U, raw, 2U) != ESP_OK) {
            return false;
        }
        int16_t counts = (int16_t)(((uint16_t)raw[0] << 8) | raw[1]);
        sum_mv += aire_ads1115_counts_to_mv(counts);
    }

    *out_mv = sum_mv / 4.0;
    return true;
}

static void acquire_ze03_pair(aire_telemetry_t *tel, aire_metric_id_t id,
                              uint8_t ads_addr, uint8_t ch_vo, uint8_t ch_corr,
                              const aire_ze03_cal_t *cal, double temp_c) {
    double v_vo = 0.0, v_corr = 0.0;
    if (!ads1115_read_channel_mv(ads_addr, ch_vo, &v_vo)) {
        return;
    }
    /* Rechaza canales flotantes o cortocircuitados fuera de ventana electroquímica activa */
    if (v_vo < 150.0 || v_vo > 3150.0) {
        return;
    }
    if (!ads1115_read_channel_mv(ads_addr, ch_corr, &v_corr) || v_corr < 150.0 || v_corr > 3150.0) {
        v_corr = cal->corr_zero_mv; /* Sin canal CORR activo, degrada a compensación térmica pura */
    }
    double conc = aire_ze03_compensate(cal, v_vo, v_corr, temp_c);
    (void)aire_telemetry_set_metric(tel, id, conc);
}

/* ==================== 6. Acústica I2S ICS-43434 ==================== */

static bool read_noise_dba(double cal_offset_db, double *out_dba) {
    if (bus_hal_i2s_init() != ESP_OK) {
        ESP_LOGW(TAG, "Fallo inicializando I2S ICS-43434");
        return false;
    }

    int32_t *chunk = (int32_t *)malloc(512U * sizeof(int32_t));
    if (!chunk) {
        bus_hal_i2s_deinit();
        return false;
    }

    /* Descarta ~85 ms iniciales para estabilizar filtro de decimación del MEMS */
    size_t n = 0U;
    for (int skip = 0; skip < 8; ++skip) {
        (void)bus_hal_i2s_read_samples(chunk, 512U, &n);
    }

    aire_dba_ctx_t ctx;
    aire_dba_init(&ctx);
    const size_t target_samples = 48000U + 24000U; /* 1.5 s @ 48 kHz */
    while (ctx.sample_count < target_samples) {
        if (bus_hal_i2s_read_samples(chunk, 512U, &n) != ESP_OK || n == 0U) {
            break;
        }
        aire_dba_process_i24(&ctx, chunk, n);
    }

    free(chunk);
    bus_hal_i2s_deinit();
    return aire_dba_compute(&ctx, cal_offset_db, out_dba);
}

/* ==================== Orquestación Completa del Ciclo ==================== */

void sensors_hal_acquire_all(aire_telemetry_t *out_tel,
                             aire_voc_state_t *rtc_voc_state,
                             bool allow_5v_sensors,
                             uint32_t warmup_ms,
                             double mic_cal_offset_db) {
    if (!out_tel) {
        return;
    }

    uint32_t t_5v_start = 0U;
    bool sps30_armed = false;
    if (allow_5v_sensors) {
        bus_hal_set_5v_sens(true);
        t_5v_start = now_ms();
        sps30_armed = sps30_start();
    }

    /* Variables ambientales para compensación interna (Costa Trujillo por defecto si falla BME280) */
    double comp_t  = 22.0;
    double comp_rh = 75.0;
    double comp_p  = 1013.0;
    bool bme_ok = read_bme280(&comp_t, &comp_rh, &comp_p);
    if (bme_ok) {
        (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_TEMPERATURE_C, comp_t);
        (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_HUMIDITY_PCT, comp_rh);
        (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_PRESSURE_HPA, comp_p);
    }

    double scd_co2 = 0.0, scd_t = 0.0, scd_rh = 0.0;
    if (read_scd41(bme_ok, comp_p, &scd_co2, &scd_t, &scd_rh)) {
        (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_CO2, scd_co2);
        if (!bme_ok) {
            comp_t  = scd_t;
            comp_rh = scd_rh;
            (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_TEMPERATURE_C, scd_t);
            (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_HUMIDITY_PCT, scd_rh);
        }
    }

    double voc = 0.0;
    if (read_sgp40(rtc_voc_state, comp_t, comp_rh, &voc)) {
        (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_VOC_INDEX, voc);
    }

    double noise_dba = 0.0;
    if (read_noise_dba(mic_cal_offset_db, &noise_dba)) {
        (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_NOISE_DBA, noise_dba);
    }

    if (allow_5v_sensors) {
        uint32_t elapsed = now_ms() - t_5v_start;
        if (elapsed < warmup_ms) {
            uint32_t rem = warmup_ms - elapsed;
            ESP_LOGI(TAG, "Estabilizando SPS30/ZE03 durante %u ms en bajo consumo", (unsigned)rem);
            sleep_energy_aware_ms(rem);
        }

        if (sps30_armed) {
            double pm25 = 0.0, pm10 = 0.0;
            if (sps30_read_and_stop(&pm25, &pm10)) {
                (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_PM25, pm25);
                (void)aire_telemetry_set_metric(out_tel, AIRE_METRIC_PM10, pm10);
            }
        }

        const aire_ze03_cal_t co_cal  = aire_ze03_default_co();
        const aire_ze03_cal_t no2_cal = aire_ze03_default_no2();
        const aire_ze03_cal_t so2_cal = aire_ze03_default_so2();
        const aire_ze03_cal_t h2s_cal = aire_ze03_default_h2s();

        acquire_ze03_pair(out_tel, AIRE_METRIC_CO,  AIRE_I2C_ADDR_ADS1115_1, 0U, 1U, &co_cal,  comp_t);
        acquire_ze03_pair(out_tel, AIRE_METRIC_NO2, AIRE_I2C_ADDR_ADS1115_1, 2U, 3U, &no2_cal, comp_t);
        acquire_ze03_pair(out_tel, AIRE_METRIC_SO2, AIRE_I2C_ADDR_ADS1115_2, 0U, 1U, &so2_cal, comp_t);
        acquire_ze03_pair(out_tel, AIRE_METRIC_H2S, AIRE_I2C_ADDR_ADS1115_2, 2U, 3U, &h2s_cal, comp_t);

        /* Apaga boost 5V antes de montar SD / transmitir por módem LTE-M */
        bus_hal_set_5v_sens(false);
    }
}
