/*
 * GIN3PS.EXE 3703:000D  回傳資料段 seg129 的段值
 * 呼叫：far，無參數。回傳 AX ＝ 0x3D22（`mov ax, seg seg129; retf`）。不讀不寫記憶體。
 * 段值 0x3D22 是重定位之後的值：映像內未重定位的欄位是 0x39D9，dosgolem 載入基底 0x0349，
 * 0x39D9 + 0x0349 ＝ 0x3D22。所以這個常數只在 dosgolem 的載入位置成立（比照 docs/internals/67 的 04F8:05C6）。
 * 用途：給 36A1:0000、36A1:0076 取 seg129 的段值（台帳 callers 欄）。seg129 的內容見 3703:0011 與 docs/re/138。
 * 輸入域：無輸入。
 * 證據：asm（b8 22 3d cb），差異測試。推論等級：回傳值已證實；「seg129」是 IDA 的段名，語意不另推論。
 */
#include "gin3_reg.h"

uint16_t f_3703_000D(void)
{
    return 0x3D22;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    (void)a;
    return f_3703_000D();
}
GIN3_REGISTER("3703_000D", 0, run)
