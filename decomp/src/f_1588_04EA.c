/*
 * GIN3PS.EXE 1588:04EA  戰場像素矩形（x, y）起的 16 × 16 範圍是否全部可通行（回 AX：1 ＝ 全部通行，0 ＝ 有一格被擋）
 *
 * 原版（135 bytes，far 呼叫，兩個堆疊字；push bp／si／di、sub sp,0Ch，所以 x 在 [bp+16h]、y 在 [bp+18h]）：
 *   x0 ← x >> 3（邏輯右移 3）            y0 ← y ÷ 10（無號 div）
 *   x1 ← (x + 15) >> 3（加法 16 位元回繞） y1 ← (y + 15) ÷ 10（加法 16 位元回繞，無號 div）
 *   對 c ＝ x0 .. x1（無號、含端點，`ja` 結束），對 r ＝ y0 .. y1（同）：
 *     1588:0621(c, r)（push ax(r) 先、push c 後）≠ 0 → 回 0（該半格被擋，不再看其餘格）
 *   全部為 0（含迴圈一次都不進入）→ 回 1。
 *   回傳 AX 是 0 或 1，AH ＝ 0（`xor ax,ax`／`mov ax,1`）。
 * 0621 的 (c, r) 是「半格」座標（x ÷ 8、y ÷ 10，docs/re/262）；本函式把像素矩形 [x, x+15] × [y, y+15] 蓋到的半格逐一問 0621。
 *   寬 16 像素最多蓋到 3 欄（x0 ... x0+2），高 16 像素最多蓋到 3 列，所以最多呼叫 9 次。
 * 沒有任何寫入（0621 也不寫）。
 * 呼叫者（台帳）：1588:0891（game-logic）、2583:022B（hw-dependent）等四處。docs/spec/024-battle-move-order.md 第 15 行
 *   「那一格過不去（`rules.BattleTerrain.Passable`，＝ 原版 `1588:04EA`）→ 取消」、docs/internals/44-battle-command-menus.md 第 95 行
 *   「`1588:04EA` 的地形判斷」、docs/re/330 第 284 行「r ← 1588:04EA(x,y)（地圖可站）」，與本檔一致。
 * 輸入域：x、y 全部 16 位元值（x + 15 與 y + 15 回繞時 x1 < x0 或 y1 < y0，迴圈不執行、回 1）；地圖與圖塊遮罩表取自記憶體。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_0621(Mem *m, uint16_t c, uint16_t r);

uint16_t f_1588_04EA(Mem *m, uint16_t x, uint16_t y)
{
    uint16_t x0 = (uint16_t)(x >> 3);
    uint16_t y0 = (uint16_t)(y / 10u);
    uint16_t x1 = (uint16_t)((uint16_t)(x + 15u) >> 3);
    uint16_t y1 = (uint16_t)((uint16_t)(y + 15u) / 10u);
    uint16_t c, r;

    for (c = x0; c <= x1; c++) {
        for (r = y0; r <= y1; r++) {
            if (f_1588_0621(m, c, r) != 0)
                return 0;
        }
    }
    return 1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_04EA(m, a[0], a[1]); }
GIN3_REGISTER("1588_04EA", 2, run)
