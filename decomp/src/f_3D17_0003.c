/*
 * 3D17:0003  正規化 far 指標（seg:off → 位移 0..15 的等價 seg:off）
 * 呼叫：far，暫存器傳參。輸入 AX＝段、BX＝位移。CX 原值保留，DX 不動。
 * 輸出：AX ＝ 新段 ＝ 線性位址 >> 4（取低 16 位元），BX ＝ 新位移 ＝ 線性位址 & 0xF。
 *   線性位址 ＝ 段 × 16 ＋ 位移，遮到 20 位元（超過 1 MB 的部分回繞到 0）。
 * 先呼叫 3D17:001F 取得 21 位元線性位址（AL:BX），再把它重新拆成段與位移。
 * 呼叫者：配置器 3CD9:0047 與 3CD9:01FC。旗標沒有被呼叫者使用，不在比對範圍。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

extern uint64_t f_3D17_001F(uint16_t ax, uint16_t bx, uint16_t cx);

uint64_t f_3D17_0003(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx)
{
    uint64_t r = f_3D17_001F(ax, bx, cx);
    uint16_t al = (uint16_t)(r >> 32);     /* 001F 回來的 AX（AH ＝ 0） */
    uint16_t lo = (uint16_t)(r >> 16);     /* 001F 回來的 BX */
    uint16_t seg = (uint16_t)((((al << 4) & 0xFFu) << 8) | (lo >> 4)); /* shl ah,4；shr bx,4；or ax,bx */
    uint16_t off = (uint16_t)(lo & 0xFu);                                /* mov bl,ch；and bx,0Fh */

    return ((uint64_t)seg << 48) | ((uint64_t)off << 32) | ((uint64_t)cx << 16) | dx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3D17_0003(a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3D17_0003", 4, run)
