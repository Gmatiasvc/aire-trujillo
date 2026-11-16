#ifndef AIRE_QUEUE_H
#define AIRE_QUEUE_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Cola persistente FIFO store-and-forward sobre MicroSD / VFS FATFS.
 * Formato resiliente a apagones con CRC32 por línea e índice atómico de lectura.
 * Ruta física esperada: /sdcard/queue.log y /sdcard/queue.idx.
 */

#define AIRE_QUEUE_MAX_TOPIC_LEN   128U
#define AIRE_QUEUE_MAX_PAYLOAD_LEN 512U

typedef struct {
    char log_path[256];
    char idx_path[256];
    size_t max_bytes;     /* Límite antes de compactar al drenar */
} aire_queue_t;

void aire_queue_init(aire_queue_t *q, const char *dir_path, size_t max_bytes);

/* Encola tópico + JSON calculando CRC32; retorna true si persistió en disco/SD */
bool aire_queue_push(aire_queue_t *q, const char *topic, const char *json_payload);

/* Inspecciona el registro válido más antiguo pendiente de envío sin eliminarlo */
bool aire_queue_peek(aire_queue_t *q,
                     char *out_topic, size_t topic_cap,
                     char *out_json, size_t json_cap,
                     size_t *out_next_offset);

/* Confirma entrega tras ACK MQTT (+SMPUB); compacta cuando la cola queda vacía */
bool aire_queue_commit_pop(aire_queue_t *q, size_t next_offset);

/* Cuenta los registros válidos actualmente pendientes en cola */
size_t aire_queue_pending_count(const aire_queue_t *q);

uint32_t aire_crc32(const void *data, size_t len);

#ifdef __cplusplus
}
#endif

#endif /* AIRE_QUEUE_H */
