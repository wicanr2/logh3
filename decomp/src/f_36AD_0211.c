/*
 * 36AD:0211  把旗標位元組 ds:8AFD 清成 0（near，無參數，DS = DGROUP）
 *
 * 原版：`mov byte ptr ds:8AFDh, 0; retn`，兩條指令，不讀記憶體，不改暫存器。
 * 呼叫者 36AD:00D2、0117、013D、0162（滑鼠游標繪製的包裝，進入時呼叫本函式，離開時呼叫 36AD:0217 設回 0xFF）。
 * DS：呼叫者 36AD:00D2 以 `mov ax, seg seg132; mov ds, ax` 載入，seg132 在 dosgolem 的值是 0x54B3，即 DGROUP
 *   （基底傾印線性 0x36BAA 的位元組 B8 B3 54 8E D8 為證）。
 *
 * 語意（假說）：ds:8AFD 是游標繪製期間的忙碌／遮蔽旗標，0 與 0xFF 兩種值；本檔只證實寫入的值與位址。
 * 輸入域：沒有輸入。函式不讀記憶體，探針沒有讀取集可突變，案例全靠 seeds：預先把 ds:8AFD 與前後相鄰位元組
 *   設成非 0，才看得出少寫、寫錯值、寫成 word 與位址差 1。
 * 推論等級：寫入的值與位址已證實（反組譯逐指令＋差異測試）；旗標語意為假說。
 */
#include "gin3_reg.h"

void f_36AD_0211(Mem *m)
{
    wr8(m, DS(0x8AFDu), 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_36AD_0211(m);
    return 0;
}
GIN3_REGISTER("36AD_0211", 0, run)
