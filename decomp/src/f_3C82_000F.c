/*
 * 3C82:000F  abs(x)（16 位元整數絕對值）
 * 呼叫：far，堆疊傳參：x（word，視為有號）。回傳 AX ＝ |x|。
 * 做法：or ax,ax；jns；neg ax。x ＝ 0x8000 時 neg 回到 0x8000（與標準 C 的 abs(INT_MIN) 同為未定義，原版回 0x8000）。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_3C82_000F(uint16_t x)
{
    return (x & 0x8000u) ? (uint16_t)(0u - x) : x;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3C82_000F(a[0]);
}
GIN3_REGISTER("3C82_000F", 1, run)
