/*
 * GIN3PS.EXE 303F:0221  電腦那條路的要塞砲橢圓判定入口：與 2264:03D9 同骨架，目標位置改取艦隊表項 ds:(C074 + A[+1] × 0x7A) 的 +0x74／+0x76
 *
 * 原版（189 bytes，far 呼叫，兩個 far 指標參數；序言 push bp／si／di、sub sp,16h、mov bp,sp，所以 B 的位移在 [bp+20h]、段在 [bp+22h]，A 的位移在 [bp+24h]、段在 [bp+26h]），
 *   回傳 AX ＝ 303F:02DE 的回傳（描述檔 ret ＝ ax）。
 * 過程（已證實：反組譯逐指令；3310:0006、2BE1:011D、2BE1:024A、303F:02DE 由其各自的差異測試證實）：
 *   P ← DS:(0xC074 ＋ (byte A[+1]) × 0x7A)                    ; 艦隊表項（16 位元位移：索引 >= 135 時 0xC074 + 乘積超過 0xFFFF 而回繞）
 *   讀 P[+0x76]、P[+0x74]、B[+0x48]、B[+0x46]（四個 word，依序 push）
 *   d ← 3310:0006(x1 ＝ B[+0x46], y1 ＝ B[+0x48], x2 ＝ P[+0x74], y2 ＝ P[+0x76])（AX:BX ＝ 高字:低字，存到 [bp+4]、[bp+2]）
 *   push B[+0x50]（word）
 *   dir ← 2BE1:011D(P[+0x74] − B[+0x46], P[+0x76] − B[+0x48])   ; 呼叫後重新讀欄位，差是 16 位元
 *   Δ ← 2BE1:024A(dir, B[+0x50])
 *   回 303F:02DE(d_lo, d_hi, Δ, ds:D616, ds:D618)                ; push 順序：ds:D618、ds:D616、Δ、d_hi、d_lo
 * 語意（強推論）：docs/re/145-fortress-gun-aim-and-fire.md 第 216 到 218 行（電腦那條路的判定在 303F:0221，骨架同 2264:03D9，量的是艦隊座標 ds:C074 表的 +0x74／+0x76），
 *   與本檔一致（該段另說「只查敵方陣營、不檢查艘數」，那是呼叫者 seg071 的事，不在本函式內）。B ＝ 行星（+0x46／+0x48 位置，+0x50 瞄準方位）；A ＝ 艦艇池項（+1 ＝ 艦隊索引，假說）。
 * 副作用：全部來自 3310:0006 與 303F:02DE 內 39B5:005C 的暫存（ds:8EDC..8EED）；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊（區域變數只是暫存）；基底狀態 fcmd-base.state，被呼叫的 39B5 C 版暫存在 SS:1F82..1F89，在探針堆疊帶內。
 * 輸入域：B 取行星表、A 取艦艇池內的指標；艦隊索引取全部 256 個位元組值（索引 >= 135 時位移回繞，C 版以 uint16_t 運算）；欄位與 ds:D616／D618 取記憶體與突變。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_024A(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_303F_02DE(Mem *m, uint16_t l_lo, uint16_t l_hi, uint16_t theta, uint16_t a, uint16_t b);

static uint16_t w0221_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_303F_0221(Mem *m, uint16_t b_off, uint16_t b_seg, uint16_t a_off, uint16_t a_seg)
{
    uint8_t k = rd8(m, LIN(a_seg, (uint16_t)(a_off + 1u)));
    uint16_t p = (uint16_t)(0xC074u + k * 0x7Au);
    uint16_t x1 = w0221_rd16s(m, b_seg, (uint16_t)(b_off + 0x46u));
    uint16_t y1 = w0221_rd16s(m, b_seg, (uint16_t)(b_off + 0x48u));
    uint16_t x2 = w0221_rd16s(m, GIN3_DS, (uint16_t)(p + 0x74u));
    uint16_t y2 = w0221_rd16s(m, GIN3_DS, (uint16_t)(p + 0x76u));
    uint32_t d = f_3310_0006(m, x1, y1, x2, y2);
    uint16_t face = w0221_rd16s(m, b_seg, (uint16_t)(b_off + 0x50u));
    uint16_t ang, delta;

    y2 = w0221_rd16s(m, GIN3_DS, (uint16_t)(p + 0x76u));
    y1 = w0221_rd16s(m, b_seg, (uint16_t)(b_off + 0x48u));
    x2 = w0221_rd16s(m, GIN3_DS, (uint16_t)(p + 0x74u));
    x1 = w0221_rd16s(m, b_seg, (uint16_t)(b_off + 0x46u));
    ang = f_2BE1_011D(m, (uint16_t)(x2 - x1), (uint16_t)(y2 - y1));
    delta = f_2BE1_024A(m, ang, face);
    return f_303F_02DE(m, (uint16_t)d, (uint16_t)(d >> 16), delta, rd16(m, DS(0xD616u)), rd16(m, DS(0xD618u)));
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_303F_0221(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("303F_0221", 4, run)
