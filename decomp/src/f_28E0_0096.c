/*
 * GIN3PS.EXE 28E0:0096  能力值換軍階等級：把一個 byte 依門檻分成 0..11（far，一個參數，回 AL，不讀寫記憶體）
 *
 * 原版：far 呼叫，一個堆疊字（args[0]），push bp、mov bp,sp，參數在 [bp+6]：
 *   args[0] ＝ v（word，只取低位元組：mov cl,[bp+6]；高位元組被忽略）
 * 無號 byte 比較（jnb），第一個成立的區間：
 *   v <  2 → 0      v <  4 → 1      v <  6 → 2      v <  8 → 3
 *   v < 12 → 4      v < 16 → 5      v < 20 → 6      v < 24 → 7
 *   v < 29 → 8      v < 34 → 9      v < 39 → 10     其餘 → 11
 *   （門檻十六進位：2 4 6 8 0C 10 14 18 1D 22 27）
 * 回傳只有 AL。「v < 2」那條路徑用 xor ax,ax 清 AX，其餘路徑只寫 AL（mov al,N），AH 保持進入時的值，
 *   不是函式的輸出；差異測試用 ret = pack:al 只比 AL（比照 2E54:088F）。呼叫端 28E0:004F、28E0:05F9 的使用方式
 *   不在本檔查證範圍。
 * 語意：docs/internals/08 第 217 行：「原版28E0:004F的下限為0、2、4、6、8、12、16、20、24、29、34、39」，
 *   與本檔 12 個區間的下限一致；docs/re/418 第 37 行起的升級流程也呼叫這一支。本檔的證據是反組譯與差異測試。
 * 副作用：無。輸入域：v 低位元組 0..255 全域，高位元組任意。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。「能力值換軍階」的用途說明是 docs 的推論。
 */
#include "gin3_reg.h"

uint8_t f_28E0_0096(Mem *m, uint16_t v_word)
{
    uint8_t v = (uint8_t)v_word;

    (void)m;
    if (v < 0x02u)
        return 0;
    if (v < 0x04u)
        return 1;
    if (v < 0x06u)
        return 2;
    if (v < 0x08u)
        return 3;
    if (v < 0x0Cu)
        return 4;
    if (v < 0x10u)
        return 5;
    if (v < 0x14u)
        return 6;
    if (v < 0x18u)
        return 7;
    if (v < 0x1Du)
        return 8;
    if (v < 0x22u)
        return 9;
    if (v < 0x27u)
        return 10;
    return 11;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_0096(m, a[0]); }
GIN3_REGISTER("28E0_0096", 1, run)
