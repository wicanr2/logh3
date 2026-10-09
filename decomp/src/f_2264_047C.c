/*
 * GIN3PS.EXE 2264:047C  橢圓判定：距離 d（16.16）是否 <= 以焦點為極點的圓錐曲線在方位 Δ 的半徑 r ＝ ℓ ÷ (1 − e × cos Δ)
 *
 * 原版（179 bytes，far 呼叫，五個 word 堆疊參數；序言 push bp／si／di、sub sp,0Ch、mov bp,sp，所以第一個參數在 [bp+16h]）：
 *   d ＝ dword（低字 [bp+16h]、高字 [bp+18h]）、Δ ＝ word [bp+1Ah]（只用低位元組）、a ＝ word [bp+1Ch]、b ＝ word [bp+1Eh]。
 *   回傳 AX ＝ 1（d <= r，有號 32 位元）或 0；描述檔 ret ＝ ax。
 * 過程（已證實：反組譯逐指令；3870:0643、3CC9:0009、3CC9:0070、39B5:000C、39B5:005C、3CC9:00B9 各由其差異測試證實）：
 *   e ← 3870:0643(a, b)                                     ; 16.16 離心率（AX:BX）
 *   ℓ ← 3CC9:0070( ((b × b) 的低字) << 16 , 0:a )            ; 3CC9:0009 取 b×b（兩個運算元的高字都 0），只留低 16 位元放到高字；有號 32 位元除，除數是 a 的零延伸
 *   c ← 查表 word ds:(0x573A + (((Δ 低位元組) + 0x40) 的低位元組) × 2)，cwd 符號延伸成 32 位元，左移 6 位   ; cos Δ 的 16.16
 *   p ← 39B5:000C(e, c)                                     ; e × cos Δ（16.16）
 *   den ← 0x10000 − p                                       ; 32 位元 sub／sbb（1.0 − e·cos Δ）
 *   r ← 39B5:005C(ℓ, den)                                   ; (ℓ << 16) ÷ den，有號，除數 0 回 0
 *   3CC9:00B9(d, r) 之後 jle：ZF ＝ 1 或 SF ≠ OF → 回 1，否則回 0（有號 32 位元 d <= r）
 * 語意：docs/re/147-fortress-gun-damage.md 第 24 到 36 行（橢圓判定，confirmed）、docs/re/145-fortress-gun-aim-and-fire.md 第 198 到 215 行、
 *   docs/internals/50-fortress-gun.md 第 80 到 86 行，與本檔逐項相符（ℓ ＝ b² × 65536 ÷ a；本檔補充：b² 只留低 16 位元再除，b >= 256 時與數學式不同，
 *   a ＝ 0 時 3CC9:0070 的除數 0 路徑，Δ 只取低位元組，cos 表索引在位元組內回繞）。ds:573A 表項「值 × 1024」（docs/re/147 第 33 行）。
 *   呼叫者 2264:03D9（d ＝ 3310:0006 的距離，Δ ＝ 2BE1:024A 的角度差，a、b 是行星砲的長短半軸）。
 * 副作用：被呼叫者在 ds:8EDC..8EED 的暫存（39B5:005C）；沒有其他寫入。
 *
 * 機器模型：C 版沒有堆疊；被呼叫者的堆疊參數與區域變數都在探針的堆疊帶內，比對時排除。
 * 輸入域：d、Δ、a、b 全部 16 位元值；cos 表 ds:573A..ds:593A（256 個 word）取基底與突變。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途見上述 docs（強推論到 confirmed）。
 */
#include "gin3_reg.h"

extern uint32_t f_3870_0643(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0070(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint32_t f_39B5_000C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F2264_ZF 0x0040u
#define F2264_SF 0x0080u
#define F2264_OF 0x0800u

uint16_t f_2264_047C(Mem *m, uint16_t d_lo, uint16_t d_hi, uint16_t delta, uint16_t a, uint16_t b)
{
    uint32_t e = f_3870_0643(m, a, b);
    uint32_t bb = f_3CC9_0009(0, b, 0, b);
    uint64_t q = f_3CC9_0070((uint16_t)bb, 0, 0, a); /* mov ax,bx；xor bx,bx：AX:BX ＝ (b×b 的低字) << 16 */
    uint32_t ell = (uint32_t)(q >> 32);
    uint16_t idx = (uint16_t)(((uint16_t)(delta & 0xFFu) + 0x40u) & 0xFFu);
    uint16_t w = rd16(m, DS((uint16_t)(0x573Au + idx * 2u)));
    uint32_t c = ((uint32_t)(int32_t)(int16_t)w) << 6; /* cwd；shl／rcl 六次 */
    uint32_t p = f_39B5_000C(m, (uint16_t)e, (uint16_t)(e >> 16), (uint16_t)c, (uint16_t)(c >> 16));
    uint32_t den = (uint32_t)(0x10000u - p);
    uint32_t r = f_39B5_005C(m, (uint16_t)ell, (uint16_t)(ell >> 16), (uint16_t)den, (uint16_t)(den >> 16));
    uint16_t f = f_3CC9_00B9(d_hi, d_lo, (uint16_t)(r >> 16), (uint16_t)r);

    if ((f & F2264_ZF) || (((f & F2264_SF) != 0) != ((f & F2264_OF) != 0)))
        return 1; /* jle */
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_2264_047C(m, a[0], a[1], a[2], a[3], a[4]);
}
GIN3_REGISTER("2264_047C", 5, run)
