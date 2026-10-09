/*
 * 36AD:0261  滑鼠按鍵狀態位元重排（near，輸入在暫存器 BL，輸出 AX；純函式）
 *
 * 原版（逐指令）：
 *   push bx; al = bl; al &= 1; al ^= 1; al = ror(al, 1);
 *   bl &= 2; bl ^= 2; bl <<= 4;（8 位元）; al |= bl; ah = 0; pop bx; retn
 * 輸入：BL 的 bit0、bit1；BL 其餘位元與 BH 不讀。
 * 輸出：AX = ((bit0 ^ 1) 的 ror 1) | (((bit1) ^ 1) << 5)，亦即
 *   bit0 = 0 → AL bit7 = 1；bit1 = 0 → AL bit5 = 1；值域 {0x00, 0x20, 0x80, 0xA0}，AH = 0。BX 由 push／pop 還原。
 * 唯一呼叫者 36AD:0076（sub_436B6）在 `int 33h` AX=3 取得滑鼠按鍵狀態 BX 之後呼叫本函式，
 *   結果 AX 寫進呼叫者參數指到的位置。BL bit0／bit1 對應滑鼠左／右鍵（INT 33h 的規格），
 *   輸出把「放開」映成設定位元（反相）；與 INT 33h 對應是文件推論，非本檔證實（強推論）。
 *
 * 輸入域：BX 的全部 16 位元值。描述檔 ret = pack:ax,bx，連同 BX 還原一併比對。
 * 推論等級：位元運算已證實（反組譯逐指令＋差異測試）；按鍵語意為強推論。
 */
#include "gin3_reg.h"

uint16_t f_36AD_0261(uint16_t bx)
{
    uint8_t bl = (uint8_t)bx;
    uint8_t al = (uint8_t)(bl & 1u);

    al = (uint8_t)(al ^ 1u);
    al = (uint8_t)((al >> 1) | (al << 7)); /* ror al,1 */
    bl = (uint8_t)(bl & 2u);
    bl = (uint8_t)(bl ^ 2u);
    bl = (uint8_t)(bl << 4); /* shl bl,4（8 位元） */
    al = (uint8_t)(al | bl);
    return al; /* xor ah,ah */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    /* pack:ax,bx：AX 在高位，BX 在低位；BX 經 push／pop 還原為輸入值 */
    return ((uint64_t)f_36AD_0261(a[0]) << 16) | a[0];
}
GIN3_REGISTER("36AD_0261", 1, run)
