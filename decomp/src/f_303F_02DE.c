/*
 * GIN3PS.EXE 303F:02DE  點是否在橢圓內：L ≤ S，S ＝ (b² ÷ a) ÷ (1 − e × cos θ)（極座標下以焦點為原點的橢圓半徑，16.16），L 是呼叫者給的 32 位元距離
 *
 * 原版（179 bytes，far 呼叫，五個 word 堆疊參數；序言 push bp／si／di、sub sp,0Ch、mov bp,sp，所以 arg_0 在 [bp+16h]、arg_2 在 [bp+18h]、arg_4 在 [bp+1Ah]、
 *   arg_6 在 [bp+1Ch]、arg_8 在 [bp+1Eh]）。回傳 AX ＝ 1 或 0。區域變數：[bp+0..3] ＝ e（低字、高字）、[bp+4..7] ＝ r、[bp+8..0Bh] ＝ S。
 *   L ＝ arg_2:arg_0（高字:低字），θ ＝ arg_4 的低位元組（256 為一圈），a ＝ arg_6，b ＝ arg_8（命名只是座標，語意見下）。
 * 過程（已證實：反組譯逐指令；3870:0643、3CC9:0009／0070／00B9、39B5:000C／005C 由其各自的差異測試證實）：
 *   e ← 3870:0643(a, b)                                       ; 16.16 離心率，AX:BX
 *   B2 ← 3CC9:0009(0:b, 0:b)                                  ; b 零延伸，取低 32 位元
 *   r ← 3CC9:0070((B2 的低字) << 16, 0:a)                      ; 有號 32 位元除法，a 零延伸；a ＝ 0 時商 0xFFFFFFFF
 *   t ← sext32(word ds:(0x573A + 2 × ((θ ＋ 0x40) 的低位元組))) << 6   ; 正弦表（docs/internals/41），θ ＋ 0x40 ＝ cos θ；sin × 1024 左移 6 ＝ 16.16
 *   p ← 39B5:000C(e, t)                                       ; (e × t) >> 16，有號
 *   den ← 0x10000 − p                                         ; 32 位元 sub／sbb（1 ＝ 1.0 的高字）
 *   S ← 39B5:005C(r, den)                                     ; (r << 16) ÷ den，有號，除數 0 回 0
 *   3CC9:00B9(L, S) 後 jle（ZF ＝ 1 或 SF ≠ OF）→ 回 1，否則回 0
 * 語意（強推論）：極座標橢圓方程 ρ ＝ (b²/a) ÷ (1 − e cos θ)，與 docs/internals/51 第 41 行（39B5:000C／005C、3870:0643／0614 是要塞砲橢圓判定用）、
 *   docs/re/145、147 的敘述相符；呼叫者 303F:0221。用途為強推論。
 * 副作用：全部來自 3870:0643 → 3310:014B → 39B5:005C 與本函式直接呼叫的 39B5:005C（ds:8EDC..8EED 共 18 bytes）；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊；被呼叫的 39B5:000C／005C 的堆疊參數放在 fcmd-base.state 進入點 SP ＝ 1F8A 的框架（f_39B5_000C.c／f_39B5_005C.c 檔頭），
 *   原版從本函式呼叫時框架更深，位置不同，但兩處都在探針的堆疊帶內，比對時排除，C 版自寫自讀，不影響結果。
 * 輸入域：五個參數全部 16 位元值；正弦表 ds:573A 起 512 bytes 取記憶體內容。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint32_t f_3870_0643(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint32_t f_39B5_000C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F02DE_ZF 0x0040u
#define F02DE_SF 0x0080u
#define F02DE_OF 0x0800u

uint16_t f_303F_02DE(Mem *m, uint16_t l_lo, uint16_t l_hi, uint16_t theta, uint16_t a, uint16_t b)
{
    uint32_t e, b2, r, rem, t, p, den, s;
    uint16_t fl, idx;

    e = f_3870_0643(m, a, b);
    b2 = f_3CC9_0009(0, b, 0, b);
    gin3_ldivmod((b2 & 0xFFFFu) << 16, (uint32_t)a, &r, &rem);
    idx = (uint16_t)((uint8_t)(theta + 0x40u));
    t = (uint32_t)(int32_t)(int16_t)rd16(m, DS((uint16_t)(0x573Au + 2u * idx))) << 6;
    p = f_39B5_000C(m, (uint16_t)e, (uint16_t)(e >> 16), (uint16_t)t, (uint16_t)(t >> 16));
    den = 0x10000u - p;
    s = f_39B5_005C(m, (uint16_t)r, (uint16_t)(r >> 16), (uint16_t)den, (uint16_t)(den >> 16));
    fl = f_3CC9_00B9(l_hi, l_lo, (uint16_t)(s >> 16), (uint16_t)s);
    if ((fl & F02DE_ZF) != 0 || ((fl & F02DE_SF) != 0) != ((fl & F02DE_OF) != 0)) /* jle */
        return 1;
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_303F_02DE(m, a[0], a[1], a[2], a[3], a[4]);
}
GIN3_REGISTER("303F_02DE", 5, run)
