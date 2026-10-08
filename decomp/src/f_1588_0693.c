/*
 * GIN3PS.EXE 1588:0693  戰場圖塊遮罩查表（ATTR.DAT 的一筆）
 *
 * 原版：far 呼叫，一個參數（push 的是一個字，函式只讀低位元組：mov cl,[bp+6]），回 AL。
 * AH 是呼叫者留下的殘值，不屬於回傳值（描述檔 ret = pack:al）。
 *
 * 語意（已證實：docs/re/262「ATTR.DAT 就是戰場的圖塊遮罩（ds:9BB4）」，唯一呼叫者 1588:0621；
 *   本檔與反組譯逐條對過）：
 *   回傳 ds:[0x9BB4 + 索引]（索引 0..255；反組譯寫成 [bx-644Ch]，即 0x9BB4 的 16 位元補數寫法）。
 *   表只有 128 筆（docs/re/262），索引 128..255 讀到表後面相鄰的位元組，本版照做。
 *
 * 輸入域：索引的低位元組 0..255；高位元組任意（原版忽略，差異測試以隨機 16 位元值驗證）。
 */
#include "gin3_reg.h"

uint8_t f_1588_0693(Mem *m, uint16_t idx)
{
    uint16_t bx = (uint16_t)(idx & 0xFFu);

    return rd8(m, DS((uint16_t)(bx + 0x9BB4u)));
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_0693(m, a[0]); }
GIN3_REGISTER("1588_0693", 1, run)
