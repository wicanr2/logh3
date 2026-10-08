/*
 * 2376:0150  畫面格表的格位址計算（near，暫存器傳參）
 *
 * 輸入：CL、CH（BL、BH 由呼叫端搬進來）；AX、DX 只被保存。
 * 輸出：DI = D81E + 0xF0 * CL + 6 * CH（16 位元回繞）；AX 與 DX 回來時不變。
 *       原版用 push ax／pop ax 保存 AX，兩個 mul 都是 8 位元乘法（AL * r/m8 → AX）。
 *       旗標被 mul／add 改過，三個呼叫者（2376:00B2、00E4、0112）都不使用。
 *
 * 語意（強推論）：D81E 是畫面格表的起點（40 欄 × 20 列、一格 6 bytes，一列 0xF0 bytes），
 *   CL 是列、CH 是欄，結果是該格在表中的位移。證據：docs/internals/42-screen-tile-table.md、
 *   docs/re/137-message-box.md 的「ds:D81E ＋ 視窗[+6]×0xF0 ＋ 視窗[+5]×6」同一條公式；
 *   呼叫者（00B2 等）以 si += 6、di += 0xF0 逐格逐列走訪，與此一致。
 * 輸入域：CX 的全部 16 位元值；本函式不讀寫記憶體。
 *
 * 這個函式沒有 docs 裡已證實的逐 byte 記載，本檔以原版雙跑為準。
 */
#include "gin3_reg.h"

uint64_t f_2376_0150(uint16_t ax, uint16_t cx, uint16_t dx)
{
    uint8_t row = (uint8_t)cx;         /* CL */
    uint8_t col = (uint8_t)(cx >> 8);  /* CH */
    uint16_t di = (uint16_t)(0xF0u * row + 6u * col + 0xD81Eu);

    /* 回傳依描述檔 pack:ax,dx,di：由高位到低位 */
    return ((uint64_t)ax << 32) | ((uint64_t)dx << 16) | di;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_2376_0150(a[0], a[1], a[2]);
}
GIN3_REGISTER("2376_0150", 3, run)
