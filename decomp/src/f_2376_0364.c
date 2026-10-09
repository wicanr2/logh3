/*
 * GIN3PS.EXE 2376:0364  畫面格表矩形操作（2376:02DE）的 far 包裝：把堆疊上的 far 指標載入 ES:SI 後呼叫 near 的 2376:02DE
 *
 * 原版（16 bytes，逐指令）：
 *   enter 0,0 ; push si ; push es
 *   les si, [bp+6]                ; 堆疊參數：far 指標（位移 [bp+6]、段 [bp+8]）→ SI ＝ 位移、ES ＝ 段
 *   call 2376:02DE                ; near，暫存器傳參 ES:SI（已 diff-tested）；AX 是進入值，02DE 不使用 AL、覆蓋 AH
 *   pop es ; pop si ; leave ; retf
 *   返回時 AX、BX、DX、旗標是 02DE（00B2）的殘值；呼叫者 35D1:0102 之後只用記憶體，描述檔 ret ＝ none，只比對記憶體寫入。
 * 參數：args[0] ＝ 物件位移、args[1] ＝ 物件段（兩個堆疊字）。
 * 副作用：全部來自 02DE → 00B2（畫面格表 ds:D81E 起，格 [0] ← 視窗鍵值 | 0x80、格 [1] |= 1，詳見 f_2376_00B2.c）。
 * 呼叫者（台帳）：35D1:0102（game-logic）。docs/re/334 第 30 行：p[+23] == 0 → 2376:0364(p) → 2376:02DE。
 * 02DE 與 02F2 讀物件的欄位位移不同（+0x24 對 +0x1F），所以本包裝若改呼叫 02F2 會被突變抓到。
 * 輸入域：同 02DE（列數或每列格數為 0 時原版繞圈、超過步數上限被探針丟棄）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern void f_2376_02DE(Mem *m, uint16_t ax, uint16_t es, uint16_t si);

void f_2376_0364(Mem *m, uint16_t obj_off, uint16_t obj_seg)
{
    /* AX 是進入值；02DE 只用 AH（被覆蓋）而不讀 AL，傳 0 */
    f_2376_02DE(m, 0, obj_seg, obj_off);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2376_0364(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2376_0364", 2, run)
