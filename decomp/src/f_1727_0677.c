/*
 * GIN3PS.EXE 1727:0677  把物件記錄上兩個世界座標換成小地圖座標，存進 ds:3C80..3C86，並畫一條線
 *
 * 原版：far 呼叫，一個參數＝物件記錄的 far 指標（位移在前、段在後），無回傳值（描述檔 ret = none）。
 *   序言 push bp／si／di、mov bp,sp（沒有 sub sp），所以參數在 [bp+0Ah]（位移）、[bp+0Ch]（段）。
 *   輸出全域：ds:3C80（x0）、ds:3C82（y0）、ds:3C84（x1）、ds:3C86（y1），各一個 word；
 *   1727:06F0 讀這四格把同一條線再畫一次擦掉。
 *
 * 語意（已證實：本檔與反組譯逐條對過。記錄的種類、欄位名稱未知；轉換式 ((v >> 4) + 0x1E0／0x24) 與
 *   docs/re/132 的小地圖點座標同式（強推論：這是小地圖座標）。呼叫者 25DB:0AF8 未在此驗證）：
 *   x0 ＝ (word obj[+0x5C] >> 4) ＋ 0x1E0     ; 邏輯右移 4，16 位元加
 *   y0 ＝ (word obj[+0x5E] >> 4) ＋ 0x24
 *   x1 ＝ (word obj[+0x64] >> 4) ＋ 0x1E0
 *   y1 ＝ (word obj[+0x66] >> 4) ＋ 0x24
 *   依序寫入 ds:3C80 ← x0、3C82 ← y0、3C84 ← x1、3C86 ← y1（讀一個、寫一個，交錯進行）
 *   若 x0 > x1（無號，jbe 不成立）：ds:3C80 ← x1、3C84 ← x0、3C82 ← y1、3C86 ← y0
 *       ——連 y 一起換，不管 y0 與 y1 的大小。
 *   39D7:0936(ds:3C80, ds:3C82, ds:3C84, ds:3C86, 4, 0xFFFF)   ; 色 4、遮罩 0xFFFF
 *
 * 原版的細節（本版照做）：
 *   - 讀與寫交錯：物件欄位若與 ds:3C80..3C86 重疊，後面的讀會讀到前面剛寫的值。本版逐一照原順序讀寫。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（rd16s），不是下一段的第一個 byte。
 *   - 傳給畫線的四個座標是交換之後從記憶體重讀的值。
 *
 * 呼叫的函式：39D7:0936（畫線，碰視訊記憶體與 I/O）。台帳的 callees 欄是 0，因為 IDA 把那個目標標成 loc_ 而不是
 *   函式（呼叫位元組 9A 36 09 D7 39）。差異測試用描述檔的 stubs 取代它：探針不執行它，只記錄參數與呼叫順序；
 *   C 版以 gin3_stub 做同樣的呼叫。**畫線本身的像素行為（含 39D7:0936 自己寫的 ds:8F0E 虛線圖樣與自我修改程式碼）
 *   不在驗證範圍。**
 *
 * 輸入域：物件指標取暫存緩衝區（0x9D00 段，內容為 0／0xFF／隨機位元組）、位移接近 0xFFFF 的邊緣值、
 *   與落在 ds:3C80..3C88 附近的重疊位置（種子）。四個欄位是 16 位元任意值。
 */
#include "gin3_reg.h"

/*
 * gin3_mem.h 的 rd16 吃線性位址，高位元組取「線性位址＋1」；真正的 CPU 位移各自回繞，seg:FFFF 的 word
 * 高位元組在 seg:0000（dosgolem internal/cpu/cpu.go read16，8086 語意）。位移可能是 0xFFFF 時用這個
 * （gin3_mem.h 沒有，回報給主工作階段）。
 */
static uint16_t rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_1727_0677(Mem *m, uint16_t obj_off, uint16_t obj_seg)
{
    uint16_t x0, y0, x1, y1;
    uint16_t a[6];

    x0 = (uint16_t)((rd16s(m, obj_seg, (uint16_t)(obj_off + 0x5Cu)) >> 4) + 0x1E0u);
    wr16(m, DS(0x3C80), x0);
    y0 = (uint16_t)((rd16s(m, obj_seg, (uint16_t)(obj_off + 0x5Eu)) >> 4) + 0x24u);
    wr16(m, DS(0x3C82), y0);
    x1 = (uint16_t)((rd16s(m, obj_seg, (uint16_t)(obj_off + 0x64u)) >> 4) + 0x1E0u);
    wr16(m, DS(0x3C84), x1);
    y1 = (uint16_t)((rd16s(m, obj_seg, (uint16_t)(obj_off + 0x66u)) >> 4) + 0x24u);
    wr16(m, DS(0x3C86), y1);
    if (x0 > x1) {
        wr16(m, DS(0x3C80), x1);
        wr16(m, DS(0x3C84), x0);
        wr16(m, DS(0x3C82), y1);
        wr16(m, DS(0x3C86), y0);
    }
    a[0] = rd16(m, DS(0x3C80));
    a[1] = rd16(m, DS(0x3C82));
    a[2] = rd16(m, DS(0x3C84));
    a[3] = rd16(m, DS(0x3C86));
    a[4] = 4u;
    a[5] = 0xFFFFu;
    (void)gin3_stub(m, "39D7_0936", 6, a);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1727_0677(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("1727_0677", 2, run)
