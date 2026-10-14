/*
 * GIN3PS.EXE 2264:0FC3  單擊算式的分母 K：目標艦種的橢圓防禦半徑 ρ ＝ ℓ ÷ (1 − e × cos Δ)（Δ ＝ 攻擊方位與目標艦隊朝向的角度差）取整數部分，乘目標艦隊的 +0x3D 再除 100
 *
 * 原版（364 bytes，far 呼叫，兩個遠指標堆疊參數 P ＝ 攻擊方艦艇記錄（位移 [bp+2Eh]、段 [bp+30h]）、S ＝ 目標艦艇記錄（位移 [bp+32h]、段 [bp+34h]）；序言 push bp／si／di、sub sp,24h、mov bp,sp），
 *   回傳 AX（描述檔 ret ＝ ax）。區域變數：[bp+0] ＝ 方位角（byte）、[bp+1] ＝ 角度差（byte）、[bp+2] ＝ 艦隊 +0x3D（byte）、[bp+4]／[bp+6] ＝ ℓ（低字／高字）、
 *   [bp+8]／[bp+0Ah] ＝ e、[bp+0Ch]／[bp+0Eh] ＝ r、[bp+10h]／[bp+12h] ＝ e × cos Δ、[bp+14h]／[bp+16h] ＝ 艦隊指標、[bp+18h]／[bp+1Ah] ＝ 武器列指標。
 * 過程（已證實：反組譯逐指令；2BE1:011D、2BE1:024A、3870:0643、3CC9:0009、3CC9:0020、3CC9:0070、39B5:000C、39B5:005C 各由其差異測試證實）：
 *   f   ← ds:(0xC074 ＋ byte S[+1] × 0x7A)                       ; 艦隊（16 位元）
 *   scale ← byte f[+0x3D]
 *   row ← ds:(0x4E3A ＋ byte S[+0] × 0xD0 ＋ byte S[+4] × 0x1A)   ; 16 位元；艦種表（強推論：陣營 × 208 ＋ 艦種 × 26，見 docs/internals/40-ship-kind-table.md）
 *   ang ← 2BE1:011D(a ＝ P[+0x46] − S[+0x34], b ＝ P[+0x48] − S[+0x36])
 *   Δ ← 2BE1:024A(ang, word f[+0x4A])                              ; 兩個 byte 角度的無號絕對差（不折回）
 *   ℓ ← 3CC9:0020( ((row[+0x18] × row[+0x18]) 的低字) << 16 , 0:row[+0x16] )   ; 無號除，除數 0 商 0xFFFFFFFF
 *   e ← 3870:0643(row[+0x16], row[+0x18])
 *   c ← 查表 word ds:(0x573A ＋ ((Δ ＋ 0x40) 的低位元組) × 2)，符號延伸成 32 位元，左移 6 位
 *   p ← 39B5:000C(e, c)；den ← 0x10000 − p（32 位元）
 *   r ← 39B5:005C(ℓ, den)                                          ; (ℓ << 16) ÷ den（有號 16.16）
 *   v ← 3CC9:0070( (r 的高字 × scale)（3CC9:0009，兩個運算元都零延伸）, 0:100 )
 *   回 v 的低字
 * 語意（強推論）：docs/re/94-attack-coefficient-is-a-conic.md（「單擊算式的分母 K 是一條圓錐曲線」，K 由 2264:0FC3 算、呼叫者 2350:0023 取 攻擊力 ÷ K，兩個參數是（攻擊方，目標），查艦種表與 +0x3D 都用目標）、docs/re/97-skirmish-path-never-runs.md、docs/spec/028-fortress-gun-fire.md 第 92 行；
 *   P、S 都是艦艇池記錄（S：+0 陣營、+1 艦隊編號、+4 艦種、+0x34／+0x36 座標；P：+0x46／+0x48 當攻擊方的座標，欄位語意只由用法推得），f[+0x4A] 是艦隊朝向、f[+0x3D] 是艦隊能力百分比；
 *   row[+0x16]／row[+0x18] 是橢圓的長半軸 a 與短半軸 b（與 2264:047C 的參數 a、b 一致，強推論）。
 * 副作用：被呼叫者在 ds:8EDC..8EED 的暫存；沒有其他寫入。
 *
 * 機器模型：C 版沒有堆疊；被呼叫者的堆疊參數與區域變數都在探針的堆疊帶內，比對時排除。
 * 輸入域：P、S 取艦艇池（或 seeds 給的位置）；艦隊、武器列、cos 表取自記憶體。word 讀取在位移 0xFFFF 時 dosgolem 的高位元組取同一段的位移 0（rd16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_024A(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3870_0643(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0020(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0070(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint32_t f_39B5_000C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);

static uint16_t f2264_0FC3_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_2264_0FC3(Mem *m, uint16_t p_off, uint16_t p_seg, uint16_t s_off, uint16_t s_seg)
{
    uint16_t f = (uint16_t)(0xC074u + (uint16_t)(rd8(m, LIN(s_seg, (uint16_t)(s_off + 1u))) * 0x7Au));
    uint8_t scale = rd8(m, DS((uint16_t)(f + 0x3Du)));
    uint16_t row = (uint16_t)(0x4E3Au + (uint16_t)(rd8(m, LIN(s_seg, s_off)) * 0xD0u)
                              + (uint16_t)(rd8(m, LIN(s_seg, (uint16_t)(s_off + 4u))) * 0x1Au));
    uint16_t b = (uint16_t)(f2264_0FC3_rd16s(m, p_seg, (uint16_t)(p_off + 0x48u)) - f2264_0FC3_rd16s(m, s_seg, (uint16_t)(s_off + 0x36u)));
    uint16_t a = (uint16_t)(f2264_0FC3_rd16s(m, p_seg, (uint16_t)(p_off + 0x46u)) - f2264_0FC3_rd16s(m, s_seg, (uint16_t)(s_off + 0x34u)));
    uint8_t ang = f_2BE1_011D(m, a, b);
    uint8_t delta = (uint8_t)f_2BE1_024A(m, ang, f2264_0FC3_rd16s(m, GIN3_DS, (uint16_t)(f + 0x4Au)));
    uint16_t r18 = f2264_0FC3_rd16s(m, GIN3_DS, (uint16_t)(row + 0x18u));
    uint32_t sq = f_3CC9_0009(0, r18, 0, r18);
    uint16_t r16 = f2264_0FC3_rd16s(m, GIN3_DS, (uint16_t)(row + 0x16u));
    uint32_t ell = (uint32_t)(f_3CC9_0020((uint16_t)sq, 0, 0, r16) >> 32);
    uint32_t e = f_3870_0643(m, f2264_0FC3_rd16s(m, GIN3_DS, (uint16_t)(row + 0x16u)), f2264_0FC3_rd16s(m, GIN3_DS, (uint16_t)(row + 0x18u)));
    uint16_t w = f2264_0FC3_rd16s(m, GIN3_DS, (uint16_t)(0x573Au + (uint16_t)((((uint16_t)delta + 0x40u) & 0xFFu) * 2u)));
    uint32_t c = ((uint32_t)(int32_t)(int16_t)w) << 6;
    uint32_t p = f_39B5_000C(m, (uint16_t)e, (uint16_t)(e >> 16), (uint16_t)c, (uint16_t)(c >> 16));
    uint32_t den = (uint32_t)(0x10000u - p);
    uint32_t r = f_39B5_005C(m, (uint16_t)ell, (uint16_t)(ell >> 16), (uint16_t)den, (uint16_t)(den >> 16));
    uint32_t prod = f_3CC9_0009(0, (uint16_t)(r >> 16), 0, scale);
    uint64_t q = f_3CC9_0070((uint16_t)(prod >> 16), (uint16_t)prod, 0, 100);

    return (uint16_t)(q >> 32);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2264_0FC3(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("2264_0FC3", 4, run)
