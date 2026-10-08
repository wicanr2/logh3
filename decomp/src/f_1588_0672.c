/*
 * GIN3PS.EXE 1588:0672  讀戰場地圖的一格圖塊
 *
 * 原版：far 呼叫，兩個參數：格 x（word）、格 y（word），回 AL（AH 是呼叫者留下的殘值，描述檔 ret = pack:al）。
 * 地圖緩衝區的遠指標在 ds:3C1A（位移）／ds:3C1C（段）；每列 128 格。
 *
 * 語意（已證實：docs/re/262「1588:0672(c>>1, r>>1)：地圖[(r>>1)×128 + (c>>1)]，遠指標 ds:3C1A」；本檔與反組譯逐條對過）：
 *   回 byte [ds:3C1C : (ds:3C1A ＋ x ＋ (y << 7))]，位移的加法與 y 左移 7 都是 16 位元（超過 0xFFFF 回繞）。
 *   不檢查 x、y 是否在 128×102 之內；超出時讀到鄰近記憶體，本版照做。
 *
 * 輸入域：x、y 取 16 位元全域；地圖指標取基底狀態（戰場地圖已載入的局面），由突變導向讀取集改動。
 */
#include "gin3_reg.h"

uint8_t f_1588_0672(Mem *m, uint16_t x, uint16_t y)
{
    uint16_t off = rd16(m, DS(0x3C1A));
    uint16_t seg = rd16(m, DS(0x3C1C));
    uint16_t si = (uint16_t)(off + x + (uint16_t)(y << 7));

    return rd8(m, LIN(seg, si));
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_0672(m, a[0], a[1]); }
GIN3_REGISTER("1588_0672", 2, run)
