/*
 * 34CB:000D  鎖住視窗範圍內「屬於本視窗」的畫面格（far，一個 far 指標參數 = 視窗結構）
 *
 * 原版：far 呼叫，參數是視窗結構的 far 指標（位移在 [bp+14h]、段在 [bp+16h]；序言 push bp／si／di、sub sp,0Ah）。
 *   唯一呼叫者 341B:0506（sub_41226 內，call 前推入視窗指標）。回傳值沒人用，描述檔 ret = none。
 *
 * 讀視窗結構（docs/internals/47-window-frame-tiles.md「結構位移」；名稱是本檔取的）：
 *   +4  槽位編號（畫面格表 +0 的低四位記擁有者）
 *   +5  格 x（欄）  +6  格 y（列）  +7  格寬  +8  格高（皆 byte）
 * 畫面格表：ds:D81E 起，40 欄 × 20 列、一格 6 bytes、一列 0xF0 bytes（docs/internals/31、47）。
 *
 * 動作（已證實：反組譯逐條對過）：
 *   起點 = D81E + (+6) × 0xF0 + (+5) × 6，16 位元回繞，段 = 呼叫時的 DS（GIN3_DS）。
 *   外層 i = 0 起，i < (+8)（無號 byte 比較；每一輪都重讀 +8）：
 *       內層 j = 0 起，j < (+7)（每一輪都重讀 +7）：
 *           若 (格[+0] & 0x0F) == (+4)（byte 比較；(+4) > 15 永遠不成立）：格[+0] |= 0x20（鎖住位元）
 *           格位移 += 6
 *       列起點 += 0xF0
 *   視窗欄位在每一輪迴圈都重讀（les bx,[bp+14h] 之後 mov cl, es:[bx+N]），不是只在迴圈前讀一次。
 *   視窗結構與畫面格表重疊時（結構放在畫面格表內），寫入會改到之後重讀的 +4／+7／+8；
 *   差異測試用 seeds 造這種別名並有對應突變。
 *
 * 語意（強推論）：0x20 在 docs/re/138 與 docs/internals/47 記為「鎖住（設了就整格跳過）」，
 *   34CB:094A 等繪製外框／內容的函式遇到鎖住的格子就不覆寫。本函式把視窗範圍內屬於本視窗的格子鎖起來。
 *   呼叫者 341B:0506 的用途（疊在其他視窗上時保護下層）未在本檔驗證：未知。
 * 輸入域：視窗結構任意 byte 值；位移全部 16 位元回繞（格位移、欄位位移）。
 *   寬、高最大 255，最多 255 × 255 格；超過探針步數上限（200000）的案例被丟棄，不在比對範圍。
 *   視窗指標的段若指到 A0000 到 BFFFF（視訊記憶體）的案例被探針丟棄。
 */
#include "gin3_reg.h"

/* 視窗結構的 byte 欄位（位移 n 以 16 位元回繞） */
static uint8_t f34CB000D_wb(const Mem *m, uint16_t seg, uint16_t off, uint16_t n)
{
    return rd8(m, LIN(seg, (uint16_t)(off + n)));
}

void f_34CB_000D(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    const uint16_t ds = GIN3_DS;
    uint16_t cur = (uint16_t)(0xD81Eu + (uint16_t)(f34CB000D_wb(m, w_seg, w_off, 6) * 0xF0u)
                              + (uint16_t)(f34CB000D_wb(m, w_seg, w_off, 5) * 6u));
    uint8_t i = 0;

    while (i < f34CB000D_wb(m, w_seg, w_off, 8)) {
        uint16_t cell = cur;
        uint8_t j = 0;

        cur = (uint16_t)(cur + 0xF0u);
        while (j < f34CB000D_wb(m, w_seg, w_off, 7)) {
            if ((rd8(m, LIN(ds, cell)) & 0x0Fu) == f34CB000D_wb(m, w_seg, w_off, 4))
                wr8(m, LIN(ds, cell), (uint8_t)(rd8(m, LIN(ds, cell)) | 0x20u));
            cell = (uint16_t)(cell + 6u);
            j++;
        }
        i++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_34CB_000D(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("34CB_000D", 2, run)
