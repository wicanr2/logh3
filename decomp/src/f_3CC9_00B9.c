/*
 * 3CC9:00B9  32 位元有號比較
 * 呼叫：far，暫存器傳參。比較 AX:BX 與 CX:DX，結果在旗標，呼叫端接 jg／jl／je 等。
 * AX 在回來時不變（函式內 push／pop ax）。注意：docs/internals/51 寫「相同時 AX 變 −1／0／1」不對，
 * 差異測試證實 AX 不變，輸出是旗標。
 *
 * 旗標的來源：
 *   高位字不同：16 位元 cmp ax,cx 的旗標。
 *   高位字相同、低位字相等：ZF=1。
 *   低位字 AX:BX 大於（無號）：CF=0 ZF=0 SF=0 OF=0。
 *   低位字小於：CF=1 SF=1 ZF=0 OF=0（dec 不改 CF）。
 * 回傳值是 OF、SF、ZF、CF 四個旗標（跳躍條件只用到這四個）。
 */
#include "gin3_reg.h"

#define F_CF 0x0001u
#define F_ZF 0x0040u
#define F_SF 0x0080u
#define F_OF 0x0800u

uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx)
{
    if (ax != cx) {
        uint16_t r = (uint16_t)(ax - cx);
        uint16_t f = 0;

        if (ax < cx)
            f |= F_CF;
        if (r & 0x8000u)
            f |= F_SF;
        if (((ax ^ cx) & (ax ^ r)) & 0x8000u)
            f |= F_OF;
        return f;
    }
    if (bx == dx)
        return F_ZF;
    if (bx > dx)
        return 0;
    return F_SF | F_CF;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3CC9_00B9(a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3CC9_00B9", 4, run)
