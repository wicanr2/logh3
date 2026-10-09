/*
 * GIN3PS.EXE 1588:0571  戰場像素 (x, y) 那一點是否可通行（回 AX：1 ＝ 通行，0 ＝ 被擋）
 *
 * 原版（54 bytes，far 呼叫，兩個堆疊字；push bp／si／di、沒有 sub sp，所以 x 在 [bp+0Ah]、y 在 [bp+0Ch]）：
 *   r ← y ÷ 10（無號 div）；c ← x >> 3（邏輯右移 3）
 *   1588:0621(c, r) ≠ 0 → 回 0；否則回 1（`xor ax,ax`／`mov ax,1`，AH ＝ 0）
 * 1588:04EA 是同型但蓋住一個 16 × 16 範圍；本函式只問一個半格。沒有任何寫入。
 * 呼叫者（台帳）：1F6F:0C11（game-pure）。docs/spec/100-fleet-combat-candidates.md 第 36 行「1588:0571／0621：X/8、Y/10，
 *   再各/2取128欄tile與四子格mask」、docs/re/403-weapon-candidates.md 第 59 行「1588:0571把世界X除8、Y除10；0621再各除2取tile，
 *   tile圖stride128」，與本檔一致。
 * 輸入域：x、y 全部 16 位元值；地圖與圖塊遮罩表取自記憶體。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_0621(Mem *m, uint16_t c, uint16_t r);

uint16_t f_1588_0571(Mem *m, uint16_t x, uint16_t y)
{
    uint16_t r = (uint16_t)(y / 10u);
    uint16_t c = (uint16_t)(x >> 3);

    return f_1588_0621(m, c, r) != 0 ? 0u : 1u;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_0571(m, a[0], a[1]); }
GIN3_REGISTER("1588_0571", 2, run)
