/*
 * GIN3PS.EXE 1588:0930  戰場鏡頭初始化：依序為座位 0、座位 1 呼叫 1588:0945（回 AX ＝ 座位 1 那次的回傳）
 *
 * 原版（21 bytes，far 呼叫，沒有堆疊參數）：
 *   xor ax,ax ; push ax ; call 1588:0945 ; (清參數)        ; 座位 0
 *   mov al,1  ; push ax ; call 1588:0945 ; (清參數)        ; 座位 1（AH 是前一次 0945 的回傳殘值，0945 只取座位的低位元組）
 *   retf                                                   ; AX 是第二次 0945 的 AX（y × 20 的低 16 位元）
 * 副作用：全部來自兩次 0945（寫 ds:8F3A ＋ 4 × 座位 的 x << 4 與 y × 20；借用 0x9D000 的暫存區淨影響為零）。
 * 呼叫者（台帳）：12AA:026C（game-logic）。docs/re/393-battle-work-control.md 第 41 行「1588:0930 依次呼叫 0945(0)／0945(1)」，與本檔一致。
 * 輸入域：同 0945（座位固定 0、1；艦隊表、在場艦隊清單 ds:9898、筆數 ds:9894、座位 → 陣營表 ds:8F25 取自記憶體）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_0945(Mem *m, uint16_t seat_word);

uint16_t f_1588_0930(Mem *m)
{
    f_1588_0945(m, 0u);
    return f_1588_0945(m, 1u);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_1588_0930(m);
}
GIN3_REGISTER("1588_0930", 0, run)
