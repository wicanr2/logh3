/*
 * GIN3PS.EXE 2AD2:0151  艦艇池挑第一個空槽（first-fit）
 *
 * 原版：far 呼叫，沒有參數，回 AX：
 *   池基底＝遠指標 (ds:CFB4 位移、ds:CFB6 段)；每筆 78 bytes（0x4E），共 0x160＝352 筆
 *   從索引 0 往上掃，第一個 池[i][+1] == 0xFF 的索引 i 就回傳；352 筆都不是 0xFF 回 0xFFFF（池滿）。
 * 迴圈條件 i < 0x160 是無號比較；位移累加是 16 位元，超過 0xFFFF 在同一段內回繞。
 * 語意見 docs/internals/64-ship-pool.md、docs/re/344 §3.1（「+0x01 == 0xFF 即空槽」：已證實）。
 * 池滿回 0xFFFF 之後，呼叫端沒有檢查（docs/re/344 §「回 0xFFFF 之後呼叫端沒檢查」，原版行為，不在本函式範圍）。
 * 副作用：無（不寫記憶體）。輸入域：池指標任何值（讀至多 352 個 byte），但池範圍落在 0xA0000..0xAFFFF（VGA 視訊記憶體）
 *   的部分，原版與基底傾印不一致（強推論：與 docs/re/445「限制」一致，未直接觀察），本版與差異測試都不涵蓋（框架限制）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2AD2_0151(Mem *m)
{
    uint16_t off = rd16(m, DS(0xCFB4));
    uint16_t seg = rd16(m, DS(0xCFB6));
    uint16_t i;

    for (i = 0; i < 0x160u; i++) {
        if (rd8(m, LIN(seg, off + 1)) == 0xFF)
            return i;
        off = (uint16_t)(off + 0x4E);
    }
    return 0xFFFF;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2AD2_0151(m);
}
GIN3_REGISTER("2AD2_0151", 0, run)
