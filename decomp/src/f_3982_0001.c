/*
 * GIN3PS.EXE 3982:0001  圖塊池取圖塊：回傳遠指標 池基底 + 索引 × 128
 *
 * 呼叫：far，一個參數 [bp+8]（函式先 push bp、push si，所以參數在 bp+8），原版只讀它的低位元組
 * （mov dl,[bp+8]; mov al,dl; xor ah,ah，零擴展），回遠指標：AX ＝ 段、BX ＝ 位移。
 * 讀 ds:F12C（池基底位移，word）與 ds:F12E（池基底段，word）；無寫入。10 個呼叫者。
 * 這個 C 版回傳 (段 << 16) | 位移，與 ret=pack:ax,bx（AX 在高位、BX 在低位）一致。
 *
 * 語意：
 *   已證實（差異測試）：位移 = (uint16_t)(ds:F12C + (索引 << 7))，段 = ds:F12E；索引零擴展、左移 7 位不會超過 16 位元，
 *   但與 ds:F12C 相加時 16 位元回繞，不進位到段。
 *   已證實（docs/re/376 proven、docs/internals/10）：DS:F12C／F12E 是 SPRITE.SPR 的圖塊池基底，每塊 128 bytes。
 *   本函式不檢查索引是否超出池的塊數。
 *   組語另外弄髒 CX（＝ 段）與 DX（低位元組 ＝ 索引），呼叫者不用（asm 讀得，差異測試只比對 AX、BX）。
 *
 * 輸入域：參數字的高位元組原版不看，差異測試以全 16 位元隨機值餵入；索引 0..255 全數涵蓋。
 */
#include "gin3_reg.h"

uint32_t f_3982_0001(Mem *m, uint8_t index)
{
    uint16_t pool_off = rd16(m, DS(0xF12C));
    uint16_t pool_seg = rd16(m, DS(0xF12E));
    uint16_t block = (uint16_t)((uint16_t)index << 7);
    uint16_t off = (uint16_t)(pool_off + block);

    return ((uint32_t)pool_seg << 16) | off;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3982_0001(m, (uint8_t)a[0]);
}
GIN3_REGISTER("3982_0001", 1, run)
