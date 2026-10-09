/*
 * 3CD5:003C  srand(seed)：設定 3CD5:000C rand() 使用的 32 位元種子
 * 呼叫：far，堆疊傳參：seed（word）。種子 ds:94DC（低字）＝ seed、ds:94DE（高字）＝ 0。
 * 回傳：AX ＝ 0、BX ＝ seed（暫存器殘值，呼叫端沒有使用，這裡一併比對）。
 * 推論等級：已證實（差異測試）。3CD5:000C 的演算法見 f_3CD5_000C.c。
 */
#include "gin3_reg.h"

uint32_t f_3CD5_003C(Mem *m, uint16_t seed)
{
    wr16(m, DS(0x94DE), 0);
    wr16(m, DS(0x94DC), seed);
    return seed; /* AX<<16 | BX ＝ 0:seed */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3CD5_003C(m, a[0]);
}
GIN3_REGISTER("3CD5_003C", 1, run)
