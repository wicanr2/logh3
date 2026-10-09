/*
 * GIN3PS.EXE 2583:052C  艦隊記錄的「短程平移」旗標：現在位置到目的地的距離小於 行程 ÷ 3 ＋ 4.0（16.16）就把 +0x48 的 bit4 設起來
 *
 * 原版（91 bytes，far 呼叫，一個 far 指標堆疊參數（位移在 [bp+0Ah]、段在 [bp+0Ch]）；序言 push bp／si／di、mov bp,sp，沒有 sub sp）。
 *   沒有有意義的回傳值（描述檔 ret ＝ none；AX 殘值不比較）。
 * 過程（已證實：反組譯逐指令；3310:0006、3CC9:0070、3CC9:00B9 由其各自的差異測試證實）：
 *   d ← 3310:0006(x1 ＝ w[+5C], y1 ＝ w[+5E], x2 ＝ w[+64], y2 ＝ w[+66])        ; 16.16 距離，AX:BX
 *   (q, r) ← 3CC9:0070(AX:BX ＝ w[+54]:w[+52]（高字:低字，有號 32 位元）, CX:DX ＝ 0:3)   ; 有號除法，商回 AX:BX
 *   thr ← q ＋ 0x40000                                                         ; add si,0；adc cx,4：高字加 4（16 位元回繞）
 *   3CC9:00B9(AX:BX ＝ d, CX:DX ＝ thr) 後 jge（SF ＝ OF）不成立，也就是 d < thr（有號 32 位元）→ b[+48] |= 0x10
 * 語意（強推論）：docs/internals/33-fleet-table.md 第 118 行（+0x48 bit4 ＝ 短程平移，唯一設定者 2583:052C：距離 < (+0x54:+0x52) ÷ 3 ＋ 4.0）、
 *   docs/re/294-short-move-strafe-and-arrival-facing.md 第 110 行起、docs/spec/027 第 38 行。+0x5C／+0x5E ＝ 現在位置，+0x64／+0x66 ＝ 目的地，
 *   +0x52 起的 long ＝ 行程（強推論，本版只照原版運算）。
 * 副作用：39B5:005C 的暫存（ds:8EDC..8EED，經 3310:0006）與 b[+48] 的 or。
 *
 * 輸入域：fleet 指標指到艦隊表內（描述檔 tables）；欄位全部 16 位元值。指標位移加欄位位移回繞 16 位元（本版用 uint16_t）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F052C_CF 0x0001u
#define F052C_SF 0x0080u
#define F052C_OF 0x0800u

void f_2583_052C(Mem *m, uint16_t off, uint16_t seg)
{
    uint32_t d, q, r, v, thr;
    uint16_t fl;

    d = f_3310_0006(m, rd16(m, LIN(seg, (uint16_t)(off + 0x5Cu))), rd16(m, LIN(seg, (uint16_t)(off + 0x5Eu))),
                    rd16(m, LIN(seg, (uint16_t)(off + 0x64u))), rd16(m, LIN(seg, (uint16_t)(off + 0x66u))));
    v = ((uint32_t)rd16(m, LIN(seg, (uint16_t)(off + 0x54u))) << 16) | rd16(m, LIN(seg, (uint16_t)(off + 0x52u)));
    gin3_ldivmod(v, 3u, &q, &r);
    (void)r;
    thr = q + 0x40000u;
    fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, (uint16_t)(thr >> 16), (uint16_t)thr);
    if (((fl & F052C_SF) != 0) != ((fl & F052C_OF) != 0)) /* jge 不成立（SF ≠ OF） */
        wr8(m, LIN(seg, (uint16_t)(off + 0x48u)), (uint8_t)(rd8(m, LIN(seg, (uint16_t)(off + 0x48u))) | 0x10u));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2583_052C(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2583_052C", 2, run)
