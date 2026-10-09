/*
 * GIN3PS.EXE 3B6B:0025  半形字元轉全形碼（Big5 全形，呼叫者 3B6B:0002 只對 0x20..0x7E 呼叫）
 *
 * 呼叫：far，一個堆疊參數 c（1 個字，整個 word 參與比較）。回傳 AX。
 * 演算法（docs/re/141 第 90 行起的描述與本版相符，常數與項數都對得上）：
 *   '0'..'9'（0x30..0x39）  → c + 0xA27F   （16 位元回繞；'0' → 0xA2AF）
 *   'a'..'z'（0x61..0x7A）  → c + 0xA288   （'a' → 0xA2E9）
 *   'A'..'Z'（0x41..0x5A）  → c + 0xA28E   （'A' → 0xA2CF）
 *   其餘：在 34 筆小表裡由前往後找 key == c：key 表在 ds:92C6（word × 34），值表在 ds:930A（word × 34），
 *         第一個符合的算數，回傳同索引的值；找不到就回傳 c 本身。
 * 三個範圍比較是有號比較（jl／jg），c >= 0x8000 時視為負數，不落在任何範圍，
 * 之後 key 比較是完整 16 位元相等（cmp cx,ax; jnz）。表在 DGROUP（DS 固定 54B3），
 * C 版從記憶體讀表，不把基底狀態的內容寫死：差異測試的隨機突變會改表項。
 * 基底狀態實測：key 表 0x20 0x21 0x60 0x23 0x24 0x25 0x26 0x22 0x27..0x2F 0x3A..0x40 0x5B 0x5C 0x5D 0x5E 0x5F 0x60 0x7B..0x7E，
 * 其中 0x60 出現兩次（索引 2 與 29，由前往後找所以索引 2 優先）；0x22 對到 0xA149（與 0x21 相同）。
 * 副作用：只讀記憶體（兩張表），不寫。暫存器：除 AX 外只改 DX、CX、BX、SI、旗標；不影響呼叫端（呼叫端 3B6B:0002 只用 AX）。
 *
 * 輸入域：全部 16 位元值。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）；Big5 全形碼的語意判讀是強推論
 * （0xA2AF 在 Big5 是全形「０」，其餘常數吻合；遊戲內碼的編碼取捨見 docs/re/141）。
 */
#include "gin3_reg.h"

#define F3B6B_KEYS 0x92C6u /* key 表：34 個 word */
#define F3B6B_VALS 0x930Au /* 值表：34 個 word */

uint16_t f_3B6B_0025(Mem *m, uint16_t c)
{
    int16_t s = (int16_t)c;
    uint16_t i;

    if (s >= 0x30 && s <= 0x39)
        return (uint16_t)(c + 0xA27F);
    if (s >= 0x61 && s <= 0x7A)
        return (uint16_t)(c + 0xA288);
    if (s >= 0x41 && s <= 0x5A)
        return (uint16_t)(c + 0xA28E);
    for (i = 0; i <= 0x21; i++) {
        if (c == rd16(m, DS((uint16_t)(F3B6B_KEYS + i * 2u))))
            return rd16(m, DS((uint16_t)(F3B6B_VALS + i * 2u)));
    }
    return c;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3B6B_0025(m, a[0]); }
GIN3_REGISTER("3B6B_0025", 1, run)
