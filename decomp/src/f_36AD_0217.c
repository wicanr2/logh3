/*
 * 36AD:0217  把旗標位元組 ds:8AFD 設成 0xFF（near，無參數，DS = DGROUP）
 *
 * 原版：`mov byte ptr ds:8AFDh, 0FFh; retn`，兩條指令，不讀記憶體，不改暫存器。
 * 與 36AD:0211 成對：呼叫者 36AD:00D2、0117、013D、0162 進入時呼叫 0211（清 0），離開時呼叫本函式（設 0xFF）。
 * DS：同 36AD:0211，是 DGROUP 0x54B3。
 *
 * 語意（假說）：ds:8AFD 是游標繪製期間的忙碌／遮蔽旗標；本檔只證實寫入的值與位址。
 * 輸入域：沒有輸入。函式不讀記憶體，案例全靠 seeds：預先把 ds:8AFD 設成非 0xFF，且前後相鄰位元組設成非 0xFF，
 *   才看得出少寫、寫錯值、寫成 word 與位址差 1。
 * 推論等級：寫入的值與位址已證實（反組譯逐指令＋差異測試）；旗標語意為假說。
 */
#include "gin3_reg.h"

void f_36AD_0217(Mem *m)
{
    wr8(m, DS(0x8AFDu), 0xFF);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_36AD_0217(m);
    return 0;
}
GIN3_REGISTER("36AD_0217", 0, run)
