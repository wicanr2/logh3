/*
 * GIN3PS.EXE 20B4:18C2  戰略層打有主行星時，把攻方艦隊的每個分隊整支揭露
 *
 * 原版：far 呼叫，一個參數 = 艦隊記錄 far 指標（位移、段）。無回傳值（描述檔 ret = none）。
 *
 * 語意（已證實：docs/re/330 §2.5 "20B4:18C2(艦隊 far)（proven）"，執行期亞斯提後艦艇 0 的
 *   +0x17／+0x18 ＝ 03 02；docs/internals/38 與 docs/internals/53 同結論；本檔與反組譯逐條對過）：
 *   for i in 0 ..< 艦隊[+0x0F]：
 *       s ← 艦艇池 ＋ 艦隊[+0x10 ＋ 4i](word) × 78
 *       s[+0x17] ← 3
 *       s[+0x18] ← 2
 *   +0x17／+0x18 在 docs 裡稱「索敵等級」，兩格各自的意義見 docs/internals/53，本函式只管寫入。
 *
 * 原版的細節（本版照做）：
 *   - 艦艇池基底在 ds:CFB4（位移）、ds:CFB6（段），每圈重讀；寫入發生在讀取之後，所以池若與
 *     艦隊記錄或這兩個全域重疊，後面的圈讀到的是被寫過的值。
 *   - 艦隊[+0x0F]（分隊數，byte）每圈重讀，計數器是 byte，無號比較。
 *   - 池索引 × 0x4E 只取低 16 位元（mul cx 後只用 AX）；位移加法 16 位元回繞。
 *
 * 輸入域：艦隊指標落在艦隊表 ds:C074 的 32 筆；分隊數與池索引取基底狀態並由突變導向讀取集改動，
 *   索引沒有上限檢查（超出池範圍的索引寫到池之外），本版同樣照做。
 */
#include "gin3_reg.h"

void f_20B4_18C2(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i = 0;
    uint16_t slot = (uint16_t)(fleet_off + 0x10);

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        uint16_t index = rd16(m, LIN(fleet_seg, slot));
        uint16_t pool_seg = rd16(m, DS(0xCFB6));
        uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(index * 0x4Eu));

        wr8(m, LIN(pool_seg, ship + 0x17), 3);
        wr8(m, LIN(pool_seg, ship + 0x18), 2);
        i++;
        slot = (uint16_t)(slot + 4);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_20B4_18C2(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("20B4_18C2", 2, run)
