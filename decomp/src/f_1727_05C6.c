/*
 * GIN3PS.EXE 1727:05C6  把 128 欄 × 102 列的戰場地圖（每格一個圖塊編號 byte）縮成 128×128 的 1 位元小圖
 *
 * 原版：far 呼叫，一個參數＝地圖的 far 指標（位移在前、段在後），無回傳值（描述檔 ret = none）。
 *   序言 push bp、sub sp,0Ch（沒有 push si／di），所以參數在 [bp+12h]（位移）、[bp+14h]（段）。
 *   輸出緩衝區由全域變數指定：ds:A422（位移）、ds:A424（段）；呼叫者 1727:0004 先配置 0x800 bytes 填進去
 *   （128 列 × 每列 16 bytes ＝ 2048 bytes），再把 ds:3C1A／3C1C 的地圖指標推入呼叫本函式。
 *   查表：ds:3CCC 起 8 個 byte ＝ 遮罩（0x80 >> k）；ds:3C8C 起以圖塊編號為索引的 byte 表。
 *
 * 語意（已證實：docs/re/314 §四 "1727:05C6(buf, 地圖)" 的偽碼：「bit = ds:3C8C[圖塊] & ds:3CCC[px & 7]，
 *   每 8 個湊一個 byte；acc = (acc + 8) % 10；acc == 0（py = 4, 9, 14, …）來源列指標退回 128，下一列重用同一列來源」；
 *   本檔與反組譯逐條對過。偽碼寫 "buf" 像是第一個參數，實際上 buf 是全域 ds:A422／A424，函式只有地圖一個參數）：
 *   acc ＝ 0；out ＝ (ds:A422, ds:A424)；src ＝ 參數
 *   for row ＝ 0; row < 128; row++：
 *       bits ＝ 0
 *       for col ＝ 0; col < 128; col++：
 *           tile ＝ byte src；src++                        ; 位移 16 位元回繞
 *           bits |= byte ds:(0x3C8C + tile) & byte ds:(0x3CCC + (col & 7))
 *           若 (col & 7) == 7：byte out ← bits；out++；bits ← 0
 *       acc ← (acc + 8) % 10                               ; 序列 8, 6, 4, 2, 0, 8, …
 *       若 acc == 0 或 row == 127：src ← src − 128（16 位元）  ; 下一列重用同一列來源
 *   即輸出第 row 列對應地圖列 row − ⌊row / 5⌋；row = 127 時對應列 102，超出 102 列的地圖，
 *   讀到地圖緩衝區之後的記憶體（docs/re/314：該列讀到的是輸出緩衝區自己的前 128 bytes 附近）。
 *
 * 原版的細節（本版照做）：
 *   - 讀與寫在同一個平面記憶體上依原順序進行；來源讀到輸出緩衝區時，讀到的是已經寫進去的位元組。
 *   - 查表 ds:3C8C + tile 最大到 ds:3D8B，超出 64 筆圖塊表的部分讀到 ds:3CCC 起的遮罩與其後的資料。
 *   - 輸出段被改到 0xF000 以上時，寫入落在 BIOS ROM 區（線性 0xF0000 起），機器忽略該寫入
 *     （dosgolem internal/machine/bios_rom.go）；本版用 wr8_skip_rom 表達。平面視訊記憶體的讀寫由探針丟棄。
 *   - 每次呼叫固定執行約 35.7 萬條指令（批次 w24a 的向量檔最大步數 356,983），超過探針預設的步數上限
 *     200,000，描述檔用 step_budget 調高。
 *
 * 輸入域：基底狀態 fcmd-standby.state（ds:A422／A424 是配置好的輸出緩衝區 7CE1:0002，ds:3C1A／3C1C 是已載入的
 *   地圖 79B0:0010）。地圖指標取地圖本身與前 8 列的位移、位移接近 0xFFFF 的邊緣值（來源讀取回繞）；
 *   圖塊內容、兩張查表與輸出指標由突變導向讀取集與種子改動。
 */
#include "gin3_reg.h"

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意（回報給主工作階段）。 */
static void wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_1727_05C6(Mem *m, uint16_t src_off, uint16_t src_seg)
{
    uint16_t out_off = rd16(m, DS(0xA422));
    uint16_t out_seg = rd16(m, DS(0xA424));
    uint8_t acc = 0;
    uint8_t row = 0;

    while (row < 0x80u) {
        uint8_t bits = 0;
        uint8_t col = 0;

        while (col < 0x80u) {
            uint8_t bit = (uint8_t)(col & 7u);
            uint8_t mask = rd8(m, DS(0x3CCCu + bit));
            uint8_t tile = rd8(m, LIN(src_seg, src_off));

            src_off = (uint16_t)(src_off + 1u);
            bits = (uint8_t)(bits | (rd8(m, DS(0x3C8Cu + tile)) & mask));
            if (bit == 7) {
                wr8_skip_rom(m, LIN(out_seg, out_off), bits);
                out_off = (uint16_t)(out_off + 1u);
                bits = 0;
            }
            col++;
        }
        acc = (uint8_t)((acc + 8u) % 10u);
        if (acc == 0 || row == 0x7Fu)
            src_off = (uint16_t)(src_off - 0x80u);
        row++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1727_05C6(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("1727_05C6", 2, run)
