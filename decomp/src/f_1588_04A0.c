/*
 * GIN3PS.EXE 1588:04A0  把視野左上角的格座標夾進戰場範圍
 *
 * 原版：far 呼叫，兩個 far 指標參數（指向 16 位元有號整數），回 AX。沒有 push si／di，參數從 [bp+6] 起。
 *
 * 語意（已證實：與反組譯逐條對過；呼叫端 1588:02CE／03A9／0945 與規則見 docs/re/326「格 x 夾 0–100、格 y 夾 0–84」）：
 *   n ← 0
 *   若 *p < 0：*p ← 0，n++
 *   若 *q < 0：*q ← 0，n++
 *   若 *p > 100：*p ← 100，n++
 *   若 *q > 84：*q ← 84，n++
 *   回 (n < 2) ? 1 : 0       （回 1 ＝ 夾了不到兩次；兩軸都被夾才回 0。docs/re/326：兩軸都撞邊那一圈不重畫）
 *   四個比較都是有號（jns／jle）。每個步驟都重新讀一次指標內容，所以 p、q 指到同一處時順序有影響，本版照做。
 *   因為「小於 0」與「大於上限」互斥，n 最多是 2。
 *
 * 輸入域：兩個指標指向可寫的 16 位元值，內容取 16 位元全域。
 */
#include "gin3_reg.h"

uint16_t f_1588_04A0(Mem *m, uint16_t p_off, uint16_t p_seg, uint16_t q_off, uint16_t q_seg)
{
    uint16_t n = 0;

    if ((int16_t)rd16(m, LIN(p_seg, p_off)) < 0) {
        wr16(m, LIN(p_seg, p_off), 0);
        n++;
    }
    if ((int16_t)rd16(m, LIN(q_seg, q_off)) < 0) {
        wr16(m, LIN(q_seg, q_off), 0);
        n++;
    }
    if ((int16_t)rd16(m, LIN(p_seg, p_off)) > 100) {
        wr16(m, LIN(p_seg, p_off), 100);
        n++;
    }
    if ((int16_t)rd16(m, LIN(q_seg, q_off)) > 84) {
        wr16(m, LIN(q_seg, q_off), 84);
        n++;
    }
    return (int16_t)n < 2 ? 1 : 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_04A0(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("1588_04A0", 4, run)
