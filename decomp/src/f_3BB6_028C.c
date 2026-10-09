/*
 * 3BB6:028C  清除模組資料段的一個位元組旗標 ds:01AB（near，無參數，無回傳值）
 *
 * 原版：mov byte ptr ds:1ABh,0；retn。只寫一個 byte（ds:01AB ← 0）；不讀記憶體。AX 等暫存器不動。
 *   DS：呼叫者 3BB6:01FB 在 call 之前 `mov ds,seg130`（dosgolem 下 0x4AFE，載入位置相依），描述檔用 ds_seg 4AFE；
 *   C 版把 ds 當參數。
 *
 * 語意（假說）：ds:01AB 是 3BB6 模組的旗標位元組：3BB6:07D7 寫 1、本函式寫 0（3BB6:01FB 先呼叫本函式再呼叫 3BB6:0003）。
 *   名稱與用途：未知。
 * 推論等級：寫入位址與值已證實（逐指令對過）；用途假說。
 * 輸入域：無輸入。基底狀態的 ds:01AB 已是 0，寫入集合為空看不出差異，描述檔的 seeds 先把 ds:01AA..01AC poke 成別的值。
 */
#include "gin3_reg.h"

void f_3BB6_028C(Mem *m, uint16_t ds)
{
    wr8(m, LIN(ds, 0x01AB), 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_3BB6_028C(m, 0x4AFEu); /* 描述檔 ds_seg：呼叫者 3BB6:01FB 載入的 seg130 */
    return 0;
}
GIN3_REGISTER("3BB6_028C", 0, run)
