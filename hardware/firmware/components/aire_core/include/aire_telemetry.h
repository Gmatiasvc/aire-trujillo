#ifndef AIRE_TELEMETRY_H
#define AIRE_TELEMETRY_H

#include <stdbool.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Contrato de telemetría Zero-Drift (docs/design/02-domain-and-data.md y
 * backend/internal/domain/models.go).
 */

typedef enum {
    AIRE_METRIC_PM25 = 0,
    AIRE_METRIC_PM10,
    AIRE_METRIC_CO,
    AIRE_METRIC_NO2,
    AIRE_METRIC_SO2,
    AIRE_METRIC_H2S,
    AIRE_METRIC_CO2,
    AIRE_METRIC_VOC_INDEX,
    AIRE_METRIC_NOISE_DBA,
    AIRE_METRIC_TEMPERATURE_C,
    AIRE_METRIC_HUMIDITY_PCT,
    AIRE_METRIC_PRESSURE_HPA,
    AIRE_METRIC_COUNT
} aire_metric_id_t;

typedef struct {
    double values[AIRE_METRIC_COUNT];
    bool   valid[AIRE_METRIC_COUNT];
    char   node_code[65];
    char   recorded_at_utc[32]; /* ISO-8601 UTC: YYYY-MM-DDTHH:MM:SSZ */
} aire_telemetry_t;

void aire_telemetry_init(aire_telemetry_t *t, const char *node_code, const char *recorded_at_utc);
bool aire_metric_in_range(aire_metric_id_t id, double val);
bool aire_telemetry_set_metric(aire_telemetry_t *t, aire_metric_id_t id, double val);
size_t aire_telemetry_valid_count(const aire_telemetry_t *t);

/* Construye "aire-trujillo/v1/nodes/{node_code}/telemetry". Retorna longitud escrita o 0 si inválido. */
size_t aire_telemetry_format_topic(const char *node_code, char *out_buf, size_t out_cap);

/*
 * Serializa el JSON exacto esperado por Mosquitto / Go Subscriber.
 * Omite cualquier métrica inválida o ausente (regla Zero-Drift: nunca emite 0 falso).
 * Retorna longitud JSON escrita o 0 si node_code/timestamp inválidos, metrics vacías o buffer insuficiente.
 */
size_t aire_telemetry_serialize_json(const aire_telemetry_t *t, char *out_buf, size_t out_cap);

const char *aire_metric_key(aire_metric_id_t id);

#ifdef __cplusplus
}
#endif

#endif /* AIRE_TELEMETRY_H */
