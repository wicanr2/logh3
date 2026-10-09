/*
 * GIN3PS.EXE 1727:03CF  在 (x, y) 畫一個 28×22 的矩形框線（四條線，色 4）
 *
 * 原版：far 呼叫，三個 word 參數（x, y, mask），無回傳值（描述檔 ret = none）。不讀寫全域記憶體。
 *   序言 push bp／si／di、sub sp,6，所以 x 在 [bp+10h]、y 在 [bp+12h]、mask 在 [bp+14h]。
 *
 * 語意（已證實：本檔與反組譯逐條對過；39D7:0936(x0, y0, x1, y1, 色, 遮罩) 的簽名：docs/internals/52。
 *   「矩形框線」的描述來自四次呼叫的座標關係；用途（框什麼）未知）：
 *   39D7:0936(x,        y,        x + 0x1B, y,        4, mask)      ; 上緣
 *   39D7:0936(x,        y + 1,    x,        y + 0x14, 4, mask)      ; 左緣（不含兩端的角）
 *   39D7:0936(x + 0x1B, y + 1,    x + 0x1B, y + 0x14, 4, mask)      ; 右緣
 *   39D7:0936(x,        y + 0x15, x + 0x1B, y + 0x15, 4, mask)      ; 下緣
 *   依這個順序呼叫。所有座標算術都是 16 位元（x + 0x1B、y + 1、y + 0x14、y + 0x15 超過 0xFFFF 回繞）。
 *   mask 原樣傳給畫線（它是畫線的最後一個參數，docs/internals/52 稱「遮罩」，0xFF ＝ 實線）。
 *
 * 呼叫的函式：39D7:0936（畫線，碰視訊記憶體與 I/O）。台帳把本函式記成 game-pure、callees 為 0，因為 IDA 把那個
 *   目標標成 loc_ 而不是函式（呼叫位元組 9A 36 09 D7 39）。差異測試用描述檔的 stubs 取代它：探針不執行它，只記錄
 *   參數與呼叫順序；C 版以 gin3_stub 做同樣的呼叫。**畫線本身的像素行為（含 39D7:0936 自己寫的 ds:8F0E 虛線圖樣與
 *   自我修改程式碼）不在驗證範圍。**
 *
 * 輸入域：x、y、mask 都是 16 位元全域（描述檔的 choices 集中在 0xFFE4..0xFFFF 與 0..0x1B 附近以驗證回繞，
 *   另混入均勻隨機值）。
 */
#include "gin3_reg.h"

void f_1727_03CF(Mem *m, uint16_t x, uint16_t y, uint16_t mask)
{
    uint16_t a[6];
    uint16_t xr = (uint16_t)(x + 0x1Bu);
    uint16_t y1 = (uint16_t)(y + 1u);
    uint16_t y2 = (uint16_t)(y + 0x14u);
    uint16_t y3 = (uint16_t)(y + 0x15u);

    a[0] = x; a[1] = y; a[2] = xr; a[3] = y; a[4] = 4u; a[5] = mask;
    (void)gin3_stub(m, "39D7_0936", 6, a);

    a[0] = x; a[1] = y1; a[2] = x; a[3] = y2; a[4] = 4u; a[5] = mask;
    (void)gin3_stub(m, "39D7_0936", 6, a);

    a[0] = xr; a[1] = y1; a[2] = xr; a[3] = y2; a[4] = 4u; a[5] = mask;
    (void)gin3_stub(m, "39D7_0936", 6, a);

    a[0] = x; a[1] = y3; a[2] = xr; a[3] = y3; a[4] = 4u; a[5] = mask;
    (void)gin3_stub(m, "39D7_0936", 6, a);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1727_03CF(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("1727_03CF", 3, run)
