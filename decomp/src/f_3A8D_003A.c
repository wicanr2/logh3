/*
 * GIN3PS.EXE 3A8D:003A  BX = y × 320 + x + 0x0C80（16 位元；CH = x、CL = y）
 *
 * 呼叫：far，暫存器傳參，沒有堆疊參數、沒有 C 序言（手寫組語，不是 Datalight C 編出來的）。
 *   輸入 CH = x（位元組）、CL = y（位元組）；其餘暫存器不看，也不讀記憶體。
 *   輸出 BX = (y × 320 + x + 0x0C80) mod 65536。唯二的呼叫者 348A:01DC、348A:0228 取 BX 存成落點位移
 *   （mov cs:[415EA],bx），CX 之後重設。
 * 無記憶體讀寫，無副作用。
 *
 * 語意：
 *   已證實（差異測試）：BX = (uint16_t)(y × 320 + x + 0x0C80)。y × 320 在 y >= 205 時超過 16 位元，原版回繞
 *   （shl 位元組進高位元組、右移兩位再相加，等於 y × 256 + y × 64，16 位元加法回繞）。
 *   與文件相符（docs/internals/20 第 52 行、docs/re/328 第 76 行）：文件記載 3A8D:003A(x, y) 的結果是
 *   「VRAM 基底 = y×320 + x + 0x0C80」，與上面差異測試證明的算術一致。
 *   強推論：CH 為 x、CL 為 y，由 348A:01DC／0228 的 mov ch,[bp+6]／mov cl,[bp+8] 與文件的 3A8D:003A(x, y) 對照得到。
 *   未知：BX 在畫面上的解讀。0x0C80 ＝ 10 × 320 與 320 是否為列寬沒有證據；文件舉的例子
 *   （x＝1、y＝4 → 像素 (8,56)）與把 y×320+x+0x0C80 當 320 寬畫面直接解讀的結果（像素 (1,14)）不吻合，
 *   原因（呼叫端是否先換算參數、VRAM 排列方式）未查證。本檔只對上面的算術負責。
 *
 * 暫存器副作用（asm 讀得，差異測試只比對 BX）：AX ＝ y << 8、CX ＝ x（零擴展）。兩個呼叫者隨後重設 CX；
 * 呼叫者之後對 AX 的使用沒有追查。
 *
 * 輸入域：CX 為全 16 位元；y 全 0..255，x 全 0..255。
 */
#include "gin3_reg.h"

uint16_t f_3A8D_003A(uint8_t x, uint8_t y)
{
    uint16_t row = (uint16_t)((uint16_t)y << 8);             /* mov ah,cl; xor al,al */
    uint16_t row_x64 = (uint16_t)(row >> 2);                 /* shr bx,1 兩次（邏輯右移） */
    uint16_t offset = (uint16_t)(row + row_x64);             /* y × 320，16 位元回繞 */

    offset = (uint16_t)(offset + x);                         /* + ch */
    return (uint16_t)(offset + 0x0C80u);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3A8D_003A((uint8_t)(a[0] >> 8), (uint8_t)(a[0] & 0xFF));
}
GIN3_REGISTER("3A8D_003A", 1, run)
