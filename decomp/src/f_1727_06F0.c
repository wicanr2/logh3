/*
 * GIN3PS.EXE 1727:06F0  把 ds:3C80..3C86 記著的那條線（小地圖上的疊畫線）再畫一次擦掉，並把四個座標設回 0xFFFF
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none）。沒有序言（沒有 push bp），直接讀全域。
 *   全域：ds:3C80（x0）、ds:3C82（y0）、ds:3C84（x1）、ds:3C86（y1），各一個 word；四個都是 0xFFFF ＝「沒有線」。
 *
 * 語意（強推論：1727:0677 把線的端點寫進這四格並畫一次（XOR），本函式再畫一次擦掉；
 *   39D7:0936 是 XOR 畫線，同一條線畫兩次會還原：docs/internals/52 "39D7:0936(x0, y0, x1, y1, 色, 遮罩)"、
 *   "每一個經過這一支的像素都是「底下的值 XOR 顏色」，同一點同一個顏色寫兩次會還原"。
 *   四個全域的欄位語意與「小地圖」這個場所是強推論：1727:0677 的座標轉換 ((v >> 4) + 0x1E0／+0x24) 與
 *   docs/re/132 的小地圖點座標同式。本檔與反組譯逐條對過，欄位名稱不是證據）：
 *   若 ds:3C80、3C82、3C84、3C86 四個 word 全部等於 0xFFFF：什麼都不做，回傳。
 *   否則：
 *       39D7:0936(ds:3C80, ds:3C82, ds:3C84, ds:3C86, 4, 0xFFFF)   ; 色 4、遮罩 0xFFFF
 *       ds:3C86 ← ds:3C84 ← ds:3C82 ← ds:3C80 ← 0xFFFF
 *   「全部相等」的判斷是逐個 word 比較（cmp reg,mem），只要有一個不等於 0xFFFF 就畫。
 *
 * 呼叫的函式：39D7:0936（畫線，碰視訊記憶體與 I/O）。台帳的 callees 欄是 0，因為 IDA 把那個目標標成 loc_ 而不是
 *   函式（呼叫位元組 9A 36 09 D7 39）。差異測試用描述檔的 stubs 取代它：探針不執行它，只記錄參數與呼叫順序；
 *   C 版以 gin3_stub 做同樣的呼叫。**畫線本身的像素行為（含 39D7:0936 自己寫的 ds:8F0E 虛線圖樣與自我修改程式碼）
 *   不在驗證範圍。**
 *
 * 輸入域：四個全域是 16 位元任意值（基底狀態是四個都 0xFFFF，即「沒有線」；其餘由突變導向讀取集與種子構造）。
 */
#include "gin3_reg.h"

void f_1727_06F0(Mem *m)
{
    uint16_t x0 = rd16(m, DS(0x3C80));

    if (x0 == 0xFFFFu && rd16(m, DS(0x3C82)) == 0xFFFFu && rd16(m, DS(0x3C84)) == 0xFFFFu
        && rd16(m, DS(0x3C86)) == 0xFFFFu)
        return;
    {
        uint16_t a[6];

        a[0] = x0;
        a[1] = rd16(m, DS(0x3C82));
        a[2] = rd16(m, DS(0x3C84));
        a[3] = rd16(m, DS(0x3C86));
        a[4] = 4u;
        a[5] = 0xFFFFu;
        (void)gin3_stub(m, "39D7_0936", 6, a);
    }
    wr16(m, DS(0x3C86), 0xFFFFu);
    wr16(m, DS(0x3C84), 0xFFFFu);
    wr16(m, DS(0x3C82), 0xFFFFu);
    wr16(m, DS(0x3C80), 0xFFFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_1727_06F0(m);
    return 0;
}
GIN3_REGISTER("1727_06F0", 0, run)
