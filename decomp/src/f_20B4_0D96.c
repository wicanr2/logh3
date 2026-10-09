/*
 * GIN3PS.EXE 20B4:0D96  艦隊打行星的傷害基準：A ＝ 攻擊力 − 攻擊力 × 距離 ÷ 750（距離 16.16），回 A ÷ 2 ＋ rand() mod (A ÷ 3)
 *
 * 原版（150 bytes，far 呼叫，兩個 far 指標堆疊參數；序言 push bp／si／di、sub sp,8、mov bp,sp，所以艦隊指標位移在 [bp+12h]、段在 [bp+14h]，
 *   行星指標位移在 [bp+16h]、段在 [bp+18h]）。回傳 AX。區域變數：[bp+0] ＝ A、[bp+2] ＝ 結果、[bp+4]／[bp+6] ＝ 距離 ÷ 150 的低字／高字。
 * 過程（已證實：反組譯逐指令；203A:0424、3310:0006、3CC9:0009／0070、3CD5:000C 由其各自的差異測試證實）：
 *   A ← 203A:0424(艦隊)                                                      ; 戰略艦隊攻擊力（16 位元）
 *   D ← 3310:0006(x1 ＝ f[+5C], y1 ＝ f[+5E], x2 ＝ p[+46], y2 ＝ p[+48])       ; 16.16 距離，AX:BX
 *   Q ← D ÷ 150                                                               ; 3CC9:0070，有號 32 位元
 *   P ← A × Q                                                                 ; 3CC9:0009，AX:BX ＝ 0:A、CX:DX ＝ Q，取低 32 位元
 *   Q2 ← P ÷ 0x50000                                                          ; mov cx,5；xor si,si；xchg dx,si → CX:DX ＝ 5:0 ＝ 5.0（16.16），3CC9:0070 有號
 *   A ← A − (Q2 的低字)                                                       ; 16 位元回繞
 *   R ← A shr 1（邏輯右移）
 *   r ← 3CD5:000C()
 *   d ← A ÷ 3（16 位元無號 div）；回 R ＋ (r mod d)（16 位元回繞）
 *   d ＝ 0（A < 3）時第二個 div 引發 #DE，屬輸入域外（探針丟棄）；C 版此時回 0，不參與比較。
 * 語意（強推論）：docs/internals/33-fleet-table.md 第 277 行（+0x5C／+0x5E ＝ 打行星時算距離用的現在座標，20B4:0D96）、
 *   docs/spec/019-planet-assault.md 第 106、123 行（B ＝ 20B4:0D96(艦隊, 行星)，「未解」）。本版只照原版運算，用途為強推論，
 *   p[+46]／p[+48] ＝ 行星座標（強推論）。
 * 副作用：ds:94DC..94DF（rand 種子）與 3310:0006 內 39B5:005C 的暫存（ds:8EDC..8EED）。
 *
 * 輸入域：艦隊指標指到艦隊表內、行星指標指到行星表內（描述檔 tables），其餘由探針突變讀取集；A < 3 的案例（#DE）不涵蓋。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint16_t f_203A_0424(Mem *m, uint16_t fleet_off, uint16_t fleet_seg);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint16_t f_3CD5_000C(Mem *m);

uint16_t f_20B4_0D96(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t p_off, uint16_t p_seg)
{
    uint16_t a, half, rnd, div;
    uint32_t d, q, q2, p, r;

    a = f_203A_0424(m, f_off, f_seg);
    d = f_3310_0006(m, rd16(m, LIN(f_seg, (uint16_t)(f_off + 0x5Cu))), rd16(m, LIN(f_seg, (uint16_t)(f_off + 0x5Eu))),
                    rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x46u))), rd16(m, LIN(p_seg, (uint16_t)(p_off + 0x48u))));
    gin3_ldivmod(d, 150u, &q, &r);
    p = f_3CC9_0009(0, a, (uint16_t)(q >> 16), (uint16_t)q);
    gin3_ldivmod(p, 0x50000u, &q2, &r);
    a = (uint16_t)(a - (uint16_t)q2);
    half = (uint16_t)(a >> 1);
    rnd = f_3CD5_000C(m);
    div = (uint16_t)(a / 3u);
    if (div == 0)
        return 0; /* #DE：輸入域外 */
    return (uint16_t)(half + (uint16_t)(rnd % div));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_20B4_0D96(m, a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("20B4_0D96", 4, run)
