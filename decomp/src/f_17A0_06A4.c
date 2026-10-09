/*
 * GIN3PS.EXE 17A0:06A4  把行星記錄的 +0x44 位元組設成指定值（補給候選順位的行星版）
 *
 * 原版：far 呼叫，兩個 word 參數（planet, value），無回傳值（描述檔 ret = none）。
 *   序言 push bp／si、sub sp,4，所以 planet 在 [bp+0Ch]、value 在 [bp+0Eh]。
 *   行星表基底由全域變數指定：ds:CFF8（位移）、ds:CFFA（段）；一筆 82 bytes。
 *
 * 語意（已證實：docs/re/131 "17A0:05B2 發順位……行星走 17A0:06A4"、"行星則是記錄 +0x44"；
 *   本檔與反組譯逐條對過；+0x44 這一格的名字「順位」是強推論，docs/internals/29 沒有記這格）：
 *   rec ＝ (ds:CFF8) ＋ ((planet & 0x7FFF) × 0x52 的低 16 位元)   ; 位移運算 16 位元回繞
 *   byte rec[+0x44] ← 低位元組(value)                             ; 高位元組被忽略
 *
 * 原版的細節（本版照做）：
 *   - planet 的 bit15 被清掉（and ah,7Fh）。但 0x52 是偶數，bit15 對 16 位元乘積的貢獻
 *     是 0x8000 × 0x52 ≡ 0 (mod 0x10000)，所以這個遮罩不改變結果。用同一個 16 位元乘積的版本
 *     無法分辨有沒有遮罩，這是等價而非遺漏。bit14 的貢獻是 0x8000，會改變結果，所以遮罩寬度可以驗。
 *   - 乘積只留低 16 位元（mul cx 之後只用 AX）。
 *
 * 輸入域：planet 與 value 都是 16 位元全域；行星表基底取基底狀態（星域 0，4 顆）並由突變改動。
 *   planet 超過筆數時寫到表後面的記憶體，本版同樣照做。段變數被改成 0xF000 以上時，寫入落在
 *   BIOS ROM 區（線性 0xF0000 起），機器忽略該寫入（dosgolem internal/machine/bios_rom.go，真機同為無效）；
 *   本版用 wr8_skip_rom 表達這一點。平面視訊記憶體（0xA0000 起）的寫入由探針丟棄，不在輸入域內。
 */
#include "gin3_reg.h"

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意（回報給主工作階段）。 */
static void wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_17A0_06A4(Mem *m, uint16_t planet, uint16_t value)
{
    uint16_t off = (uint16_t)(rd16(m, DS(0xCFF8)) + (uint16_t)((planet & 0x7FFFu) * 0x52u));
    uint16_t seg = rd16(m, DS(0xCFFA));

    wr8_skip_rom(m, LIN(seg, off + 0x44u), (uint8_t)value);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_17A0_06A4(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("17A0_06A4", 2, run)
