/*
 * GIN3PS.EXE 332E:000B  空函式（IDA 名稱 nullsub_4，只有一個 retf）
 *
 * 原版：far 呼叫，沒有參數、沒有回傳值；函式本體只有一個位元組 CB（retf）。
 * 不讀、不寫任何記憶體，也不動 AX BX CX DX。
 * 唯一的呼叫者是 0374:0001 內的 call far 332E:000B（台帳 callers＝1）。docs/ 沒有這個位址的記載。
 * 呼叫端為什麼留一個空呼叫未知（假說：拿掉的功能留下的空位，沒有證據）。
 * 位元組與 042F:0191 相同（同一個 sha256_final_16 383e5d7d58caa41c），是另一個獨立的 nullsub。
 *
 * 退化情況：差異測試能驗證的只有兩件事，一是寫入集合為空，二是暫存器原樣返回。
 * 描述檔用 kind=reg 把 AX BX CX DX 填成隨機值，ret 是 pack:ax,bx,cx,dx，C 版把四個輸入原樣打包回傳。
 *
 * 輸入域：全部可到達（沒有輸入）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為未知。
 */
#include "gin3_reg.h"

void f_332E_000B(Mem *m)
{
    (void)m; /* 不讀不寫 */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_332E_000B(m);
    /* AX BX CX DX 原樣返回，高位到低位 */
    return ((uint64_t)a[0] << 48) | ((uint64_t)a[1] << 32) | ((uint64_t)a[2] << 16) | (uint64_t)a[3];
}
GIN3_REGISTER("332E_000B", 4, run)
