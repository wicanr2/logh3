/*
 * 讀 fleet_attack_probe 產生的向量檔，用 C 重寫版重算，與原版輸出逐筆比對。
 *
 * 向量檔格式（文字，一案一行；# 開頭是 metadata）：
 *   C <id> <艦隊記錄 hex 244 字元> <池筆數> <索引>:<艦艇記錄 hex 156 字元> ... <AX hex4> <DX hex4> <純函式 0|1> <步數>
 * 退出碼：全部相符且原版呼叫沒有動到堆疊以外的記憶體才是 0。
 */
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "gin3_model.h"

static int hexval(int c)
{
    if (c >= '0' && c <= '9') return c - '0';
    if (c >= 'a' && c <= 'f') return c - 'a' + 10;
    if (c >= 'A' && c <= 'F') return c - 'A' + 10;
    return -1;
}

static int unhex(const char *s, uint8_t *out, size_t n)
{
    size_t i;
    for (i = 0; i < n; i++) {
        int hi = hexval(s[2 * i]);
        int lo = hexval(s[2 * i + 1]);
        if (hi < 0 || lo < 0) return -1;
        out[i] = (uint8_t)(hi * 16 + lo);
    }
    return s[2 * n] == '\0' ? 0 : -1;
}

int main(int argc, char **argv)
{
    static char line[16384];
    static uint8_t pool[GIN3_POOL_SHIPS][GIN3_SHIP_SIZE];
    uint8_t fleet[GIN3_FLEET_SIZE];
    FILE *f;
    long total = 0, mismatch = 0, impure = 0, bad = 0;

    if (argc != 2) {
        fprintf(stderr, "用法: %s 向量檔\n", argv[0]);
        return 2;
    }
    f = fopen(argv[1], "r");
    if (!f) {
        perror(argv[1]);
        return 2;
    }

    while (fgets(line, sizeof line, f)) {
        char *save = NULL, *tok;
        long id, npool, k;
        unsigned orig_ax, orig_dx;
        int pure;
        uint16_t got;

        if (line[0] == '#' || line[0] == '\n') continue;
        tok = strtok_r(line, " \n", &save);
        if (!tok || strcmp(tok, "C") != 0) { bad++; continue; }
        id = atol(strtok_r(NULL, " \n", &save));
        tok = strtok_r(NULL, " \n", &save);
        if (!tok || unhex(tok, fleet, GIN3_FLEET_SIZE) != 0) { bad++; continue; }
        npool = atol(strtok_r(NULL, " \n", &save));
        memset(pool, 0, sizeof pool);
        for (k = 0; k < npool; k++) {
            char *rec;
            long idx;
            tok = strtok_r(NULL, " \n", &save);
            rec = tok ? strchr(tok, ':') : NULL;
            if (!rec) { bad++; goto next; }
            *rec++ = '\0';
            idx = atol(tok);
            if (idx < 0 || idx >= GIN3_POOL_SHIPS || unhex(rec, pool[idx], GIN3_SHIP_SIZE) != 0) { bad++; goto next; }
        }
        orig_ax = (unsigned)strtoul(strtok_r(NULL, " \n", &save), NULL, 16);
        orig_dx = (unsigned)strtoul(strtok_r(NULL, " \n", &save), NULL, 16);
        pure = atoi(strtok_r(NULL, " \n", &save));

        total++;
        got = gin3ps_fleet_attack(fleet, &pool[0][0]);
        if (got != (uint16_t)orig_ax) {
            if (mismatch < 20)
                printf("不符 case %ld: C=%u 原版 AX=%u（DX=%04X）\n", id, got, orig_ax, orig_dx);
            mismatch++;
        }
        if (!pure) impure++;
    next:;
    }
    fclose(f);

    printf("案數 %ld，不符 %ld，原版動到堆疊以外記憶體 %ld，格式錯誤 %ld\n", total, mismatch, impure, bad);
    if (total == 0) {
        printf("沒有任何案，不能當通過\n");
        return 1;
    }
    return (mismatch == 0 && impure == 0 && bad == 0) ? 0 : 1;
}
