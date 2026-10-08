/*
 * 2376:0B43  空函式（far，單一 retf，位元組 CB）
 *
 * 原版只有一個 retf：不讀寫記憶體、不改任何暫存器。呼叫者 321F:06FC、321F:0897。
 * 比對範圍：AX、BX、CX、DX 原樣回來（描述檔 pack:ax,bx,cx,dx）與全部記憶體寫入為空。
 *   ret 最多串 4 個暫存器，其餘（SI、DI、BP、旗標；ES 不在可選範圍）只由反組譯確認「只有 retf」。
 * 語意（已證實）：docs/re/407-fleet-blocks.md 與 docs/spec/095-fleet-blocks.md 記為「原 byte CB，只返回」，
 *   先前「呈現／buffer 搬移」的假說已撤回。
 * 輸入域：四個暫存器的全部 16 位元值。
 */
#include "gin3_reg.h"

uint64_t f_2376_0B43(Mem *m, uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx)
{
    (void)m;
    return ((uint64_t)ax << 48) | ((uint64_t)bx << 32) | ((uint64_t)cx << 16) | dx;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2376_0B43(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("2376_0B43", 4, run)
