/*
 * GIN3PS.EXE 3ABF:003B  free(p) 的 far 包裝：把兩個堆疊參數原樣轉給 3CD9:01FC，回傳它的 AX
 *
 * 呼叫：far，兩個堆疊參數：p 位移 [bp+6]、p 段 [bp+8]（與 3CD9:01FC 的參數佈局相同）。
 * 本體：push bp; mov bp,sp; push [bp+8]; push [bp+6]; call far 3CD9:01FC; add sp,4; pop bp; retf。
 * 回傳 AX＝3CD9:01FC 的回傳（0 成功，含 p＝0:0 的空操作；0xFFFF 失敗，見該檔檔頭）。
 * 其餘暫存器沒有保證（BX、CX、DX、SI、DI 由 3CD9:01FC 決定，不屬於本函式的輸出，差異測試不比）。
 * 3CD9:01FC 是 free（強推論，docs/internals/03 §5；已 diff-tested），本函式因此是 free 的 far 包裝；
 * 台帳 callers＝19，docs/re/314 第 51 行把它寫成「釋放地圖與緩衝區」。
 * 該 call 的目標段有 1 個重定位（台帳 relocs＝1，sha256_16 與 sha256_final_16 不同），
 * 描述檔的 code_sha256_16 取重定位後的值（sha256_final_16）。
 *
 * 記憶體副作用：全部來自 3CD9:01FC（自由串列的節點與 rover），本函式自己不讀寫記憶體。
 * 輸入域：同 3CD9:01FC（全部 16 位元值，串列要是良構；損壞的串列原版會無窮迴圈，以步數上限丟棄）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD9_01FC(Mem *m, uint16_t off, uint16_t seg);

uint16_t f_3ABF_003B(Mem *m, uint16_t off, uint16_t seg)
{
    return f_3CD9_01FC(m, off, seg);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3ABF_003B(m, a[0], a[1]); }
GIN3_REGISTER("3ABF_003B", 2, run)
