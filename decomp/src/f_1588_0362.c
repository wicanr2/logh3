/*
 * GIN3PS.EXE 1588:0362  戰場畫面座標換成世界座標（含範圍檢查）
 *
 * 原版：far 呼叫，四個參數：畫面 x（word）、畫面 y（word）、x 的輸出指標（far）、y 的輸出指標（far）。回 AX。
 * 沒有 push si／di，序言只有 push bp：參數從 [bp+6] 起。
 *
 * 語意（已證實：與反組譯逐條對過；呼叫端用途見 docs/re/121「畫面 → 世界；回 0 ＝ 點在戰場那一塊外面」）：
 *   若 x < 0x10 或 y < 0x14 或 x >= 0x1D0 或 y >= 0x17C（四個比較都是無號）：回 0，輸出指標不寫。
 *   否則 *out_x ← x − 0x10 ＋ ds:9BF4（16 位元回繞）；*out_y ← y − 0x14 ＋ ds:9BF6（16 位元回繞）；回 1。
 *   寫入順序：先 x 後 y（兩個輸出指標指到同一處時，最後留下 y 的值）。
 *   ds:9BF4／ds:9BF6 是戰場視野左上角的世界座標（docs/re/326：9BF4 ＝ 格×16、9BF6 ＝ 格×20）。
 *   (0x10, 0x14) 是戰場視窗在滑鼠座標裡的左上角（docs/re/326 §二）；0x1D0／0x17C 即 16＋448、20＋360。
 *
 * 輸入域：x、y 取 16 位元全域；輸出指標指向可寫的緩衝區（差異測試以暫存緩衝區與一組重疊指標驗證）。
 */
#include "gin3_reg.h"

uint16_t f_1588_0362(Mem *m, uint16_t x, uint16_t y, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off, uint16_t oy_seg)
{
    if (x < 0x10u || y < 0x14u || x >= 0x1D0u || y >= 0x17Cu)
        return 0;
    wr16(m, LIN(ox_seg, ox_off), (uint16_t)(x + 0xFFF0u + rd16(m, DS(0x9BF4))));
    wr16(m, LIN(oy_seg, oy_off), (uint16_t)(y + 0xFFECu + rd16(m, DS(0x9BF6))));
    return 1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_0362(m, a[0], a[1], a[2], a[3], a[4], a[5]); }
GIN3_REGISTER("1588_0362", 6, run)
