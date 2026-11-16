#ifndef SENSORS_HAL_H
#define SENSORS_HAL_H

#include <stdbool.h>
#include "aire_telemetry.h"
#include "aire_calibration.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Adquiere todas las variables físicas disponibles respetando las fallas puntuales sin abortar el ciclo */
void sensors_hal_acquire_all(aire_telemetry_t *out_tel,
                             aire_voc_state_t *rtc_voc_state,
                             bool allow_5v_sensors,
                             uint32_t warmup_ms,
                             double mic_cal_offset_db);

#ifdef __cplusplus
}
#endif

#endif /* SENSORS_HAL_H */
