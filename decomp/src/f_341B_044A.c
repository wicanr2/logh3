/*
 * GIN3PS.EXE 341B:044A  畫面座標 (x, y) 換成「視窗內的格座標」，寫成兩個 byte；視窗有框時再扣掉框並夾成 0xFF
 *
 * 原版：far 呼叫，序言 push bp／si／di、mov bp,sp（沒有區域變數）。堆疊參數（由低位址到高位址）：
 *   win（far 指標，位移 [bp+0Ah]／段 [bp+0Ch]）、x（word，[bp+0Eh]）、y（word，[bp+10h]）、
 *   out_x（far 指標，[bp+12h]／[bp+14h]）、out_y（far 指標，[bp+16h]／[bp+18h]）。
 *   沒有回傳值；AX 是殘值，本版不比對（描述檔 ret ＝ none）。
 *
 * 視窗節點欄位（341B:039D 同一種記錄）：+0x03 byte ＝ 非 0 表示有框（強推論：docs/re/09 §座標是格子，
 * docs/internals/data/dropdown-menus.json 的 hit_test 註解「0xFF ＝ 在外面」）；+0x05／+0x06 ＝ 視窗左上角格 X／Y；
 * +0x07／+0x08 ＝ 視窗寬／高（格）。格寬 16、格高 20 像素。
 *
 * 語意（已證實，本檔與反組譯逐條對過）：
 *   *out_x ＝ 低位元組( 低位元組(x >> 4) − win[+5] )                 ; 先寫
 *   *out_y ＝ 低位元組( 低位元組(y ÷ 20) − win[+6] )                 ; 後寫
 *   若 win[+3] == 0 結束。否則（有框）對 out_x、out_y 各做一次夾限：
 *       a ＝ *out_x（從記憶體重讀，零延伸成 16 位元）
 *       若 a == 0 或 !( (win[+7] − 1) > a )（16 位元無號，win[+7] ＝ 0 時 win[+7]−1 ＝ 0xFFFF）：*out_x ＝ 0xFF
 *       否則 *out_x ＝ *out_x − 1                                       ; 也就是扣掉框的一格
 *     out_y 同理，用 win[+8]。夾限先做 x 再做 y。
 *   結果：有框時，格座標 0（框本身）與 >= 寬−1（右／下框與視窗外）都變成 0xFF，內部格座標減 1。
 * 原版的細節（本版照做）：
 *   - 所有欄位與輸出都是記憶體讀寫，順序固定：讀 win[+5]、寫 out_x、讀 win[+6]、寫 out_y、讀 win[+3]、
 *     重讀 out_x 與 win[+7]、寫 out_x、重讀 out_y 與 win[+8]、寫 out_y。輸出指標與視窗別名（例如 out_x 指到 win[+6]）時，
 *     後面的讀取看得到前面的寫入，本版逐步照做。
 *   - 夾限的 a == 0 分支與「dec 後得 0xFF」結果相同，是原版的冗餘；本版保留兩個分支以對應指令。
 *   - 寫入落在線性 0xF0000 以上（ROM 區）時 dosgolem 安靜丟棄，之後的重讀讀到的是 ROM 原內容；本版自己模型化。
 * 輸入域：全部 16 位元值；win 或輸出指標指向 A0000 到 BFFFF 的案例被探針丟棄。
 * 呼叫者（台帳）：0C0A:02A2、0C0A:0351、0C7F:0293、0C7F:03CB、0D05:027D、0D05:03D2、0FF4:087A、1373:014D、1373:026E、
 * 1629:01D5、3667:00F2、37A9:01F9、37A9:0470。
 */
#include "gin3_reg.h"

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意。 */
static void f341B044A_wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_341B_044A(Mem *m, uint16_t woff, uint16_t wseg, uint16_t x, uint16_t y,
                 uint16_t xo, uint16_t xs, uint16_t yo, uint16_t ys)
{
    uint8_t cell_x, cell_y;
    uint16_t a;

    cell_x = (uint8_t)((uint8_t)(x >> 4) - rd8(m, LIN(wseg, (uint16_t)(woff + 5u))));
    f341B044A_wr8_skip_rom(m, LIN(xs, xo), cell_x);
    cell_y = (uint8_t)((uint8_t)(y / 20u) - rd8(m, LIN(wseg, (uint16_t)(woff + 6u))));
    f341B044A_wr8_skip_rom(m, LIN(ys, yo), cell_y);

    if (rd8(m, LIN(wseg, (uint16_t)(woff + 3u))) == 0)
        return;

    a = rd8(m, LIN(xs, xo));
    if (a != 0 && (uint16_t)(rd8(m, LIN(wseg, (uint16_t)(woff + 7u))) - 1u) > a)
        f341B044A_wr8_skip_rom(m, LIN(xs, xo), (uint8_t)(rd8(m, LIN(xs, xo)) - 1u));
    else
        f341B044A_wr8_skip_rom(m, LIN(xs, xo), 0xFFu);

    a = rd8(m, LIN(ys, yo));
    if (a != 0 && (uint16_t)(rd8(m, LIN(wseg, (uint16_t)(woff + 8u))) - 1u) > a)
        f341B044A_wr8_skip_rom(m, LIN(ys, yo), (uint8_t)(rd8(m, LIN(ys, yo)) - 1u));
    else
        f341B044A_wr8_skip_rom(m, LIN(ys, yo), 0xFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_341B_044A(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7]);
    return 0;
}
GIN3_REGISTER("341B_044A", 8, run)
