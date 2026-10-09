/*
 * 3C9C:000A  memset(dst, val, count)（Datalight C runtime 的位元組填充）
 * 呼叫：far，堆疊傳參（由低位址到高位址）：dst 位移、dst 段、val（word，只用低位元組 AL）、count（word）。
 * 回傳：AX＝dst 段、BX＝dst 位移（不是 DX:AX，也不是填完之後的 DI）。
 * 副作用：從 dst 起連續寫 count 個位元組為 val 的低位元組；位移是 16 位元，超過 0xFFFF 在同一段內回繞
 * （rep stosb 的 DI 回繞）。count 為 0 不寫任何東西。DF 清除（cld），其餘暫存器不保證。
 * 語意等價於標準 C 的 memset，但回傳的是 seg:off 對（AX:BX），不是單一指標。
 * 輸入域：全部 16 位元值。count 很大時會寫到呼叫者看不到的鄰近段，本版與原版一樣照寫。
 * 證據：asm（les di,[bp+6]；mov bx,di；mov ax,[bp+0Ah]；mov cx,[bp+0Ch]；cld；rep stosb；mov ax,es），
 * 推論等級：已證實（差異測試）。docs/re 內有許多 3C9C:000A 的呼叫點，用法都是填 0 或 0xFF。
 */
#include "gin3_reg.h"

uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count)
{
    uint16_t i;

    for (i = 0; i < count; i++)
        wr8(m, LIN(seg, (uint16_t)(off + i)), (uint8_t)val);
    return ((uint32_t)seg << 16) | off; /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3C9C_000A(m, a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3C9C_000A", 4, run)
