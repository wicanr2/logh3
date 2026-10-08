/*
 * GIN3PS.EXE 2703:1433  從星域的待命存量領出一組編制的艘數（TakeRegionShips）
 *
 * 原版：far 呼叫，三個參數（各佔一個字，只讀低位元組）：星域編號、陣營、艦種；回 AX ＝ 領出的艘數。
 * 語意（已證實：反組譯逐指令；docs/re/215 一致）：
 *   欄位 ＝ ds:A79C + 星域 × 126 + 0x30 + 陣營 × 28 + 艦種 × 4     （dword 存量，位移 16 位元回繞）
 *   滿編 ＝ word(ds:4E3A + 陣營 × 208 + 艦種 × 26)                  （艦種表的 +0x00）
 *   存量高字 != 0     → n ＝ 滿編
 *   存量高字 == 0     → n ＝ 存量低字 < 滿編 ? 存量低字 : 滿編（無號）
 *   dword(欄位) −= n（sub／sbb，高字吃借位）；回傳 n
 *   等價於 min(存量, 滿編)，但高字非 0 的路徑與借位是反組譯的真實形狀。
 * 欄位名稱：星域 +0x30 起是 2 陣營 × 7 艦種 × 4 bytes 的待命存量（docs/internals/22）；
 *   艦種表 ds:4E3A 一個陣營 208 bytes、一個艦種 26 bytes，+0x00 是滿編艘數（docs/re/215、345）。
 * 副作用：只寫那一個 dword。
 * 輸入域：星域、陣營、艦種都按 byte 零擴充，不檢查範圍。呼叫端只傳艦種 0..6（艦種 7 不呼叫這支，
 *   docs/re/215），測試仍涵蓋 0..7 與更大的值（位址照 16 位元算式落在 DS 其他位置）。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2703_1433(Mem *m, uint16_t region, uint16_t side, uint16_t kind)
{
    uint8_t r = (uint8_t)region, s = (uint8_t)side, k = (uint8_t)kind;
    uint16_t field = (uint16_t)(0xA79C + (uint16_t)(r * 0x7E));
    uint16_t cap, lo, hi, take;

    field = (uint16_t)(field + 0x30);
    field = (uint16_t)(field + (uint16_t)(s * 0x1C));
    field = (uint16_t)(field + (uint16_t)(k << 2));
    cap = rd16(m, DS((uint16_t)(0x4E3A + (uint16_t)(s * 0xD0) + (uint16_t)(k * 0x1A))));
    lo = rd16(m, DS(field));
    hi = rd16(m, DS((uint16_t)(field + 2)));

    if (hi != 0)
        take = cap;
    else if (lo < cap)
        take = lo;
    else
        take = cap;

    wr16(m, DS(field), (uint16_t)(lo - take));
    wr16(m, DS((uint16_t)(field + 2)), (uint16_t)(hi - (lo < take)));
    return take;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2703_1433(m, a[0], a[1], a[2]); }
GIN3_REGISTER("2703_1433", 3, run)
