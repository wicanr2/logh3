/*
 * GIN3PS.EXE 39D7:0B4B  空函式（IDA 名稱 nullsub_9，只有一個 near 返回）
 *
 * 原版：near 呼叫，沒有參數、沒有回傳值；函式本體只有一個位元組 C3（retn）。
 * 位元組取自解包映像（gin3ps-final.bin）偏移 0x3742B，前一個位元組是 CB（retf，屬於前一段程式），
 * 後一段是 B0 C0（mov al,0C0h，繼續是 out 7Ch,al 的 PC-98 風格輸出碼，屬於別的函式）。
 * 不讀、不寫任何記憶體，也不動任何暫存器。
 * 唯一的呼叫者在同一模組（39D7，畫線模組）的 call near 39D7:0B4B；
 * 呼叫端為什麼留一個空呼叫未知（推測是依顯示裝置分派的空位，假說）。
 *
 * 這是退化情況：差異測試能驗證的只有兩件事，一是寫入集合為空，二是暫存器原樣返回。
 * 為了驗第二件，描述檔用 kind=reg 把 AX BX CX DX 填成隨機值，ret 是 pack:ax,bx,cx,dx，
 * C 版把四個輸入原樣打包回傳。本函式自己不碰視訊記憶體、I/O 與 int。
 *
 * 輸入域：全部可到達（沒有輸入）。
 * 推論等級：函式只是一個 retn：已證實（反組譯逐指令＋差異測試）；用途：未知。
 */
#include "gin3_reg.h"

void f_39D7_0B4B(Mem *m)
{
    (void)m; /* 不讀不寫 */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_39D7_0B4B(m);
    /* AX BX CX DX 原樣返回，高位到低位 */
    return ((uint64_t)a[0] << 48) | ((uint64_t)a[1] << 32) | ((uint64_t)a[2] << 16) | (uint64_t)a[3];
}
GIN3_REGISTER("39D7_0B4B", 4, run)
