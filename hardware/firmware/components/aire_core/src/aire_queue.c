#include "aire_queue.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

uint32_t aire_crc32(const void *data, size_t len) {
    const uint8_t *p = (const uint8_t *)data;
    uint32_t crc = 0xFFFFFFFFU;
    for (size_t i = 0U; i < len; ++i) {
        crc ^= p[i];
        for (int b = 0; b < 8; ++b) {
            uint32_t mask = (uint32_t)(-(int32_t)(crc & 1U));
            crc = (crc >> 1) ^ (0xEDB88320U & mask);
        }
    }
    return ~crc;
}

void aire_queue_init(aire_queue_t *q, const char *dir_path, size_t max_bytes) {
    if (!q || !dir_path) {
        return;
    }
    memset(q, 0, sizeof(*q));
    (void)snprintf(q->log_path, sizeof(q->log_path), "%s/queue.log", dir_path);
    (void)snprintf(q->idx_path, sizeof(q->idx_path), "%s/queue.idx", dir_path);
    q->max_bytes = (max_bytes == 0U) ? (4U * 1024U * 1024U) : max_bytes;
}

static size_t read_head_offset(const aire_queue_t *q) {
    FILE *f = fopen(q->idx_path, "rb");
    if (!f) {
        return 0U;
    }
    unsigned long long off = 0ULL;
    if (fscanf(f, "%llu", &off) != 1) {
        off = 0ULL;
    }
    fclose(f);
    return (size_t)off;
}

static bool write_head_offset(const aire_queue_t *q, size_t offset) {
    char tmp_path[272];
    int n = snprintf(tmp_path, sizeof(tmp_path), "%s.tmp", q->idx_path);
    if (n < 0 || (size_t)n >= sizeof(tmp_path)) {
        return false;
    }
    FILE *f = fopen(tmp_path, "wb");
    if (!f) {
        return false;
    }
    bool ok = (fprintf(f, "%llu\n", (unsigned long long)offset) > 0) && (fflush(f) == 0);
    fclose(f);
    if (!ok) {
        remove(tmp_path);
        return false;
    }
    remove(q->idx_path);
    return rename(tmp_path, q->idx_path) == 0;
}

static bool parse_line_crc(const char *line, size_t len,
                           char *out_topic, size_t topic_cap,
                           char *out_json, size_t json_cap) {
    if (len < 12U || line[0] != '|' || line[9] != '|') {
        return false;
    }
    unsigned long expected_crc = 0UL;
    char hex_buf[9];
    memcpy(hex_buf, &line[1], 8U);
    hex_buf[8] = '\0';
    for (int i = 0; i < 8; ++i) {
        char c = hex_buf[i];
        bool is_hex = (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F');
        if (!is_hex) {
            return false;
        }
    }
    expected_crc = strtoul(hex_buf, NULL, 16);

    const char *body = &line[10];
    size_t body_len = len - 10U;
    while (body_len > 0U && (body[body_len - 1U] == '\n' || body[body_len - 1U] == '\r')) {
        --body_len;
    }
    if (body_len == 0U || aire_crc32(body, body_len) != (uint32_t)expected_crc) {
        return false;
    }

    const char *tab = memchr(body, '\t', body_len);
    if (!tab) {
        return false;
    }
    size_t tlen = (size_t)(tab - body);
    size_t jlen = body_len - tlen - 1U;
    if (tlen == 0U || jlen == 0U || tlen >= topic_cap || jlen >= json_cap) {
        return false;
    }
    memcpy(out_topic, body, tlen);
    out_topic[tlen] = '\0';
    memcpy(out_json, tab + 1, jlen);
    out_json[jlen] = '\0';
    return true;
}

static void compact_log_for_append(aire_queue_t *q, size_t incoming_bytes) {
    FILE *f = fopen(q->log_path, "rb");
    if (!f) {
        return;
    }
    fseek(f, 0L, SEEK_END);
    long file_len_l = ftell(f);
    if (file_len_l <= 0) {
        fclose(f);
        return;
    }
    size_t file_len = (size_t)file_len_l;
    if (file_len + incoming_bytes <= q->max_bytes) {
        fclose(f);
        return;
    }

    size_t keep_off = read_head_offset(q);
    if (keep_off > file_len) {
        keep_off = file_len;
    }
    fseek(f, (long)keep_off, SEEK_SET);

    char line[AIRE_QUEUE_MAX_TOPIC_LEN + AIRE_QUEUE_MAX_PAYLOAD_LEN + 32U];
    while ((file_len - keep_off) + incoming_bytes > q->max_bytes &&
           fgets(line, (int)sizeof(line), f) != NULL) {
        long pos = ftell(f);
        keep_off = (pos >= 0) ? (size_t)pos : file_len;
    }

    char tmp_log[272];
    int tn = snprintf(tmp_log, sizeof(tmp_log), "%s.tmp", q->log_path);
    if (tn < 0 || (size_t)tn >= sizeof(tmp_log)) {
        fclose(f);
        return;
    }

    FILE *out = fopen(tmp_log, "wb");
    if (!out) {
        fclose(f);
        return;
    }

    fseek(f, (long)keep_off, SEEK_SET);
    bool copy_ok = true;
    char buf[512];
    size_t nread = 0U;
    while ((nread = fread(buf, 1U, sizeof(buf), f)) > 0U) {
        if (fwrite(buf, 1U, nread, out) != nread) {
            copy_ok = false;
            break;
        }
    }
    if (fflush(out) != 0) {
        copy_ok = false;
    }
    fclose(out);
    fclose(f);

    if (copy_ok) {
        remove(q->log_path);
        if (rename(tmp_log, q->log_path) == 0) {
            (void)write_head_offset(q, 0U);
            return;
        }
    }
    remove(tmp_log);
}

bool aire_queue_push(aire_queue_t *q, const char *topic, const char *json_payload) {
    if (!q || !topic || !json_payload) {
        return false;
    }
    size_t tlen = strlen(topic);
    size_t jlen = strlen(json_payload);
    if (tlen == 0U || jlen == 0U || tlen >= AIRE_QUEUE_MAX_TOPIC_LEN || jlen >= AIRE_QUEUE_MAX_PAYLOAD_LEN) {
        return false;
    }
    if (strchr(topic, '\t') || strchr(topic, '\n') || strchr(json_payload, '\n')) {
        return false;
    }

    char body[AIRE_QUEUE_MAX_TOPIC_LEN + AIRE_QUEUE_MAX_PAYLOAD_LEN + 2U];
    int bn = snprintf(body, sizeof(body), "%s\t%s", topic, json_payload);
    if (bn <= 0 || (size_t)bn >= sizeof(body)) {
        return false;
    }
    uint32_t crc = aire_crc32(body, (size_t)bn);

    /* Compacta y descarta registros más antiguos si supera max_bytes */
    compact_log_for_append(q, (size_t)bn + 12U);

    FILE *f = fopen(q->log_path, "ab");
    if (!f) {
        return false;
    }
    bool ok = (fprintf(f, "|%08lx|%s\n", (unsigned long)crc, body) > 0) && (fflush(f) == 0);
    fclose(f);
    return ok;
}

bool aire_queue_peek(aire_queue_t *q,
                     char *out_topic, size_t topic_cap,
                     char *out_json, size_t json_cap,
                     size_t *out_next_offset) {
    if (!q || !out_topic || !out_json || !out_next_offset) {
        return false;
    }
    FILE *f = fopen(q->log_path, "rb");
    if (!f) {
        return false;
    }

    size_t off = read_head_offset(q);
    fseek(f, 0L, SEEK_END);
    long file_len = ftell(f);
    if (file_len < 0 || off >= (size_t)file_len) {
        fclose(f);
        return false;
    }
    fseek(f, (long)off, SEEK_SET);

    char line[AIRE_QUEUE_MAX_TOPIC_LEN + AIRE_QUEUE_MAX_PAYLOAD_LEN + 32U];
    while (fgets(line, (int)sizeof(line), f) != NULL) {
        long cur_pos = ftell(f);
        size_t l = strlen(line);
        if (parse_line_crc(line, l, out_topic, topic_cap, out_json, json_cap)) {
            *out_next_offset = (cur_pos >= 0) ? (size_t)cur_pos : off + l;
            fclose(f);
            return true;
        }
        /* Línea corrupta o truncada tras corte de energía: avanza cursor automáticamente */
        if (cur_pos >= 0) {
            off = (size_t)cur_pos;
            (void)write_head_offset(q, off);
        }
    }

    fclose(f);
    return false;
}

bool aire_queue_commit_pop(aire_queue_t *q, size_t next_offset) {
    if (!q) {
        return false;
    }
    FILE *f = fopen(q->log_path, "rb");
    if (!f) {
        return false;
    }
    fseek(f, 0L, SEEK_END);
    long file_len = ftell(f);
    fclose(f);

    if (file_len >= 0 && next_offset >= (size_t)file_len) {
        /* Cola drenada por completo: compacta eliminando log histórico */
        remove(q->log_path);
        return write_head_offset(q, 0U);
    }
    return write_head_offset(q, next_offset);
}

size_t aire_queue_pending_count(const aire_queue_t *q) {
    if (!q) {
        return 0U;
    }
    FILE *f = fopen(q->log_path, "rb");
    if (!f) {
        return 0U;
    }
    size_t off = read_head_offset(q);
    fseek(f, 0L, SEEK_END);
    long file_len = ftell(f);
    if (file_len < 0 || off >= (size_t)file_len) {
        fclose(f);
        return 0U;
    }
    fseek(f, (long)off, SEEK_SET);

    char line[AIRE_QUEUE_MAX_TOPIC_LEN + AIRE_QUEUE_MAX_PAYLOAD_LEN + 32U];
    char dt[AIRE_QUEUE_MAX_TOPIC_LEN];
    char dj[AIRE_QUEUE_MAX_PAYLOAD_LEN];
    size_t count = 0U;
    while (fgets(line, (int)sizeof(line), f) != NULL) {
        if (parse_line_crc(line, strlen(line), dt, sizeof(dt), dj, sizeof(dj))) {
            ++count;
        }
    }
    fclose(f);
    return count;
}
