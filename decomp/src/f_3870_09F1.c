/*
 * GIN3PS.EXE 3870:09F1  Cohen–Sutherland 的 outcode：點 (x, y) 相對裁切矩形 ds:8D6A、8D6C、8D6E、8D70 的位置碼
 * 呼叫：far，兩個 word 參數 x（[bp+6]）、y（[bp+8]）。矩形：xmin ＝ ds:8D6A、ymin ＝ ds:8D6C、xmax ＝ ds:8D6E、ymax ＝ ds:8D70
 *   （docs/internals/52 第 93 到 95 行，兩端含）。矩形設定者是 3870:089D。
 * 語意（比較全部是有號 16 位元，jge／jle）：
 *   x <  xmin → 位元 0x01（左）；否則 x > xmax → 0x02（右）
 *   y <  ymin → 位元 0x08（上）；否則 y > ymax → 0x04（下）
 * 回傳 AX：低位元組是 outcode；高位元組是 y 的高位元組（最後一次載入 AX 是 `mov ax,dx`，也就是 y，之後只改 AL），
 *   所以 AX ＝ (y & 0xFF00) | outcode。呼叫者 3870:08FF 與 3870:0A30 只用 AL（強推論，描述檔仍比對整個 AX）。
 * 不寫記憶體。
 * 輸入域：全部 16 位元的 x、y 與矩形值。推論等級：已證實（asm 與差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_3870_09F1(Mem *m, uint16_t x, uint16_t y)
{
    uint8_t code = 0;

    if ((int16_t)x < (int16_t)rd16(m, DS(0x8D6A)))
        code |= 1;
    else if ((int16_t)x > (int16_t)rd16(m, DS(0x8D6E)))
        code |= 2;
    if ((int16_t)y < (int16_t)rd16(m, DS(0x8D6C)))
        code |= 8;
    else if ((int16_t)y > (int16_t)rd16(m, DS(0x8D70)))
        code |= 4;
    return (uint16_t)((y & 0xFF00u) | code);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3870_09F1(m, a[0], a[1]);
}
GIN3_REGISTER("3870_09F1", 2, run)
