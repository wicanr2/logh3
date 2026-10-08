/*
 * GIN3PS.EXE 25DB:0F6B  把全域顯示物件 C 的座標減鏡頭後寫回（不寫旗標）
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none；AX 在返回時是殘值，兩個呼叫者 25DB:0E78、
 *   25DB:0FCE 的反組譯在呼叫後都沒有讀 AX）。沒有序言，直接用 DS 全域變數：
 *   物件 C：指標 ds:A78C（位移）／ds:A78E（段）；世界座標 X ds:A798、Y ds:A79A
 *   鏡頭：ds:9BF4（X）、ds:9BF6（Y）
 *
 * 語意（位移為已證實：本檔與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   若 指標的位移與段都是 0（`mov ax,段; or ax,位移; jz`）：什麼都不做。
 *   否則（兩次寫入各自重新讀指標）：
 *     x ← ds:A798 − ds:9BF4（16 位元相減）;  obj[+6] ← x
 *     y ← ds:A79A − ds:9BF6（16 位元相減）;  obj[+8] ← y
 *   與 25DB:0DA3 的樣式相同，但沒有 obj[+1] ← 0xFF 這一步。讀寫順序與原版一致：
 *     1. 讀 ds:A798、ds:9BF4，讀指標，寫 obj[+6]
 *     2. 讀 ds:A79A、ds:9BF6（在 obj[+6] 寫入之後），重讀指標，寫 obj[+8]
 *     obj 若蓋到這些全域變數，後面的讀取會看到前面的寫入。
 *
 * 欄位語意（強推論）：ds:A78C／A78E 是 25DB:0E78 配置的 40-byte 暫時顯示物件（`3ABF:0000(0x28)` 的回傳存進
 *   這兩個字，docs/internals/34-battle-screen.md 第 344 到 346 行）；ds:A798／A79A 是它要顯示的世界座標。
 *   寫入者（反組譯，已證實）：25DB:0EDD 把艦艇[+0x34]／[+0x36] 抄進去，25DB:1029 把自己的兩個參數存進去。
 *   docs/internals/34 第 345 行的「物件+6/+8 ＝ SHIP+34/+36 世界座標減鏡頭」是語意層描述：本函式本身
 *   讀的是 ds:A798／A79A，不讀艦艇記錄。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000），put8／put16 不動作。
 *   - 字的存取跨越位移 0xFFFF 時，本版與探針都按線性位址連續處理，不涵蓋真機上回繞到段首的行為。
 *
 * 輸入域：基底是戰略層（fcmd-base.state），ds:A78C 到 A79B 全 0，沒有戰場顯示物件；
 *   指標與座標全由突變與 seeds 供給，obj 可以指到任意位址，包括這幾個全域變數自己。
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

void f_25DB_0F6B(Mem *m)
{
    uint16_t x, y;
    uint16_t a_off, a_seg, b_off, b_seg;

    if ((rd16(m, DS(0xA78E)) | rd16(m, DS(0xA78C))) == 0)
        return;

    x = (uint16_t)(rd16(m, DS(0xA798)) - rd16(m, DS(0x9BF4)));
    a_off = rd16(m, DS(0xA78C));
    a_seg = rd16(m, DS(0xA78E));
    put16(m, LIN(a_seg, a_off + 6), x);

    y = (uint16_t)(rd16(m, DS(0xA79A)) - rd16(m, DS(0x9BF6)));
    b_off = rd16(m, DS(0xA78C));
    b_seg = rd16(m, DS(0xA78E));
    put16(m, LIN(b_seg, b_off + 8), y);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_25DB_0F6B(m);
    return 0;
}
GIN3_REGISTER("25DB_0F6B", 0, run)
