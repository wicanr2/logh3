/*
 * GIN3PS.EXE 3A8D:0004  空函式（near，只有 retn）
 *
 * 呼叫：near，無參數，無回傳；函式本體只有一個位元組 C3（retn）。
 * 不讀、不寫記憶體。差異測試把 AX、BX、CX、DX 以隨機值放進暫存器，要求執行後原封不動
 * （ret=pack:ax,bx,cx,dx），並比對全 1 MB 的寫入集合為空（堆疊區除外）。
 * 其他暫存器與旗標沒有列入比對；retn 不改它們是指令本身的語意，不是差異測試的結論。
 *
 * 來源與推論等級：
 *   已證實（差異測試）：AX、BX、CX、DX 不變、記憶體不寫、retn 後 IP 回到 near 哨兵。
 *   用途：未知。IDA 命名為 nullsub_10。
 *   IDA 反組譯：靜態上唯一的引用是 3A8D 段內一段不屬於任何台帳函式的片段（call nullsub_10; retf），
 *   該片段能否被執行到沒有驗證；片段前後是 out 7Ch／7Eh 的碼。docs/ 沒有這個位址的記載。
 *   假說：平台相關鉤子被關掉後留下的空殼，沒有證據。
 *   位址 3A8D:0003 也是 C3（前一段碼的 retn），本函式只涵蓋 0004 這一個位元組。
 */
#include "gin3_reg.h"

void f_3A8D_0004(Mem *m)
{
    (void)m; /* retn */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3A8D_0004(m);
    return ((uint64_t)a[0] << 48) | ((uint64_t)a[1] << 32) | ((uint64_t)a[2] << 16) | (uint64_t)a[3];
}
GIN3_REGISTER("3A8D_0004", 4, run)
