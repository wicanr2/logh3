/*
 * GIN3PS.EXE 2703:0FD7  星域「民心」項（未安定度公式的 f2）
 *
 * 原版：far 呼叫，一個參數＝星域記錄的 far 指標（位移在前、段在後），回 AX。
 * 語意（強推論：名稱「民心」來自 docs/re/34、43，欄位 +0x7C 的本義未獨立驗證）：
 *   星域 +0x12 == 0（擁有者＝帝國）：取 +0x7C
 *   其餘擁有者（含 2＝無主）        ：取 200 - +0x7C（16 位元有號，+0x7C > 200 時為負）
 *   回傳 ＝ 該值 ÷ 10，有號除法、商向零取整（idiv）。
 * 星域記錄 ds:A79C 起 32 筆 × 126 bytes（docs/internals/22）。
 * 副作用：無。只讀 +0x12 與 +0x7C。
 * 輸入域：全部可到達（除數是常數 10，不會 #DE）。dividend 範圍 -55..255。
 * 證據：反組譯逐指令（docs/re/43 f2 一致）。
 */
#include "gin3_reg.h"

uint16_t f_2703_0FD7(Mem *m, uint16_t region_off, uint16_t region_seg)
{
    uint8_t owner = rd8(m, LIN(region_seg, region_off + 0x12));
    uint8_t value = rd8(m, LIN(region_seg, region_off + 0x7C));
    int16_t base;

    if (owner == 0)
        base = (int16_t)value;
    else
        base = (int16_t)(0xC8 - value);
    return (uint16_t)(int16_t)(base / 10);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2703_0FD7(m, a[0], a[1]); }
GIN3_REGISTER("2703_0FD7", 2, run)
