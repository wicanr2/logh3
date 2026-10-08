/*
 * GIN3PS.EXE 20B4:15E1  目前星域切片內每一顆行星的防衛力回復一步
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none）。切片由全域變數指定：
 *   ds:CFF8（位移）、ds:CFFA（段）＝ 行星切片的第一筆；ds:CFFC（byte）＝ 筆數。一筆 82 bytes。
 *
 * 語意（已證實：docs/re/330 §2.3 "20B4:15E1()（proven）"；docs/internals/29 +0x14／+0x16／+0x18；
 *   本檔與反組譯逐條對過）：
 *   對切片每一顆行星 p：
 *     inc   ＝ (p[+0x16] × p[+0x18]) ÷ 100     ; 32 位元乘（3CC9:0009）、無號除（3CC9:0020），只取商的低字
 *     p[+0x14] ← (p[+0x14] + inc) 的低 16 位元   ; 16 位元加，會繞回，不先判溢位
 *     若 p[+0x14] >= p[+0x16]：p[+0x14] ← p[+0x16]   ; 夾上限（cmp ax,cx ＋ jb，無號）
 *   +0x14 防衛力、+0x16 防衛力上限、+0x18 每次回復的百分比。
 *
 * 原版的細節（本版照做）：
 *   - 筆數 ds:CFFC 每圈重讀，計數器是 byte 且無號比較；切片指標只在開頭讀一次。
 *   - 上限 +0x16 在乘法前讀一次、夾值比較時再讀一次（兩次之間只寫 +0x14，不會變）。
 *   - 位移運算 16 位元回繞（add bx,16h／add si,14h／add [bp+2],52h）。
 *   - 夾上限的條件是 cur >= cap；與 cur > cap 在行為上無法區分（相等時寫回同一個值），本版寫成 >=。
 *   - 16 位元加繞回後可能小於上限，此時不夾（例如 0xF0FF ＋ 0xFF00 → 0xEFFF）。這是原版行為，不是本版加的。
 *
 * 呼叫的函式：3CC9:0009、3CC9:0020 已 diff-tested，這裡用 gin3_lmul／gin3_uldiv 直接表達數學結果。
 *
 * 輸入域：切片指標與筆數取基底狀態（星域 0，4 顆）並由突變導向讀取集改動；筆數可到 255，
 *   超出 81 顆時讀寫行星表之後的記憶體，本版同樣照做（平面視訊記憶體的讀取除外，見 docs/re/445 限制）。
 */
#include "gin3_reg.h"

void f_20B4_15E1(Mem *m)
{
    uint16_t off = rd16(m, DS(0xCFF8));
    uint16_t seg = rd16(m, DS(0xCFFA));
    uint8_t i = 0;

    while (i < rd8(m, DS(0xCFFC))) {
        uint16_t cap_off = (uint16_t)(off + 0x16);
        uint16_t cur_off = (uint16_t)(off + 0x14);
        uint16_t cap = rd16(m, LIN(seg, cap_off));
        uint16_t rate = rd16(m, LIN(seg, off + 0x18));
        uint16_t inc = (uint16_t)gin3_uldiv(gin3_lmul(cap, rate), 100u);
        uint16_t cur = (uint16_t)(rd16(m, LIN(seg, cur_off)) + inc);

        wr16(m, LIN(seg, cur_off), cur);
        if (cur >= rd16(m, LIN(seg, cap_off))) {
            wr16(m, LIN(seg, cur_off), rd16(m, LIN(seg, cap_off)));
        }
        i++;
        off = (uint16_t)(off + 0x52);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_20B4_15E1(m);
    return 0;
}
GIN3_REGISTER("20B4_15E1", 0, run)
