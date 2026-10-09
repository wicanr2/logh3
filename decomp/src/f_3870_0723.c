/*
 * GIN3PS.EXE 3870:0723  畫矩形外框：兩角 (x0, y0)、(x1, y1) 各軸（有號）排好後，用 3870:07C9 畫四條邊（上、下、左、右）
 *
 * 原版（166 bytes，far 呼叫，六個堆疊字；序言 push bp／si／di、sub sp,6、mov bp,sp，所以 x0 在 [bp+10h]、y0 在 [bp+12h]、x1 在 [bp+14h]、
 *   y1 在 [bp+16h]、色 word 在 [bp+18h]、最後 word 在 [bp+1Ah]）。沒有有意義的回傳（描述檔 ret ＝ none）。
 * 過程（已證實：反組譯逐指令；3870:07C9 由其差異測試證實）：
 *   si ← x0；di ← y0
 *   si > x1（有號 jle 反向）→ si 與 [bp+14h] 對調（只對調 x，不連動 y）
 *   di > y1（有號）          → di 與 [bp+16h] 對調（只對調 y）
 *   之後 xl ＝ si、xh ＝ x1（對調後）、yl ＝ di、yh ＝ y1（對調後）。四次 3870:07C9(x0, y0, x1, y1, 色 word, 最後 word)：
 *     上邊：(xl, yl, xh, yl)
 *     下邊：(xl, yh, xh, yh)
 *     左邊：(xl, yl + 1, xl, yh − 1)            ; 16 位元的 inc／dec
 *     右邊：(xh, yl + 1, xh, yh − 1)
 *   色 word 與最後 word 原樣傳給每次 07C9（07C9 只用色 word 的低位元組）。
 * 語意（強推論）：docs/re/314-fleet-standby-planet.md 第 156 行「3870:0723(0x97, 0xB3, 0x118, 0x134, 7, 0xFFFF)  外框 (151,219)-(280,348)，四邊含端點」、
 *   docs/re/320-region-info-window.md 第 47 行「3870:0723(23, 44, 152, 173, 7) 小圖外框」，與本檔一致。
 *   左右兩邊縮進一像素，避免與上下邊重畫角點（異或線重畫會抵消，見 3870:0833 檔頭）。
 *   x0 > x1 時「對調改成 >=」是等價改動（相等時對調同一個值），所以 x 的比較用 > 或 >= 無法區分。
 * 呼叫的函式：3870:07C9 內呼叫畫線 39D7:0647（IDA 沒切成函式，台帳 callees 看不到），差異測試用描述檔 stubs 取代，
 *   **畫線本身的像素行為不在驗證範圍**。
 * 副作用：只經由 07C9（寫本函式堆疊上的端點副本與替身）；本函式不寫全域記憶體。
 *
 * 機器模型：C 版沒有堆疊（本函式的區域變數 [bp+0..5] 只存不傳指標，C 版省略）。f_3870_07C9 把端點放在原版的 SS:1F7E..1F85（見其檔頭），
 *   落在探針的堆疊帶內，寫入集合比對時雙方都排除。
 * 輸入域：六個參數全部 16 位元值；裁切矩形取記憶體 ds:8D6A..8D70；07C9 內 08FF 在原版不終止的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_3870_07C9(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);

void f_3870_0723(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word)
{
    uint16_t xl = x0;
    uint16_t yl = y0;
    uint16_t t;

    if ((int16_t)xl > (int16_t)x1) {
        t = xl;
        xl = x1;
        x1 = t;
    }
    if ((int16_t)yl > (int16_t)y1) {
        t = yl;
        yl = y1;
        y1 = t;
    }
    f_3870_07C9(m, xl, yl, x1, yl, color_word, last_word);
    f_3870_07C9(m, xl, y1, x1, y1, color_word, last_word);
    f_3870_07C9(m, xl, (uint16_t)(yl + 1u), xl, (uint16_t)(y1 - 1u), color_word, last_word);
    f_3870_07C9(m, x1, (uint16_t)(yl + 1u), x1, (uint16_t)(y1 - 1u), color_word, last_word);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3870_0723(m, a[0], a[1], a[2], a[3], a[4], a[5]);
    return 0;
}
GIN3_REGISTER("3870_0723", 6, run)
