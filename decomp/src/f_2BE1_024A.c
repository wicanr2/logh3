/*
 * GIN3PS.EXE 2BE1:024A  兩個 byte 角度的無號絕對差（不折回）
 *
 * 原版：far 呼叫，兩個堆疊字（push bp、沒有 sub sp，所以參數在 [bp+6] 與 [bp+8]），回 AX。
 *   只取兩個參數的低位元組（mov dl,[bp+6]、mov bl,[bp+8]），高位元組被忽略。
 *   a >= b（無號 byte 比較，jb 不成立）→ 回 a − b
 *   a <  b → 回 b − a
 *   結果 0 到 255，AH = 0（兩條路徑都先 xor ah,ah 再相減）。
 *
 * 輸入域：兩個參數的全部 16 位元值（只有低位元組有效）。純函式，不讀不寫記憶體。
 * 證據：asm 逐指令。docs/re/88 第 24 行「無號的絕對差」、docs/re/330 第 315 行「|a−b|（byte，不折回）」與本檔一致。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2BE1_024A(Mem *m, uint16_t a_word, uint16_t b_word)
{
    uint8_t a = (uint8_t)a_word;
    uint8_t b = (uint8_t)b_word;

    (void)m;
    if (a >= b)
        return (uint16_t)(a - b);
    return (uint16_t)(b - a);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2BE1_024A(m, a[0], a[1]); }
GIN3_REGISTER("2BE1_024A", 2, run)
