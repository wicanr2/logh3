/*
 * GIN3PS.EXE 07F9:02CB  星域的艦種存量夠不夠（far，三個參數，回 AX = 0 或 1，不寫記憶體）
 *
 * 參數（堆疊字，由低位址到高位址）：
 *   region  far 指標（位移、段）：星域記錄（ds:A79C 起，一筆 126 bytes，docs/internals/22）
 *   faction 一個字，只用低位元組（mov bl,[bp+..]；xor bh,bh）
 *   type    一個字，只用低位元組
 * 算式（asm 的乘法全是 mul，位移運算 16 位元回繞）：
 *   stock = 32 位元 @ region:[ +0x30 + faction × 28 + type × 4 ]   （低字在前、高字在 +2）
 *   thr   = 16 位元 @ ds:[ 0x4E3A + faction × 208 + type × 26 ]    （無號 word，高字視為 0）
 *   回 (stock >= thr) ? 1 : 0     （stock 高字非 0 直接成立；高字為 0 才比低字，jnb 無號）
 * 位移相加用 uint16_t，與原版 16 位元暫存器相同；stock 的高字位移是 (位移 + 2) 同樣 16 位元回繞。
 *
 * 推論等級
 *   已證實（差異測試）：上列位移與比較。
 *   強推論（docs/re/182、docs/internals/29）：region+0x30 起每陣營 28 bytes = 7 個 32 位元艦種存量，
 *     ds:4E3A 是艦種表（一筆 26 bytes、一陣營 8 筆 = 208 bytes），thr 是艦種表第一個 word（欄位語意未驗）。
 *     faction／type 的名稱來自呼叫端 07F9:023B（陣營 = 提督[+0x01]、艦種 = 提督[+0x08] & 0Fh）。
 * 輸入域：region 取星域表 32 筆；faction、type 的低位元組 0..255 全域，高位元組任意。
 *   控制流不依賴 faction、type 的值，抽樣取等價類代表（0、1、2、邊界 255、7/8/15/16、高位元組有值的變體）。
 */
#include "gin3_reg.h"

uint16_t f_07F9_02CB(Mem *m, uint16_t region_off, uint16_t region_seg, uint16_t faction_w, uint16_t type_w)
{
    uint16_t faction = (uint16_t)(faction_w & 0xFFu);
    uint16_t type = (uint16_t)(type_w & 0xFFu);
    uint16_t slot = (uint16_t)(region_off + 0x30u + faction * 28u + type * 4u);
    uint16_t stock_lo = rd16(m, LIN(region_seg, slot));
    uint16_t stock_hi = rd16(m, LIN(region_seg, (uint16_t)(slot + 2u)));
    uint16_t thr = rd16(m, DS((uint16_t)(0x4E3Au + faction * 208u + type * 26u)));

    if (stock_hi != 0)
        return 1;
    return stock_lo >= thr ? 1 : 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_07F9_02CB(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("07F9_02CB", 4, run)
