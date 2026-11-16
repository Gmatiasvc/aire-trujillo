#ifndef BUS_HAL_H
#define BUS_HAL_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include "esp_err.h"
#include "aire_modem_at.h"

#ifdef __cplusplus
extern "C" {
#endif

esp_err_t bus_hal_power_init(void);
void      bus_hal_set_3v3_sens(bool enable);
void      bus_hal_set_5v_sens(bool enable);
void      bus_hal_set_led(bool on);
void      bus_hal_read_charger_status(bool *charging, bool *done);

esp_err_t bus_hal_adc_read_power_mv(uint32_t *out_vbat_mv, uint32_t *out_vsolar_mv);

esp_err_t bus_hal_i2c_init(void);
void      bus_hal_i2c_deinit(void);
esp_err_t bus_hal_i2c_write(uint8_t addr7, const uint8_t *tx, size_t len);
esp_err_t bus_hal_i2c_read(uint8_t addr7, uint8_t *rx, size_t len);
esp_err_t bus_hal_i2c_write_read(uint8_t addr7, const uint8_t *tx, size_t tx_len, uint8_t *rx, size_t rx_len);

esp_err_t bus_hal_i2s_init(void);
void      bus_hal_i2s_deinit(void);
esp_err_t bus_hal_i2s_read_samples(int32_t *out_24bit_signed, size_t count, size_t *out_read);

bool      bus_hal_sdcard_mount(const char *mount_point);
void      bus_hal_sdcard_unmount(const char *mount_point);

esp_err_t        bus_hal_modem_uart_init(void);
void             bus_hal_modem_uart_deinit(void);
aire_modem_hal_t bus_hal_get_modem_ops(void);

void bus_hal_prepare_deep_sleep(uint64_t sleep_us);

#ifdef __cplusplus
}
#endif

#endif /* BUS_HAL_H */
