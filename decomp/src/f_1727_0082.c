/*
 * GIN3PS.EXE 1727:0082  戰場小地圖視野框的更新（固定「顯示」）：以參數 1 呼叫 1727:035E，回 AX = 1
 *
 * 原版：far 呼叫，無參數、無區域變數、沒有 push bp（12 bytes）：
 *   mov ax,1；push ax；call 1727:035E；inc sp；inc sp；retf
 *   回傳 AX 是 035E 的 AX：035E 最後一條寫 AX 的指令是 `mov ax,[bp+arg_0]`，所以回傳恆為 1。
 * 與 035E 的關係：035E 的參數 ＝ 0 時只擦舊框、不畫新框（隱藏）；參數非 0 時擦舊框並畫新框。本函式固定傳 1，
 *   所以 ds:A427 一律寫成 1（035E 寫 `byte ds:A427 ← arg 的低位元組`）。
 *
 * 呼叫者（台帳）：1588:02CE、1588:03A9（皆 hw-dependent，戰場載入流程）。
 * 呼叫的函式：1727:035E（已 diff-tested）；它的 1727:03CF 畫線 39D7:0936 碰視訊記憶體，差異測試以 stubs 取代，
 *   像素行為不在驗證範圍（見 f_1727_035E.c、f_1727_03CF.c）。
 *
 * 輸入域：ds:9BF4／9BF6、ds:A426..A42C 任意；03CF 的呼叫序列與參數由軌跡比對。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；「顯示／隱藏」的語意是強推論（035E 檔頭）。
 */
#include "gin3_reg.h"

extern uint16_t f_1727_035E(Mem *m, uint16_t arg);

uint16_t f_1727_0082(Mem *m)
{
    return f_1727_035E(m, 1u);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_1727_0082(m);
}
GIN3_REGISTER("1727_0082", 0, run)
