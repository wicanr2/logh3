/*
 * GIN3PS.EXE 2376:0354  顯示物件的新舊矩形更新的 far 包裝：把堆疊上的 far 指標載入 ES:SI，呼叫 near 的 2376:01E6
 *
 * 原版（16 bytes，逐指令）：
 *   enter 0,0 ; push si ; push es
 *   les si, [bp+6]                ; 堆疊參數：far 指標（位移在低位址 [bp+6]、段在 [bp+8]）→ SI ＝ 位移、ES ＝ 段
 *   call 2376:01E6                ; near，暫存器傳參 ES:SI（已 diff-tested）
 *   pop es ; pop si ; leave ; retf
 *   回傳 AX 是 01E6 的 AX（pop／leave 不碰 AX）。
 * 參數：args[0] ＝ 物件位移、args[1] ＝ 物件段（兩個堆疊字，farptr 的順序）。
 * 回傳：AX（01E6 的回傳值；分離路徑 AL ＝ 1、AH ＝ cbw 殘值，其餘路徑 AH ＝ 0）。
 * 副作用：全部來自 01E6（寫物件 +16..+27、+1E..+22，並在堆疊帶內放區域變數）。ES、SI 在返回前還原，不影響記憶體。
 * 呼叫者（台帳）：35D1:0009（unknown）、35D1:0102（game-logic）。docs/re/334 第 30 行：35D1:0102(p) → 2376:0354(p) → 2376:01E6。
 *
 * 堆疊上的區域變數：01E6 的 C 版把它們寫到「進入點 SP − 4」往下的位置（f_2376_01E6.c 檔頭），本包裝的真實框架更深
 * （多了 far 回位址、bp、si、es 與近呼叫哨兵），位置不同，但兩者都落在堆疊帶（進入點 SP 以下 0x400 bytes）內，
 * 比對時排除，不影響寫入集合與回傳值。
 * 輸入域：同 01E6（物件不蓋到堆疊帶；ES 在 ROM 區的情形不涵蓋）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_2376_01E6(Mem *m, uint16_t es, uint16_t si);

uint16_t f_2376_0354(Mem *m, uint16_t obj_off, uint16_t obj_seg)
{
    return f_2376_01E6(m, obj_seg, obj_off);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2376_0354(m, a[0], a[1]); }
GIN3_REGISTER("2376_0354", 2, run)
