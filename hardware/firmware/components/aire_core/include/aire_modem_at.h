#ifndef AIRE_MODEM_AT_H
#define AIRE_MODEM_AT_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Máquina de estados AT / LTE-M / GNSS / MQTT para SIMCom SIM7080G (U8).
 * Desacoplada de hardware mediante callbacks HAL para pruebas unitarias en host y ESP-IDF.
 */

typedef struct {
    void   (*delay_ms)(uint32_t ms, void *ctx);
    void   (*set_pwrkey_level)(bool active_low_pull, void *ctx); /* true = MOSFET Q3 conduce -> PWRKEY a GND */
    void   (*set_dtr_sleep)(bool sleep_high, void *ctx);
    bool   (*read_status_pin)(void *ctx);
    void   (*uart_flush_rx)(void *ctx);
    size_t (*uart_write)(const uint8_t *data, size_t len, void *ctx);
    int    (*uart_read_byte)(uint32_t timeout_ms, void *ctx);     /* -1 si timeout */
    void  *ctx;
} aire_modem_hal_t;

typedef struct {
    const char *apn;
    const char *mqtt_host;
    uint16_t    mqtt_port;
    const char *mqtt_client_id;
    const char *mqtt_username; /* Opcional NULL o "" */
    const char *mqtt_password; /* Opcional NULL o "" */
    uint32_t    net_timeout_ms;
} aire_modem_cfg_t;

bool aire_modem_power_on(const aire_modem_hal_t *hal);
bool aire_modem_power_off(const aire_modem_hal_t *hal);
bool aire_modem_connect_ltem(const aire_modem_hal_t *hal, const aire_modem_cfg_t *cfg);
bool aire_modem_sync_utc(const aire_modem_hal_t *hal, char *out_iso_utc, size_t cap);

bool aire_modem_mqtt_open(const aire_modem_hal_t *hal, const aire_modem_cfg_t *cfg);
bool aire_modem_mqtt_publish_qos1(const aire_modem_hal_t *hal, const char *topic, const char *json_payload);
void aire_modem_mqtt_close(const aire_modem_hal_t *hal);

/* Funciones puras de parseo de tramas AT expuestas para pruebas deterministas */
bool aire_modem_parse_cgnsinf_utc(const char *line, char *out_iso_utc, size_t cap);
bool aire_modem_parse_cclk_utc(const char *line, char *out_iso_utc, size_t cap);

#ifdef __cplusplus
}
#endif

#endif /* AIRE_MODEM_AT_H */
