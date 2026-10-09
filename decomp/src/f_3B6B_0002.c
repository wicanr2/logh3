/*
 * GIN3PS.EXE 3B6B:0002  字元轉全形碼的守衛：c 在 0x20..0x7E（有號比較）就交給 3B6B:0025，否則原值返回
 *
 * 原版：far 呼叫，一個堆疊參數 c（整個 word，[bp+8]；序言 push bp／si、mov bp,sp，沒有 sub sp），回 AX（35 bytes）：
 *   si ← c；若 `cmp ax,20h; jl` 或 `cmp ax,7Eh; jg`（有號）→ 回 AX ＝ c（原值）；
 *   否則 push c；call far 3B6B:0025；回 AX ＝ 0025 的回傳。
 * 有號比較的邊界效果：c >= 0x8000 視為負數，小於 0x20，直接原值返回；等價於無號比較 c > 0x7E 也原值返回，
 *   所以本函式的行為與 c 當有號或無號無關，只有「整個 word 落在 0x20..0x7E」才轉換（c = 0x0120 之類高位元組非 0 的值不轉換）。
 * 呼叫的函式：3B6B:0025（半形轉全形碼，已 diff-tested；三段範圍與 34 筆小表的細節見該檔）。
 * 呼叫者：台帳 callers ＝ 1，由 seg085:031F 呼叫（IDA 未把它辨識成函式）；用途判斷同 0025（強推論）。
 *
 * 輸入域：c 的全部 16 位元值。副作用：只讀兩張表（由 0025 讀），不寫記憶體。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3B6B_0025(Mem *m, uint16_t c);

uint16_t f_3B6B_0002(Mem *m, uint16_t c)
{
    int16_t s = (int16_t)c;

    if (s < 0x20 || s > 0x7E)
        return c;
    return f_3B6B_0025(m, c);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3B6B_0002(m, a[0]); }
GIN3_REGISTER("3B6B_0002", 1, run)
