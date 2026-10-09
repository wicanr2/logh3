/*
 * GIN3PS.EXE 2D54:0163  戰略態勢碼（0 到 11）
 *
 * 原版：far 呼叫，無參數，回 AX = 0 到 11。讀三個 byte：
 *   a ＝ ds:AE02 + 0x12 ＝ ds:AE14（星域表 ds:A79C 第 13 筆的 +0x12；星域記錄 126 bytes，13 × 126 ＝ 0x666）
 *   f ＝ ds:8F48
 *   b ＝ ds:AE80 + 0x12 ＝ ds:AE92（星域表第 14 筆的 +0x12）
 * 決策樹（全部是 byte 相等比較，沒有大小比較）：
 *   a == 0：f == 0 → 0；否則 b == 0 → 1、b == 1 → 2、其他 → 3
 *   a == 1：f == 0 → 4；否則 b == 0 → 5、b == 1 → 6、其他 → 7
 *   其他   ：f == 0 → 8；否則 b == 0 → 9、b == 1 → 10、其他 → 11
 * 純函式，不寫記憶體。b 只有在 f != 0 時才讀（f == 0 的路徑不讀 b）。
 *
 * 欄位語意（強推論）：星域 +0x12 ＝ 擁有者（0 帝國、1 同盟、其他；docs/internals/22）；第 13 筆是伊謝爾倫、第 14 筆是費沙；
 *   ds:8F48 ＝ 費沙事件已發生的旗標（docs/re/179 第 46 到 60 行、docs/internals/54 第 451 行）。位移與比較值已證實，名稱是強推論。
 * 證據：asm 逐指令；docs/re/179 第 12 到 25 行的 12 種表格與本檔一致。
 * 輸入域：全部三個 byte 的任意值；描述檔 seeds 對 12 種組合各跑一次。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2D54_0163(Mem *m)
{
    uint8_t a = rd8(m, DS(0xAE02 + 0x12));
    uint8_t f;
    uint8_t b;
    uint16_t base;

    if (a == 0)
        base = 0;
    else if (a == 1)
        base = 4;
    else
        base = 8;
    f = rd8(m, DS(0x8F48));
    if (f == 0)
        return base;
    b = rd8(m, DS(0xAE80 + 0x12));
    if (b == 0)
        return (uint16_t)(base + 1);
    if (b == 1)
        return (uint16_t)(base + 2);
    return (uint16_t)(base + 3);
}

static uint64_t run(Mem *m, const uint16_t *a) { (void)a; return f_2D54_0163(m); }
GIN3_REGISTER("2D54_0163", 0, run)
