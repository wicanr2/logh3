/*
 * GIN3PS.EXE 2C5D:06E2  ds:8F16 的低位元組等於 1 時，依序清戰場艦艇的小地圖標記（1727:011E）與移除在場艦隊的顯示物件（25DB:0221）
 *
 * 原版（18 bytes，far 呼叫，無參數，無回傳值，描述檔 ret ＝ none）：
 *   cmp byte ptr ds:8F16h, 1 ; jnz 返回        ; 只比 byte，ds:8F17 不讀
 *   call far 1727:011E                          ; 無參數
 *   call far 25DB:0221                          ; 無參數
 *   retf
 * 過程（已證實：反組譯逐指令；1727:011E、25DB:0221 由其各自的差異測試證實）：
 *   byte ds:8F16 ≠ 1 → 什麼都不做
 *   否則 先 1727:011E，後 25DB:0221
 * 語意（強推論）：呼叫者 2C5D:061A（台帳，hw-dependent）；ds:8F16 的語意未知，只知道值為 1 時才做這兩件事。
 *   docs 沒有 2C5D:06E2 本身的記載（見回報）。
 * 副作用：兩個被呼叫者的全部寫入（見各自檔頭）。
 * 輸入域：基底狀態 fcmd-base.state；與 25DB:0221 相同，艦隊序號清單、池基底、艦隊與艦艇的物件指標列入 freeze，由 seeds 給定。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern void f_1727_011E(Mem *m);
extern void f_25DB_0221(Mem *m);

void f_2C5D_06E2(Mem *m)
{
    if (rd8(m, DS(0x8F16u)) != 1u)
        return;
    f_1727_011E(m);
    f_25DB_0221(m);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2C5D_06E2(m);
    return 0;
}
GIN3_REGISTER("2C5D_06E2", 0, run)
