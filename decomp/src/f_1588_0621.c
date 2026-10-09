/*
 * GIN3PS.EXE 1588:0621  戰場格 (c, r)（半格座標）的圖塊遮罩，取四個四分之一格中該角落的那一位（回 AX）
 *
 * 原版：far 呼叫，兩個堆疊字（序言 push bp／si／di、sub sp,6，所以 c 在 [bp+10h]、r 在 [bp+12h]），回 AX。
 * 逐條對應反組譯：
 *   t ← 1588:0693(1588:0672(c >> 1, r >> 1))          ; 地圖 (c>>1, r>>1) 那格的圖塊編號 → 圖塊遮罩表（ds:9BB4）一筆
 *                                                    ; c、r 是邏輯右移（shr）；0672 只回 AL（AH 是殘值，0693 只取低位元組）
 *   k ← ((r & 1) << 1) + (c & 1)                      ; 0..3，該格內四個四分之一格的哪一個
 *   回 AX ← t & byte ds:[0x3C3E + k]                  ; `mov cl,[bx+3C3Eh]; xor ch,ch; and ax,cx`，AX 的高位元組為 0
 * 本函式不寫記憶體（兩個被呼叫者也沒有）。
 * 註：r 的右移是邏輯或算術沒有差別（0672 對 y 做 `<< 7` 的 16 位元運算，bit15 被丟掉）；c 的右移則有差別（c >> 1 直接加進位移）。
 *
 * 呼叫者：1588:04EA、1588:0571（docs/re/262「1588:0621(c, r) 通行判定」、docs/internals/43）。
 *
 * 推論等級
 *   已證實（差異測試）：移位方向與種類、奇偶索引的組成（r 為高位）、表基底 0x3C3E、AND、回傳的高位元組為 0。
 *   強推論（docs/re/262、internals/43）：ds:3C3E 起的 4 bytes（基底狀態是 00 02 04 08）是四個四分之一格的位元遮罩；
 *     ds:9BB4 起 128 bytes 是 ATTR.DAT 的圖塊遮罩（0F ＝ 通行）。名稱為強推論。
 * 輸入域：c、r 為 16 位元全域（右移後仍可超出 128 × 102，0672 照原版讀鄰近記憶體，位移 16 位元回繞）；地圖指標取基底狀態。
 */
#include "gin3_reg.h"

extern uint8_t f_1588_0672(Mem *m, uint16_t x, uint16_t y);
extern uint8_t f_1588_0693(Mem *m, uint16_t idx);

uint16_t f_1588_0621(Mem *m, uint16_t c, uint16_t r)
{
    uint8_t t = f_1588_0693(m, f_1588_0672(m, (uint16_t)(c >> 1), (uint16_t)(r >> 1)));
    uint16_t k = (uint16_t)((uint16_t)((r & 1u) << 1) + (c & 1u));

    return (uint16_t)(t & rd8(m, DS((uint16_t)(0x3C3Eu + k))));
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_0621(m, a[0], a[1]); }
GIN3_REGISTER("1588_0621", 2, run)
