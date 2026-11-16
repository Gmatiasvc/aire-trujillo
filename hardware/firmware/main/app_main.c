#include "sdkconfig.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/time.h>
#include <time.h>

#include "esp_log.h"
#include "esp_sleep.h"
#include "nvs_flash.h"

#include "aire_board.h"
#include "aire_calibration.h"
#include "aire_modem_at.h"
#include "aire_queue.h"
#include "aire_telemetry.h"
#include "bus_hal.h"
#include "sensors_hal.h"

static const char *TAG = "aire_main";

RTC_DATA_ATTR static uint32_t         s_boot_count      = 0U;
RTC_DATA_ATTR static aire_voc_state_t s_voc_state       = { .baseline_sraw = 0.0, .initialized = false };
RTC_DATA_ATTR static bool             s_rtc_time_synced = false;

/* Algoritmo civil inverso a epoch POSIX UTC (independiente de locale y zona horaria TZ) */
static int64_t civil_to_days(int y, unsigned m, unsigned d) {
    y -= (m <= 2U) ? 1 : 0;
    int era = (y >= 0 ? y : y - 399) / 400;
    unsigned yoe = (unsigned)(y - era * 400);
    unsigned doy = (153U * (m + (m > 2U ? (unsigned)-3 : 9U)) + 2U) / 5U + d - 1U;
    unsigned doe = yoe * 365U + yoe / 4U - yoe / 100U + doy;
    return (int64_t)era * 146097LL + (int64_t)doe - 719468LL;
}

static bool iso_utc_to_epoch(const char *iso_utc, time_t *out_epoch) {
    if (!iso_utc || !out_epoch) {
        return false;
    }
    int yr = 0, mo = 0, dy = 0, hh = 0, mm = 0, ss = 0;
    if (sscanf(iso_utc, "%4d-%2d-%2dT%2d:%2d:%2dZ", &yr, &mo, &dy, &hh, &mm, &ss) != 6) {
        return false;
    }
    if (yr < 2024 || mo < 1 || mo > 12 || dy < 1 || dy > 31) {
        return false;
    }
    int64_t days = civil_to_days(yr, (unsigned)mo, (unsigned)dy);
    *out_epoch = (time_t)(days * 86400LL + (int64_t)hh * 3600LL + (int64_t)mm * 60LL + (int64_t)ss);
    return true;
}

static bool format_current_rtc_utc(char *out_iso, size_t cap) {
    if (!out_iso || cap < 21U || !s_rtc_time_synced) {
        return false;
    }
    time_t now = time(NULL);
    if (now < 1735689600LL) { /* < 2025-01-01T00:00:00Z -> reloj no válido */
        return false;
    }
    struct tm tm_utc;
    if (!gmtime_r(&now, &tm_utc)) {
        return false;
    }
    int n = snprintf(out_iso, cap, "%04d-%02d-%02dT%02d:%02d:%02dZ",
                     tm_utc.tm_year + 1900, tm_utc.tm_mon + 1, tm_utc.tm_mday,
                     tm_utc.tm_hour, tm_utc.tm_min, tm_utc.tm_sec);
    return n == 20;
}

static void sync_system_rtc_from_iso(const char *iso_utc) {
    time_t epoch = 0;
    if (iso_utc_to_epoch(iso_utc, &epoch)) {
        struct timeval tv = { .tv_sec = epoch, .tv_usec = 0 };
        settimeofday(&tv, NULL);
        s_rtc_time_synced = true;
    }
}

void app_main(void) {
    esp_err_t nvs_err = nvs_flash_init();
    if (nvs_err == ESP_ERR_NVS_NO_FREE_PAGES || nvs_err == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        ESP_ERROR_CHECK(nvs_flash_erase());
        nvs_err = nvs_flash_init();
    }
    ESP_ERROR_CHECK(nvs_err);
    ESP_ERROR_CHECK(bus_hal_power_init());

    s_boot_count++;
    bus_hal_set_led(true);

    uint32_t vbat_mv = 0U, vsolar_mv = 0U;
    bool chg_active = false, chg_done = false;
    (void)bus_hal_adc_read_power_mv(&vbat_mv, &vsolar_mv);
    bus_hal_read_charger_status(&chg_active, &chg_done);

    const aire_power_plan_t plan = aire_power_evaluate_budget(
        vbat_mv, vsolar_mv, (uint32_t)CONFIG_AIRE_SAMPLE_INTERVAL_SEC);

    ESP_LOGI(TAG, "Ciclo #%u | VBAT=%u mV | VSOL=%u mV | CHG=%d DONE=%d | cadencia=%u s | wake=%d",
             (unsigned)s_boot_count, (unsigned)vbat_mv, (unsigned)vsolar_mv,
             chg_active ? 1 : 0, chg_done ? 1 : 0, (unsigned)plan.interval_sec,
             (int)esp_sleep_get_wakeup_cause());

    if (plan.critical_low_batt) {
        ESP_LOGW(TAG, "Batería crítica (%u mV < 3250 mV); entrando a deep-sleep protector", (unsigned)vbat_mv);
        bus_hal_prepare_deep_sleep((uint64_t)plan.interval_sec * 1000000ULL);
    }

    /* 1. Adquisición de sensores en rieles conmutados 3V3_SENS y 5V_SENS */
    bus_hal_set_3v3_sens(true);
    if (bus_hal_i2c_init() != ESP_OK) {
        ESP_LOGE(TAG, "Error inicializando I2C principal");
    }

    aire_telemetry_t tel;
    aire_telemetry_init(&tel, CONFIG_AIRE_NODE_CODE, "");

    const double mic_cal_db = strtod(CONFIG_AIRE_MIC_CAL_OFFSET_DB, NULL);
    sensors_hal_acquire_all(&tel,
                            &s_voc_state,
                            plan.allow_5v_sensors,
                            (uint32_t)CONFIG_AIRE_WARMUP_5V_MS,
                            mic_cal_db);

    const size_t valid_metrics = aire_telemetry_valid_count(&tel);
    ESP_LOGI(TAG, "Métricas válidas adquiridas: %u/12", (unsigned)valid_metrics);

    /* 2. Almacenamiento local store-and-forward + Enlace LTE-M / GNSS / MQTT QoS 1 */
    const bool sd_ok = bus_hal_sdcard_mount("/sdcard");
    aire_queue_t queue;
    if (sd_ok) {
        aire_queue_init(&queue, "/sdcard", 8U * 1024U * 1024U);
    }

    bool modem_on = false;
    bool lte_up   = false;
    const aire_modem_hal_t modem_hal = bus_hal_get_modem_ops();
    const aire_modem_cfg_t modem_cfg = {
        .apn            = CONFIG_AIRE_CELLULAR_APN,
        .mqtt_host      = CONFIG_AIRE_MQTT_BROKER_HOST,
        .mqtt_port      = (uint16_t)CONFIG_AIRE_MQTT_BROKER_PORT,
        .mqtt_client_id = CONFIG_AIRE_NODE_CODE,
        .mqtt_username  = CONFIG_AIRE_MQTT_USERNAME,
        .mqtt_password  = CONFIG_AIRE_MQTT_PASSWORD,
        .net_timeout_ms = 45000U,
    };

    if (plan.allow_lte_tx && bus_hal_modem_uart_init() == ESP_OK) {
        modem_on = aire_modem_power_on(&modem_hal);
        if (modem_on) {
            lte_up = aire_modem_connect_ltem(&modem_hal, &modem_cfg);
            char utc_iso[32] = { 0 };
            if (aire_modem_sync_utc(&modem_hal, utc_iso, sizeof(utc_iso))) {
                snprintf(tel.recorded_at_utc, sizeof(tel.recorded_at_utc), "%s", utc_iso);
                sync_system_rtc_from_iso(utc_iso);
            }
        }
    }

    if (tel.recorded_at_utc[0] == '\0') {
        if (!format_current_rtc_utc(tel.recorded_at_utc, sizeof(tel.recorded_at_utc))) {
            ESP_LOGW(TAG, "Sin hora UTC válida (GNSS/NITZ/RTC); muestra descartada para proteger IDW");
        }
    }

    char topic[AIRE_QUEUE_MAX_TOPIC_LEN];
    char payload[AIRE_QUEUE_MAX_PAYLOAD_LEN];
    bool enqueued = false;

    if (valid_metrics > 0U && tel.recorded_at_utc[0] != '\0') {
        size_t tlen = aire_telemetry_format_topic(CONFIG_AIRE_NODE_CODE, topic, sizeof(topic));
        size_t jlen = aire_telemetry_serialize_json(&tel, payload, sizeof(payload));
        if (tlen > 0U && jlen > 0U) {
            ESP_LOGI(TAG, "Payload %s -> %s", topic, payload);
            if (sd_ok) {
                enqueued = aire_queue_push(&queue, topic, payload);
            }
        }
    }

    if (lte_up && aire_modem_mqtt_open(&modem_hal, &modem_cfg)) {
        if (sd_ok) {
            char q_topic[AIRE_QUEUE_MAX_TOPIC_LEN];
            char q_json[AIRE_QUEUE_MAX_PAYLOAD_LEN];
            size_t next_off = 0U;
            int drained = 0;
            /* Drena hasta 32 tramas por ciclo para equilibrar recuperación vs. energía */
            while (drained < 32 &&
                   aire_queue_peek(&queue, q_topic, sizeof(q_topic), q_json, sizeof(q_json), &next_off)) {
                if (!aire_modem_mqtt_publish_qos1(&modem_hal, q_topic, q_json)) {
                    ESP_LOGW(TAG, "Fallo en +SMPUB QoS1 tras drenar %d mensajes; conserva backlog SD", drained);
                    break;
                }
                (void)aire_queue_commit_pop(&queue, next_off);
                drained++;
            }
            ESP_LOGI(TAG, "Backlog drenado=%d | Pendientes SD=%u",
                     drained, (unsigned)aire_queue_pending_count(&queue));
        } else if (!enqueued && valid_metrics > 0U && tel.recorded_at_utc[0] != '\0') {
            /* Sin MicroSD montada: intenta envío directo de la lectura viva actual */
            (void)aire_modem_mqtt_publish_qos1(&modem_hal, topic, payload);
        }
        aire_modem_mqtt_close(&modem_hal);
    }

    if (modem_on) {
        (void)aire_modem_power_off(&modem_hal);
    }
    if (sd_ok) {
        bus_hal_sdcard_unmount("/sdcard");
    }

    ESP_LOGI(TAG, "Entrando a deep-sleep por %u s", (unsigned)plan.interval_sec);
    bus_hal_prepare_deep_sleep((uint64_t)plan.interval_sec * 1000000ULL);
}
