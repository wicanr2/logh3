/*
 * GIN3PS.EXE 3ACA:0944  printf 的正負號前綴選擇：回傳指向 "-"、"+"、" "、"" 其中之一的 far 指標
 *
 * 呼叫：far，兩個堆疊參數（各 1 個字）：
 *   [bp+6]  neg   整個 word 測是否為 0（or cx,cx）；非 0 表示負數，回 "-"
 *   [bp+8]  flags 只測低位元組（mov ax,dx; test al,2／test al,4），高位元組不看
 *             bit 1（0x02）：要求顯示正號，回 "+"
 *             bit 2（0x04）：要求以空白代替正號，回 " "
 * 回傳 AX:BX 是 far 指標（AX＝DS，BX＝位移），不是 DX:AX。
 * 位移 DS:90FC／90FE／9100／9102 是四個 2 位元組的 C 字串，基底狀態實測內容為
 * "-\0"、"+\0"、" \0"、"\0\0"（od 讀 fcmd-base.state 傾印，已證實）。本函式只回指標，不讀這些位元組。
 *
 * 判斷順序（反組譯）：neg != 0 → 90FC；否則 flags 低位元組 bit1 → 90FE；否則 bit2 → 9100；否則 9102。
 * 兩個 bit 同時設時 bit1 優先（回 "+"）。
 * 呼叫者：只有 3ACA:0079（printf 本體），3 處。
 *
 * 輸入域：全部 16 位元值。無記憶體讀寫，無副作用。
 * 推論等級：反組譯行為已證實（差異測試）；「printf 的正負號前綴」是用途推斷（強推論：四個字串的內容與旗標語意吻合 C 的 '+'、' ' 旗標）。
 */
#include "gin3_reg.h"

/* 回傳打包成 (AX<<16)|BX */
uint32_t f_3ACA_0944(uint16_t neg, uint16_t flags)
{
    uint16_t bx;

    if (neg != 0)
        bx = 0x90FC;
    else if ((uint8_t)flags & 2)
        bx = 0x90FE;
    else if ((uint8_t)flags & 4)
        bx = 0x9100;
    else
        bx = 0x9102;
    return ((uint32_t)GIN3_DS << 16) | bx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3ACA_0944(a[0], a[1]);
}
GIN3_REGISTER("3ACA_0944", 2, run)
