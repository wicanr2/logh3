/*
 * GIN3PS.EXE 37A9:05DA  free(p) 的薄包裝：把兩個參數原樣轉給 3CD9:01FC 並回傳它的結果
 * 呼叫：far，一個 far 指標參數 p（位移在 [bp+6]、段在 [bp+8]）。序言 push bp、mov bp,sp；
 *   push [bp+8]（段）、push [bp+6]（位移）、call far 3CD9:01FC、mov sp,bp、pop bp、retf。
 * 回傳 AX：3CD9:01FC 的回傳（0 成功，含 p ＝ 0:0 的空操作；0xFFFF 失敗）。其餘暫存器是 free 的出口值，不比對。
 * 副作用：全部是 3CD9:01FC 的副作用（自由串列、rover ds:94E6／94E8），C 版直接呼叫已 diff-tested 的 f_3CD9_01FC。
 * 用途：4 個呼叫者（042F:0655、090A:02AF、090A:0415、0F7A:04C7）用這支釋放物件。函式名「free」是 3CD9:01FC 的強推論（docs/internals/67），
 *   本函式是它的遊戲端包裝（強推論）。
 * 輸入域：同 3CD9:01FC（全部 16 位元的 p；串列要良構）。推論等級：已證實（asm 與差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD9_01FC(Mem *m, uint16_t off, uint16_t seg);

uint16_t f_37A9_05DA(Mem *m, uint16_t off, uint16_t seg)
{
    return f_3CD9_01FC(m, off, seg);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_37A9_05DA(m, a[0], a[1]);
}
GIN3_REGISTER("37A9_05DA", 2, run)
