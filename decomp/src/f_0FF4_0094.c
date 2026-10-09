/*
 * GIN3PS.EXE 0FF4:0094  艦隊編成視窗開場：初始化格盤工作表，記下這支艦隊與它所在的星域（無回傳值）
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄（122 bytes）的 far 指標；序言 push bp／si、沒有 sub sp，
 * 參數在 [bp+8]（位移）與 [bp+0Ah]（段）。沒有回傳值（描述檔 ret=none）。
 *
 * 反組譯逐條對應（依序）：
 *   ds:8F1D == 0 → ds:984D ← 0x2C、ds:984E ← 0x24；否則 ds:984D ← 0x4A、ds:984E ← 0x42（byte）
 *   memset(ds:977A, 0xFF, 0xA4)                         （呼叫 3C9C:000A，已 diff-tested）
 *   ds:9878 ← 艦隊段、ds:9876 ← 艦隊位移（先段後位移）
 *   k ← 艦隊[+6]（byte，在上面兩次寫入之後才讀；艦隊指標若指到 ds:977A 起的工作表或 ds:9876／9878 本身，
 *       讀到的是 memset 與上面寫入之後的值）
 *   ds:9874 ← DS、ds:9872 ← 0xA79C ＋ k × 0x7E            （16 位元回繞；k 是 byte、零延伸）
 *
 * 推論等級
 *   已證實（差異測試）：上列順序、常數、讀寫位址與寬度；k 零延伸（k ＝ 0xFF 時 0xA79C ＋ 0x7D82 回繞成 0x251E）。
 *   強推論（docs/internals/56-fleet-composition-panel.md 第 18、22、23、157 行）：ds:977A 是 41 × 4 bytes 的格盤
 *     工作表、ds:984D／984E 是空位與有部隊圖塊的基底（依陣營 ds:8F1D 選）、ds:9872 是艦隊所在星域記錄
 *     （星域表 ds:A79C，每筆 126 bytes）的 far 指標、艦隊[+6] 是所在星域編號。
 *   未知：ds:9876／9878 在編成視窗裡的名稱（艦隊 far 指標，docs/re/224 稱「這支艦隊」）。
 *
 * 輸入域：艦隊指標取艦隊表任一筆（32 筆）與 seeds 指定的位址（含指到工作表與 ds:9872 起八個 byte 的別名）；
 *   k 為任意 byte。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);

void f_0FF4_0094(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint16_t kind;

    if (rd8(m, DS(0x8F1D)) == 0) {
        wr8(m, DS(0x984D), 0x2C);
        wr8(m, DS(0x984E), 0x24);
    } else {
        wr8(m, DS(0x984D), 0x4A);
        wr8(m, DS(0x984E), 0x42);
    }
    f_3C9C_000A(m, 0x977A, GIN3_DS, 0xFF, 0xA4);
    wr16(m, DS(0x9878), fleet_seg);
    wr16(m, DS(0x9876), fleet_off);
    kind = rd8(m, LIN(fleet_seg, fleet_off + 6));
    wr16(m, DS(0x9874), GIN3_DS);
    wr16(m, DS(0x9872), (uint16_t)(0xA79Cu + (uint16_t)(kind * 0x7Eu)));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_0FF4_0094(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("0FF4_0094", 2, run)
