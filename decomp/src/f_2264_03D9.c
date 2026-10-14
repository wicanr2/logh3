/*
 * GIN3PS.EXE 2264:03D9  要塞砲的橢圓判定入口：距離 d ＝ 3310:0006(P1 位置, P2 位置)、方位差 Δ ＝ 2BE1:024A(2BE1:011D(P2 − P1), P1[+0x50])，
 *   回 2264:047C(d, Δ, ds:A772, ds:A774)
 *
 * 原版（163 bytes，far 呼叫，兩個 far 指標參數；序言 push bp／si／di、sub sp,12h、mov bp,sp，所以 P1 的位移在 [bp+1Ch]、段在 [bp+1Eh]，P2 的位移在 [bp+20h]、段在 [bp+22h]），
 *   回傳 AX ＝ 2264:047C 的回傳（描述檔 ret ＝ ax）。
 * 過程（已證實：反組譯逐指令；3310:0006、2BE1:011D、2BE1:024A、2264:047C 由其各自的差異測試證實）：
 *   讀 P2[+0x36]、P2[+0x34]、P1[+0x48]、P1[+0x46]（四個 word，依序 push）
 *   d ← 3310:0006(x1 ＝ P1[+0x46], y1 ＝ P1[+0x48], x2 ＝ P2[+0x34], y2 ＝ P2[+0x36])（AX:BX ＝ 高字:低字，存到 [bp+4]、[bp+2]）
 *   push P1[+0x50]（word）
 *   dir ← 2BE1:011D(P2[+0x34] − P1[+0x46], P2[+0x36] − P1[+0x48])        ; 呼叫後重新讀欄位，差是 16 位元
 *   Δ ← 2BE1:024A(dir, P1[+0x50])                                          ; AX（高位元組 0）；dir 的 AH 是殘值，024A 只取低位元組
 *   回 2264:047C(d_lo, d_hi, Δ, ds:A772, ds:A774)                           ; push 順序：ds:A774、ds:A772、Δ、d_hi、d_lo
 * 語意（強推論）：docs/re/145-fortress-gun-aim-and-fire.md、docs/re/147-fortress-gun-damage.md（橢圓判定）。P1 ＝ 行星（+0x46／+0x48 位置，+0x50 瞄準方位），
 *   P2 ＝ 艦艇（+0x34／+0x36 位置，docs/internals/38-battle-units.md）；ds:A772／A774 ＝ 橢圓長、短半軸（假說，docs/re/145 第 83 行：長半軸、短半軸）。名稱不是證據。
 * 副作用：全部來自 3310:0006 與 2264:047C 內 39B5:005C 的暫存（ds:8EDC..8EED）；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊（區域變數只是暫存）；基底狀態 fcmd-base.state，被呼叫的 39B5 C 版暫存在 SS:1F82..1F89，在探針堆疊帶內。
 * 輸入域：P1、P2 取行星表與艦艇池內的指標；欄位與 ds:A772／A774 取記憶體與突變；word 欄位在位移 0xFFFF 時高位元組取同一段的位移 0（本表不會遇到）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_024A(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2264_047C(Mem *m, uint16_t d_lo, uint16_t d_hi, uint16_t delta, uint16_t a, uint16_t b);

static uint16_t w03D9_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_2264_03D9(Mem *m, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off, uint16_t p2_seg)
{
    uint16_t x1 = w03D9_rd16s(m, p1_seg, (uint16_t)(p1_off + 0x46u));
    uint16_t y1 = w03D9_rd16s(m, p1_seg, (uint16_t)(p1_off + 0x48u));
    uint16_t x2 = w03D9_rd16s(m, p2_seg, (uint16_t)(p2_off + 0x34u));
    uint16_t y2 = w03D9_rd16s(m, p2_seg, (uint16_t)(p2_off + 0x36u));
    uint32_t d = f_3310_0006(m, x1, y1, x2, y2);
    uint16_t face = w03D9_rd16s(m, p1_seg, (uint16_t)(p1_off + 0x50u));
    uint16_t ang, delta;

    y2 = w03D9_rd16s(m, p2_seg, (uint16_t)(p2_off + 0x36u));
    y1 = w03D9_rd16s(m, p1_seg, (uint16_t)(p1_off + 0x48u));
    x2 = w03D9_rd16s(m, p2_seg, (uint16_t)(p2_off + 0x34u));
    x1 = w03D9_rd16s(m, p1_seg, (uint16_t)(p1_off + 0x46u));
    ang = f_2BE1_011D(m, (uint16_t)(x2 - x1), (uint16_t)(y2 - y1));
    delta = f_2BE1_024A(m, ang, face);
    return f_2264_047C(m, (uint16_t)d, (uint16_t)(d >> 16), delta, rd16(m, DS(0xA772u)), rd16(m, DS(0xA774u)));
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2264_03D9(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("2264_03D9", 4, run)
