/*
 * GIN3PS.EXE 25DB:0714  把艦艇的戰場世界座標減鏡頭，寫進它的顯示物件
 *
 * 原版：far 呼叫，一個參數＝艦艇記錄（78 bytes）的 far 指標，無回傳值（描述檔 ret = none）。
 *   堆疊：序言 push bp／si、sub sp,2，所以第一個參數在 [bp+0Ah]（位移）與 [bp+0Ch]（段）。
 *
 * 語意（位移為已證實：本檔與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   obj ← 艦艇[+0x38] 的 far 指標（位移 +0x38、段 +0x3A）。三次寫入各自重新讀這個指標：
 *     x ← 艦艇[+0x34] − ds:9BF4（16 位元相減）;      obj[+6] ← x
 *     y ← 艦艇[+0x36] − ds:9BF6（16 位元相減）;      obj[+8] ← y
 *     obj[+1] ← 0xFF
 *   讀寫順序與原版一致，因為 obj 若與艦艇記錄或 ds:9BF4／9BF6 重疊，後面的讀取會看到前面的寫入：
 *     1. 讀 艦艇[+0x34]、ds:9BF4，讀 艦艇[+0x38..0x3B]，寫 obj[+6]
 *     2. 讀 艦艇[+0x36]、ds:9BF6（在 obj[+6] 寫入之後），重讀 艦艇[+0x38..0x3B]，寫 obj[+8]
 *     3. 重讀 艦艇[+0x38..0x3B]，寫 obj[+1]
 *
 * 欄位語意（強推論）：艦艇[+0x34]／[+0x36] 是戰場世界座標 X／Y（docs/internals/38-battle-units.md 第 470 行：
 *   「25DB:0714 將 SHIP 現在 XY 減 view 寫顯示物件 +6／+8，無條件設物件 +1 為 FF」；docs/internals/33 的
 *   艦艇表）；ds:9BF4／9BF6 是鏡頭（視野左上角，1588:0362 的座標換算同樣用它們）；艦艇[+0x38] 是這艘
 *   艦艇的圖示顯示物件指標；obj[+1] 是「需要重畫」旗標之類的旗標（名稱為假說，值 0xFF 為已證實）。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000）。基底狀態裡大部分艦艇
 *     的 +0x38 是 FFFF:FFFF，寫入正好落在 ROM，所以 put8／put16 對該區不動作，與執行器一致。
 *   - 字的存取跨越位移 0xFFFF（例如 obj 位移是 0xFFF9）時，本版與探針都按線性位址連續處理，
 *     不涵蓋真機上回繞到段首的行為；描述檔沒有刻意造這種案例。
 *
 * 輸入域：艦艇指標取艦艇池任一筆（352 筆）；基底是戰略層（fcmd-base.state），沒有戰場顯示物件，
 *   物件指標全由基底值、突變與 seeds 供給，因此 obj 可以指到任意位址，包括艦艇記錄自己與 ds:9BF4。
 */
#include "gin3_reg.h"

static void put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void put16(Mem *m, uint32_t lin, uint16_t v)
{
    put8(m, lin, (uint8_t)v);
    put8(m, lin + 1, (uint8_t)(v >> 8));
}

void f_25DB_0714(Mem *m, uint16_t ship_off, uint16_t ship_seg)
{
    uint16_t x, y;
    uint16_t a_off, a_seg, b_off, b_seg, c_off, c_seg;

    /* 步驟 1：x 與第一次讀 obj */
    x = (uint16_t)(rd16(m, LIN(ship_seg, ship_off + 0x34)) - rd16(m, DS(0x9BF4)));
    a_off = rd16(m, LIN(ship_seg, ship_off + 0x38));
    a_seg = rd16(m, LIN(ship_seg, ship_off + 0x3A));
    put16(m, LIN(a_seg, a_off + 6), x);
    y = (uint16_t)(rd16(m, LIN(ship_seg, ship_off + 0x36)) - rd16(m, DS(0x9BF6)));

    /* 步驟 2：第二次讀 obj，寫 obj[+8]（y 在 obj[+6] 寫入之後讀，見上一行） */
    b_off = rd16(m, LIN(ship_seg, ship_off + 0x38));
    b_seg = rd16(m, LIN(ship_seg, ship_off + 0x3A));
    put16(m, LIN(b_seg, b_off + 8), y);

    /* 步驟 3：第三次讀 obj，寫旗標 */
    c_off = rd16(m, LIN(ship_seg, ship_off + 0x38));
    c_seg = rd16(m, LIN(ship_seg, ship_off + 0x3A));
    put8(m, LIN(c_seg, c_off + 1), 0xFF);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_25DB_0714(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("25DB_0714", 2, run)
