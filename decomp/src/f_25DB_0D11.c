/*
 * GIN3PS.EXE 25DB:0D11  戰場全域顯示物件 B（指標 ds:A784）的初始化：旗標、顏色、尺寸、圖塊指標，並抄行星座標
 *
 * 原版：far 呼叫，一個參數＝行星編號 n（word，整個 16 位元參與乘法）。序言 push bp／si／di、sub sp,4，所以參數在 [bp+0Eh]。
 *   無回傳值（描述檔 ret = none；AX 是殘值，唯一呼叫者 25DB:0AF8 之後不讀它）。
 *   用到的全域變數：行星表 far 指標 ds:CFF8（位移）／ds:CFFA（段），一筆 82（0x52）bytes；物件 B 指標 ds:A784（位移）／ds:A786（段）；
 *   ds:9B9E（byte）；輸出 ds:A790、ds:A792、ds:A788、ds:A78A；圖塊池 ds:F12C／F12E（3982:0001 讀）。
 *
 * 語意（位移與順序為已證實：與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   planet ← (ds:CFFA):(ds:CFF8 ＋ n × 0x52 的低 16 位元)，在一切寫入之前算好（放在堆疊區域變數）。
 *   以下每一步都重新讀 ds:A784／A786 取物件指標（原版每個寫入前都有 `les bx, ds:A784`）：
 *     obj[+1] ← 0xFF；dl ← ds:9B9E（這時才讀）；obj[+2] ← dl；obj[+3] ← 0xFF；obj[+0x0A] ← 0x10；obj[+0x0B] ← 0x10；obj[+0x0C] ← 0x0A。
 *   (seg, off) ← 3982:0001(0x0A)（`push ax` 時 AL = 0x0A；被呼叫者只讀低位元組）：obj[+0x10] ← seg（word）；obj[+0x0E] ← off（word）。
 *   ds:A790 ← word planet[+0x46]；ds:A792 ← word planet[+0x48]（兩者各在前一步寫入之後才讀）；ds:A788 ← 0；ds:A78A ← 0。
 *   物件若蓋到這些全域變數、物件指標自己或行星記錄，後面的讀取看得到前面的寫入，本版照原版的讀寫順序逐項重讀。
 *
 * 欄位語意（強推論）：docs/internals/34-battle-screen.md 第 344 行記「25DB:0E78／0EDD：40-byte 暫時顯示物件，優先值 FF；
 *   16×16 圖塊 10；無詳細資料文字」。0D11 的寫入序列與 0EDD 逐項相同，因此 obj[+3] ← 0xFF 是優先值、obj[+0x0A]／[+0x0B] ← 0x10
 *   是 16×16 的尺寸、obj[+0x0C] ← 0x0A 是圖塊編號 10；obj[+0x0E]／[+0x10] 是該圖塊的遠指標（3982:0001 的回傳，同 25DB:08A2）。
 *   obj[+1] ← 0xFF 是標髒（同 25DB:0865）。ds:9B9E 是什麼（byte）：未知。planet[+0x46]／[+0x48] 依 docs/internals/29 是行星在星域
 *   地圖上的位置 X／Y（1727、17A0 的註解同）；ds:A790／A792 是物件 B 的地圖座標、ds:A788／A78A 是偏移歸零，名稱為假說。
 *   25DB:0EDD 是同型，對象是 ds:A78C 的物件 C 與艦艇記錄（抄艦艇 +0x34／+0x36）。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000），w0D11_put8 不動作；讀仍照基底傾印。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 行為，w0D11_rd16／put16）。
 *
 * 輸入域：n 的全部 16 位元值（乘積取低 16 位元，行星記錄在行星切片的段內回繞）；基底戰略層的 ds:A784 是 0:0（寫入落在中斷向量表），
 *   物件指標由種子與突變供給，可以指到任意位址，包括全域變數、物件指標自己與行星記錄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint32_t f_3982_0001(Mem *m, uint8_t index);

static uint16_t w0D11_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0D11_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w0D11_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w0D11_put8(m, LIN(seg, off), (uint8_t)v);
    w0D11_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* les bx, ds:A784：位移在 ds:A784、段在 ds:A786 */
static void w0D11_obj(const Mem *m, uint16_t *off, uint16_t *seg)
{
    *off = w0D11_rd16(m, GIN3_DS, 0xA784u);
    *seg = w0D11_rd16(m, GIN3_DS, 0xA786u);
}

void f_25DB_0D11(Mem *m, uint16_t n)
{
    uint16_t planet_seg = w0D11_rd16(m, GIN3_DS, 0xCFFAu);
    uint16_t planet_off = (uint16_t)(w0D11_rd16(m, GIN3_DS, 0xCFF8u) + (uint16_t)(n * 0x52u));
    uint16_t off, seg;
    uint8_t dl;
    uint32_t tile;

    w0D11_obj(m, &off, &seg);
    w0D11_put8(m, LIN(seg, (uint16_t)(off + 1u)), 0xFF);
    dl = rd8(m, DS(0x9B9Eu));
    w0D11_obj(m, &off, &seg);
    w0D11_put8(m, LIN(seg, (uint16_t)(off + 2u)), dl);
    w0D11_obj(m, &off, &seg);
    w0D11_put8(m, LIN(seg, (uint16_t)(off + 3u)), 0xFF);
    w0D11_obj(m, &off, &seg);
    w0D11_put8(m, LIN(seg, (uint16_t)(off + 0x0Au)), 0x10);
    w0D11_obj(m, &off, &seg);
    w0D11_put8(m, LIN(seg, (uint16_t)(off + 0x0Bu)), 0x10);
    w0D11_obj(m, &off, &seg);
    w0D11_put8(m, LIN(seg, (uint16_t)(off + 0x0Cu)), 0x0A);
    tile = f_3982_0001(m, 0x0A); /* AX = 段（高 16 位元）、BX = 位移（低 16 位元） */
    w0D11_obj(m, &off, &seg);
    w0D11_put16(m, seg, (uint16_t)(off + 0x10u), (uint16_t)(tile >> 16));
    w0D11_put16(m, seg, (uint16_t)(off + 0x0Eu), (uint16_t)tile);
    wr16(m, DS(0xA790u), w0D11_rd16(m, planet_seg, (uint16_t)(planet_off + 0x46u)));
    wr16(m, DS(0xA792u), w0D11_rd16(m, planet_seg, (uint16_t)(planet_off + 0x48u)));
    wr16(m, DS(0xA788u), 0);
    wr16(m, DS(0xA78Au), 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_25DB_0D11(m, a[0]);
    return 0;
}
GIN3_REGISTER("25DB_0D11", 1, run)
