/*
 * 3CC9:0020  32 位元無號除法
 * 呼叫：far，暫存器傳參。AX:BX ÷ CX:DX，商回 AX:BX，餘數回 CX:DX。
 * 除數為 0：商 0xFFFFFFFF、餘數 0。語意見 docs/internals/51。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

uint64_t f_3CC9_0020(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx)
{
    uint32_t q, r;

    gin3_uldivmod(((uint32_t)ax << 16) | bx, ((uint32_t)cx << 16) | dx, &q, &r);
    return ((uint64_t)q << 32) | r; /* AX:BX:CX:DX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3CC9_0020(a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3CC9_0020", 4, run)
