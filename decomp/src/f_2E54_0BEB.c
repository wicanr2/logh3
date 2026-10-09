/*
 * GIN3PS.EXE 2E54:0BEB  抽亂數決定「只對已到達的艦隊下出擊令」或「對全部艦隊下出擊令」
 *
 * 原版：far 呼叫，四個堆疊字（push bp 之後，第一個參數在 [bp+6]），無回傳值（AX 是殘留值）：
 *   args[0..1] ＝ 艦隊 far 指標陣列的位移、段    args[2] ＝ 筆數 count    args[3] ＝ 目標星域 target
 * 過程（已證實：反組譯逐指令）：
 *   r ＝ rand()（3CD5:000C，寫 ds:94DC 種子一次；回 0..0x7FFF）
 *   r ÷ 5 的餘數（cwd；idiv cx，有號）== 0  → 2E54:0C27(list, count, target)   ; 只處理 +0x79 == 6 的艦隊
 *   其餘                                    → 2E54:0C83(list, count, target)   ; 處理陣列裡全部的艦隊
 * 兩個被呼叫者收到的四個字原樣來自本函式的參數（push 順序 target、count、段、位移）。
 * r 不會是負數（rand 最多 0x7FFF），有號除法與無號除法的餘數是否為 0 結論相同。
 * 語意（強推論，docs/re/170、docs/internals/54）：約 1/5 的機率只讓「已到達」的艦隊再出擊，其餘 4/5 讓全部出擊；
 *   名稱不是證據，只有 ÷ 5、餘數 0、分支對應的被呼叫者位址是證據。
 * 副作用：ds:94DC..94DF（亂數種子）寫一次，加上被呼叫者的寫入（艦隊 +7、+8、+0x79）。
 * 輸入域：同 2E54:0C27／0C83（陣列取暫存區或 seeds，count／target 只用低位元組）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD5_000C(Mem *m);
extern void f_2E54_0C27(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t target_w);
extern void f_2E54_0C83(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t target_w);

void f_2E54_0BEB(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t target_w)
{
    int16_t r = (int16_t)f_3CD5_000C(m);

    if (r % 5 == 0)
        f_2E54_0C27(m, list_off, list_seg, count_w, target_w);
    else
        f_2E54_0C83(m, list_off, list_seg, count_w, target_w);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2E54_0BEB(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("2E54_0BEB", 4, run)
