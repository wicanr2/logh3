/*
 * GIN3PS.EXE 203A:0214  艦艇受擊的橢圓裝甲值：以來襲方位對目標艦種的橢圓半徑 ρ（16.16）的整數部分，乘目標艦隊的某項能力再除以 100
 *
 * 原版（412 bytes，far 呼叫，兩個 far 指標堆疊參數；序言 push bp／si／di、sub sp,24h、mov bp,sp，所以 P1 位移在 [bp+2Eh]、段在 [bp+30h]，
 *   P2 位移在 [bp+32h]、段在 [bp+34h]）。回傳 AX。區域變數：[bp+0] ＝ 方位 byte、[bp+1] ＝ 相對方位 byte、[bp+2] ＝ stat（byte，見下）、
 *   [bp+4..7] ＝ r、[bp+8..0Bh] ＝ e、[bp+0Ch..0Fh] ＝ ρ。
 *   P1、P2 都指向艦艇池記錄（78 bytes；+0x34／+0x36 ＝ 座標，同 203A:03B0 的描述檔），P1 是目標、P2 是來襲方（強推論，見語意）。
 * 過程（已證實：反組譯逐指令；2BE1:011D／024A、3870:0643、39B5:000C／005C、3CC9:0009／0020／0070 由其各自的差異測試證實）：
 *   F ← ds:(0xC074 ＋ P1[+1] × 0x7A)                               ; 艦隊記錄（P1[+1] ＝ 艦隊索引，強推論）
 *   stat ← F[+0x3D]（P2[+0x2E] ＝ 0）、F[+0x3E]（＝ 1）、F[+0x3F]（＝ 2）；其他值不寫 [bp+2]，stat ＝ 堆疊上原有的殘值（見機器模型）
 *   U ← ds:(0x4E3A ＋ P1[0] × 0xD0 ＋ P1[+4] × 0x1A)               ; 艦種能力表的一筆（16 位元回繞）；v ＝ U[+0x16]、w ＝ U[+0x18]
 *   ang ← 2BE1:011D(P2[+0x34] − P1[+0x34], P2[+0x36] − P1[+0x36])    ; 16 位元相減，來襲方位 byte
 *   rel ← 2BE1:024A(ang, F[+0x4A])                                  ; 無號絕對差（byte）
 *   r ← 3CC9:0020(((w × w) 的低字) << 16, 0:v)                       ; 無號，w 零延伸；v ＝ 0 時商 0xFFFFFFFF
 *   e ← 3870:0643(v, w)
 *   t ← sext32(word ds:(0x573A ＋ 2 × ((rel ＋ 0x40) 的低位元組))) << 6
 *   p ← 39B5:000C(e, t)；den ← 0x10000 − p；ρ ← 39B5:005C(r, den)
 *   回 3CC9:0070(0:ρ高字 × stat（3CC9:0009），0:100) 的商低字
 * 語意（強推論）：docs/internals/38-battle-units.md 第 376 行「203A:0214：目標艦種裝甲 +16／+18、來襲方位與目標 FLEET+4A 取正差，半徑取整後乘目標相應能力除 100」
 *   與本檔一致；極座標橢圓 ρ ＝ (b²/a) ÷ (1 − e cos θ)，與 303F:02DE 同形。呼叫者 1DD6:0C54、1F6F:0714。
 * 副作用：全部來自 3870:0643 與本函式直接呼叫的 39B5:005C（ds:8EDC..8EED 共 18 bytes）；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊。被呼叫者的堆疊參數放在探針的堆疊帶內，比對時排除。stat 在 P2[+0x2E] > 2 時是未初始化的區域變數：
 *   基底狀態 fcmd-base.state（SS ＝ 0x63EB，進入點 SP ＝ 1F8A）下，探針推 4 個參數字 8 bytes、返回位址 4、push bp／si／di 6、sub sp,24h，
 *   所以 bp ＝ 1F54，[bp+2] 在 SS:1F56。該位址不被本函式寫（[bp+0]、[bp+1] 是 byte 存取，被呼叫者的框架在更低位址），所以它是基底傾印
 *   （加上描述檔 pokes）的內容，C 版從那裡讀。堆疊帶的位元組不是探針突變的對象，只能由 seeds 的 pokes 改。
 * 輸入域：P1、P2 指到艦艇池記錄（描述檔 table ship）；其餘欄位由探針突變讀取集。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_024A(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3870_0643(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0020(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint32_t f_39B5_000C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);

#define F0214_SS 0x63EBu    /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F0214_STAT 0x1F56u  /* 進入點 SP 1F8A − 8 − 4 − 6 − 0x24 ＝ bp(1F54)，stat 在 bp+2 */

uint16_t f_203A_0214(Mem *m, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off, uint16_t p2_seg)
{
    uint16_t f, u, v, w, rel16, idx, dx_w, dy_w;
    uint8_t stat, sel, ang, rel;
    uint32_t b2, r, e, t, p, den, rho, prod, q, rem;

    f = (uint16_t)(0xC074u + (uint16_t)(rd8(m, LIN(p1_seg, p1_off + 1u)) * 0x7Au));
    sel = rd8(m, LIN(p2_seg, (uint16_t)(p2_off + 0x2Eu)));
    if (sel == 0)
        stat = rd8(m, DS((uint16_t)(f + 0x3Du)));
    else if (sel == 1)
        stat = rd8(m, DS((uint16_t)(f + 0x3Eu)));
    else if (sel == 2)
        stat = rd8(m, DS((uint16_t)(f + 0x3Fu)));
    else
        stat = rd8(m, LIN(F0214_SS, F0214_STAT)); /* 未初始化的區域變數 */

    u = (uint16_t)(0x4E3Au + (uint16_t)(rd8(m, LIN(p1_seg, p1_off)) * 0xD0u) +
                   (uint16_t)(rd8(m, LIN(p1_seg, (uint16_t)(p1_off + 4u))) * 0x1Au));
    dy_w = (uint16_t)(rd16(m, LIN(p2_seg, (uint16_t)(p2_off + 0x36u))) - rd16(m, LIN(p1_seg, (uint16_t)(p1_off + 0x36u))));
    dx_w = (uint16_t)(rd16(m, LIN(p2_seg, (uint16_t)(p2_off + 0x34u))) - rd16(m, LIN(p1_seg, (uint16_t)(p1_off + 0x34u))));
    ang = f_2BE1_011D(m, dx_w, dy_w);
    rel16 = f_2BE1_024A(m, ang, rd16(m, DS((uint16_t)(f + 0x4Au))));
    rel = (uint8_t)rel16;

    w = rd16(m, DS((uint16_t)(u + 0x18u)));
    v = rd16(m, DS((uint16_t)(u + 0x16u)));
    b2 = f_3CC9_0009(0, w, 0, w);
    r = (uint32_t)(f_3CC9_0020((uint16_t)b2, 0, 0, v) >> 32); /* AX:BX ＝ (b2 低字):0 */
    e = f_3870_0643(m, v, w);
    idx = (uint16_t)((uint8_t)(rel + 0x40u));
    t = (uint32_t)(int32_t)(int16_t)rd16(m, DS((uint16_t)(0x573Au + 2u * idx))) << 6;
    p = f_39B5_000C(m, (uint16_t)e, (uint16_t)(e >> 16), (uint16_t)t, (uint16_t)(t >> 16));
    den = 0x10000u - p;
    rho = f_39B5_005C(m, (uint16_t)r, (uint16_t)(r >> 16), (uint16_t)den, (uint16_t)(den >> 16));
    prod = f_3CC9_0009(0, (uint16_t)(rho >> 16), 0, stat);
    gin3_ldivmod(prod, 100u, &q, &rem);
    return (uint16_t)q;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_203A_0214(m, a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("203A_0214", 4, run)
