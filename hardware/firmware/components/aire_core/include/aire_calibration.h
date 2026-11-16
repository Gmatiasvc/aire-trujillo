#ifndef AIRE_CALIBRATION_H
#define AIRE_CALIBRATION_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* CRC-8 Sensirion (polinomio 0x31, init 0xFF) usado en SCD41, SGP40 y SPS30 */
uint8_t aire_crc8_sensirion(const uint8_t *data, size_t len);
bool aire_sensirion_check_words(const uint8_t *buf, size_t word_count);

/* Datos de calibración de fábrica Bosch BME280 (registros 0x88..0xA1 y 0xE1..0xE7) */
typedef struct {
    uint16_t dig_T1;
    int16_t  dig_T2;
    int16_t  dig_T3;
    uint16_t dig_P1;
    int16_t  dig_P2;
    int16_t  dig_P3;
    int16_t  dig_P4;
    int16_t  dig_P5;
    int16_t  dig_P6;
    int16_t  dig_P7;
    int16_t  dig_P8;
    int16_t  dig_P9;
    uint8_t  dig_H1;
    int16_t  dig_H2;
    uint8_t  dig_H3;
    int16_t  dig_H4;
    int16_t  dig_H5;
    int8_t   dig_H6;
} aire_bme280_calib_t;

bool aire_bme280_parse_calib(const uint8_t reg_88_a1[26], const uint8_t reg_e1_e7[7], aire_bme280_calib_t *out);
bool aire_bme280_compensate(const aire_bme280_calib_t *c,
                            int32_t adc_T, int32_t adc_P, int32_t adc_H,
                            double *out_temp_c, double *out_press_hpa, double *out_hum_pct);

/* Conversión SCD41 & SGP40 & SPS30 */
void aire_scd41_decode(const uint16_t words[3], double *out_co2_ppm, double *out_temp_c, double *out_rh_pct);
void aire_sgp40_encode_ticks(double temp_c, double rh_pct, uint8_t out_6bytes[6]);

typedef struct {
    double baseline_sraw;
    bool   initialized;
} aire_voc_state_t;

double aire_sgp40_calc_voc_index(aire_voc_state_t *state, uint16_t sraw_voc);

bool aire_sps30_decode_pm(const uint8_t raw_60bytes[60], double *out_pm25, double *out_pm10);

/* ADS1115 (PGA +-4.096V -> 0.125 mV/LSB) & Winsen ZE03 compensado */
double aire_ads1115_counts_to_mv(int16_t raw_counts);

typedef struct {
    double v_zero_mv;         /* Salida típica en aire limpio (ej. 400 mV) */
    double v_span_mv;         /* Excursión fondo de escala (ej. 1600 mV) */
    double full_scale_ppm;    /* Fondo de escala del módulo ZE03 (ppm) */
    double molar_mass_g_mol;  /* CO=28.01, NO2=46.01, SO2=64.07, H2S=34.08 */
    double corr_zero_mv;      /* Tensión nominal de referencia del canal libre CORR */
    double corr_gain;         /* Factor de sustracción de deriva del canal CORR (0..1) */
    double temp_coeff_per_c;  /* Corrección térmica relativa respecto a 25 C */
    bool   output_in_ug_m3;   /* false -> mg/m3 (CO), true -> ug/m3 (NO2/SO2/H2S) */
} aire_ze03_cal_t;

aire_ze03_cal_t aire_ze03_default_co(void);
aire_ze03_cal_t aire_ze03_default_no2(void);
aire_ze03_cal_t aire_ze03_default_so2(void);
aire_ze03_cal_t aire_ze03_default_h2s(void);

/* Convierte canales VO + CORR (mV) + temperatura ambiente a mg/m3 o ug/m3 canónicos */
double aire_ze03_compensate(const aire_ze03_cal_t *cal, double v_vo_mv, double v_corr_mv, double temp_c);

/* DSP Acústico I2S InvenSense ICS-43434 (24-bit signed @ 48 kHz -> Ponderación A -> dBA) */
typedef struct {
    double x1, x2, y1, y2;
} aire_biquad_t;

typedef struct {
    aire_biquad_t stages[4]; /* HP 20Hz + A-weight bilineal 48 kHz (3 etapas + DC blocker) */
    double        prev_x;
    double        prev_y;
    double        energy_sum;
    uint64_t      sample_count;
} aire_dba_ctx_t;

void aire_dba_init(aire_dba_ctx_t *ctx);
void aire_dba_process_i24(aire_dba_ctx_t *ctx, const int32_t *samples_24bit_signed, size_t count);
bool aire_dba_compute(const aire_dba_ctx_t *ctx, double cal_offset_db, double *out_dba);

/* Presupuesto Energético Solar / Li-Ion 1S */
typedef struct {
    uint32_t interval_sec;
    bool     allow_5v_sensors;
    bool     allow_lte_tx;
    bool     critical_low_batt;
} aire_power_plan_t;

uint32_t aire_adc_divider_mv(uint32_t adc_pin_mv, float ratio);
aire_power_plan_t aire_power_evaluate_budget(uint32_t vbat_mv, uint32_t vsolar_mv, uint32_t base_interval_sec);

#ifdef __cplusplus
}
#endif

#endif /* AIRE_CALIBRATION_H */
