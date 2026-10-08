/*
 * GIN3PS.EXE 25DB:0F63  回傳第一個參數（8 bytes，恆等函式）
 *
 * 原版：far 呼叫，一個 word 參數，回 AX = 該參數。
 *   push bp; mov bp,sp; mov ax,[bp+6]; pop bp; retf
 *   堆疊：序言只 push bp，所以參數在 [bp+6]（[bp+2] ＝ 返回位移、[bp+4] ＝ 返回段）。
 *
 * 語意（已證實：本檔與反組譯逐條對過）：
 *   AX ← 整個 16 位元參數，不截成 byte，不做符號處理；不讀不寫記憶體，不碰旗標以外的暫存器。
 *
 * 用途（強推論，來自呼叫者的反組譯，未另行驗證）：唯一的呼叫者 25DB:0E78 在 `si` 的 bit15 為 1 時
 *   `push si; call 25DB:0F63`，呼叫後直接丟掉 AX（`mov sp,bp` 後沒有讀 AX），所以它在遊戲內是一個
 *   「什麼都不做」的替代路徑；本函式只證明回傳值等於參數。
 *
 * 輸入域：任意 16 位元參數。
 */
#include "gin3_reg.h"

uint16_t f_25DB_0F63(Mem *m, uint16_t v)
{
    (void)m;
    return v;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_25DB_0F63(m, a[0]); }
GIN3_REGISTER("25DB_0F63", 1, run)
