/*
 * 3C5B:0002  空函式（IDA 名稱 nullsub_6）：只有一個 retf（0xCB）
 * 呼叫：far。不讀寫任何記憶體，AX、BX、CX、DX 原值保留（差異測試比對這四個暫存器與整個 1 MB 的寫入集合）。
 * 呼叫者只有 3931:0002（台帳）；用途未知，可能是占位。
 * 旗標、SI、DI、BP 與段暫存器沒有比對（retf 不動它們，由反組譯得知，強推論）。推論等級：已證實（差異測試，僅限上述範圍）。
 */
#include "gin3_reg.h"

uint64_t f_3C5B_0002(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx)
{
    return ((uint64_t)ax << 48) | ((uint64_t)bx << 32) | ((uint64_t)cx << 16) | dx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3C5B_0002(a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3C5B_0002", 4, run)
