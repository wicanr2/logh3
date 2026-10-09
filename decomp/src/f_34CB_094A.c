/*
 * 34CB:094A  視窗外框：在視窗矩形外圈的格子寫外框字碼 0xFF00 + 起始碼 + n（只寫「沒鎖住」的格子）
 *
 * 原版：far 呼叫，參數是視窗結構的 far 指標（位移 [bp+18h]、段 [bp+1Ah]；序言 push bp／si／di、sub sp,0Eh）。
 *   唯一呼叫者 34CB:01A4（sub_419C4，+0x43），在視窗 +3（外框起始碼）非 0 時先畫外框（docs/re/138）。
 *   回傳值沒人用，描述檔 ret = none。
 *
 * 讀視窗結構（docs/internals/47-window-frame-tiles.md「結構位移」）：
 *   +1 顏色  +3 外框樣式起始碼  +4 槽位編號  +5 格 x  +6 格 y  +7 格寬  +8 格高（皆 byte）
 *   （+7、+8 含外框那兩格，docs/internals/47）。
 * 畫面格表：ds:D81E 起，40 欄 × 20 列、一格 6 bytes、一列 0xF0 bytes（docs/internals/31、47）。
 *
 * 動作（已證實：反組譯逐條對過；每個「寫一格」是同一組動作 put(cell, n)）：
 *   put(cell, n)：格[+0] = attr；格[+4] = (+1)（重讀）；格[+5] = 0；格 word[+2] = base + n（16 位元）。
 *       attr = (+4) | 0x90（髒 0x80 ＋ 走字模 0x10 ＋ 擁有者）、base = 0xFF00 + (+3)，兩者在進函式時算一次。
 *   「沒鎖住」= (格[+0] & 0x20) == 0 才 put（鎖住的格子整格跳過，docs/re/138）。
 *   格起點 = D81E + (+6) × 0xF0 + (+5) × 6（16 位元回繞），段 = 呼叫時的 DS（GIN3_DS）。
 *   順序（W = (+7) − 1、H = (+8) − 1，都是 16 位元運算，每次重讀；k、r 是 byte 計數器）：
 *     1. 左上 put(cell0, 0)；格 += 6。
 *     2. k = 1 起，W > k（無號 16 位元比較）：上緣 put(·, 1)；格 += 6；k++。
 *     3. 右上 put(·, 2)。
 *     4. r = 1 起，H > r：列起點 += 0xF0；左緣 put(列起點, 3)；格 = 列起點 + W × 6（16 位元）；右緣 put(·, 4)；r++。
 *     5. 列起點 += 0xF0；左下 put(·, 5)；格 += 6；k = 1 起，W > k：下緣 put(·, 6)；格 += 6；k++。
 *     6. 右下 put(·, 7)。
 *   寬為 1 時 W = 0：上緣迴圈不執行，右上落在左上右邊一格（視窗矩形外）；左緣與右緣同一格（W × 6 = 0，右緣後寫）；
 *   右下同理落在左下右邊一格。高為 1 時 H = 0：側邊迴圈不執行，步驟 5 的列起點 += 0xF0 照做，下緣那一列
 *   落在視窗矩形下方一列。原版如此，本版照做。
 *
 * 輸入域：全部 byte 值，位移 16 位元回繞。寬或高為 0 時 W 或 H = 0xFFFF，byte 計數器回繞，迴圈條件永遠成立，
 *   只有走訪時寫到視窗結構自己的 +7／+8（結構在 DS 段內，被某一格的字碼覆寫）才會停；沒走到就不返回，
 *   步數超過上限的案例被探針丟棄。本版逐步照做，同樣可能不返回（域外）。格位移恆為偶數
 *   （D81E、0xF0、6 都是偶數），所以格 word[+2] 的寫入不會跨過位移 0xFFFF；本檔仍用段內回繞的 word 寫入。
 *   視窗指標段指到 A0000 到 BFFFF 的案例被探針丟棄。
 *   視窗結構與畫面格表重疊時，put 的寫入會改到之後重讀的 +1／+7／+8；差異測試用 seeds 造這種別名並有對應突變。
 *
 * 語意：八個字碼 base + 0..7 = 左上、上、右上、左、右、左下、下、右下（docs/re/138、docs/internals/47）。
 *   與 34CB:0BA0 互為鏡像：本函式在「沒鎖住」時寫，0BA0 在「鎖住」時寫，其餘逐指令相同。
 */
#include "gin3_reg.h"

static uint8_t f34CB094A_wb(const Mem *m, uint16_t seg, uint16_t off, uint16_t n)
{
    return rd8(m, LIN(seg, (uint16_t)(off + n)));
}

static void f34CB094A_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* 沒鎖住才寫一格外框字碼 */
static void f34CB094A_put(Mem *m, uint16_t w_off, uint16_t w_seg, uint16_t cell, uint8_t attr, uint16_t code)
{
    const uint16_t ds = GIN3_DS;

    if ((rd8(m, LIN(ds, cell)) & 0x20u) == 0) {
        wr8(m, LIN(ds, cell), attr);
        wr8(m, LIN(ds, (uint16_t)(cell + 4u)), f34CB094A_wb(m, w_seg, w_off, 1));
        wr8(m, LIN(ds, (uint16_t)(cell + 5u)), 0);
        f34CB094A_wr16s(m, ds, (uint16_t)(cell + 2u), code);
    }
}

void f_34CB_094A(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    uint16_t cur = (uint16_t)(0xD81Eu + (uint16_t)(f34CB094A_wb(m, w_seg, w_off, 6) * 0xF0u)
                              + (uint16_t)(f34CB094A_wb(m, w_seg, w_off, 5) * 6u));
    const uint8_t attr = (uint8_t)(f34CB094A_wb(m, w_seg, w_off, 4) | 0x90u);
    const uint16_t base = (uint16_t)(f34CB094A_wb(m, w_seg, w_off, 3) + 0xFF00u);
    uint16_t cell = cur;
    uint8_t k, r;

    f34CB094A_put(m, w_off, w_seg, cell, attr, base);
    cell = (uint16_t)(cell + 6u);
    k = 1;
    while ((uint16_t)(f34CB094A_wb(m, w_seg, w_off, 7) - 1u) > k) {
        f34CB094A_put(m, w_off, w_seg, cell, attr, (uint16_t)(base + 1u));
        cell = (uint16_t)(cell + 6u);
        k++;
    }
    f34CB094A_put(m, w_off, w_seg, cell, attr, (uint16_t)(base + 2u));

    r = 1;
    while ((uint16_t)(f34CB094A_wb(m, w_seg, w_off, 8) - 1u) > r) {
        cur = (uint16_t)(cur + 0xF0u);
        cell = cur;
        f34CB094A_put(m, w_off, w_seg, cell, attr, (uint16_t)(base + 3u));
        cell = (uint16_t)(cell + (uint16_t)((uint16_t)(f34CB094A_wb(m, w_seg, w_off, 7) - 1u) * 6u));
        f34CB094A_put(m, w_off, w_seg, cell, attr, (uint16_t)(base + 4u));
        r++;
    }

    cur = (uint16_t)(cur + 0xF0u);
    cell = cur;
    f34CB094A_put(m, w_off, w_seg, cell, attr, (uint16_t)(base + 5u));
    cell = (uint16_t)(cell + 6u);
    k = 1;
    while ((uint16_t)(f34CB094A_wb(m, w_seg, w_off, 7) - 1u) > k) {
        f34CB094A_put(m, w_off, w_seg, cell, attr, (uint16_t)(base + 6u));
        cell = (uint16_t)(cell + 6u);
        k++;
    }
    f34CB094A_put(m, w_off, w_seg, cell, attr, (uint16_t)(base + 7u));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_34CB_094A(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("34CB_094A", 2, run)
