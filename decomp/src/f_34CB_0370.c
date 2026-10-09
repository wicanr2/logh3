/*
 * 34CB:0370  清掉視窗範圍內「屬於本視窗」的畫面格：旗標設成 0xA0、字碼低位元組清 0
 *
 * 原版：far 呼叫，參數是視窗結構的 far 指標（位移 [bp+14h]、段 [bp+16h]；序言 push bp／si／di、sub sp,0Ah）。
 *   唯一呼叫者 341B:01FA 內的 sub_40F1A（+0x33）。回傳值沒人用，描述檔 ret = none。
 *
 * 讀視窗結構（docs/internals/47-window-frame-tiles.md「結構位移」；名稱是本檔取的）：
 *   +4 槽位編號  +5 格 x  +6 格 y  +7 格寬  +8 格高（皆 byte）。
 * 畫面格表：ds:D81E 起，40 欄 × 20 列、一格 6 bytes、一列 0xF0 bytes（docs/internals/31、47）。
 *
 * 動作（已證實：反組譯逐條對過）：
 *   格起點 = D81E + (+6) × 0xF0 + (+5) × 6（16 位元回繞），段 = 呼叫時的 DS（GIN3_DS）。
 *   外層 i = 0 起，i < (+8)（每一輪重讀 +8）：列起點 = 目前格位移；格位移 += 0xF0；
 *       內層 j = 0 起，j < (+7)（每一輪重讀 +7）：
 *           若 (格[+0] & 0x0F) == (+4)（byte 比較，每次重讀 +4）：格[+0] = 0xA0；格[+2] = 0
 *           格位移 += 6
 *   0xA0 = 髒（bit7）＋ 鎖住（bit5）（docs/re/138、docs/internals/47 的位元表）；擁有者（低四位）
 *   被清成 0，所以這些格子不再屬於本視窗。格[+3]（字碼高位元組）不動，只有字碼低位元組清 0。
 *   視窗結構與畫面格表重疊時，寫入會改到之後重讀的 +4／+7／+8；差異測試用 seeds 造這種別名並有對應突變。
 *
 * 語意（強推論）：關閉視窗時（341B:01FA）把自己佔的格子清掉並標髒，讓刷新常式重畫底下的內容。
 *   0xA0 兩個位元的用意（髒 ＋ 0x20）與 341B:01FA 的整體流程未在本檔驗證：未知。
 * 輸入域：視窗結構任意 byte 值，位移全部 16 位元回繞。寬 255、高 255 時超過探針步數上限的案例被丟棄；
 *   視窗指標段指到 A0000 到 BFFFF（視訊記憶體）的案例被探針丟棄。
 */
#include "gin3_reg.h"

static uint8_t f34CB0370_wb(const Mem *m, uint16_t seg, uint16_t off, uint16_t n)
{
    return rd8(m, LIN(seg, (uint16_t)(off + n)));
}

void f_34CB_0370(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    const uint16_t ds = GIN3_DS;
    uint16_t cur = (uint16_t)(0xD81Eu + (uint16_t)(f34CB0370_wb(m, w_seg, w_off, 6) * 0xF0u)
                              + (uint16_t)(f34CB0370_wb(m, w_seg, w_off, 5) * 6u));
    uint8_t i = 0;

    while (i < f34CB0370_wb(m, w_seg, w_off, 8)) {
        uint16_t cell = cur;
        uint8_t j = 0;

        cur = (uint16_t)(cur + 0xF0u);
        while (j < f34CB0370_wb(m, w_seg, w_off, 7)) {
            if ((rd8(m, LIN(ds, cell)) & 0x0Fu) == f34CB0370_wb(m, w_seg, w_off, 4)) {
                wr8(m, LIN(ds, cell), 0xA0);
                wr8(m, LIN(ds, (uint16_t)(cell + 2u)), 0);
            }
            cell = (uint16_t)(cell + 6u);
            j++;
        }
        i++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_34CB_0370(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("34CB_0370", 2, run)
