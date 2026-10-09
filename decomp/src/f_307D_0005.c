/*
 * GIN3PS.EXE 307D:0005  把 ds:D6EC 起 32 個位元組填成 1（memset 的薄包裝）
 *
 * 原版：far 呼叫，沒有參數。本體是 push 0x20；push 1；push ds；push 0xD6EC；call far 3C9C:000A；add sp,8；retf，
 * 也就是 memset(ds:D6EC, 1, 0x20)。呼叫者 3009:0003（台帳 hw-dependent）。
 * 副作用：ds:D6EC .. D70B 共 32 個位元組寫成 0x01，沒有別的寫入。
 * 回傳：本函式沒有自己的回傳值。AX:BX 是 3C9C:000A 留下來的殘值，AX ＝ 目的段（DGROUP 0x54B3）、
 * BX ＝ 目的位移（0xD6EC）；呼叫者是否使用未驗證。描述檔 ret 取 pack:ax,bx 把殘值也一併比對。
 *
 * 資料版面（強推論，來自相鄰函式 307D:0C0F 與 docs/re/287 §八個選項）：
 *   ds:D69C 起 8 個 dword，ds:D6BC 起 8 筆 × 6 bytes，ds:D6EC 起 8 個 dword（本函式的目標），共 0x70 bytes 緊鄰。
 *   填 1 的語意（八個選項各自的 dword 初值為 1）是假說，沒有讀到消費者。
 *
 * 被呼叫者 3C9C:000A（runtime memset）已通過差異測試，這裡直接呼叫 f_3C9C_000A。
 * 輸入域：沒有輸入。基底狀態 ds:D6EC 起全 0，描述檔 seeds 另外把前後各 8 bytes 與目標區預填成不同的值，
 * 驗證寫入範圍剛好是 32 bytes。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);

uint32_t f_307D_0005(Mem *m)
{
    /* 回傳值是 3C9C:000A 的 AX<<16 | BX */
    return f_3C9C_000A(m, 0xD6ECu, GIN3_DS, 1u, 0x20u);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_307D_0005(m);
}
GIN3_REGISTER("307D_0005", 0, run)
