/*
 * 3CA8:0002  字元分類表查詢的高位元：ctype 風格的 table[c] & 0x80
 * 呼叫：far，堆疊傳參：c（word）。回傳 AX ＝ ds:[c + 0x93D7] 的位元 7（0 或 0x80），AH＝0。
 * 位移運算在 DGROUP 內 16 位元回繞（c ＝ 0xFFFF 時讀 ds:93D6，即 EOF 之前的表項）。表在 ds:93D7 起，內容由遊戲啟動時決定，
 * 這個函式只讀它。推論等級：查表與遮罩已證實（差異測試）；表項的語意：基底狀態下 ds:93D7+0x30..0x39（字元 0 到 9）的表項是 0x80、其他字元沒有 bit 7，所以像 isdigit（強推論，只有基底狀態的觀察）。
 */
#include "gin3_reg.h"

uint16_t f_3CA8_0002(Mem *m, uint16_t c)
{
    return (uint16_t)(rd8(m, DS((uint16_t)(c + 0x93D7u))) & 0x80u);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3CA8_0002(m, a[0]);
}
GIN3_REGISTER("3CA8_0002", 1, run)
