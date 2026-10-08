/*
 * 3CC9:0009  32 位元乘法（Datalight 的 long 乘法輔助函式）
 * 呼叫：far，暫存器傳參。AX:BX（高:低）× CX:DX（高:低），結果回 AX:BX，取低 32 位元。
 * 其他暫存器（SI、DI、DX）被破壞，呼叫端不依賴它們（150 個呼叫者）。語意見 docs/internals/51。
 */
#include "gin3_reg.h"

uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx)
{
    uint32_t a = ((uint32_t)ax << 16) | bx;
    uint32_t b = ((uint32_t)cx << 16) | dx;

    return gin3_lmul(a, b); /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3CC9_0009(a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3CC9_0009", 4, run)
