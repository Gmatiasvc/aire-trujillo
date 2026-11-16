#include "aire_calibration.h"
#include "aire_modem_at.h"
#include "aire_queue.h"
#include "aire_telemetry.h"

#include <assert.h>
#include <math.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <unistd.h>

static const double k_pi = 3.14159265358979323846;

#define EXPECT_TRUE(cond)                                                              \
    do {                                                                               \
        if (!(cond)) {                                                                 \
            fprintf(stderr, "FALLA EXPECT_TRUE en %s:%d: %s\n", __FILE__, __LINE__,    \
                    #cond);                                                            \
            exit(1);                                                                   \
        }                                                                              \
    } while (0)

#define EXPECT_NEAR(actual, expected, tol)                                             \
    do {                                                                               \
        double _a = (actual);                                                          \
        double _e = (expected);                                                        \
        double _t = (tol);                                                             \
        if (!isfinite(_a) || fabs(_a - _e) > _t) {                                     \
            fprintf(stderr,                                                            \
                    "FALLA EXPECT_NEAR en %s:%d: got %.6f, want %.6f +- %.6f\n",      \
                    __FILE__, __LINE__, _a, _e, _t);                                   \
            exit(1);                                                                   \
        }                                                                              \
    } while (0)

/* ==================== 1. Contrato Telemetría Zero-Drift ==================== */

static void build_sample_telemetry(aire_telemetry_t *t) {
    aire_telemetry_init(t, "TRJ-HUANCHACO-01", "2026-11-15T22:10:00Z");
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_PM25, 18.42));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_PM10, 31.75));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_CO, 0.845));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_NO2, 42.10));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_SO2, 11.20));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_H2S, 3.40));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_CO2, 512.0));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_VOC_INDEX, 118.0));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_NOISE_DBA, 63.8));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_TEMPERATURE_C, 21.35));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_HUMIDITY_PCT, 78.90));
    EXPECT_TRUE(aire_telemetry_set_metric(t, AIRE_METRIC_PRESSURE_HPA, 1013.45));
}

static void test_telemetry_contract(void) {
    char topic[128];
    EXPECT_TRUE(aire_telemetry_format_topic("TRJ-CENTRO-02", topic, sizeof(topic)) > 0U);
    EXPECT_TRUE(strcmp(topic, "aire-trujillo/v1/nodes/TRJ-CENTRO-02/telemetry") == 0);
    /* Rechaza comodines o separadores en node_code */
    EXPECT_TRUE(aire_telemetry_format_topic("TRJ/+#", topic, sizeof(topic)) == 0U);
    EXPECT_TRUE(aire_telemetry_format_topic("", topic, sizeof(topic)) == 0U);

    aire_telemetry_t full;
    build_sample_telemetry(&full);
    EXPECT_TRUE(aire_telemetry_valid_count(&full) == 12U);

    char json[512];
    size_t jlen = aire_telemetry_serialize_json(&full, json, sizeof(json));
    EXPECT_TRUE(jlen > 0U);
    EXPECT_TRUE(strstr(json, "\"node_code\":\"TRJ-HUANCHACO-01\"") != NULL);
    EXPECT_TRUE(strstr(json, "\"recorded_at\":\"2026-11-15T22:10:00Z\"") != NULL);
    EXPECT_TRUE(strstr(json, "\"pm25\":18.42") != NULL);
    EXPECT_TRUE(strstr(json, "\"pressure_hpa\":1013.45") != NULL);

    /* Zero-Drift: omite métricas ausentes o fuera de rango en lugar de emitir 0 falso */
    aire_telemetry_t partial;
    aire_telemetry_init(&partial, "TRJ-TEST-01", "2026-11-15T22:15:00Z");
    EXPECT_TRUE(!aire_telemetry_set_metric(&partial, AIRE_METRIC_CO2, 150.0));  /* < 400 ppm */
    EXPECT_TRUE(!aire_telemetry_set_metric(&partial, AIRE_METRIC_PM25, -1.0));  /* < 0 */
    EXPECT_TRUE(!aire_telemetry_set_metric(&partial, AIRE_METRIC_NOISE_DBA, NAN));
    /* Sin métricas válidas rechaza serialización */
    EXPECT_TRUE(aire_telemetry_serialize_json(&partial, json, sizeof(json)) == 0U);

    EXPECT_TRUE(aire_telemetry_set_metric(&partial, AIRE_METRIC_TEMPERATURE_C, 22.5));
    jlen = aire_telemetry_serialize_json(&partial, json, sizeof(json));
    EXPECT_TRUE(jlen > 0U);
    EXPECT_TRUE(strcmp(json,
                       "{\"node_code\":\"TRJ-TEST-01\",\"recorded_at\":\"2026-11-15T22:15:00Z\","
                       "\"metrics\":{\"temperature_c\":22.50}}") == 0);
}

/* ==================== 2. CRC8, BME280, SCD41, SGP40, SPS30 ==================== */

static void test_digital_sensors(void) {
    /* Vector estándar hoja de datos Sensirion: 0xBE 0xEF -> CRC 0x92 */
    const uint8_t sensirion_vec[2] = { 0xBEU, 0xEFU };
    EXPECT_TRUE(aire_crc8_sensirion(sensirion_vec, 2U) == 0x92U);

    /* Vector oficial Bosch BME280 Datasheet Sección 4.2.3 (temp ~25.08 C, press ~1006.53 hPa) */
    const uint8_t reg_88_a1[26] = {
        0x70, 0x6B, /* T1 = 27504 */
        0x43, 0x67, /* T2 = 26435 */
        0x18, 0xFC, /* T3 = -1000 */
        0x7D, 0x8E, /* P1 = 36477 */
        0x43, 0xD6, /* P2 = -10685 */
        0xD0, 0x0B, /* P3 = 3024 */
        0x27, 0x0B, /* P4 = 2855 */
        0x8C, 0x00, /* P5 = 140 */
        0xF9, 0xFF, /* P6 = -7 */
        0x8C, 0x3C, /* P7 = 15500 */
        0xF8, 0xC6, /* P8 = -14600 */
        0x70, 0x17, /* P9 = 6000 */
        0x00, 75    /* res, H1 = 75 */
    };
    const uint8_t reg_e1_e7[7] = {
        0x6A, 0x01, /* H2 = 362 */
        0x00,       /* H3 = 0 */
        0x12, 0x05, /* H4 = 293, H5 = 48? */
        0x03,       /* H5 MSB */
        0x1E        /* H6 = 30 */
    };
    aire_bme280_calib_t cal;
    EXPECT_TRUE(aire_bme280_parse_calib(reg_88_a1, reg_e1_e7, &cal));
    double temp = 0.0, press = 0.0, hum = 0.0;
    EXPECT_TRUE(aire_bme280_compensate(&cal, 519888, 415148, 29000, &temp, &press, &hum));
    EXPECT_NEAR(temp, 25.08, 0.02);
    EXPECT_NEAR(press, 1006.53, 0.05);
    EXPECT_TRUE(hum >= 0.0 && hum <= 100.0);

    /* SCD41 decode */
    const uint16_t scd_words[3] = { 620U, 25132U, 45875U }; /* CO2=620, T~22.1 C, RH~70.0 % */
    double co2 = 0.0, scd_t = 0.0, scd_rh = 0.0;
    aire_scd41_decode(scd_words, &co2, &scd_t, &scd_rh);
    EXPECT_NEAR(co2, 620.0, 0.01);
    EXPECT_NEAR(scd_t, 22.11, 0.1);
    EXPECT_NEAR(scd_rh, 70.0, 0.1);

    /* SGP40 ticks + CRC + Índice VOC centrado en 100 */
    uint8_t sgp_ticks[6];
    aire_sgp40_encode_ticks(25.0, 50.0, sgp_ticks);
    EXPECT_TRUE(aire_sensirion_check_words(sgp_ticks, 2U));
    aire_voc_state_t voc_state = { 0 };
    EXPECT_NEAR(aire_sgp40_calc_voc_index(&voc_state, 30000U), 100.0, 0.1);
    EXPECT_NEAR(aire_sgp40_calc_voc_index(&voc_state, 30000U), 100.0, 0.1);
    EXPECT_TRUE(aire_sgp40_calc_voc_index(&voc_state, 28500U) > 180.0); /* Baja sraw => sube VOC */
    EXPECT_TRUE(aire_sgp40_calc_voc_index(&voc_state, 31500U) < 100.0); /* Sube sraw => aire más limpio */

    /* SPS30 payload IEEE-754 big-endian de 60 bytes (20 palabras con CRC-8) */
    uint8_t sps_raw[60];
    memset(sps_raw, 0, sizeof(sps_raw));
    /* Palabra 2..3 (offset 6): PM2.5 = 24.5f = 0x41C40000; Palabra 6..7 (offset 18): PM10 = 42.25f = 0x42290000 */
    sps_raw[6] = 0x41U; sps_raw[7] = 0xC4U; sps_raw[9]  = 0x00U; sps_raw[10] = 0x00U;
    sps_raw[18] = 0x42U; sps_raw[19] = 0x29U; sps_raw[21] = 0x00U; sps_raw[22] = 0x00U;
    for (size_t w = 0U; w < 20U; ++w) {
        sps_raw[w * 3U + 2U] = aire_crc8_sensirion(&sps_raw[w * 3U], 2U);
    }
    double pm25 = 0.0, pm10 = 0.0;
    EXPECT_TRUE(aire_sps30_decode_pm(sps_raw, &pm25, &pm10));
    EXPECT_NEAR(pm25, 24.5, 0.01);
    EXPECT_NEAR(pm10, 42.25, 0.01);

    /* Corrupción de 1 bit en CRC8 debe invalidar toda la trama */
    sps_raw[8] ^= 0x01U;
    EXPECT_TRUE(!aire_sps30_decode_pm(sps_raw, &pm25, &pm10));
}

/* ==================== 3. ADS1115 + 4x Winsen ZE03 ==================== */

static void test_ads1115_and_ze03(void) {
    /* 0.125 mV / LSB con PGA +-4.096 V */
    EXPECT_NEAR(aire_ads1115_counts_to_mv(3200), 400.0, 0.001);
    EXPECT_NEAR(aire_ads1115_counts_to_mv(9600), 1200.0, 0.001);

    /* CO @ 25 C: VO=560 mV, CORR=400 mV (sin deriva) => net=160 mV => 10 ppm => 10*28.01/24.45 = 11.456 mg/m3 */
    aire_ze03_cal_t co = aire_ze03_default_co();
    double co_mg = aire_ze03_compensate(&co, 560.0, 400.0, 25.0);
    EXPECT_NEAR(co_mg, 11.456, 0.02);

    /* Con deriva compartida de +30 mV en VO y CORR debe cancelar la deriva exactamente */
    double co_drift_comp = aire_ze03_compensate(&co, 590.0, 430.0, 25.0);
    EXPECT_NEAR(co_drift_comp, co_mg, 0.01);

    /* NO2 @ 25 C: VO=480 mV, CORR=400 mV => net=80 mV => 0.1 ppm => 0.1*46.01/24.45*1000 = 188.18 ug/m3 */
    aire_ze03_cal_t no2 = aire_ze03_default_no2();
    EXPECT_NEAR(aire_ze03_compensate(&no2, 480.0, 400.0, 25.0), 188.18, 0.5);
}

/* ==================== 4. DSP Acústico IEC 61672 @ 48 kHz ==================== */

static double synth_sine_dba(double freq_hz, double amp_fs, double dc_bias) {
    aire_dba_ctx_t ctx;
    aire_dba_init(&ctx);

    const size_t total = 48000U; /* 1 s @ 48 kHz */
    int32_t buf[512];
    size_t idx = 0U;
    while (idx < total) {
        size_t n = (total - idx > 512U) ? 512U : (total - idx);
        for (size_t i = 0U; i < n; ++i) {
            double t = (double)(idx + i) / 48000.0;
            double s = amp_fs * sin(2.0 * k_pi * freq_hz * t) + dc_bias;
            buf[i] = (int32_t)llround(s * 8388608.0);
        }
        aire_dba_process_i24(&ctx, buf, n);
        idx += n;
    }

    double dba = 0.0;
    EXPECT_TRUE(aire_dba_compute(&ctx, 0.0, &dba));
    return dba;
}

static void test_acoustic_a_weighting(void) {
    /* Referencia ICS-43434: 1 kHz a -26 dBFS (amp = 10^(-26/20) = 0.0501187) = 94.0 dBA */
    double dba_1k = synth_sine_dba(1000.0, pow(10.0, -26.0 / 20.0), 0.0);
    EXPECT_NEAR(dba_1k, 94.0, 0.35);

    /* Con fuerte offset DC simulado del MEMS debe conservar ~94.0 dBA gracias al DC blocker */
    double dba_dc = synth_sine_dba(1000.0, pow(10.0, -26.0 / 20.0), 0.08);
    EXPECT_NEAR(dba_dc, 94.0, 0.5);

    /* Ponderación A @ 100 Hz debe atenuar ~19.1 dB respecto a 1 kHz */
    double dba_100 = synth_sine_dba(100.0, pow(10.0, -26.0 / 20.0), 0.0);
    EXPECT_NEAR(dba_100, 94.0 - 19.17, 0.7);
}

/* ==================== 5. Planificador Solar / Li-Ion ==================== */

static void test_power_planner(void) {
    aire_power_plan_t crit = aire_power_evaluate_budget(3180U, 0U, 300U);
    EXPECT_TRUE(crit.critical_low_batt);
    EXPECT_TRUE(!crit.allow_5v_sensors);
    EXPECT_TRUE(!crit.allow_lte_tx);
    EXPECT_TRUE(crit.interval_sec == 1200U);

    aire_power_plan_t night = aire_power_evaluate_budget(3420U, 100U, 300U);
    EXPECT_TRUE(!night.critical_low_batt);
    EXPECT_TRUE(night.allow_5v_sensors);
    EXPECT_TRUE(night.allow_lte_tx);
    EXPECT_TRUE(night.interval_sec == 600U);

    aire_power_plan_t nom = aire_power_evaluate_budget(3950U, 5800U, 300U);
    EXPECT_TRUE(!nom.critical_low_batt);
    EXPECT_TRUE(nom.interval_sec == 300U);
}

/* ==================== 6. Cola FIFO Store-and-Forward MicroSD ==================== */

static void test_sd_queue(void) {
    char dir[128];
    snprintf(dir, sizeof(dir), "/tmp/aire_fw_queue_%ld", (long)getpid());
    mkdir(dir, 0755);

    char logf[256], idxf[256];
    snprintf(logf, sizeof(logf), "%s/queue.log", dir);
    snprintf(idxf, sizeof(idxf), "%s/queue.idx", dir);
    remove(logf);
    remove(idxf);

    aire_queue_t q;
    aire_queue_init(&q, dir, 4096U);

    EXPECT_TRUE(aire_queue_push(&q, "aire-trujillo/v1/nodes/N1/telemetry", "{\"id\":1}"));
    EXPECT_TRUE(aire_queue_push(&q, "aire-trujillo/v1/nodes/N1/telemetry", "{\"id\":2}"));
    /* Simula apagón a mitad de escritura sin salto de línea/CRC válido */
    FILE *corrupt = fopen(logf, "ab");
    EXPECT_TRUE(corrupt != NULL);
    fputs("|deadbeef|TRUNCA_CORRUPTA\n", corrupt);
    fclose(corrupt);
    EXPECT_TRUE(aire_queue_push(&q, "aire-trujillo/v1/nodes/N1/telemetry", "{\"id\":3}"));

    EXPECT_TRUE(aire_queue_pending_count(&q) == 3U);

    char tbuf[128], jbuf[512];
    size_t next_off = 0U;
    EXPECT_TRUE(aire_queue_peek(&q, tbuf, sizeof(tbuf), jbuf, sizeof(jbuf), &next_off));
    EXPECT_TRUE(strcmp(jbuf, "{\"id\":1}") == 0);
    EXPECT_TRUE(aire_queue_commit_pop(&q, next_off));

    EXPECT_TRUE(aire_queue_peek(&q, tbuf, sizeof(tbuf), jbuf, sizeof(jbuf), &next_off));
    EXPECT_TRUE(strcmp(jbuf, "{\"id\":2}") == 0);
    EXPECT_TRUE(aire_queue_commit_pop(&q, next_off));

    /* Debe saltar automáticamente la línea corrupta y recuperar id=3 intacto */
    EXPECT_TRUE(aire_queue_peek(&q, tbuf, sizeof(tbuf), jbuf, sizeof(jbuf), &next_off));
    EXPECT_TRUE(strcmp(jbuf, "{\"id\":3}") == 0);
    EXPECT_TRUE(aire_queue_commit_pop(&q, next_off));
    EXPECT_TRUE(aire_queue_pending_count(&q) == 0U);

    /* Compactación en offline acotado por max_bytes */
    aire_queue_t bounded;
    aire_queue_init(&bounded, dir, 220U);
    for (int i = 1; i <= 6; ++i) {
        char payload[64];
        snprintf(payload, sizeof(payload), "{\"seq\":%d}", i);
        EXPECT_TRUE(aire_queue_push(&bounded, "aire-trujillo/v1/nodes/N1/telemetry", payload));
    }
    EXPECT_TRUE(aire_queue_peek(&bounded, tbuf, sizeof(tbuf), jbuf, sizeof(jbuf), &next_off));
    EXPECT_TRUE(strcmp(jbuf, "{\"seq\":4}") == 0 || strcmp(jbuf, "{\"seq\":5}") == 0 ||
                strcmp(jbuf, "{\"seq\":6}") == 0);

    FILE *szf = fopen(logf, "rb");
    EXPECT_TRUE(szf != NULL);
    fseek(szf, 0L, SEEK_END);
    long final_sz = ftell(szf);
    fclose(szf);
    EXPECT_TRUE(final_sz <= 220L);

    remove(logf);
    remove(idxf);
    rmdir(dir);
}

/* ==================== 7. Máquina de Estados AT SIM7080G ==================== */

typedef struct {
    char   tx_buf[1024];
    size_t tx_len;
    char   rx_buf[1024];
    size_t rx_pos;
    size_t rx_len;
    bool   awaiting_pub_body;
    bool   powered;
    char   last_published_json[1024];
} mock_modem_t;

static void mock_feed(mock_modem_t *m, const char *s) {
    size_t l = strlen(s);
    if (m->rx_len + l < sizeof(m->rx_buf)) {
        memcpy(m->rx_buf + m->rx_len, s, l);
        m->rx_len += l;
    }
}

static void mock_delay(uint32_t ms, void *ctx) { (void)ms; (void)ctx; }
static void mock_pwrkey(bool active, void *ctx) {
    mock_modem_t *m = (mock_modem_t *)ctx;
    if (active) {
        m->powered = true;
    }
}
static void mock_dtr(bool sleep_hi, void *ctx) { (void)sleep_hi; (void)ctx; }
static bool mock_status(void *ctx) { return ((mock_modem_t *)ctx)->powered; }
static void mock_flush(void *ctx) {
    mock_modem_t *m = (mock_modem_t *)ctx;
    m->rx_pos = 0U;
    m->rx_len = 0U;
}
static int mock_read(uint32_t timeout_ms, void *ctx) {
    (void)timeout_ms;
    mock_modem_t *m = (mock_modem_t *)ctx;
    if (m->rx_pos >= m->rx_len) {
        return -1;
    }
    return (unsigned char)m->rx_buf[m->rx_pos++];
}

static size_t mock_write(const uint8_t *data, size_t len, void *ctx) {
    mock_modem_t *m = (mock_modem_t *)ctx;
    for (size_t i = 0U; i < len; ++i) {
        if (m->tx_len + 1U < sizeof(m->tx_buf)) {
            m->tx_buf[m->tx_len++] = (char)data[i];
            m->tx_buf[m->tx_len]   = '\0';
        }
    }

    if (m->awaiting_pub_body && m->tx_buf[0] == '{') {
        snprintf(m->last_published_json, sizeof(m->last_published_json), "%s", m->tx_buf);
        m->awaiting_pub_body = false;
        m->tx_len = 0U;
        m->tx_buf[0] = '\0';
        mock_feed(m, "\r\nOK\r\n");
        return len;
    }

    if (strstr(m->tx_buf, "\r\n") != NULL) {
        if (strstr(m->tx_buf, "AT+CPIN?") != NULL) {
            mock_feed(m, "\r\n+CPIN: READY\r\nOK\r\n");
        } else if (strstr(m->tx_buf, "AT+CGATT?") != NULL) {
            mock_feed(m, "\r\n+CGATT: 1\r\nOK\r\n");
        } else if (strstr(m->tx_buf, "AT+CGNSINF") != NULL) {
            /* Simula GNSS sin fix para ejercitar fallback a NITZ +CCLK */
            mock_feed(m, "\r\n+CGNSINF: 1,0,,,,,,,,,,,,,,,,,,,\r\nOK\r\n");
        } else if (strstr(m->tx_buf, "AT+CCLK?") != NULL) {
            /* 21:45:10 hora Perú (-20 = UTC-5) => 02:45:10 del día siguiente en UTC */
            mock_feed(m, "\r\n+CCLK: \"26/11/15,21:45:10-20\"\r\nOK\r\n");
        } else if (strstr(m->tx_buf, "AT+SMPUB=") != NULL) {
            m->awaiting_pub_body = true;
            mock_feed(m, "\r\n> ");
        } else if (strstr(m->tx_buf, "AT+CPOWD=1") != NULL) {
            m->powered = false;
            mock_feed(m, "\r\nNORMAL POWER DOWN\r\n");
        } else {
            mock_feed(m, "\r\nOK\r\n");
        }
        m->tx_len = 0U;
        m->tx_buf[0] = '\0';
    }
    return len;
}

static void test_modem_state_machine(void) {
    char iso[32];
    /* Pruebas puras de parseo GNSS y CCLK con cruce de medianoche y año bisiesto */
    EXPECT_TRUE(aire_modem_parse_cgnsinf_utc("+CGNSINF: 1,1,20261115235959.000,-8.111,-79.028", iso, sizeof(iso)));
    EXPECT_TRUE(strcmp(iso, "2026-11-15T23:59:59Z") == 0);
    EXPECT_TRUE(!aire_modem_parse_cgnsinf_utc("+CGNSINF: 1,0,,,,", iso, sizeof(iso)));

    EXPECT_TRUE(aire_modem_parse_cclk_utc("+CCLK: \"28/02/29,23:30:00-20\"", iso, sizeof(iso)));
    EXPECT_TRUE(strcmp(iso, "2028-03-01T04:30:00Z") == 0);

    mock_modem_t mock;
    memset(&mock, 0, sizeof(mock));
    const aire_modem_hal_t hal = {
        .delay_ms         = mock_delay,
        .set_pwrkey_level = mock_pwrkey,
        .set_dtr_sleep    = mock_dtr,
        .read_status_pin  = mock_status,
        .uart_flush_rx    = mock_flush,
        .uart_write       = mock_write,
        .uart_read_byte   = mock_read,
        .ctx              = &mock,
    };
    const aire_modem_cfg_t cfg = {
        .apn            = "movistar.pe",
        .mqtt_host      = "mqtt.aire-trujillo.local",
        .mqtt_port      = 1883U,
        .mqtt_client_id = "TRJ-HUANCHACO-01",
        .mqtt_username  = "",
        .mqtt_password  = "",
        .net_timeout_ms = 5000U,
    };

    EXPECT_TRUE(aire_modem_power_on(&hal));
    EXPECT_TRUE(mock.powered);
    EXPECT_TRUE(aire_modem_connect_ltem(&hal, &cfg));
    EXPECT_TRUE(aire_modem_sync_utc(&hal, iso, sizeof(iso)));
    EXPECT_TRUE(strcmp(iso, "2026-11-16T02:45:10Z") == 0);

    aire_telemetry_t tel;
    build_sample_telemetry(&tel);
    snprintf(tel.recorded_at_utc, sizeof(tel.recorded_at_utc), "%s", iso);

    char topic[128], json[512];
    EXPECT_TRUE(aire_telemetry_format_topic(tel.node_code, topic, sizeof(topic)) > 0U);
    EXPECT_TRUE(aire_telemetry_serialize_json(&tel, json, sizeof(json)) > 0U);

    EXPECT_TRUE(aire_modem_mqtt_open(&hal, &cfg));
    EXPECT_TRUE(aire_modem_mqtt_publish_qos1(&hal, topic, json));
    EXPECT_TRUE(strcmp(mock.last_published_json, json) == 0);
    aire_modem_mqtt_close(&hal);
    EXPECT_TRUE(aire_modem_power_off(&hal));
    EXPECT_TRUE(!mock.powered);
}

int main(int argc, char **argv) {
    if (argc > 1 && strcmp(argv[1], "--emit-json") == 0) {
        aire_telemetry_t tel;
        build_sample_telemetry(&tel);
        char json[512];
        size_t n = aire_telemetry_serialize_json(&tel, json, sizeof(json));
        if (n == 0U) {
            return 1;
        }
        printf("%s\n", json);
        return 0;
    }

    test_telemetry_contract();
    test_digital_sensors();
    test_ads1115_and_ze03();
    test_acoustic_a_weighting();
    test_power_planner();
    test_sd_queue();
    test_modem_state_machine();
    printf("ALL AIRE-TRUJILLO FIRMWARE HOST TESTS PASSED\n");
    return 0;
}
