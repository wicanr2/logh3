/*
 * 3CD5:000C  rand()：線性同餘亂數
 * 呼叫：far，無參數，回 AX。32 位元種子在 ds:94DC（低字）與 ds:94DE（高字）：
 *   seed ← seed × 0x41C64E6D ＋ 0x3039（32 位元回繞）
 *   回傳 ← 新種子的高 16 位元 & 0x7FFF
 * 常數與演算法由專案自己量出（docs/re/87），不是 Datalight 的原始碼。
 */
#include "gin3_reg.h"

uint16_t f_3CD5_000C(Mem *m)
{
    uint32_t seed = rd32(m, DS(0x94DC));

    seed = gin3_lmul(seed, 0x41C64E6Du) + 0x3039u;
    wr32(m, DS(0x94DC), seed);
    return (uint16_t)((seed >> 16) & 0x7FFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_3CD5_000C(m);
}
GIN3_REGISTER("3CD5_000C", 0, run)
