/*
 * GIN3PS.EXE 341B:04DB  畫面座標 (x, y) 換成格座標，各寫一個 byte 到兩個輸出位址
 *
 * 原版：far 呼叫，序言 push bp／si／di、mov bp,sp（沒有區域變數）。堆疊參數（由低位址到高位址）：
 *   x（word，[bp+0Ah]）、y（word，[bp+0Ch]）、out_x（far 指標，位移 [bp+0Eh]／段 [bp+10h]）、out_y（far 指標，[bp+12h]／[bp+14h]）。
 *   沒有回傳值；AX 是除法留下的殘值，本版不比對（描述檔 ret ＝ none）。
 *
 * 語意（已證實，本檔與反組譯逐條對過）：
 *   *out_x ＝ 低位元組(x >> 4)       ; shr ax,4（16 位元邏輯右移）後只存 AL
 *   *out_y ＝ 低位元組(y ÷ 20)       ; xor dx,dx；div 20（16 位元無號，商最大 3276）後只存 AL
 *   先寫 out_x、再寫 out_y；兩個指標相同時以 out_y 為準。
 *   格寬 16、格高 20 像素：docs/re/117 §位置是格座標（「341B:04DB 的兩個輸出是 byte」）、docs/internals/data/battle-menu.json
 *   （「格 ＝ 游標所在格 − (1,1)」另有 3982:0198 夾限）。本函式沒有減 1、沒有夾限。
 *
 * 原版的細節（本版照做）：
 *   - 位移運算在 16 位元下：x 先右移 4 再取低位元組（x >= 4096 時格號回繞），不是先取低位元組。
 *   - 寫入落在線性 0xF0000 以上（ROM 區）時 dosgolem 安靜丟棄（internal/machine/bios_rom.go），本版自己模型化。
 *   - 輸出指標的位移不需要回繞處理（只有 byte 存取）。
 *
 * 輸入域：全部 16 位元值；輸出指標指向 A0000 到 BFFFF 的案例被探針丟棄（視訊記憶體不比對）。
 * 呼叫者（台帳）：1CC5:070D、1CC5:09C6、1D9E:01F1、1ED2:03E9、1ED2:0616。
 */
#include "gin3_reg.h"

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意。 */
static void f341B04DB_wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_341B_04DB(Mem *m, uint16_t x, uint16_t y, uint16_t xo, uint16_t xs, uint16_t yo, uint16_t ys)
{
    f341B04DB_wr8_skip_rom(m, LIN(xs, xo), (uint8_t)(x >> 4));
    f341B04DB_wr8_skip_rom(m, LIN(ys, yo), (uint8_t)(y / 20u));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_341B_04DB(m, a[0], a[1], a[2], a[3], a[4], a[5]);
    return 0;
}
GIN3_REGISTER("341B_04DB", 6, run)
