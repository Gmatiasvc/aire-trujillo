#include "aire_telemetry.h"

#include <math.h>
#include <stdio.h>
#include <string.h>

typedef struct {
    const char *key;
    double      min_val;
    double      max_val;
    const char *fmt;
} aire_metric_meta_t;

/* Tabla única coincidente carácter por carácter con backend/internal/domain/models.go */
static const aire_metric_meta_t k_metric_meta[AIRE_METRIC_COUNT] = {
    [AIRE_METRIC_PM25]          = { "pm25",          0.0,   1000.0,  "%.2f" },
    [AIRE_METRIC_PM10]          = { "pm10",          0.0,   1000.0,  "%.2f" },
    [AIRE_METRIC_CO]            = { "co",            0.0,   100.0,   "%.3f" },
    [AIRE_METRIC_NO2]           = { "no2",           0.0,   2000.0,  "%.2f" },
    [AIRE_METRIC_SO2]           = { "so2",           0.0,   2000.0,  "%.2f" },
    [AIRE_METRIC_H2S]           = { "h2s",           0.0,   2000.0,  "%.2f" },
    [AIRE_METRIC_CO2]           = { "co2",           400.0, 10000.0, "%.1f" },
    [AIRE_METRIC_VOC_INDEX]     = { "voc_index",     1.0,   500.0,   "%.0f" },
    [AIRE_METRIC_NOISE_DBA]     = { "noise_dba",     20.0,  130.0,   "%.1f" },
    [AIRE_METRIC_TEMPERATURE_C] = { "temperature_c", -10.0, 60.0,    "%.2f" },
    [AIRE_METRIC_HUMIDITY_PCT]  = { "humidity_pct",  0.0,   100.0,   "%.2f" },
    [AIRE_METRIC_PRESSURE_HPA]  = { "pressure_hpa",  800.0, 1100.0,  "%.2f" },
};

static bool safe_copy_str(char *dst, size_t dst_cap, const char *src) {
    if (!dst || dst_cap == 0U || !src) {
        return false;
    }
    size_t len = strlen(src);
    if (len >= dst_cap) {
        return false;
    }
    memcpy(dst, src, len + 1U);
    return true;
}

static bool append_json_escaped(char *buf, size_t cap, size_t *pos, const char *src) {
    for (const unsigned char *p = (const unsigned char *)src; *p != '\0'; ++p) {
        unsigned char c = *p;
        if (c == '"' || c == '\\') {
            if (*pos + 2U >= cap) {
                return false;
            }
            buf[(*pos)++] = '\\';
            buf[(*pos)++] = (char)c;
        } else if (c >= 0x20U && c < 0x7FU) {
            if (*pos + 1U >= cap) {
                return false;
            }
            buf[(*pos)++] = (char)c;
        } else {
            return false; /* Caracteres de control no permitidos en node_code/timestamps */
        }
    }
    return true;
}

void aire_telemetry_init(aire_telemetry_t *t, const char *node_code, const char *recorded_at_utc) {
    if (!t) {
        return;
    }
    memset(t, 0, sizeof(*t));
    if (node_code) {
        (void)safe_copy_str(t->node_code, sizeof(t->node_code), node_code);
    }
    if (recorded_at_utc) {
        (void)safe_copy_str(t->recorded_at_utc, sizeof(t->recorded_at_utc), recorded_at_utc);
    }
}

bool aire_metric_in_range(aire_metric_id_t id, double val) {
    if (id < 0 || id >= AIRE_METRIC_COUNT || !isfinite(val)) {
        return false;
    }
    return val >= k_metric_meta[id].min_val && val <= k_metric_meta[id].max_val;
}

bool aire_telemetry_set_metric(aire_telemetry_t *t, aire_metric_id_t id, double val) {
    if (!t || !aire_metric_in_range(id, val)) {
        return false;
    }
    t->values[id] = val;
    t->valid[id]  = true;
    return true;
}

size_t aire_telemetry_valid_count(const aire_telemetry_t *t) {
    if (!t) {
        return 0U;
    }
    size_t n = 0U;
    for (int i = 0; i < AIRE_METRIC_COUNT; ++i) {
        if (t->valid[i] && aire_metric_in_range((aire_metric_id_t)i, t->values[i])) {
            ++n;
        }
    }
    return n;
}

const char *aire_metric_key(aire_metric_id_t id) {
    if (id < 0 || id >= AIRE_METRIC_COUNT) {
        return "";
    }
    return k_metric_meta[id].key;
}

size_t aire_telemetry_format_topic(const char *node_code, char *out_buf, size_t out_cap) {
    if (!node_code || node_code[0] == '\0' || !out_buf || out_cap == 0U) {
        return 0U;
    }
    for (const char *p = node_code; *p != '\0'; ++p) {
        if (*p == '/' || *p == '+' || *p == '#' || (unsigned char)*p <= 0x20U) {
            return 0U; /* Protege comodines o niveles extra en tópico MQTT */
        }
    }
    int n = snprintf(out_buf, out_cap, "aire-trujillo/v1/nodes/%s/telemetry", node_code);
    if (n < 0 || (size_t)n >= out_cap) {
        return 0U;
    }
    return (size_t)n;
}

size_t aire_telemetry_serialize_json(const aire_telemetry_t *t, char *out_buf, size_t out_cap) {
    if (!t || !out_buf || out_cap < 32U) {
        return 0U;
    }
    if (t->node_code[0] == '\0' || t->recorded_at_utc[0] == '\0') {
        return 0U;
    }
    if (aire_telemetry_valid_count(t) == 0U) {
        return 0U; /* Subscriber backend rechaza metrics vacías */
    }

    size_t pos = 0U;
    int n = snprintf(out_buf, out_cap, "{\"node_code\":\"");
    if (n < 0 || (size_t)n >= out_cap) {
        return 0U;
    }
    pos = (size_t)n;

    if (!append_json_escaped(out_buf, out_cap, &pos, t->node_code)) {
        return 0U;
    }

    n = snprintf(out_buf + pos, out_cap - pos, "\",\"recorded_at\":\"");
    if (n < 0 || pos + (size_t)n >= out_cap) {
        return 0U;
    }
    pos += (size_t)n;

    if (!append_json_escaped(out_buf, out_cap, &pos, t->recorded_at_utc)) {
        return 0U;
    }

    n = snprintf(out_buf + pos, out_cap - pos, "\",\"metrics\":{");
    if (n < 0 || pos + (size_t)n >= out_cap) {
        return 0U;
    }
    pos += (size_t)n;

    bool first = true;
    for (int i = 0; i < AIRE_METRIC_COUNT; ++i) {
        if (!t->valid[i] || !aire_metric_in_range((aire_metric_id_t)i, t->values[i])) {
            continue;
        }
        char num_buf[32];
        int fn = snprintf(num_buf, sizeof(num_buf), k_metric_meta[i].fmt, t->values[i]);
        if (fn <= 0 || (size_t)fn >= sizeof(num_buf)) {
            return 0U;
        }
        int wn = snprintf(out_buf + pos, out_cap - pos, "%s\"%s\":%s",
                          first ? "" : ",", k_metric_meta[i].key, num_buf);
        if (wn < 0 || pos + (size_t)wn >= out_cap) {
            return 0U;
        }
        pos += (size_t)wn;
        first = false;
    }

    if (pos + 3U > out_cap) {
        return 0U;
    }
    out_buf[pos++] = '}';
    out_buf[pos++] = '}';
    out_buf[pos]   = '\0';
    return pos;
}
