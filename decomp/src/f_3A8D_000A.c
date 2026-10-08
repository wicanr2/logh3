/*
 * GIN3PS.EXE 3A8D:000A  空函式（far，只有 retf）
 *
 * 呼叫：far，無參數，無回傳；函式本體只有一個位元組 CB（retf）。
 * 不讀、不寫記憶體。差異測試把 AX、BX、CX、DX 以隨機值放進暫存器，要求執行後原封不動
 * （ret=pack:ax,bx,cx,dx），並比對全 1 MB 的寫入集合為空（堆疊區除外）。
 * 其他暫存器與旗標沒有列入比對；retf 不改它們是指令本身的語意，不是差異測試的結論。
 *
 * 來源與推論等級：
 *   已證實（差異測試）：AX、BX、CX、DX 不變、記憶體不寫、retf 後回到 far 哨兵。
 *   已證實（IDA 反組譯）：唯一呼叫者是 3931:0002（IDA sub_45E82）的初始化序列，在 call sub_45EC3 之後
 *   以 far call 呼叫本函式，不傳參數、不使用回傳值。
 *   用途：未知。IDA 命名為 nullsub_5。docs/ 沒有這個位址的記載。
 *   注意：docs/re/324 寫的 `3A8D:000B(0)` 是下一個位元組開始的另一支函式（55 8B EC 開頭的有框架函式），
 *   不是這支空函式；兩者位址相差 1。
 *   假說：平台相關鉤子被關掉後留下的空殼（前一支 3A8D:0005 是 xor al,al; out 7Ch,al; retn），沒有證據。
 */
#include "gin3_reg.h"

void f_3A8D_000A(Mem *m)
{
    (void)m; /* retf */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3A8D_000A(m);
    return ((uint64_t)a[0] << 48) | ((uint64_t)a[1] << 32) | ((uint64_t)a[2] << 16) | (uint64_t)a[3];
}
GIN3_REGISTER("3A8D_000A", 4, run)
