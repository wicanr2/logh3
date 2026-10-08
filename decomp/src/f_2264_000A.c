/*
 * GIN3PS.EXE 2264:000A  要塞砲倒數：每回合把有要塞且倒數非 0 的行星 +0x45 減 1
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none）。行星切片由全域變數指定：
 *   ds:CFF8（位移）、ds:CFFA（段）＝ 切片第一筆；ds:CFFC（byte）＝ 筆數。一筆 82 bytes。
 *
 * 語意（已證實：docs/re/145 "+0x45 是倒數，不是旗標" 的偽碼；docs/internals/50、docs/internals/29 +0x45；
 *   docs/spec/028 第 21 行同結論；本檔與反組譯逐條對過）：
 *   對切片每一顆行星 p：
 *       若 p[+0x12] != 0（有要塞）且 p[+0x45] != 0：p[+0x45] ← p[+0x45] − 1
 *   +0x12 是要塞旗標，+0x45 是要塞砲狀態與倒數（3 下一次結算開火、2／1 填充中、0 待命）。
 *
 * 原版的細節（本版照做）：
 *   - 筆數 ds:CFFC 每圈重讀，計數器是 byte 且無號比較；切片指標只在開頭讀一次。
 *   - 減 1 是 byte 的 dec，+0x45 為 0 時不執行（有 != 0 的守衛），所以不會從 0 繞回 0xFF。
 *   - 位移 16 位元回繞（bx+12h／bx+45h／add [bp],52h）。
 *
 * 輸入域：切片指標與筆數取基底狀態（星域 0，4 顆）並由突變導向讀取集改動；筆數可到 255，
 *   超出 81 顆時讀寫行星表之後的記憶體，本版同樣照做。
 */
#include "gin3_reg.h"

void f_2264_000A(Mem *m)
{
    uint16_t off = rd16(m, DS(0xCFF8));
    uint16_t seg = rd16(m, DS(0xCFFA));
    uint8_t i = 0;

    while (i < rd8(m, DS(0xCFFC))) {
        if (rd8(m, LIN(seg, off + 0x12)) != 0 && rd8(m, LIN(seg, off + 0x45)) != 0) {
            wr8(m, LIN(seg, off + 0x45), (uint8_t)(rd8(m, LIN(seg, off + 0x45)) - 1));
        }
        i++;
        off = (uint16_t)(off + 0x52);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2264_000A(m);
    return 0;
}
GIN3_REGISTER("2264_000A", 0, run)
