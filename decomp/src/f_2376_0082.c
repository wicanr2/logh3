/*
 * 2376:0082  清空畫面格表（far，無參數，DS = DGROUP）
 *
 * 原版：ES = DS，DI = 0xD81E，CX = 0x960，`cld; xor ax,ax; rep stosw`，把 ds:D81E 起 0x960 個 word
 *   （4800 bytes）全部寫成 0。DI 由 push／pop 保存，其餘暫存器的結果沒人用（呼叫者 341B:0004 之後
 *   直接呼叫下一個函式），描述檔 ret = none，只比對記憶體寫入。
 *   ES 被改成 DS 沒有還原：這是原版的副作用，不在比對範圍（ES 不是可比對的暫存器）。
 * 輸出：ds:D81E 到 ds:EADD 全部變 0。不讀記憶體。
 *
 * 語意：ds:D81E 是 40 欄 × 20 列、一格 6 bytes 的畫面格表（docs/internals/42-screen-tile-table.md、
 *   docs/internals/47-window-frame-tiles.md 第 87 行：「畫面格表 ds:D81E（40 × 20 格，一格 6 bytes）」），
 *   40 × 20 × 6 = 4800 = 0x12C0 bytes = 0x960 個 word，與本函式清除的範圍吻合（已證實）。
 *   這個函式是開機初始化（341B:0004）的一步。
 *
 * 輸入域：沒有輸入。差異測試用種子預先把表內外的 byte 設成非 0，才看得出少寫、多寫與起點偏移。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；表的用途是文件結論。
 */
#include "gin3_reg.h"

void f_2376_0082(Mem *m)
{
    uint16_t di = 0xD81Eu;
    uint16_t cx;

    for (cx = 0x960u; cx != 0; cx--) {
        wr16(m, DS(di), 0);
        di = (uint16_t)(di + 2u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2376_0082(m);
    return 0;
}
GIN3_REGISTER("2376_0082", 0, run)
