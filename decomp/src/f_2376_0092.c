/*
 * 2376:0092  畫面格表的旗標整理（far，無參數，DS = DGROUP）
 *
 * 原版：對 ds:D81E 起 800（0x320）筆、每筆 6 bytes 的畫面格表逐筆做
 *     al = [si] & 0x9F; [si] = al;       旗標清掉 bit 5 與 bit 6（保留 bit 7、bit 4、bit 3..0）
 *     [si+1] &= 0x01;                    只留 bit 0
 *     若 (al & 0x10) == 0：[si+3] = 0     （bit 4 沒設才清）
 *   al 是 `and 9Fh` 之後的值；bit 4 不在遮罩內，所以測 and 前後的 bit 4 結果相同。
 * 暫存器：SI 由 push／pop 保存，CX 結束時是 0，AX 是最後一筆的暫用值。五個呼叫者
 *   （341B:0004、0145、01CF、02E5、35D1:026D）之後都不使用它們，描述檔 ret = none，只比對記憶體寫入。
 * 輸出：ds:D81E 到 ds:EADD 內每筆的 +0、+1，以及 bit 4 沒設時的 +3。+2、+4、+5 不讀不寫。
 *
 * 語意：docs/internals/42-screen-tile-table.md 與 47-window-frame-tiles.md 把 ds:D81E 記為 40 × 20 格、
 *   一格 6 bytes 的畫面格表，+0 是旗標（bit 7、6 要有、bit 4 要沒有才畫），+3 要是 0 才畫；
 *   42 號記 2376:0390 那一段畫完一格後 `and byte es:[bx], 1Fh`（清掉旗標的 bit 7..5）。本函式清的位元
 *   與那一段不同（清 bit 6、bit 5，保留 bit 7 與 bit 4），兩者各自的語意名稱：未知。這個函式的位元語意
 *   沒有文件記載，本檔以原版的差異測試為準。
 *
 * 輸入域：表內全部 4800 bytes 的任意值（+0、+1、+3 是讀取集，突變會改它們）。
 * 推論等級：位移、遮罩與條件已證實（反組譯逐指令＋差異測試）；各位元的用途：未知。
 */
#include "gin3_reg.h"

void f_2376_0092(Mem *m)
{
    uint16_t si = 0xD81Eu;
    uint16_t cx;

    for (cx = 0x320u; cx != 0; cx--) {
        uint8_t al = (uint8_t)(rd8(m, DS(si)) & 0x9Fu);

        wr8(m, DS(si), al);
        wr8(m, DS((uint16_t)(si + 1u)), (uint8_t)(rd8(m, DS((uint16_t)(si + 1u))) & 0x01u));
        if ((al & 0x10u) == 0)
            wr8(m, DS((uint16_t)(si + 3u)), 0);
        si = (uint16_t)(si + 6u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2376_0092(m);
    return 0;
}
GIN3_REGISTER("2376_0092", 0, run)
