/*
 * GIN3PS.EXE 39C8:00CA  計時器中斷常式的每次計時處理：兩個倒數計時遞減、32 位元計時計數加 1
 *
 * 呼叫：near，沒有參數。呼叫者 seg099:00A3（計時器 ISR 內，docs/re/331 第 242 行：
 *   「ISR：ds:8EF2 每 10 次歸零並呼叫 39C8:00CA：ds:8EFC、ds:8EFE 非 0 就各減 1，
 *    ds:8EF8:8EFA 32 位元加 1」，與本版相符）。台帳 callers＝1，該呼叫者在 ISR 內，不在函式台帳內。
 *
 * 行為（全是 DGROUP 內的 word，DS 固定 54B3）：
 *   ds:8EFC  word  倒數計時 A：非 0 才減 1（不減到負數）
 *   ds:8EFE  word  倒數計時 B：非 0 才減 1
 *   ds:8EF8  word  32 位元計時計數的低字：加 1，進位帶到 8EFA
 *   ds:8EFA  word  計數的高字：加進位（adc），32 位元回繞（0xFFFFFFFF → 0）
 * 回傳：AX 的殘值是 adc 之後的高字（新的 ds:8EFA）。ISR 呼叫者不用它，
 *   但差異測試比整個 AX，所以 C 版也回傳。它不是函式的語意輸出。
 * 39C8:0078（等待 ds:8EFC 歸零的忙等迴圈）靠這支遞減才會結束；它那邊另行處理，本檔不涵蓋。
 *
 * 暫存器與旗標：只動 AX 與旗標，其餘暫存器不變。
 * 副作用：只寫 ds:8EFC／8EFE（有條件）與 ds:8EF8／8EFA（必寫）。不碰 I/O、int、視訊記憶體。
 * 輸入域：全部 16 位元值（計數、兩個倒數計時都不限範圍）。
 * 欄位語意：倒數計時 A、B 的消費者（誰把它們設成非 0）：39C8:0078 設 8EFC；8EFE 的設定者本檔未查，未知。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）；「計時器 ISR 的每次計時處理」來自 docs/re/331（強推論）。
 */
#include "gin3_reg.h"

uint16_t f_39C8_00CA(Mem *m)
{
    uint16_t ax, carry;

    if (rd16(m, DS(0x8EFC)) != 0)
        wr16(m, DS(0x8EFC), (uint16_t)(rd16(m, DS(0x8EFC)) - 1));
    if (rd16(m, DS(0x8EFE)) != 0)
        wr16(m, DS(0x8EFE), (uint16_t)(rd16(m, DS(0x8EFE)) - 1));
    ax = rd16(m, DS(0x8EF8));
    ax = (uint16_t)(ax + 1);
    carry = (ax == 0) ? 1 : 0; /* add ax,1 的進位：0xFFFF + 1 */
    wr16(m, DS(0x8EF8), ax);
    ax = (uint16_t)(rd16(m, DS(0x8EFA)) + carry);
    wr16(m, DS(0x8EFA), ax);
    return ax;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_39C8_00CA(m);
}
GIN3_REGISTER("39C8_00CA", 0, run)
