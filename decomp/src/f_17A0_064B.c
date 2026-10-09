/*
 * GIN3PS.EXE 17A0:064B  把艦艇記錄的 +0x30 設成指定值；若是旗艦分隊，一併寫進所屬艦隊的 +0x6B
 *
 * 原版：far 呼叫，兩個 word 參數（ship, value），無回傳值（描述檔 ret = none）。
 *   序言 push bp／si／di、sub sp,4，所以 ship 在 [bp+0Eh]、value 在 [bp+10h]。
 *   艦艇池基底由全域變數指定：ds:CFB4（位移）、ds:CFB6（段）；一筆 78（0x4E）bytes。
 *   艦隊表在 ds:C074，一筆 122（0x7A）bytes。
 *
 * 語意（已證實：docs/re/131 "艦艇走 17A0:064B（寫 +0x30，旗艦另外把值傳進艦隊記錄）"；
 *   docs/internals/33 "+0x6B 戰場圖示的繪製順位：17A0:0699 寫艦艇 +0x30 時，那一筆是旗艦
 *   （艦艇[+0x03] == 0）就一併寫進所屬艦隊的這一格"；docs/internals/38 艦艇 +0x01 ＝ 所屬艦隊索引、
 *   +0x03 ＝ 分隊在艦隊裡的序號、+0x30 ＝ 顯示／選擇順位；本檔與反組譯逐條對過）：
 *   rec  ＝ (ds:CFB4) ＋ (ship × 0x4E 的低 16 位元)       ; 池段 ds:CFB6
 *   byte rec[+0x30] ← 低位元組(value)
 *   若 byte rec[+0x03] == 0：
 *       fleet ＝ 0xC074 ＋ byte rec[+0x01] × 0x7A         ; 16 位元加，超過 0xFFFF 回繞
 *       byte ds:fleet[+0x6B] ← 低位元組(value)
 *   （docs/internals/33 引用的 17A0:0699（線性 0x18099）不是另一個入口：本函式 17A0:064B 起點線性 0x1804B、
 *   長 89 bytes，0x18099 是函式內偏移 0x4E 的 `mov es:[si+6Bh],al`，也就是寫 +0x6B 的那一條指令。）
 *
 * 原版的細節（本版照做）：
 *   - 先寫 +0x30、再讀 +0x03 與 +0x01（同一筆記錄，不同欄位，順序不影響結果）。
 *   - 艦隊位移＝0xC074＋索引×0x7A＋0x6B 全部 16 位元回繞：艦艇 +0x01 ≥ 134 時超過 0xFFFF。
 *   - value 的高位元組被忽略。
 *
 * 輸入域：ship 與 value 是 16 位元全域；池基底取基底狀態並由突變改動；艦艇 +0x01 為 0..255。
 *   池段變數被改成 0xF000 以上時，艦艇欄位的寫入落在 BIOS ROM 區（線性 0xF0000 起），機器忽略該寫入
 *   （dosgolem internal/machine/bios_rom.go）；本版用 wr8_skip_rom 表達。平面視訊記憶體的寫入由探針丟棄。
 */
#include "gin3_reg.h"

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意（回報給主工作階段）。 */
static void wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_17A0_064B(Mem *m, uint16_t ship, uint16_t value)
{
    uint16_t rec = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(ship * 0x4Eu));
    uint16_t seg = rd16(m, DS(0xCFB6));

    wr8_skip_rom(m, LIN(seg, rec + 0x30u), (uint8_t)value);
    if (rd8(m, LIN(seg, rec + 3u)) == 0) {
        uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(rd8(m, LIN(seg, rec + 1u)) * 0x7Au));

        wr8(m, DS(fleet + 0x6Bu), (uint8_t)value);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_17A0_064B(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("17A0_064B", 2, run)
