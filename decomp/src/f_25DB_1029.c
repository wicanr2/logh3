/*
 * GIN3PS.EXE 25DB:1029  戰場全域顯示物件 C（指標 ds:A78C）的初始化：旗標、顏色、尺寸、圖塊指標，並存入兩個參數當世界座標
 *
 * 原版：far 呼叫，兩個參數（word）。序言 push bp／si／di、mov bp,sp（沒有區域變數），所以參數在 [bp+0Ah]（x）與 [bp+0Ch]（y）。
 *   無回傳值（描述檔 ret = none；AX 是殘值，唯一呼叫者 25DB:0FCE 之後不讀它）。
 *   用到的全域變數：物件 C 指標 ds:A78C（位移）／ds:A78E（段）；ds:9B9E（byte）；輸出 ds:A798、ds:A79A；圖塊池 ds:F12C／F12E（3982:0001 讀）。
 *
 * 語意（位移與順序為已證實：與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   與 25DB:0EDD 的物件部分逐項相同（每個寫入前都重新讀 ds:A78C／A78E 取物件指標）：
 *     obj[+1] ← 0xFF；dl ← ds:9B9E（這時才讀）；obj[+2] ← dl；obj[+3] ← 0xFF；obj[+0x0A] ← 0x10；obj[+0x0B] ← 0x10；obj[+0x0C] ← 0x0A。
 *   (seg, off) ← 3982:0001(0x0A)：obj[+0x10] ← seg（word）；obj[+0x0E] ← off（word）。
 *   差別在最後：ds:A798 ← x；ds:A79A ← y（直接存參數，不讀艦艇記錄）。
 *   物件若蓋到全域變數或物件指標自己，後面的讀取看得到前面的寫入，本版照原版的讀寫順序逐項重讀。
 *
 * 欄位語意（強推論）：同 25DB:0EDD（docs/internals/34-battle-screen.md 第 344 行；docs/internals/67 第 48 行：25DB:0FCE 路徑由
 *   25DB:1029 存自己的兩個參數，0F6B 隨後把 ds:A798／A79A 減鏡頭寫進 obj[+6]／[+8]）。參數 x、y 是世界座標（假說）。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000），w1029_put8 不動作；讀仍照基底傾印。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 行為，w1029_rd16／put16）。
 *
 * 輸入域：x、y 的全部 16 位元值；基底戰略層的 ds:A78C 是 0:0（寫入落在中斷向量表），物件指標由種子與突變供給，
 *   可以指到任意位址，包括全域變數與物件指標自己。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint32_t f_3982_0001(Mem *m, uint8_t index);

static uint16_t w1029_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w1029_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w1029_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w1029_put8(m, LIN(seg, off), (uint8_t)v);
    w1029_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* les bx, ds:A78C：位移在 ds:A78C、段在 ds:A78E */
static void w1029_obj(const Mem *m, uint16_t *off, uint16_t *seg)
{
    *off = w1029_rd16(m, GIN3_DS, 0xA78Cu);
    *seg = w1029_rd16(m, GIN3_DS, 0xA78Eu);
}

void f_25DB_1029(Mem *m, uint16_t x, uint16_t y)
{
    uint16_t off, seg;
    uint8_t cl;
    uint32_t tile;

    w1029_obj(m, &off, &seg);
    w1029_put8(m, LIN(seg, (uint16_t)(off + 1u)), 0xFF);
    cl = rd8(m, DS(0x9B9Eu));
    w1029_obj(m, &off, &seg);
    w1029_put8(m, LIN(seg, (uint16_t)(off + 2u)), cl);
    w1029_obj(m, &off, &seg);
    w1029_put8(m, LIN(seg, (uint16_t)(off + 3u)), 0xFF);
    w1029_obj(m, &off, &seg);
    w1029_put8(m, LIN(seg, (uint16_t)(off + 0x0Au)), 0x10);
    w1029_obj(m, &off, &seg);
    w1029_put8(m, LIN(seg, (uint16_t)(off + 0x0Bu)), 0x10);
    w1029_obj(m, &off, &seg);
    w1029_put8(m, LIN(seg, (uint16_t)(off + 0x0Cu)), 0x0A);
    tile = f_3982_0001(m, 0x0A); /* AX = 段（高 16 位元）、BX = 位移（低 16 位元） */
    w1029_obj(m, &off, &seg);
    w1029_put16(m, seg, (uint16_t)(off + 0x10u), (uint16_t)(tile >> 16));
    w1029_put16(m, seg, (uint16_t)(off + 0x0Eu), (uint16_t)tile);
    wr16(m, DS(0xA798u), x);
    wr16(m, DS(0xA79Au), y);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_25DB_1029(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("25DB_1029", 2, run)
