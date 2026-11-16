#include "aire_modem_at.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static bool at_expect(const aire_modem_hal_t *hal, const char *cmd, const char *expected,
                      uint32_t timeout_ms, char *resp_buf, size_t resp_cap) {
    if (!hal || !hal->uart_write || !hal->uart_read_byte) {
        return false;
    }
    if (cmd && cmd[0] != '\0') {
        if (hal->uart_flush_rx) {
            hal->uart_flush_rx(hal->ctx);
        }
        hal->uart_write((const uint8_t *)cmd, strlen(cmd), hal->ctx);
        hal->uart_write((const uint8_t *)"\r\n", 2U, hal->ctx);
    }

    char local_buf[512];
    char *dst  = (resp_buf && resp_cap > 0U) ? resp_buf : local_buf;
    size_t cap = (resp_buf && resp_cap > 0U) ? resp_cap : sizeof(local_buf);
    size_t pos = 0U;
    dst[0] = '\0';

    /* Ventana por paso de byte usando timeout de lectura UART */
    uint32_t step_to = timeout_ms;
    while (pos + 1U < cap) {
        int ch = hal->uart_read_byte(step_to, hal->ctx);
        if (ch < 0) {
            break;
        }
        step_to = 150U; /* Bytes subsiguientes de la misma trama AT llegan rápido */
        dst[pos++] = (char)ch;
        dst[pos]   = '\0';
        if (expected && strstr(dst, expected) != NULL) {
            return true;
        }
        if (strstr(dst, "\r\nERROR\r\n") != NULL || strstr(dst, "+CME ERROR:") != NULL) {
            return false;
        }
    }
    return expected == NULL;
}

bool aire_modem_power_on(const aire_modem_hal_t *hal) {
    if (!hal || !hal->delay_ms || !hal->set_pwrkey_level) {
        return false;
    }
    if (hal->set_dtr_sleep) {
        hal->set_dtr_sleep(false, hal->ctx); /* Despierta UART del módem */
    }
    if (hal->read_status_pin && hal->read_status_pin(hal->ctx)) {
        if (at_expect(hal, "AT", "OK", 1000U, NULL, 0U)) {
            return true;
        }
    }
    /* Pulso SIM7080G PWRKEY > 1.2s vía BSS138 Q3 */
    hal->set_pwrkey_level(true, hal->ctx);
    hal->delay_ms(1250U, hal->ctx);
    hal->set_pwrkey_level(false, hal->ctx);
    hal->delay_ms(1800U, hal->ctx);

    for (int i = 0; i < 6; ++i) {
        if (at_expect(hal, "ATE0", "OK", 1000U, NULL, 0U)) {
            return true;
        }
        hal->delay_ms(400U, hal->ctx);
    }
    return false;
}

bool aire_modem_power_off(const aire_modem_hal_t *hal) {
    if (!hal) {
        return false;
    }
    (void)at_expect(hal, "AT+SMDISC", "OK", 1500U, NULL, 0U);
    (void)at_expect(hal, "AT+CNACT=0,0", "OK", 2000U, NULL, 0U);
    bool ok = at_expect(hal, "AT+CPOWD=1", "NORMAL POWER DOWN", 3000U, NULL, 0U);
    if (hal->set_dtr_sleep) {
        hal->set_dtr_sleep(true, hal->ctx);
    }
    return ok;
}

bool aire_modem_connect_ltem(const aire_modem_hal_t *hal, const aire_modem_cfg_t *cfg) {
    if (!hal || !cfg || !cfg->apn) {
        return false;
    }
    char resp[256];
    if (!at_expect(hal, "AT+CPIN?", "+CPIN: READY", 2000U, resp, sizeof(resp))) {
        return false;
    }
    if (!at_expect(hal, "AT+CNMP=38", "OK", 1500U, NULL, 0U)) { /* Solo LTE */
        return false;
    }
    if (!at_expect(hal, "AT+CMNB=1", "OK", 1500U, NULL, 0U)) {  /* LTE-M Cat-M1 */
        return false;
    }

    char apn_cmd[160];
    (void)snprintf(apn_cmd, sizeof(apn_cmd), "AT+CGDCONT=1,\"IP\",\"%s\"", cfg->apn);
    if (!at_expect(hal, apn_cmd, "OK", 2000U, NULL, 0U)) {
        return false;
    }

    bool attached = false;
    uint32_t polls = (cfg->net_timeout_ms / 1000U) + 1U;
    if (polls > 30U) {
        polls = 30U;
    }
    for (uint32_t i = 0U; i < polls; ++i) {
        if (at_expect(hal, "AT+CGATT?", "+CGATT: 1", 1500U, resp, sizeof(resp))) {
            attached = true;
            break;
        }
        if (hal->delay_ms) {
            hal->delay_ms(1000U, hal->ctx);
        }
    }
    if (!attached) {
        return false;
    }

    (void)snprintf(apn_cmd, sizeof(apn_cmd), "AT+CNCFG=0,1,\"%s\"", cfg->apn);
    if (!at_expect(hal, apn_cmd, "OK", 2000U, NULL, 0U)) {
        return false;
    }
    return at_expect(hal, "AT+CNACT=0,1", "OK", 8000U, resp, sizeof(resp));
}

bool aire_modem_parse_cgnsinf_utc(const char *line, char *out_iso_utc, size_t cap) {
    if (!line || !out_iso_utc || cap < 21U) {
        return false;
    }
    const char *p = strstr(line, "+CGNSINF:");
    if (!p) {
        return false;
    }
    /* +CGNSINF: <run>,<fix>,<yyyyMMddHHmmss.sss>,... */
    int run = 0, fix = 0;
    char ts[32] = {0};
    if (sscanf(p, "+CGNSINF: %d,%d,%31[^,]", &run, &fix, ts) != 3) {
        return false;
    }
    if (fix != 1 || strlen(ts) < 14U) {
        return false;
    }
    int yr = (ts[0] - '0') * 1000 + (ts[1] - '0') * 100 + (ts[2] - '0') * 10 + (ts[3] - '0');
    if (yr < 2025) {
        return false;
    }
    int n = snprintf(out_iso_utc, cap, "%.4s-%.2s-%.2sT%.2s:%.2s:%.2sZ",
                     &ts[0], &ts[4], &ts[6], &ts[8], &ts[10], &ts[12]);
    return n == 20;
}

static int64_t days_from_civil(int y, int m, int d) {
    y -= (m <= 2);
    int era = (y >= 0 ? y : y - 399) / 400;
    unsigned yoe = (unsigned)(y - era * 400);
    unsigned doy = (unsigned)((153 * (m + (m > 2 ? -3 : 9)) + 2) / 5 + d - 1);
    unsigned doe = yoe * 365U + yoe / 4U - yoe / 100U + doy;
    return (int64_t)era * 146097LL + (int64_t)doe - 719468LL;
}

static void civil_from_days(int64_t z, int *out_y, int *out_m, int *out_d) {
    z += 719468LL;
    int64_t era = (z >= 0 ? z : z - 146096LL) / 146097LL;
    int64_t doe = z - era * 146097LL;
    int64_t yoe = (doe - doe / 1460LL + doe / 36524LL - doe / 146096LL) / 365LL;
    int64_t y   = yoe + era * 400LL;
    int64_t doy = doe - (365LL * yoe + yoe / 4LL - yoe / 100LL);
    int64_t mp  = (5LL * doy + 2LL) / 153LL;
    *out_d = (int)(doy - (153LL * mp + 2LL) / 5LL + 1LL);
    *out_m = (int)(mp + (mp < 10LL ? 3LL : -9LL));
    *out_y = (int)(y + (*out_m <= 2));
}

bool aire_modem_parse_cclk_utc(const char *line, char *out_iso_utc, size_t cap) {
    if (!line || !out_iso_utc || cap < 21U) {
        return false;
    }
    const char *p = strstr(line, "+CCLK:");
    if (!p) {
        return false;
    }
    /* +CCLK: "26/11/15,21:45:10-20" -> tz en cuartos de hora respecto a UTC (-20 => UTC-5) */
    int yy = 0, mm = 0, dd = 0, hh = 0, mi = 0, ss = 0, tz_q = 0;
    char sign_ch = '+';
    int matched = sscanf(p, "+CCLK: \"%d/%d/%d,%d:%d:%d%c%d\"", &yy, &mm, &dd, &hh, &mi, &ss, &sign_ch, &tz_q);
    if (matched < 6) {
        return false;
    }
    if (yy < 25 || mm < 1 || mm > 12 || dd < 1 || dd > 31) {
        return false;
    }
    int64_t offset_sec = 0;
    if (matched == 8 && (sign_ch == '+' || sign_ch == '-')) {
        offset_sec = (int64_t)tz_q * 900LL * (sign_ch == '-' ? -1LL : 1LL);
    }
    int64_t local_sec = days_from_civil(2000 + yy, mm, dd) * 86400LL +
                        (int64_t)hh * 3600LL + (int64_t)mi * 60LL + (int64_t)ss;
    int64_t utc_sec   = local_sec - offset_sec;
    int64_t utc_days  = (utc_sec >= 0) ? (utc_sec / 86400LL) : ((utc_sec - 86399LL) / 86400LL);
    int64_t rem       = utc_sec - utc_days * 86400LL;

    int uy = 0, um = 0, ud = 0;
    civil_from_days(utc_days, &uy, &um, &ud);
    int uh = (int)(rem / 3600LL);
    int umi = (int)((rem % 3600LL) / 60LL);
    int us = (int)(rem % 60LL);

    int n = snprintf(out_iso_utc, cap, "%04d-%02d-%02dT%02d:%02d:%02dZ", uy, um, ud, uh, umi, us);
    return n == 20;
}

bool aire_modem_sync_utc(const aire_modem_hal_t *hal, char *out_iso_utc, size_t cap) {
    if (!hal || !out_iso_utc || cap < 21U) {
        return false;
    }
    char resp[256];
    (void)at_expect(hal, "AT+CGNSPWR=1", "OK", 1500U, NULL, 0U);
    if (at_expect(hal, "AT+CGNSINF", "OK", 1500U, resp, sizeof(resp))) {
        if (aire_modem_parse_cgnsinf_utc(resp, out_iso_utc, cap)) {
            return true;
        }
    }
    if (at_expect(hal, "AT+CCLK?", "OK", 1500U, resp, sizeof(resp))) {
        return aire_modem_parse_cclk_utc(resp, out_iso_utc, cap);
    }
    return false;
}

bool aire_modem_mqtt_open(const aire_modem_hal_t *hal, const aire_modem_cfg_t *cfg) {
    if (!hal || !cfg || !cfg->mqtt_host || !cfg->mqtt_client_id) {
        return false;
    }
    char cmd[224];
    (void)snprintf(cmd, sizeof(cmd), "AT+SMCONF=\"URL\",\"%s\",%u", cfg->mqtt_host, (unsigned)cfg->mqtt_port);
    if (!at_expect(hal, cmd, "OK", 1500U, NULL, 0U)) {
        return false;
    }
    (void)snprintf(cmd, sizeof(cmd), "AT+SMCONF=\"CLIENTID\",\"%s\"", cfg->mqtt_client_id);
    if (!at_expect(hal, cmd, "OK", 1500U, NULL, 0U)) {
        return false;
    }
    if (!at_expect(hal, "AT+SMCONF=\"KEEPTIME\",60", "OK", 1500U, NULL, 0U)) {
        return false;
    }
    if (cfg->mqtt_username && cfg->mqtt_username[0] != '\0') {
        (void)snprintf(cmd, sizeof(cmd), "AT+SMCONF=\"USERNAME\",\"%s\"", cfg->mqtt_username);
        if (!at_expect(hal, cmd, "OK", 1500U, NULL, 0U)) {
            return false;
        }
        if (cfg->mqtt_password && cfg->mqtt_password[0] != '\0') {
            (void)snprintf(cmd, sizeof(cmd), "AT+SMCONF=\"PASSWORD\",\"%s\"", cfg->mqtt_password);
            if (!at_expect(hal, cmd, "OK", 1500U, NULL, 0U)) {
                return false;
            }
        }
    }
    return at_expect(hal, "AT+SMCONN", "OK", 10000U, NULL, 0U);
}

bool aire_modem_mqtt_publish_qos1(const aire_modem_hal_t *hal, const char *topic, const char *json_payload) {
    if (!hal || !topic || !json_payload) {
        return false;
    }
    size_t jlen = strlen(json_payload);
    if (jlen == 0U || jlen > 1024U) {
        return false;
    }
    char cmd[256];
    int n = snprintf(cmd, sizeof(cmd), "AT+SMPUB=\"%s\",%u,1,0", topic, (unsigned)jlen);
    if (n < 0 || (size_t)n >= sizeof(cmd)) {
        return false;
    }
    if (!at_expect(hal, cmd, ">", 5000U, NULL, 0U)) {
        return false;
    }
    hal->uart_write((const uint8_t *)json_payload, jlen, hal->ctx);
    return at_expect(hal, NULL, "OK", 8000U, NULL, 0U);
}

void aire_modem_mqtt_close(const aire_modem_hal_t *hal) {
    if (hal) {
        (void)at_expect(hal, "AT+SMDISC", "OK", 1500U, NULL, 0U);
    }
}
