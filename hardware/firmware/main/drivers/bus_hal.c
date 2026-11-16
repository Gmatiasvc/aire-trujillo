#include "bus_hal.h"
#include "aire_board.h"
#include "aire_calibration.h"

#include <string.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "driver/gpio.h"
#include "driver/i2c_master.h"
#include "driver/i2s_std.h"
#include "driver/uart.h"
#include "driver/spi_master.h"
#include "driver/sdspi_host.h"
#include "esp_adc/adc_oneshot.h"
#include "esp_adc/adc_cali.h"
#include "esp_adc/adc_cali_scheme.h"
#include "esp_vfs_fat.h"
#include "sdmmc_cmd.h"
#include "esp_sleep.h"
#include "esp_log.h"

static const char *TAG = "bus_hal";

static i2c_master_bus_handle_t s_i2c_bus = NULL;
static i2s_chan_handle_t       s_i2s_rx  = NULL;
static sdmmc_card_t           *s_sd_card = NULL;
static bool                    s_uart_up = false;

esp_err_t bus_hal_power_init(void) {
    /* Desactiva retención previa al despertar de deep-sleep */
    gpio_hold_dis(AIRE_GPIO_EN_SENS_3V3);
    gpio_hold_dis(AIRE_GPIO_PWR_5V_EN);
    gpio_hold_dis(AIRE_GPIO_EN_SENS_5V);

    const gpio_config_t out_cfg = {
        .pin_bit_mask = (1ULL << AIRE_GPIO_EN_SENS_3V3) |
                        (1ULL << AIRE_GPIO_PWR_5V_EN)   |
                        (1ULL << AIRE_GPIO_EN_SENS_5V)  |
                        (1ULL << AIRE_GPIO_LED_SYS)     |
                        (1ULL << AIRE_GPIO_MODEM_PWRKEY)|
                        (1ULL << AIRE_GPIO_MODEM_DTR),
        .mode = GPIO_MODE_OUTPUT,
        .pull_up_en = GPIO_PULLUP_DISABLE,
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type = GPIO_INTR_DISABLE,
    };
    esp_err_t err = gpio_config(&out_cfg);
    if (err != ESP_OK) {
        return err;
    }

    gpio_set_level(AIRE_GPIO_EN_SENS_3V3, 0);
    gpio_set_level(AIRE_GPIO_PWR_5V_EN, 0);
    gpio_set_level(AIRE_GPIO_EN_SENS_5V, 0);
    gpio_set_level(AIRE_GPIO_LED_SYS, 0);
    gpio_set_level(AIRE_GPIO_MODEM_PWRKEY, 0); /* Gate Q3 en 0 -> no pulsa PWRKEY */
    gpio_set_level(AIRE_GPIO_MODEM_DTR, 0);

    const gpio_config_t in_cfg = {
        .pin_bit_mask = (1ULL << AIRE_GPIO_CHG_STAT) |
                        (1ULL << AIRE_GPIO_CHG_DONE) |
                        (1ULL << AIRE_GPIO_ADS_ALERT)|
                        (1ULL << AIRE_GPIO_MODEM_STATUS)|
                        (1ULL << AIRE_GPIO_SD_CD),
        .mode = GPIO_MODE_INPUT,
        .pull_up_en = GPIO_PULLUP_DISABLE, /* Pull-ups externos R14/R15/R44/R9 */
        .pull_down_en = GPIO_PULLDOWN_DISABLE,
        .intr_type = GPIO_INTR_DISABLE,
    };
    return gpio_config(&in_cfg);
}

void bus_hal_set_3v3_sens(bool enable) {
    gpio_set_level(AIRE_GPIO_EN_SENS_3V3, enable ? 1 : 0);
    if (enable) {
        vTaskDelay(pdMS_TO_TICKS(AIRE_SETTLE_3V3_SENS_MS));
    }
}

void bus_hal_set_5v_sens(bool enable) {
    if (enable) {
        gpio_set_level(AIRE_GPIO_PWR_5V_EN, 1);
        vTaskDelay(pdMS_TO_TICKS(AIRE_SETTLE_5V_BOOST_MS));
        gpio_set_level(AIRE_GPIO_EN_SENS_5V, 1);
        vTaskDelay(pdMS_TO_TICKS(AIRE_SETTLE_5V_SENS_MS));
    } else {
        gpio_set_level(AIRE_GPIO_EN_SENS_5V, 0);
        gpio_set_level(AIRE_GPIO_PWR_5V_EN, 0);
    }
}

void bus_hal_set_led(bool on) {
    gpio_set_level(AIRE_GPIO_LED_SYS, on ? 1 : 0);
}

void bus_hal_read_charger_status(bool *charging, bool *done) {
    /* Open-drain activos en bajo de CN3791 */
    if (charging) {
        *charging = (gpio_get_level(AIRE_GPIO_CHG_STAT) == 0);
    }
    if (done) {
        *done = (gpio_get_level(AIRE_GPIO_CHG_DONE) == 0);
    }
}

esp_err_t bus_hal_adc_read_power_mv(uint32_t *out_vbat_mv, uint32_t *out_vsolar_mv) {
    adc_oneshot_unit_handle_t adc1 = NULL;
    const adc_oneshot_unit_init_cfg_t unit_cfg = { .unit_id = ADC_UNIT_1 };
    esp_err_t err = adc_oneshot_new_unit(&unit_cfg, &adc1);
    if (err != ESP_OK) {
        return err;
    }

    adc_unit_t u_bat, u_sol;
    adc_channel_t ch_bat, ch_sol;
    adc_oneshot_io_to_channel(AIRE_GPIO_VBAT_ADC, &u_bat, &ch_bat);
    adc_oneshot_io_to_channel(AIRE_GPIO_VSOLAR_ADC, &u_sol, &ch_sol);

    const adc_oneshot_chan_cfg_t ch_cfg = {
        .atten    = ADC_ATTEN_DB_12,
        .bitwidth = ADC_BITWIDTH_12,
    };
    adc_oneshot_config_channel(adc1, ch_bat, &ch_cfg);
    adc_oneshot_config_channel(adc1, ch_sol, &ch_cfg);

    adc_cali_handle_t cali = NULL;
    bool has_cali = false;
#if ADC_CALI_SCHEME_CURVE_FITTING_SUPPORTED
    const adc_cali_curve_fitting_config_t cal_cfg = {
        .unit_id  = ADC_UNIT_1,
        .atten    = ADC_ATTEN_DB_12,
        .bitwidth = ADC_BITWIDTH_12,
    };
    has_cali = (adc_cali_create_scheme_curve_fitting(&cal_cfg, &cali) == ESP_OK);
#endif

    int sum_bat = 0, sum_sol = 0;
    for (int i = 0; i < 8; ++i) {
        int raw_b = 0, raw_s = 0, mv = 0;
        if (adc_oneshot_read(adc1, ch_bat, &raw_b) == ESP_OK) {
            if (has_cali && adc_cali_raw_to_voltage(cali, raw_b, &mv) == ESP_OK) {
                sum_bat += mv;
            } else {
                sum_bat += (raw_b * 3300) / 4095;
            }
        }
        if (adc_oneshot_read(adc1, ch_sol, &raw_s) == ESP_OK) {
            if (has_cali && adc_cali_raw_to_voltage(cali, raw_s, &mv) == ESP_OK) {
                sum_sol += mv;
            } else {
                sum_sol += (raw_s * 3300) / 4095;
            }
        }
    }

    if (out_vbat_mv) {
        *out_vbat_mv = aire_adc_divider_mv((uint32_t)(sum_bat / 8), AIRE_VBAT_DIV_NUM);
    }
    if (out_vsolar_mv) {
        *out_vsolar_mv = aire_adc_divider_mv((uint32_t)(sum_sol / 8), AIRE_VSOLAR_DIV_NUM);
    }

#if ADC_CALI_SCHEME_CURVE_FITTING_SUPPORTED
    if (has_cali) {
        adc_cali_delete_scheme_curve_fitting(cali);
    }
#endif
    adc_oneshot_del_unit(adc1);
    return ESP_OK;
}

esp_err_t bus_hal_i2c_init(void) {
    if (s_i2c_bus) {
        return ESP_OK;
    }
    const i2c_master_bus_config_t cfg = {
        .i2c_port   = I2C_NUM_0,
        .sda_io_num = AIRE_GPIO_I2C_SDA,
        .scl_io_num = AIRE_GPIO_I2C_SCL,
        .clk_source = I2C_CLK_SRC_DEFAULT,
        .glitch_ignore_cnt = 7,
        .flags.enable_internal_pullup = false, /* Pull-ups R40/R41 externos a 3V3_SENS */
    };
    return i2c_new_master_bus(&cfg, &s_i2c_bus);
}

void bus_hal_i2c_deinit(void) {
    if (s_i2c_bus) {
        i2c_del_master_bus(s_i2c_bus);
        s_i2c_bus = NULL;
    }
}

static esp_err_t i2c_with_dev(uint8_t addr7,
                              const uint8_t *tx, size_t tx_len,
                              uint8_t *rx, size_t rx_len) {
    if (!s_i2c_bus) {
        return ESP_ERR_INVALID_STATE;
    }
    const i2c_device_config_t dcfg = {
        .dev_addr_length = I2C_ADDR_BIT_LEN_7,
        .device_address  = addr7,
        .scl_speed_hz    = AIRE_I2C_FREQ_HZ,
    };
    i2c_master_dev_handle_t dev = NULL;
    esp_err_t err = i2c_master_bus_add_device(s_i2c_bus, &dcfg, &dev);
    if (err != ESP_OK) {
        return err;
    }
    if (tx && tx_len > 0U && rx && rx_len > 0U) {
        err = i2c_master_transmit_receive(dev, tx, tx_len, rx, rx_len, 250);
    } else if (tx && tx_len > 0U) {
        err = i2c_master_transmit(dev, tx, tx_len, 250);
    } else if (rx && rx_len > 0U) {
        err = i2c_master_receive(dev, rx, rx_len, 250);
    }
    i2c_master_bus_rm_device(dev);
    return err;
}

esp_err_t bus_hal_i2c_write(uint8_t addr7, const uint8_t *tx, size_t len) {
    return i2c_with_dev(addr7, tx, len, NULL, 0U);
}

esp_err_t bus_hal_i2c_read(uint8_t addr7, uint8_t *rx, size_t len) {
    return i2c_with_dev(addr7, NULL, 0U, rx, len);
}

esp_err_t bus_hal_i2c_write_read(uint8_t addr7, const uint8_t *tx, size_t tx_len, uint8_t *rx, size_t rx_len) {
    return i2c_with_dev(addr7, tx, tx_len, rx, rx_len);
}

esp_err_t bus_hal_i2s_init(void) {
    if (s_i2s_rx) {
        return ESP_OK;
    }
    i2s_chan_config_t chan_cfg = I2S_CHANNEL_DEFAULT_CONFIG(I2S_NUM_0, I2S_ROLE_MASTER);
    esp_err_t err = i2s_new_channel(&chan_cfg, NULL, &s_i2s_rx);
    if (err != ESP_OK) {
        return err;
    }

    i2s_std_config_t std_cfg = {
        .clk_cfg  = I2S_STD_CLK_DEFAULT_CONFIG(48000),
        .slot_cfg = I2S_STD_PHILIPS_SLOT_DEFAULT_CONFIG(I2S_DATA_BIT_WIDTH_32BIT, I2S_SLOT_MODE_MONO),
        .gpio_cfg = {
            .mclk = I2S_GPIO_UNUSED,
            .bclk = AIRE_GPIO_I2S_SCK,
            .ws   = AIRE_GPIO_I2S_WS,
            .dout = I2S_GPIO_UNUSED,
            .din  = AIRE_GPIO_I2S_SD,
            .invert_flags = { .mclk_inv = false, .bclk_inv = false, .ws_inv = false },
        },
    };
    std_cfg.slot_cfg.slot_mask = I2S_STD_SLOT_LEFT; /* ICS-43434 L/R a GND */
    err = i2s_channel_init_std_mode(s_i2s_rx, &std_cfg);
    if (err != ESP_OK) {
        i2s_del_channel(s_i2s_rx);
        s_i2s_rx = NULL;
        return err;
    }
    return i2s_channel_enable(s_i2s_rx);
}

void bus_hal_i2s_deinit(void) {
    if (s_i2s_rx) {
        i2s_channel_disable(s_i2s_rx);
        i2s_del_channel(s_i2s_rx);
        s_i2s_rx = NULL;
    }
}

esp_err_t bus_hal_i2s_read_samples(int32_t *out_24bit_signed, size_t count, size_t *out_read) {
    if (!s_i2s_rx || !out_24bit_signed) {
        return ESP_ERR_INVALID_STATE;
    }
    size_t bytes_read = 0U;
    esp_err_t err = i2s_channel_read(s_i2s_rx, out_24bit_signed, count * sizeof(int32_t), &bytes_read, 500);
    size_t n = bytes_read / sizeof(int32_t);
    for (size_t i = 0U; i < n; ++i) {
        out_24bit_signed[i] >>= 8; /* 24-bit justificados MSB en palabra de 32 bits */
    }
    if (out_read) {
        *out_read = n;
    }
    return err;
}

bool bus_hal_sdcard_mount(const char *mount_point) {
    if (s_sd_card) {
        return true;
    }
    const sdmmc_host_t host = SDSPI_HOST_DEFAULT();
    const spi_bus_config_t bus_cfg = {
        .mosi_io_num     = AIRE_GPIO_SD_MOSI,
        .miso_io_num     = AIRE_GPIO_SD_MISO,
        .sclk_io_num     = AIRE_GPIO_SD_CLK,
        .quadwp_io_num   = -1,
        .quadhd_io_num   = -1,
        .max_transfer_sz = 4096,
    };
    if (spi_bus_initialize((spi_host_device_t)host.slot, &bus_cfg, SPI_DMA_CH_AUTO) != ESP_OK) {
        return false;
    }
    sdspi_device_config_t slot_cfg = SDSPI_DEVICE_CONFIG_DEFAULT();
    slot_cfg.gpio_cs  = AIRE_GPIO_SD_CS;
    slot_cfg.host_id  = (spi_host_device_t)host.slot;

    const esp_vfs_fat_sdmmc_mount_config_t mcfg = {
        .format_if_mount_failed = false,
        .max_files              = 4,
        .allocation_unit_size   = 16 * 1024,
    };
    esp_err_t err = esp_vfs_fat_sdspi_mount(mount_point, &host, &slot_cfg, &mcfg, &s_sd_card);
    if (err != ESP_OK) {
        ESP_LOGW(TAG, "MicroSD no montada (%s)", esp_err_to_name(err));
        spi_bus_free((spi_host_device_t)host.slot);
        s_sd_card = NULL;
        return false;
    }
    return true;
}

void bus_hal_sdcard_unmount(const char *mount_point) {
    if (s_sd_card) {
        esp_vfs_fat_sdcard_unmount(mount_point, s_sd_card);
        s_sd_card = NULL;
        spi_bus_free(SPI2_HOST);
    }
}

/* ==================== UART Módem SIM7080G + Callbacks HAL ==================== */

esp_err_t bus_hal_modem_uart_init(void) {
    if (s_uart_up) {
        return ESP_OK;
    }
    const uart_config_t ucfg = {
        .baud_rate  = 115200,
        .data_bits  = UART_DATA_8_BITS,
        .parity     = UART_PARITY_DISABLE,
        .stop_bits  = UART_STOP_BITS_1,
        .flow_ctrl  = UART_HW_FLOWCTRL_DISABLE,
        .source_clk = UART_SCLK_DEFAULT,
    };
    esp_err_t err = uart_driver_install(UART_NUM_1, 2048, 0, 0, NULL, 0);
    if (err != ESP_OK) {
        return err;
    }
    uart_param_config(UART_NUM_1, &ucfg);
    uart_set_pin(UART_NUM_1, AIRE_GPIO_MODEM_TXD, AIRE_GPIO_MODEM_RXD,
                 UART_PIN_NO_CHANGE, UART_PIN_NO_CHANGE);
    s_uart_up = true;
    return ESP_OK;
}

void bus_hal_modem_uart_deinit(void) {
    if (s_uart_up) {
        uart_driver_delete(UART_NUM_1);
        s_uart_up = false;
    }
}

static void hal_delay_ms(uint32_t ms, void *ctx) { (void)ctx; vTaskDelay(pdMS_TO_TICKS(ms)); }
static void hal_pwrkey(bool active, void *ctx) { (void)ctx; gpio_set_level(AIRE_GPIO_MODEM_PWRKEY, active ? 1 : 0); }
static void hal_dtr(bool sleep_hi, void *ctx) { (void)ctx; gpio_set_level(AIRE_GPIO_MODEM_DTR, sleep_hi ? 1 : 0); }
static bool hal_status(void *ctx) { (void)ctx; return gpio_get_level(AIRE_GPIO_MODEM_STATUS) != 0; }
static void hal_flush(void *ctx) { (void)ctx; uart_flush_input(UART_NUM_1); }
static size_t hal_write(const uint8_t *d, size_t l, void *ctx) {
    (void)ctx;
    int n = uart_write_bytes(UART_NUM_1, (const char *)d, l);
    return n > 0 ? (size_t)n : 0U;
}
static int hal_read(uint32_t to_ms, void *ctx) {
    (void)ctx;
    uint8_t b = 0U;
    int n = uart_read_bytes(UART_NUM_1, &b, 1U, pdMS_TO_TICKS(to_ms));
    return (n == 1) ? (int)b : -1;
}

aire_modem_hal_t bus_hal_get_modem_ops(void) {
    return (aire_modem_hal_t){
        .delay_ms         = hal_delay_ms,
        .set_pwrkey_level = hal_pwrkey,
        .set_dtr_sleep    = hal_dtr,
        .read_status_pin  = hal_status,
        .uart_flush_rx    = hal_flush,
        .uart_write       = hal_write,
        .uart_read_byte   = hal_read,
        .ctx              = NULL,
    };
}

void bus_hal_prepare_deep_sleep(uint64_t sleep_us) {
    bus_hal_i2s_deinit();
    bus_hal_i2c_deinit();
    bus_hal_modem_uart_deinit();

    bus_hal_set_5v_sens(false);
    bus_hal_set_3v3_sens(false);
    bus_hal_set_led(false);

    /* Retiene apagados los load switches TPS22919 y boost TPS61023 durante deep-sleep */
    gpio_hold_en(AIRE_GPIO_EN_SENS_3V3);
    gpio_hold_en(AIRE_GPIO_PWR_5V_EN);
    gpio_hold_en(AIRE_GPIO_EN_SENS_5V);
    gpio_deep_sleep_hold_en();

    esp_sleep_enable_timer_wakeup(sleep_us);
    esp_deep_sleep_start();
}
