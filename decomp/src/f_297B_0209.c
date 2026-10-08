/*
 * GIN3PS.EXE 297B:0209  提督等級 → 艦隊編制槽數（應有組數）
 *
 * 原版：far 呼叫，一個參數＝一個 word（呼叫端推入 es:[提督+6] 整個 word，函式只讀低位元組）。
 * 回傳：只寫 AL，AH 保持進入時的值（mov al,dl 之後 retf）；兩個呼叫者（297B:0066、297B:05B2）
 *   都只取 AL（mov es:[..+0Eh],al）。差異測試比較整個 AX，所以用 kind=reg 的 ax_in 把進入時的 AX
 *   也餵給 C 版，回傳 (ax_in & 0xFF00) | AL，忠實表達「AH 原樣穿過」。
 * 語意（已證實：反組譯逐指令，與 docs/internals/33 的階梯一致）：
 *   等級（低位元組，無號比較 jnb）：<2→1、<6→2、<16→3、<20→4、<24→6、<29→7、<34→8、<39→9、其餘→10（沒有 5）
 * args 順序：[0]＝ax_in（暫存器），[1]＝堆疊上的 word。
 * 副作用：無，不讀記憶體。輸入域：全部 16 位元值。
 */
#include "gin3_reg.h"

uint16_t f_297B_0209(Mem *m, uint16_t ax_in, uint16_t level_word)
{
    uint8_t level = (uint8_t)level_word;
    uint8_t slots;

    (void)m;
    if (level < 2)
        slots = 1;
    else if (level < 6)
        slots = 2;
    else if (level < 16)
        slots = 3;
    else if (level < 20)
        slots = 4;
    else if (level < 24)
        slots = 6;
    else if (level < 29)
        slots = 7;
    else if (level < 34)
        slots = 8;
    else if (level < 39)
        slots = 9;
    else
        slots = 10;
    return (uint16_t)((ax_in & 0xFF00u) | slots);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_0209(m, a[0], a[1]); }
GIN3_REGISTER("297B_0209", 2, run)
