/*
 * 35D1:0002  把戰場圖示物件清單的筆數 ds:EADE 清成 0（far，無參數，DS = DGROUP）
 *
 * 原版：`mov word ptr ds:0EADEh, 0; retf`，兩條指令，不讀記憶體，不改暫存器。
 * 呼叫者 341B:0004（開機初始化序列的一步，sub_40D24+0x2C）。
 * ds:EADE 是 word；docs/re/376、377、394 記為「DS:EADE 筆數、DS:EAE0 起是遠指標陣列」（清單內容不被本函式清除，
 *   只清筆數）。35D1:0194、01DE 兩支走這份清單。
 *
 * 輸入域：沒有輸入。函式不讀記憶體，案例全靠 seeds：預先把 ds:EADE 的兩個位元組與相鄰位元組設成非 0，
 *   才看得出少寫（只寫低位元組）、寫錯值與位址差 1。
 * 推論等級：寫入的值、寬度與位址已證實（反組譯逐指令＋差異測試）；清單語意見上述 docs。
 */
#include "gin3_reg.h"

void f_35D1_0002(Mem *m)
{
    wr16(m, DS(0xEADEu), 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_35D1_0002(m);
    return 0;
}
GIN3_REGISTER("35D1_0002", 0, run)
