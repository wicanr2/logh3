/*
 * 3CA9:0002  memmove(dst, src, count)（Datalight C runtime，重疊安全的位元組搬移）
 * 呼叫：far，堆疊傳參（由低位址到高位址）：dst 位移、dst 段、src 位移、src 段、count（word）。
 * 回傳：AX＝dst 段、BX＝dst 位移。
 * 方向由「兩個位移」的無號比較決定，不比較段，也不比較線性位址：
 *   src 位移 >= dst 位移：由低往高逐位元組複製（cld；rep movsb）。
 *   src 位移 <  dst 位移：先把兩個位移各加 (count−1)，再由高往低逐位元組複製（std；rep movsb）。
 * 位移是 16 位元，在各自的段內回繞。段不同而線性位址重疊時，方向仍照位移決定，結果不一定是
 * 「正確的 memmove」，本版逐位元組照做，與原版一致。
 * 語意等價於標準 C 的 memmove（同段時）。count 為 0 不寫任何東西。
 * 輸入域：全部 16 位元值。
 * 證據：asm（cmp si,di；jnb；add si,cx；dec si；add di,cx；dec di；std；rep movsb）。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

uint32_t f_3CA9_0002(Mem *m, uint16_t doff, uint16_t dseg, uint16_t soff, uint16_t sseg, uint16_t count)
{
    uint16_t si = soff, di = doff, i;

    if (si >= di) {
        for (i = 0; i < count; i++)
            wr8(m, LIN(dseg, (uint16_t)(di + i)), rd8(m, LIN(sseg, (uint16_t)(si + i))));
    } else {
        si = (uint16_t)(si + count - 1);
        di = (uint16_t)(di + count - 1);
        for (i = 0; i < count; i++)
            wr8(m, LIN(dseg, (uint16_t)(di - i)), rd8(m, LIN(sseg, (uint16_t)(si - i))));
    }
    return ((uint32_t)dseg << 16) | doff; /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3CA9_0002(m, a[0], a[1], a[2], a[3], a[4]);
}
GIN3_REGISTER("3CA9_0002", 5, run)
