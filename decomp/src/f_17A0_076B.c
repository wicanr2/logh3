/*
 * GIN3PS.EXE 17A0:076B  兩點是否靠得夠近（x 與 y 的距離都小於 16）
 *
 * 原版：far 呼叫，四個 word 參數（x0, y0, x1, y1），回 AX = 1 或 0，不讀寫記憶體。
 *   堆疊：序言 push bp／si／di、sub sp,4，所以第一個參數在 [bp+0Eh]。
 *   [bp+0Eh]＝x0、[bp+10h]＝y0、[bp+12h]＝x1、[bp+14h]＝y1。
 *
 * 語意（已證實：本檔與反組譯逐條對過；呼叫者 17A0:06D7 的用途為強推論，未在此驗證）：
 *   dx ＝ |x0 − x1|、dy ＝ |y0 − y1|，兩個都是「大減小」的無號差（jb 比較後相減），不是有號差。
 *   dx <  16 且 dy < 16 回 1，其餘回 0。
 *   兩個比較都是無號（jnb／jb）。
 *
 * 輸入域：四個參數全部 16 位元。差異測試的參數集中在 0..47 與 0xFFF0..0xFFFF（讓距離落在 16 的兩側），
 *   另混入均勻隨機值。因為是無號差，0 與 0xFFF8 的距離是 0xFFF8，不是 8。
 */
#include "gin3_reg.h"

uint16_t f_17A0_076B(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1)
{
    uint16_t dx, dy;

    (void)m;
    dx = x0 < x1 ? (uint16_t)(x1 - x0) : (uint16_t)(x0 - x1);
    dy = y0 < y1 ? (uint16_t)(y1 - y0) : (uint16_t)(y0 - y1);
    if (dx >= 0x10u)
        return 0;
    return dy < 0x10u ? 1 : 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_17A0_076B(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("17A0_076B", 4, run)
