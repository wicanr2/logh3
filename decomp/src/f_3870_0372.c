/*
 * GIN3PS.EXE 3870:0372  佔位函式：把前兩個堆疊參數分別放進 AX 與 CX 就返回（什麼都不畫）
 * 呼叫：far。呼叫者 3870:0296 推入 4 個 word（x、y、物件 far 指標的位移與段），本函式只讀前兩個
 *   a0（[bp+6]）、a1（[bp+8]），描述檔也只宣告這兩個。
 * 回傳：AX ＝ a0、CX ＝ a1，其餘暫存器不保證。不讀不寫記憶體。序言 push bp、mov bp,sp，結尾 pop bp、retf，BP 還原。
 * 用途：3870:0296 依物件 [+6]（0 到 5）查 ds:8D7A 的跳躍表，種類 0 呼叫本函式；docs/re/135 第 221 行記種類 0 為
 *   「什麼都不畫」的佔位。AX／CX 是否被呼叫者使用未查證（呼叫後 3870:0296 直接 mov sp,bp 並跳出 switch，不讀 AX、CX）。
 * 描述檔以 pack:cx,ax 比對兩個暫存器（高 16 位元是 CX、低 16 位元是 AX）。
 * 輸入域：全部 16 位元值。推論等級：暫存器行為已證實（差異測試）；用途為強推論（docs/re/135 加呼叫者 asm）。
 */
#include "gin3_reg.h"

uint32_t f_3870_0372(uint16_t a0, uint16_t a1)
{
    return ((uint32_t)a1 << 16) | a0; /* CX<<16 | AX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3870_0372(a[0], a[1]);
}
GIN3_REGISTER("3870_0372", 2, run)
