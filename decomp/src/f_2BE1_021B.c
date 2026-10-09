/*
 * GIN3PS.EXE 2BE1:021B  兩個 byte 角度的有號差（方位 − 朝向），正規化到約 −128 到 128
 *
 * 原版：far 呼叫，兩個堆疊字（push bp、push si、沒有 sub sp，所以參數在 [bp+8] 與 [bp+0Ah]），回 AX。
 *   只取兩個參數的低位元組（mov dl,[bp+8]、mov bl,[bp+0Ah]），高位元組被忽略。
 *   d ← a − b（兩個零延伸的 byte 相減，16 位元；範圍 −255 到 255）
 *   d >= 0x80（有號比較，jl）→ d ← d + 0xFF00（即 d − 256）
 *   否則 d > −128（有號比較，jg，比較值 0xFF80）→ d 不變
 *   否則（d <= −128）→ d ← d + 0x100
 *   回 d（16 位元；負值以補數表示，高位元組為 0xFF）。
 *   注意邊界不對稱：d = 128 → −128，d = −128 → +128。
 *
 * 輸入域：兩個參數的全部 16 位元值（只有低位元組有效）。純函式，不讀不寫記憶體。
 * 證據：asm 逐指令。docs/re/129 第 44 行「≥ 0x80 減 0x100、≤ 0xFF80 加 0x100」與本檔一致；
 *   同檔第 34 行的範圍註記「−128..127」不精確：輸出範圍是 −128 到 128，d = −128（例如 a=0、b=128）回 +128（0x0080）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2BE1_021B(Mem *m, uint16_t a_word, uint16_t b_word)
{
    uint16_t d = (uint16_t)((uint16_t)(uint8_t)a_word - (uint16_t)(uint8_t)b_word);

    (void)m;
    if ((int16_t)d >= 0x80)
        return (uint16_t)(d + 0xFF00u);
    if ((int16_t)d > (int16_t)0xFF80)
        return d;
    return (uint16_t)(d + 0x100u);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2BE1_021B(m, a[0], a[1]); }
GIN3_REGISTER("2BE1_021B", 2, run)
