/*
 * GIN3PS.EXE 37A9:05EC  清單元件的底圖：把視窗內容緩衝區裡一個 w × h 的矩形填成空白格，並在矩形右側畫一條捲軸
 * 呼叫：far，六個 word 參數（序言 push bp／si／di、sub sp,0Ah，第一個參數在 [bp+14h]）：
 *   win  [bp+14h]（位移）、[bp+16h]（段）：視窗物件的 far 指標
 *   x    [bp+18h]  矩形左上角的欄（內容格），只取低位元組
 *   y    [bp+1Ah]  矩形左上角的列，只取低位元組
 *   w    [bp+1Ch]  矩形寬（格），只取低位元組
 *   h    [bp+1Eh]  矩形高（列），只取低位元組
 * 視窗物件欄位（docs/re/138、docs/re/48）：+0x0B byte ＝ 內容緩衝區每列的格數（stride）；+0x0E word、+0x10 word ＝ 內容緩衝區 far 指標的位移與段。
 * 內容格是 4 bytes：+0 word 字碼、+2 byte 顏色、+3 byte 旗標（docs/re/138「內容格的旗標 ＝ 裝飾位元」）。格 (c, r) 的位址 ＝ 緩衝區 + (r × stride + c) × 4。
 * 無回傳值（AX 是殘值，描述檔 ret ＝ none）。
 *
 * 副作用（asm 逐條，順序也照做，因為緩衝區可能蓋到視窗物件自己）：
 *   第一段，i ＝ 0 起、i < h（無號）：
 *     格位址 ＝ 緩衝區位移 + x × 4 + ((y + i) × stride 的低 16 位元) × 4（全部 16 位元回繞）；每次迴圈重讀 stride、緩衝區段與位移。
 *     第一格：word ← 0x2020、旗標 ← 0x80、顏色 ← 0x87（寫入順序：word、+3、+2）。
 *     之後 j ＝ 1 起、j < w（無號）的每格：word ← 0x2020、旗標 ← 0x08、顏色 ← 0x87。所以 w ＝ 0 時每列仍寫第一格。
 *   第二段（捲軸，在矩形右邊一欄 x + w，起始列 y）：
 *     位址 ＝ 緩衝區位移 + x × 4 + w × 4 + (y × stride) × 4（16 位元）；stride 取自 win+0x0B，每次前進前重讀。
 *     word 0xFF11（上箭頭）；下移一列（+ stride × 4）word 0xFF12；
 *     然後 k ＝ 2 起，只要 (h − 2) > k（16 位元有號）就寫 word 0xFF13 並下移一列；
 *     最後 word 0xFF14、下移一列、word 0xFF15。
 *   這四個 word 值是字碼，捲軸字元的意義未查證（強推論：FF11 ▲、FF12 與 FF13 桿、FF14 與 FF15 ▼ 的圖塊，見 docs/re/323 的 ▲／▼）。
 * 呼叫者：04F8:0355、056A:011C、0C0A:0086、0C7F:007C、0D05:008B、1373:007E；docs/re/323 第 76 行 `37A9:05EC(win, 0, 1, 0x11, 0x10)`。
 *
 * 機器模型：寫入落在線性 0xF0000 以上（BIOS ROM）時，dosgolem 安靜丟棄（internal/machine/bios_rom.go）；word 寫入的兩個位元組位移各自在段內回繞。
 *   本檔以 put8／put16 表達。平面視訊記憶體（A0000 到 BFFFF）的寫入由探針丟棄。
 * 輸入域：視窗取 ds:9518 與 ds:962C（基底狀態 fcmd-base.state 兩個都開著，緩衝區 6D48:0006 與 7944:0006）；
 *   stride、緩衝區指標可被突變成任意值（緩衝區可指到任何位置，包括視窗物件自己）。w × h 很大時超過探針步數上限的案例被丟棄。
 *   不涵蓋線性位址超過 1 MB（A20）的位址。
 * 推論等級：位移、算術、值與順序已證實（asm 與差異測試）；視窗欄位與「清單底圖／捲軸」用途為強推論。
 */
#include "gin3_reg.h"

static void f_37A9_05EC_put8(Mem *m, uint16_t seg, uint16_t off, uint8_t v)
{
    uint32_t lin = LIN(seg, off);

    if (lin < 0xF0000u)
        wr8(m, lin, v);
}

static void f_37A9_05EC_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f_37A9_05EC_put8(m, seg, off, (uint8_t)v);
    f_37A9_05EC_put8(m, seg, (uint16_t)(off + 1u), (uint8_t)(v >> 8));
}

static uint16_t f_37A9_05EC_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_37A9_05EC(Mem *m, uint16_t win_off, uint16_t win_seg, uint16_t x, uint16_t y, uint16_t w, uint16_t h)
{
    uint16_t i, j, k;
    uint16_t cell_off, cell_seg;
    uint16_t stride_off = (uint16_t)(win_off + 0x0Bu);
    uint16_t x4 = (uint16_t)((x & 0xFFu) << 2);
    uint16_t bar_off, bar_seg;
    uint16_t stride1; /* 捲軸起算時讀到的 stride（之後每次前進都重讀，不沿用這個值） */

    for (i = 0; i < (h & 0xFFu); i++) {
        uint16_t row = (uint16_t)((y & 0xFFu) + i);
        uint16_t stride = rd8(m, LIN(win_seg, stride_off));
        uint16_t rowoff = (uint16_t)((uint16_t)(row * stride) << 2);

        cell_seg = f_37A9_05EC_rd16s(m, win_seg, (uint16_t)(win_off + 0x10u));
        cell_off = f_37A9_05EC_rd16s(m, win_seg, (uint16_t)(win_off + 0x0Eu));
        cell_off = (uint16_t)(cell_off + x4);
        cell_off = (uint16_t)(cell_off + rowoff);
        f_37A9_05EC_put16(m, cell_seg, cell_off, 0x2020);
        f_37A9_05EC_put8(m, cell_seg, (uint16_t)(cell_off + 3u), 0x80);
        f_37A9_05EC_put8(m, cell_seg, (uint16_t)(cell_off + 2u), 0x87);
        cell_off = (uint16_t)(cell_off + 4u);
        for (j = 1; j < (w & 0xFFu); j++) {
            f_37A9_05EC_put16(m, cell_seg, cell_off, 0x2020);
            f_37A9_05EC_put8(m, cell_seg, (uint16_t)(cell_off + 3u), 0x08);
            f_37A9_05EC_put8(m, cell_seg, (uint16_t)(cell_off + 2u), 0x87);
            cell_off = (uint16_t)(cell_off + 4u);
        }
    }

    {
        uint16_t rowoff;
        uint16_t w4 = (uint16_t)((w & 0xFFu) << 2);

        stride1 = rd8(m, LIN(win_seg, stride_off));
        rowoff = (uint16_t)((uint16_t)((y & 0xFFu) * stride1) << 2);

        bar_seg = f_37A9_05EC_rd16s(m, win_seg, (uint16_t)(win_off + 0x10u));
        bar_off = f_37A9_05EC_rd16s(m, win_seg, (uint16_t)(win_off + 0x0Eu));
        bar_off = (uint16_t)(bar_off + x4);
        bar_off = (uint16_t)(bar_off + w4);
        bar_off = (uint16_t)(bar_off + rowoff);
    }
    f_37A9_05EC_put16(m, bar_seg, bar_off, 0xFF11);
    bar_off = (uint16_t)(bar_off + (uint16_t)((uint16_t)rd8(m, LIN(win_seg, stride_off)) << 2));
    f_37A9_05EC_put16(m, bar_seg, bar_off, 0xFF12);
    bar_off = (uint16_t)(bar_off + (uint16_t)((uint16_t)rd8(m, LIN(win_seg, stride_off)) << 2));
    for (k = 2; (int16_t)(uint16_t)((h & 0xFFu) + 0xFFFEu) > (int16_t)k; k++) {
        f_37A9_05EC_put16(m, bar_seg, bar_off, 0xFF13);
        bar_off = (uint16_t)(bar_off + (uint16_t)((uint16_t)rd8(m, LIN(win_seg, stride_off)) << 2));
    }
    f_37A9_05EC_put16(m, bar_seg, bar_off, 0xFF14);
    bar_off = (uint16_t)(bar_off + (uint16_t)((uint16_t)rd8(m, LIN(win_seg, stride_off)) << 2));
    f_37A9_05EC_put16(m, bar_seg, bar_off, 0xFF15);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_37A9_05EC(m, a[0], a[1], a[2], a[3], a[4], a[5]);
    return 0;
}
GIN3_REGISTER("37A9_05EC", 6, run)
