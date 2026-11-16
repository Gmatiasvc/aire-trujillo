#include "aire_calibration.h"

#include <math.h>
#include <string.h>

/* ==================== 1. CRC-8 Sensirion (Poly 0x31, Init 0xFF) ==================== */

uint8_t aire_crc8_sensirion(const uint8_t *data, size_t len) {
    uint8_t crc = 0xFFU;
    for (size_t i = 0U; i < len; ++i) {
        crc ^= data[i];
        for (int bit = 0; bit < 8; ++bit) {
            crc = (crc & 0x80U) ? (uint8_t)((crc << 1) ^ 0x31U) : (uint8_t)(crc << 1);
        }
    }
    return crc;
}

bool aire_sensirion_check_words(const uint8_t *buf, size_t word_count) {
    if (!buf) {
        return false;
    }
    for (size_t i = 0U; i < word_count; ++i) {
        const uint8_t *w = &buf[i * 3U];
        if (aire_crc8_sensirion(w, 2U) != w[2]) {
            return false;
        }
    }
    return true;
}

/* ==================== 2. Bosch BME280 (Compensación Oficial Datasheet) ==================== */

static uint16_t u16_le(const uint8_t *b) {
    return (uint16_t)b[0] | ((uint16_t)b[1] << 8);
}

static int16_t s16_le(const uint8_t *b) {
    return (int16_t)u16_le(b);
}

bool aire_bme280_parse_calib(const uint8_t reg_88_a1[26], const uint8_t reg_e1_e7[7], aire_bme280_calib_t *out) {
    if (!reg_88_a1 || !reg_e1_e7 || !out) {
        return false;
    }
    out->dig_T1 = u16_le(&reg_88_a1[0]);
    out->dig_T2 = s16_le(&reg_88_a1[2]);
    out->dig_T3 = s16_le(&reg_88_a1[4]);
    out->dig_P1 = u16_le(&reg_88_a1[6]);
    out->dig_P2 = s16_le(&reg_88_a1[8]);
    out->dig_P3 = s16_le(&reg_88_a1[10]);
    out->dig_P4 = s16_le(&reg_88_a1[12]);
    out->dig_P5 = s16_le(&reg_88_a1[14]);
    out->dig_P6 = s16_le(&reg_88_a1[16]);
    out->dig_P7 = s16_le(&reg_88_a1[18]);
    out->dig_P8 = s16_le(&reg_88_a1[20]);
    out->dig_P9 = s16_le(&reg_88_a1[22]);
    out->dig_H1 = reg_88_a1[25];

    out->dig_H2 = s16_le(&reg_e1_e7[0]);
    out->dig_H3 = reg_e1_e7[2];
    const uint16_t raw_h4 = (uint16_t)(((uint16_t)reg_e1_e7[3] << 4) | (uint16_t)(reg_e1_e7[4] & 0x0FU));
    const uint16_t raw_h5 = (uint16_t)(((uint16_t)reg_e1_e7[5] << 4) | (uint16_t)(reg_e1_e7[4] >> 4));
    out->dig_H4 = (raw_h4 & 0x800U) ? (int16_t)(raw_h4 | 0xF000U) : (int16_t)raw_h4;
    out->dig_H5 = (raw_h5 & 0x800U) ? (int16_t)(raw_h5 | 0xF000U) : (int16_t)raw_h5;
    out->dig_H6 = (int8_t)reg_e1_e7[6];
    return out->dig_T1 != 0U && out->dig_P1 != 0U;
}

bool aire_bme280_compensate(const aire_bme280_calib_t *c,
                            int32_t adc_T, int32_t adc_P, int32_t adc_H,
                            double *out_temp_c, double *out_press_hpa, double *out_hum_pct) {
    if (!c || c->dig_T1 == 0U || c->dig_P1 == 0U) {
        return false;
    }

    double var1 = (((double)adc_T) / 16384.0 - ((double)c->dig_T1) / 1024.0) * ((double)c->dig_T2);
    double var2 = ((((double)adc_T) / 131072.0 - ((double)c->dig_T1) / 8192.0) *
                   (((double)adc_T) / 131072.0 - ((double)c->dig_T1) / 8192.0)) * ((double)c->dig_T3);
    double t_fine = var1 + var2;
    double temp_c = t_fine / 5120.0;

    var1 = (t_fine / 2.0) - 64000.0;
    var2 = var1 * var1 * ((double)c->dig_P6) / 32768.0;
    var2 = var2 + var1 * ((double)c->dig_P5) * 2.0;
    var2 = (var2 / 4.0) + (((double)c->dig_P4) * 65536.0);
    var1 = (((double)c->dig_P3) * var1 * var1 / 524288.0 + ((double)c->dig_P2) * var1) / 524288.0;
    var1 = (1.0 + var1 / 32768.0) * ((double)c->dig_P1);
    if (fabs(var1) < 1e-9) {
        return false;
    }
    double p_pa = 1048576.0 - (double)adc_P;
    p_pa = (p_pa - (var2 / 4096.0)) * 6250.0 / var1;
    var1 = ((double)c->dig_P9) * p_pa * p_pa / 2147483648.0;
    var2 = p_pa * ((double)c->dig_P8) / 32768.0;
    p_pa = p_pa + (var1 + var2 + ((double)c->dig_P7)) / 16.0;

    double h = t_fine - 76800.0;
    h = (adc_H - (((double)c->dig_H4) * 64.0 + ((double)c->dig_H5) / 16384.0 * h)) *
        (((double)c->dig_H2) / 65536.0 * (1.0 + ((double)c->dig_H6) / 67108864.0 * h *
         (1.0 + ((double)c->dig_H3) / 67108864.0 * h)));
    h = h * (1.0 - ((double)c->dig_H1) * h / 524288.0);
    if (h > 100.0) {
        h = 100.0;
    } else if (h < 0.0) {
        h = 0.0;
    }

    if (out_temp_c) {
        *out_temp_c = temp_c;
    }
    if (out_press_hpa) {
        *out_press_hpa = p_pa / 100.0;
    }
    if (out_hum_pct) {
        *out_hum_pct = h;
    }
    return true;
}

/* ==================== 3. SCD41, SGP40 & SPS30 ==================== */

void aire_scd41_decode(const uint16_t words[3], double *out_co2_ppm, double *out_temp_c, double *out_rh_pct) {
    if (!words) {
        return;
    }
    if (out_co2_ppm) {
        *out_co2_ppm = (double)words[0];
    }
    if (out_temp_c) {
        *out_temp_c = -45.0 + 175.0 * ((double)words[1] / 65535.0);
    }
    if (out_rh_pct) {
        *out_rh_pct = 100.0 * ((double)words[2] / 65535.0);
    }
}

void aire_sgp40_encode_ticks(double temp_c, double rh_pct, uint8_t out_6bytes[6]) {
    if (!out_6bytes) {
        return;
    }
    double clamped_rh = rh_pct < 0.0 ? 0.0 : (rh_pct > 100.0 ? 100.0 : rh_pct);
    double clamped_t  = temp_c < -45.0 ? -45.0 : (temp_c > 130.0 ? 130.0 : temp_c);
    uint16_t rh_tick  = (uint16_t)llround(clamped_rh * 65535.0 / 100.0);
    uint16_t t_tick   = (uint16_t)llround((clamped_t + 45.0) * 65535.0 / 175.0);

    out_6bytes[0] = (uint8_t)(rh_tick >> 8);
    out_6bytes[1] = (uint8_t)(rh_tick & 0xFFU);
    out_6bytes[2] = aire_crc8_sensirion(&out_6bytes[0], 2U);

    out_6bytes[3] = (uint8_t)(t_tick >> 8);
    out_6bytes[4] = (uint8_t)(t_tick & 0xFFU);
    out_6bytes[5] = aire_crc8_sensirion(&out_6bytes[3], 2U);
}

double aire_sgp40_calc_voc_index(aire_voc_state_t *state, uint16_t sraw_voc) {
    /*
     * El índice SGP40 MOX disminuye sraw al subir compuestos reductores (VOC).
     * Línea base adaptativa EWMA en RTC mapeada mediante sigmoide centrada en 100.
     */
    double sraw = (double)sraw_voc;
    if (!state) {
        return 100.0;
    }
    if (!state->initialized || state->baseline_sraw < 1000.0) {
        state->baseline_sraw = (sraw > 5000.0) ? sraw : 30000.0;
        state->initialized   = true;
        return 100.0;
    }
    state->baseline_sraw = 0.98 * state->baseline_sraw + 0.02 * sraw;
    double delta = state->baseline_sraw - sraw;
    double voc = 500.0 / (1.0 + exp(-(delta / 850.0) + log(4.0)));
    if (voc < 1.0) {
        voc = 1.0;
    } else if (voc > 500.0) {
        voc = 500.0;
    }
    return voc;
}

static float be_bytes_to_float(const uint8_t w4[6]) {
    uint32_t raw = ((uint32_t)w4[0] << 24) |
                   ((uint32_t)w4[1] << 16) |
                   ((uint32_t)w4[3] << 8)  |
                   ((uint32_t)w4[4]);
    float f = 0.0f;
    memcpy(&f, &raw, sizeof(f));
    return f;
}

bool aire_sps30_decode_pm(const uint8_t raw_60bytes[60], double *out_pm25, double *out_pm10) {
    if (!raw_60bytes || !aire_sensirion_check_words(raw_60bytes, 20U)) {
        return false;
    }
    /* Orden IEEE754 SPS30: [0] PM1.0, [1] PM2.5 (offset 6), [2] PM4.0, [3] PM10 (offset 18) */
    double pm25 = (double)be_bytes_to_float(&raw_60bytes[6]);
    double pm10 = (double)be_bytes_to_float(&raw_60bytes[18]);
    if (!isfinite(pm25) || !isfinite(pm10)) {
        return false;
    }
    if (out_pm25) {
        *out_pm25 = pm25;
    }
    if (out_pm10) {
        *out_pm10 = pm10;
    }
    return true;
}

/* ==================== 4. ADS1115 & Winsen ZE03 Compensado ==================== */

double aire_ads1115_counts_to_mv(int16_t raw_counts) {
    return (double)raw_counts * (4096.0 / 32768.0); /* 0.125 mV / LSB */
}

aire_ze03_cal_t aire_ze03_default_co(void) {
    return (aire_ze03_cal_t){
        .v_zero_mv = 400.0, .v_span_mv = 1600.0, .full_scale_ppm = 100.0,
        .molar_mass_g_mol = 28.01, .corr_zero_mv = 400.0, .corr_gain = 1.0,
        .temp_coeff_per_c = 0.002, .output_in_ug_m3 = false
    };
}

aire_ze03_cal_t aire_ze03_default_no2(void) {
    return (aire_ze03_cal_t){
        .v_zero_mv = 400.0, .v_span_mv = 1600.0, .full_scale_ppm = 2.0,
        .molar_mass_g_mol = 46.01, .corr_zero_mv = 400.0, .corr_gain = 1.0,
        .temp_coeff_per_c = 0.003, .output_in_ug_m3 = true
    };
}

aire_ze03_cal_t aire_ze03_default_so2(void) {
    return (aire_ze03_cal_t){
        .v_zero_mv = 400.0, .v_span_mv = 1600.0, .full_scale_ppm = 2.0,
        .molar_mass_g_mol = 64.07, .corr_zero_mv = 400.0, .corr_gain = 1.0,
        .temp_coeff_per_c = 0.0025, .output_in_ug_m3 = true
    };
}

aire_ze03_cal_t aire_ze03_default_h2s(void) {
    return (aire_ze03_cal_t){
        .v_zero_mv = 400.0, .v_span_mv = 1600.0, .full_scale_ppm = 2.0,
        .molar_mass_g_mol = 34.08, .corr_zero_mv = 400.0, .corr_gain = 1.0,
        .temp_coeff_per_c = 0.0025, .output_in_ug_m3 = true
    };
}

double aire_ze03_compensate(const aire_ze03_cal_t *cal, double v_vo_mv, double v_corr_mv, double temp_c) {
    if (!cal || cal->v_span_mv <= 1.0 || !isfinite(v_vo_mv) || !isfinite(v_corr_mv)) {
        return -1.0;
    }
    /* Compensa deriva de línea base con el canal libre CORR y sensibilidad térmica (ref 25 C) */
    double baseline_shift_mv = (v_corr_mv - cal->corr_zero_mv) * cal->corr_gain;
    double v_net_mv = (v_vo_mv - baseline_shift_mv) - cal->v_zero_mv;
    if (v_net_mv < 0.0) {
        v_net_mv = 0.0;
    }
    double dt = isfinite(temp_c) ? (temp_c - 25.0) : 0.0;
    double temp_factor = 1.0 + cal->temp_coeff_per_c * dt;
    if (temp_factor < 0.5) {
        temp_factor = 0.5;
    }
    double ppm = (v_net_mv / cal->v_span_mv) * cal->full_scale_ppm / temp_factor;
    /* Conversión estándar a 25 C, 1 atm (24.45 L/mol): mg/m3 = ppm * MW / 24.45 */
    double mg_m3 = ppm * (cal->molar_mass_g_mol / 24.45);
    return cal->output_in_ug_m3 ? (mg_m3 * 1000.0) : mg_m3;
}

/* ==================== 5. DSP Acústico IEC 61672 Ponderación A (48 kHz) ==================== */

/* Coeficientes bilineales exactos IEC 61672 @ 48000 Hz (0.000 dB @ 1 kHz) */
static const double k_aweight_sos[3][5] = {
    { 0.997309039125, -1.994618078249, 0.997309039125, -1.994614452734, 0.994621703764 },
    { 0.947223501132, -1.894447002264, 0.947223501132, -1.893801875992, 0.895092128537 },
    { 0.321844675996,  0.643689351991, 0.321844675996,  0.025424315232, 0.000161598951 },
};

void aire_dba_init(aire_dba_ctx_t *ctx) {
    if (!ctx) {
        return;
    }
    memset(ctx, 0, sizeof(*ctx));
}

static double step_biquad(aire_biquad_t *s, const double c[5], double x) {
    double y = c[0] * x + c[1] * s->x1 + c[2] * s->x2 - c[3] * s->y1 - c[4] * s->y2;
    s->x2 = s->x1;
    s->x1 = x;
    s->y2 = s->y1;
    s->y1 = y;
    return y;
}

void aire_dba_process_i24(aire_dba_ctx_t *ctx, const int32_t *samples_24bit_signed, size_t count) {
    if (!ctx || !samples_24bit_signed || count == 0U) {
        return;
    }
    /* Normaliza rango completo 24 bits con signo a [-1.0, +1.0) */
    const double inv_fs = 1.0 / 8388608.0;
    for (size_t i = 0U; i < count; ++i) {
        double x = (double)samples_24bit_signed[i] * inv_fs;
        /* Bloqueador DC de primer orden (alpha = 0.995) */
        double dc = x - ctx->prev_x + 0.995 * ctx->prev_y;
        ctx->prev_x = x;
        ctx->prev_y = dc;

        double y = dc;
        for (int s = 0; s < 3; ++s) {
            y = step_biquad(&ctx->stages[s], k_aweight_sos[s], y);
        }
        ctx->energy_sum += y * y;
    }
    ctx->sample_count += (uint64_t)count;
}

bool aire_dba_compute(const aire_dba_ctx_t *ctx, double cal_offset_db, double *out_dba) {
    if (!ctx || !out_dba || ctx->sample_count == 0U) {
        return false;
    }
    double rms = sqrt(ctx->energy_sum / (double)ctx->sample_count);
    if (rms < 1e-9) {
        rms = 1e-9;
    }
    /* ICS-43434: 94.0 dB SPL senoidal = -26.0 dBFS respecto a RMS senoidal (1/sqrt(2)) => +120.0 dB */
    const double sine_rms_fs = 0.7071067811865475;
    double dbfs = 20.0 * log10(rms / sine_rms_fs);
    double dba  = dbfs + 120.0 + cal_offset_db;
    if (dba < 20.0) {
        dba = 20.0;
    } else if (dba > 130.0) {
        dba = 130.0;
    }
    *out_dba = dba;
    return true;
}

/* ==================== 6. Presupuesto Solar y Batería ==================== */

uint32_t aire_adc_divider_mv(uint32_t adc_pin_mv, float ratio) {
    return (uint32_t)llround((double)adc_pin_mv * (double)ratio);
}

aire_power_plan_t aire_power_evaluate_budget(uint32_t vbat_mv, uint32_t vsolar_mv, uint32_t base_interval_sec) {
    uint32_t interval = base_interval_sec < 60U ? 300U : base_interval_sec;
    aire_power_plan_t plan = {
        .interval_sec      = interval,
        .allow_5v_sensors  = true,
        .allow_lte_tx      = true,
        .critical_low_batt = false,
    };

    if (vbat_mv < 3250U) {
        /* Protege celda Li-Ion y evita brownout del boost 5V o pico TX SIM7080G */
        plan.critical_low_batt = true;
        plan.allow_5v_sensors  = false;
        plan.allow_lte_tx      = false;
        plan.interval_sec      = interval * 4U;
    } else if (vbat_mv < 3500U && vsolar_mv < 4500U) {
        /* Modo ahorro nocturno con batería media-baja: extiende intervalo y conserva I2C + SD */
        plan.interval_sec     = interval * 2U;
        plan.allow_5v_sensors = true;
        plan.allow_lte_tx     = true;
    }
    return plan;
}
