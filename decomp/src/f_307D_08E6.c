/*
 * GIN3PS.EXE 307D:08E6  抽籤決定艦隊的目標：機率式地「複製另一支艦隊的既有位置／方位」或交給 307D:097E「偏移一小段」
 *
 * 原版：far 呼叫，五個 far 指標參數（序言 push bp／si／di、sub sp,2、mov bp,sp，所以第一個參數在 [bp+0Ch]），沒有有意義的回傳值（描述檔 ret ＝ none）：
 *   args[0..1] ＝ Fa（艦隊記錄，讀 [+0x44]）  args[2..3] ＝ Fb（艦隊記錄，讀 [+0x74]、[+0x76]、[+0x78]，並原樣傳給 097E）
 *   args[4..5] ＝ P1（word 輸出）  args[6..7] ＝ P2（word 輸出）  args[8..9] ＝ P3（byte 輸出）
 * 過程（已證實：反組譯逐指令；3CD5:000C（rand）、307D:097E 由其各自的差異測試證實）：
 *   t ← 0x28 ＋ (rand() mod 50)（有號 cwd／idiv，rand 回 0..0x7FFF 所以餘數 ≥ 0）
 *   rand() 四次（結果丟棄，只為了推進亂數種子）
 *   t ← t ＋ (rand() mod 50)                                       ; 共抽 6 次 rand
 *   byte Fa[+0x44]（零延伸）< t（無號，jb）→ 307D:097E(Fb, P1, P2, P3)      ; 偏移一小段（見 f_307D_097E.c）
 *   否則：*P1（word）← Fb[+0x74]；*P2（word）← Fb[+0x76]；*P3（byte）← Fb[+0x78]            ; 每一步寫完才讀下一個欄位
 * 語意（強推論）：t 的範圍是 40..138，Fa[+0x44] 是 0..255 的某項能力值，值愈小愈傾向「偏移」（097E）；值夠大就沿用 Fb 既有的三個欄位
 *   （x ＝ [+0x74]、y ＝ [+0x76]、方位 ＝ [+0x78]）。呼叫者 307D:0521（unknown）。docs 沒有任何記載（grep 307D:08E6 只有台帳呼叫關係）。
 *   欄位名稱不是證據，位移與比較才是。
 * 副作用：ds:94DC（亂數種子，六次 rand，加上 097E 內再抽的次數）；複製分支寫 P1、P2、P3；097E 分支的寫入見 f_307D_097E.c。
 *   線性 0xF0000 以上（ROM）的寫入在 dosgolem 被安靜丟棄，本檔的複製分支按 byte 模型化；word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0。
 * 輸入域：Fa、Fb 取艦隊表；P1、P2、P3 取暫存區與 seeds。097E 分支繼承 f_307D_097E.c 的限制（其呼叫的 2BE1:026F 用線性位址寫 word，
 *   所以 P1、P2 的位移不涵蓋 0xFFFF）。讀到 A0000 到 BFFFF 的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD5_000C(Mem *m);
extern void f_307D_097E(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off,
                        uint16_t p2_seg, uint16_t p3_off, uint16_t p3_seg);

static uint16_t w08E6_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void w08E6_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w08E6_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w08E6_put8(m, LIN(seg, off), (uint8_t)v);
    w08E6_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_307D_08E6(Mem *m, uint16_t fa_off, uint16_t fa_seg, uint16_t fb_off, uint16_t fb_seg, uint16_t p1_off,
                 uint16_t p1_seg, uint16_t p2_off, uint16_t p2_seg, uint16_t p3_off, uint16_t p3_seg)
{
    uint16_t t = (uint16_t)(0x28u + (uint16_t)((int16_t)f_3CD5_000C(m) % 50));

    f_3CD5_000C(m);
    f_3CD5_000C(m);
    f_3CD5_000C(m);
    f_3CD5_000C(m);
    t = (uint16_t)(t + (uint16_t)((int16_t)f_3CD5_000C(m) % 50));
    if ((uint16_t)rd8(m, LIN(fa_seg, (uint16_t)(fa_off + 0x44u))) < t) {
        f_307D_097E(m, fb_off, fb_seg, p1_off, p1_seg, p2_off, p2_seg, p3_off, p3_seg);
    } else {
        w08E6_wr16s(m, p1_seg, p1_off, w08E6_rd16s(m, fb_seg, (uint16_t)(fb_off + 0x74u)));
        w08E6_wr16s(m, p2_seg, p2_off, w08E6_rd16s(m, fb_seg, (uint16_t)(fb_off + 0x76u)));
        w08E6_put8(m, LIN(p3_seg, p3_off), rd8(m, LIN(fb_seg, (uint16_t)(fb_off + 0x78u))));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_307D_08E6(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9]);
    return 0;
}
GIN3_REGISTER("307D_08E6", 10, run)
