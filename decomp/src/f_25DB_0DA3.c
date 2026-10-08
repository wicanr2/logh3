/*
 * GIN3PS.EXE 25DB:0DA3  把兩個全域顯示物件的座標減鏡頭後寫回，並標記旗標
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none）。沒有序言，直接用 DS 全域變數：
 *   物件 A：指標 ds:A784（位移）／ds:A786（段）；世界座標 X ds:A790、Y ds:A792
 *   物件 B：指標 ds:A788（位移）／ds:A78A（段）；世界座標 X ds:A794、Y ds:A796
 *   鏡頭：ds:9BF4（X）、ds:9BF6（Y）
 *
 * 語意（位移為已證實：本檔與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   對 A 與 B 依序各做一次（先 A 後 B）：
 *     若 指標的位移與段都是 0（`mov ax,段; or ax,位移; jz`）：整塊跳過。
 *     否則（三次寫入各自重新讀這個指標）：
 *       x ← 座標X − ds:9BF4（16 位元相減）;  obj[+6] ← x
 *       y ← 座標Y − ds:9BF6（16 位元相減）;  obj[+8] ← y
 *       obj[+1] ← 0xFF
 *   讀寫順序與原版一致，因為 obj 若蓋到這些全域變數，後面的讀取會看到前面的寫入：
 *     1. 讀 座標X、ds:9BF4，讀指標，寫 obj[+6]
 *     2. 讀 座標Y、ds:9BF6（在 obj[+6] 寫入之後），重讀指標，寫 obj[+8]
 *     3. 重讀指標，寫 obj[+1]
 *     B 的判斷與讀取在 A 的三次寫入之後，所以 A 的寫入若蓋到 B 的指標或座標，B 會看到新值。
 *
 * 欄位語意（強推論）：兩個物件是戰場的暫時顯示物件；ds:A78C／A78E 那一個由 25DB:0E78 配置、
 *   25DB:0F6B 更新（docs/internals/34-battle-screen.md 第 344 到 346 行的 25DB:0E78／0F6B／0F93）；
 *   A、B 兩個物件的建立者與用途不在本函式範圍，未查證。obj[+1] 旗標的名稱為假說，值 0xFF 為已證實。
 *   呼叫者是 25DB:0547 與 25DB:0AF8。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000），put8／put16 不動作。
 *   - 字的存取跨越位移 0xFFFF 時，本版與探針都按線性位址連續處理，不涵蓋真機上回繞到段首的行為。
 *
 * 輸入域：基底是戰略層（fcmd-base.state），ds:A784 到 A79B 全 0，沒有戰場顯示物件；
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

static void update_block(Mem *m, uint16_t ptr_off, uint16_t x_off, uint16_t y_off)
{
    uint16_t x, y;
    uint16_t a_off, a_seg, b_off, b_seg, c_off, c_seg;

    if ((rd16(m, DS(ptr_off + 2)) | rd16(m, DS(ptr_off))) == 0)
        return;

    x = (uint16_t)(rd16(m, DS(x_off)) - rd16(m, DS(0x9BF4)));
    a_off = rd16(m, DS(ptr_off));
    a_seg = rd16(m, DS(ptr_off + 2));
    put16(m, LIN(a_seg, a_off + 6), x);

    y = (uint16_t)(rd16(m, DS(y_off)) - rd16(m, DS(0x9BF6)));
    b_off = rd16(m, DS(ptr_off));
    b_seg = rd16(m, DS(ptr_off + 2));
    put16(m, LIN(b_seg, b_off + 8), y);

    c_off = rd16(m, DS(ptr_off));
    c_seg = rd16(m, DS(ptr_off + 2));
    put8(m, LIN(c_seg, c_off + 1), 0xFF);
}

void f_25DB_0DA3(Mem *m)
{
    update_block(m, 0xA784, 0xA790, 0xA792);
    update_block(m, 0xA788, 0xA794, 0xA796);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_25DB_0DA3(m);
    return 0;
}
GIN3_REGISTER("25DB_0DA3", 0, run)
