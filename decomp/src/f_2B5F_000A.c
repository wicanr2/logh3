/*
 * GIN3PS.EXE 2B5F:000A  清空 ds:8F9C 起 0x38 bytes（兩個陣營各 7 艦種 × 32 位元的交付計劃）
 *
 * 原版：far 呼叫，沒有參數，只做一件事：memset(ds:8F9C, 0, 0x38)（呼叫 3C9C:000A）。
 * ds:8F9C[陣營] 是次月底交付的艘數計劃，每陣營 0x1C bytes（7 艦種 × 32 位元），共 2 個陣營＝0x38
 * （docs/re/264、docs/re/265：已證實）。只在新遊戲初始化跑一次（docs/re/264）。
 * 回傳：沒有回傳值。離開時 AX、BX 是 memset 的殘值：AX＝DS、BX＝0x8F9C（3C9C:000A 回 AX＝段、BX＝位移）。
 *   唯一的呼叫者 0374:023F 呼叫後直接 push 下一個呼叫的參數，不使用（反組譯）。差異測試比較 AX:BX（pack:ax,bx）。
 * 呼叫：3C9C:000A（已 diff-tested）。
 * 副作用：寫 ds:8F9C..8FD3 共 0x38 個 byte 為 0。不讀記憶體。輸入域：無參數，全部可到達。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);

uint32_t f_2B5F_000A(Mem *m)
{
    return f_3C9C_000A(m, 0x8F9C, GIN3_DS, 0, 0x38);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2B5F_000A(m);
}
GIN3_REGISTER("2B5F_000A", 0, run)
