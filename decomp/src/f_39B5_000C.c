/*
 * GIN3PS.EXE 39B5:000C  16.16 定點有號乘法：回傳 (a × b) >> 16 的低 32 位元（有號，截斷），a、b 各是 32 位元有號數
 *
 * 原版：far 呼叫，四個 word 堆疊參數（push bp／si／di，沒有 sub sp，BP ＝ push bp 之後的 SP，所以 arg_0 在 [bp+6]）：
 *   args[0..1] ＝ a（低字、高字）  args[2..3] ＝ b（低字、高字）
 *   回傳 AX ＝ 結果高字、BX ＝ 結果低字（pack:ax,bx）。
 * 過程（已證實：反組譯逐指令）：
 *   AX ← 39B5:010C()          ; near，BP ＝ 本框架；a、b 就地取絕對值（寫回 [bp+6..0Dh]），回 0x8000 表示兩者異號，否則 0
 *   以無號 16 位元 mul 做 32 × 32 → 取中間 32 位元（第 16 到 47 位元）：
 *     di ← 高字(a.lo × b.lo)；si ← 高字(a.lo × b.hi)；di += 低字(a.lo × b.hi)（進位 → si += 1）
 *     di += 低字(a.hi × b.lo)；si += 高字(a.hi × b.lo) + 進位
 *     si += 低字(a.hi × b.hi)                          ; 只加低字，a.hi × b.hi 的高字與更高位元被截掉
 *   符號位元（010C 的回傳）有設 → (si:di) ← 0 − (si:di)（32 位元二補數）；AX ← si、BX ← di
 *   a.lo × b.lo 的低字、a.hi × b.hi 的高字不進結果。
 * 用途（假說）：16.16 定點乘法（docs/re/129 第 73 行 `r = 1.0 + (q × q >> 16) ; 39B5:000C，截斷`；docs/re/124 第 29 行）。
 *   呼叫者（台帳）：1F6F:056C、203A:0214、2264:047C、2264:0FC3、303F:02DE、3310:0006。
 * 副作用：a、b 在呼叫者框架（堆疊）內被改成絕對值；沒有堆疊以外的寫入。
 *
 * 機器模型：C 版沒有堆疊。四個參數放在原版的實際位置 SS:1F82..1F89（基底狀態 SP ＝ 1F8A，SS 固定 0x63EB，diffprobe 無法設定；
 *   進入點 SP ＝ 1F8A − 8，push 返回位址 4 bytes、push bp 2 bytes，BP ＝ 1F7C），交給 f_39B5_010C(bp) 就地取絕對值，再讀回。
 *   這個位址在探針的堆疊帶（SS:1B8A..1F89）內，寫入集合比對時雙方都排除，與原版把絕對值寫回堆疊一致。
 * 輸入域：四個參數全部 16 位元值（含 0x80000000 這種取絕對值後仍為負的數；結果照原版運算）。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint16_t f_39B5_010C(Mem *m, uint16_t bp);

#define F000C_SS 0x63EBu  /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F000C_BP 0x1F7Cu  /* 進入點 SP 1F8A − 參數 8 − 返回位址 4 − push bp 2 */

static void f000C_wr16s(Mem *m, uint16_t off, uint16_t v)
{
    wr8(m, LIN(F000C_SS, off), (uint8_t)v);
    wr8(m, LIN(F000C_SS, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

static uint16_t f000C_rd16s(const Mem *m, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(F000C_SS, off)) | (rd8(m, LIN(F000C_SS, (uint16_t)(off + 1u))) << 8));
}

uint32_t f_39B5_000C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi)
{
    uint16_t sign, di, si, ax, bx;
    uint32_t p, t;

    f000C_wr16s(m, (uint16_t)(F000C_BP + 6u), a_lo);
    f000C_wr16s(m, (uint16_t)(F000C_BP + 8u), a_hi);
    f000C_wr16s(m, (uint16_t)(F000C_BP + 0x0Au), b_lo);
    f000C_wr16s(m, (uint16_t)(F000C_BP + 0x0Cu), b_hi);

    sign = f_39B5_010C(m, F000C_BP);

    a_lo = f000C_rd16s(m, (uint16_t)(F000C_BP + 6u));
    a_hi = f000C_rd16s(m, (uint16_t)(F000C_BP + 8u));
    b_lo = f000C_rd16s(m, (uint16_t)(F000C_BP + 0x0Au));
    b_hi = f000C_rd16s(m, (uint16_t)(F000C_BP + 0x0Cu));

    di = (uint16_t)(((uint32_t)a_lo * b_lo) >> 16);
    p = (uint32_t)a_lo * b_hi;
    si = (uint16_t)(p >> 16);
    t = (uint32_t)di + (uint16_t)p;                 /* add di,ax */
    di = (uint16_t)t;
    si = (uint16_t)(si + (t >> 16));                /* adc si,0 */
    p = (uint32_t)a_hi * b_lo;
    t = (uint32_t)di + (uint16_t)p;                 /* add di,ax */
    di = (uint16_t)t;
    si = (uint16_t)(si + (uint16_t)(p >> 16) + (t >> 16)); /* adc si,dx */
    p = (uint32_t)a_hi * b_hi;
    si = (uint16_t)(si + (uint16_t)p);              /* add si,ax */

    if (sign & 0x8000u) {
        bx = (uint16_t)(0u - di);
        ax = (uint16_t)(0u - si - (di != 0u ? 1u : 0u));
    } else {
        bx = di;
        ax = si;
    }
    return ((uint32_t)ax << 16) | bx;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_39B5_000C(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("39B5_000C", 4, run)
