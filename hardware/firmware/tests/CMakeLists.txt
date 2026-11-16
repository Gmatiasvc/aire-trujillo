cmake_minimum_required(VERSION 3.16)
project(aire_firmware_host_tests C)

set(CMAKE_C_STANDARD 11)
set(CMAKE_C_STANDARD_REQUIRED ON)
set(CMAKE_C_EXTENSIONS OFF)

add_executable(test_firmware_core
    test_firmware_core.c
    ../components/aire_core/src/aire_telemetry.c
    ../components/aire_core/src/aire_calibration.c
    ../components/aire_core/src/aire_queue.c
    ../components/aire_core/src/aire_modem_at.c
)

target_include_directories(test_firmware_core PRIVATE
    ../components/aire_core/include
)

target_compile_options(test_firmware_core PRIVATE
    -Wall -Wextra -Werror -Wpedantic -Wconversion -Wshadow -O2
)

target_link_libraries(test_firmware_core PRIVATE m)

enable_testing()
add_test(NAME aire_core_suite COMMAND test_firmware_core)
