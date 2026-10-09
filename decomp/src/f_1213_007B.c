/*
 * GIN3PS.EXE 1213:007B  費沙易主登記（把 ds:988E 設 1，無回傳值）
 *
 * 原版：far 呼叫，一個參數＝星域記錄（126 bytes）的 far 指標；序言只有 push bp，參數在 [bp+6]（位移）與 [bp+8]（段）。
 * 沒有回傳值（AX 保留進入值，描述檔 ret=none）。
 *
 * 反組譯逐條對應：
 *   ds:8F48 != 0                       → 什麼都不做
 *   星域[+0x11] != 0x0E                → 什麼都不做
 *   否則 ds:988E ← 1（byte）
 * 比較都是 byte 對 byte；ds:988E 以外沒有任何寫入。
 *
 * 推論等級
 *   已證實（差異測試）：上列分支、常數 0x0E、旗標位址 ds:8F48／ds:988E、星域[+0x11]／[+0x12] 沒有第二種讀法。
 *   強推論（docs/internals/57-phezzan-event.md、docs/re/296、docs/re/416）：星域[+0x11] 是星域編號（14 ＝ 費沙），
 *     ds:988E 是「這個回合費沙易主了」待辦旗標，ds:8F48 非 0 時登記被擋掉（放行旗標，與 2E54:0F93 讀同一格）。
 *   未知：ds:8F48 的寫入者（docs/re/178、docs/re/296 同樣記為未查）。
 *
 * 輸入域：星域指標取星域表任一筆（32 筆）與 seeds 指定的位址；其餘全是記憶體。
 */
#include "gin3_reg.h"

void f_1213_007B(Mem *m, uint16_t rec_off, uint16_t rec_seg)
{
    if (rd8(m, DS(0x8F48)) != 0)
        return;
    if (rd8(m, LIN(rec_seg, rec_off + 0x11)) != 0x0E)
        return;
    wr8(m, DS(0x988E), 1);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1213_007B(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("1213_007B", 2, run)
