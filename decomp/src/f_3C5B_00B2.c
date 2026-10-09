/*
 * 3C5B:00B2  y × 80 ＋ (x >> 3) ＋ 0x0C80 的位址運算（暫存器傳參）
 * 呼叫：far，暫存器傳參：BX ＝ x、CX ＝ y；AX 的輸入值不使用。
 * 輸出：BX ＝ ((BX >> 3) ＋ CX × 80 ＋ 0x0C80) mod 64K；AX ＝ CX × 80 mod 64K；CX ＝ CX × 16 mod 64K。
 *   （三個暫存器都被改寫，呼叫端不一定用到 AX、CX，這裡全部比對。）
 * 全部是 16 位元回繞運算，shr 是邏輯右移。呼叫者都在 video-dos 類函式，BX 看起來是點陣列位址（每列 80 位元組、
 * 每位元組 8 像素、基底 0x0C80），但本函式只證實算術，用途未證實，推論等級：算術已證實（差異測試）、用途假說。
 */
#include "gin3_reg.h"

uint64_t f_3C5B_00B2(uint16_t bx, uint16_t cx)
{
    uint16_t ax = (uint16_t)(cx * 16u);
    uint16_t cx16 = ax;

    ax = (uint16_t)(ax * 4u);
    ax = (uint16_t)(ax + cx16); /* ＝ 原 CX × 80 */
    bx = (uint16_t)((bx >> 3) + ax + 0x0C80u);
    return ((uint64_t)ax << 32) | ((uint64_t)bx << 16) | cx16; /* AX:BX:CX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3C5B_00B2(a[0], a[1]);
}
GIN3_REGISTER("3C5B_00B2", 2, run)
