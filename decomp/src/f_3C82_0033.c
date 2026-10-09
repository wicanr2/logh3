/*
 * 3C82:0033  atoi(s)（Datalight C runtime，16 位元結果）
 * 呼叫：far，堆疊傳參：s 位移、s 段。回傳 AX ＝ 16 位元整數（有號解讀）。
 * 做法：
 *   1. 略過前導空白：0x20 與 0x09..0x0D。
 *   2. 一個可選的 '+' 或 '-'（'-' 設負號旗標）。
 *   3. 連續十進位數字 '0'..'9'：累加值 acc ← acc × 10 − 數字（16 位元回繞）。原版累加的是負值，
 *      最後正數結果取補數（neg）、負數結果直接回傳；所以 16 位元溢位時的回繞結果與一般 atoi 不同處由此而來，
 *      本版逐步照做（mul bx 取低 16 位元）。
 *   4. 遇到 NUL 或任何非數字字元就停止。
 * 位移在段內 16 位元回繞（lodsb 的 SI 回繞）。沒有任何數字時回 0。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_3C82_0033(Mem *m, uint16_t off, uint16_t seg)
{
    uint16_t si = off, cx = 0, neg = 0;
    uint8_t al;

    for (;;) {
        al = rd8(m, LIN(seg, si));
        si++;
        if (al == 0x20)
            continue;
        if (al < 9)
            break;
        if (al <= 0x0D)
            continue;
        break;
    }
    if (al == '+' || al == '-') {
        if (al == '-')
            neg = 1;
        al = rd8(m, LIN(seg, si));
        si++;
    }
    for (;;) {
        uint8_t d;

        if (al == 0)
            break;
        if (al < '0')
            break;
        d = (uint8_t)(al - '0');
        if (d > 9)
            break;
        cx = (uint16_t)((uint16_t)(cx * 10u) - d);
        al = rd8(m, LIN(seg, si));
        si++;
    }
    return neg ? cx : (uint16_t)(0u - cx);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3C82_0033(m, a[0], a[1]);
}
GIN3_REGISTER("3C82_0033", 2, run)
