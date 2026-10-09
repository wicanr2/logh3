/*
 * GIN3PS.EXE 39B5:005C  16.16 定點有號除法：回傳 (a << 16) ÷ b 的低 32 位元（有號，截斷），除數 0 回 0
 *
 * 原版：far 呼叫，四個 word 堆疊參數（push bp／si／di／es，沒有 sub sp，BP ＝ push bp 之後的 SP，所以 arg_0 在 [bp+6]）：
 *   args[0..1] ＝ a（被除數，低字、高字）  args[2..3] ＝ b（除數，低字、高字）
 *   回傳 AX ＝ 結果高字、BX ＝ 結果低字（pack:ax,bx）。
 * 過程（已證實：反組譯逐指令）：
 *   AX ← 39B5:010C()          ; near，BP ＝ 本框架；a、b 就地取絕對值（寫回 [bp+6..0Dh]），回 0x8000 表示兩者異號，否則 0；先 push 起來
 *   ds:8EDC 起 9 個 word 清 0（rep stosw，ES ＝ DS，DF ＝ 0）：三組 48 位元數
 *     R ＝ ds:8EDC..8EE1（餘數／被除數）：[8EDE] ← |a|.lo、[8EE0] ← |a|.hi，所以 R ＝ |a| × 65536
 *     D ＝ ds:8EE2..8EE7（除數）：[8EE2] ← |b|.lo、[8EE4] ← |b|.hi，[8EE6] ＝ 0
 *     Q ＝ ds:8EE8..8EED（商）：初值 0
 *   |b| ＝ 0（`mov ax,[bp+0Ch]; or ax,[bp+0Ah]`）→ 不做除法，Q ＝ 0（回 0，不引發 #DE）
 *   否則（cx ＝ 1）：
 *     對齊：D 的第 47 位元沒設 → 若 R < D（48 位元無號，jb）停；否則 cx++、D <<= 1（48 位元），重複
 *     迴圈 cx 次：R_old ≥ D → R ← R − D；Q ← (Q << 1) | (R_old ≥ D)（48 位元 rcl）；D >>= 1（邏輯右移 48 位元）
 *   BX ← Q 低字 [8EE8]、AX ← Q 第二個 word [8EEA]；010C 的回傳有設 bit15 → 32 位元取負：BX ← 0 − [8EE8]，AX ← 0 − [8EEA] − 借位
 * 用途（假說）：16.16 定點有號除法（docs/re/129 第 72 行 `q = (m << 16) / M ; 39B5:005C，截斷`；docs/re/124 第 26 行；
 *   docs/re/147 第 12 行 `16.16 除（移位相減）`）。呼叫者（台帳）：1F6F:056C、203A:00EC、203A:0214、203A:03B0、2264:047C、
 *   2264:0FC3、303F:02DE、3310:0006、3310:014B。
 * 副作用：ds:8EDC..8EED 共 18 bytes（R 最終餘數、D 最終移位值、Q 商）；a、b 在呼叫者框架（堆疊）內被改成絕對值；
 *   ES 還原，其他暫存器不比對。商超過 32 位元時（|b| 小、|a| 大）第三個 word [8EEC] 非 0，但回傳只取低 32 位元。
 *
 * 機器模型：C 版沒有堆疊。四個參數放在原版的實際位置 SS:1F82..1F89（基底狀態 SP ＝ 1F8A，SS 固定 0x63EB，diffprobe 無法設定；
 *   進入點 SP ＝ 1F8A − 8，push 返回位址 4 bytes、push bp 2 bytes，BP ＝ 1F7C），交給 f_39B5_010C(bp) 就地取絕對值，再讀回。
 *   這個位址在探針的堆疊帶（SS:1B8A..1F89）內，寫入集合比對時雙方都排除，與原版把絕對值寫回堆疊一致。
 *   DS 固定 DGROUP（0x54B3），ds:8EDC..8EED 的位移全為偶數，word 寫入不會跨過位移 0xFFFF。
 * 輸入域：四個參數全部 16 位元值（含 0x80000000 這種取絕對值後仍為負的數：以無號 48 位元運算，照原版）。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint16_t f_39B5_010C(Mem *m, uint16_t bp);

#define F005C_SS 0x63EBu  /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F005C_BP 0x1F7Cu  /* 進入點 SP 1F8A − 參數 8 − 返回位址 4 − push bp 2 */
#define F005C_M48 0xFFFFFFFFFFFFull

static void f005C_wr16s(Mem *m, uint16_t off, uint16_t v)
{
    wr8(m, LIN(F005C_SS, off), (uint8_t)v);
    wr8(m, LIN(F005C_SS, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

static uint16_t f005C_rd16s(const Mem *m, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(F005C_SS, off)) | (rd8(m, LIN(F005C_SS, (uint16_t)(off + 1u))) << 8));
}

/* 把 48 位元數寫成三個 word（低字在低位址）到 ds:off */
static void f005C_put48(Mem *m, uint16_t off, uint64_t v)
{
    wr16(m, DS(off), (uint16_t)v);
    wr16(m, DS((uint16_t)(off + 2u)), (uint16_t)(v >> 16));
    wr16(m, DS((uint16_t)(off + 4u)), (uint16_t)(v >> 32));
}

uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi)
{
    uint16_t sign, ax, bx, q0, q1;
    uint64_t r, d, q = 0;
    unsigned cx;
    int i;

    f005C_wr16s(m, (uint16_t)(F005C_BP + 6u), a_lo);
    f005C_wr16s(m, (uint16_t)(F005C_BP + 8u), a_hi);
    f005C_wr16s(m, (uint16_t)(F005C_BP + 0x0Au), b_lo);
    f005C_wr16s(m, (uint16_t)(F005C_BP + 0x0Cu), b_hi);

    sign = f_39B5_010C(m, F005C_BP);

    a_lo = f005C_rd16s(m, (uint16_t)(F005C_BP + 6u));
    a_hi = f005C_rd16s(m, (uint16_t)(F005C_BP + 8u));
    b_lo = f005C_rd16s(m, (uint16_t)(F005C_BP + 0x0Au));
    b_hi = f005C_rd16s(m, (uint16_t)(F005C_BP + 0x0Cu));

    r = ((uint64_t)a_hi << 32) | ((uint64_t)a_lo << 16);
    d = ((uint64_t)b_hi << 16) | b_lo;

    if ((uint16_t)(b_hi | b_lo) != 0) {
        cx = 1;
        while (!(d & (1ull << 47))) {
            if (r < d)
                break;
            cx++;
            d = (d << 1) & F005C_M48;
        }
        for (i = 0; i < (int)cx; i++) {
            int ge = r >= d;

            if (ge)
                r -= d;
            q = ((q << 1) | (uint64_t)ge) & F005C_M48;
            d >>= 1;
        }
    }
    f005C_put48(m, 0x8EDCu, r);
    f005C_put48(m, 0x8EE2u, d);
    f005C_put48(m, 0x8EE8u, q);

    q0 = (uint16_t)q;
    q1 = (uint16_t)(q >> 16);
    if (sign & 0x8000u) {
        bx = (uint16_t)(0u - q0);
        ax = (uint16_t)(0u - q1 - (q0 != 0u ? 1u : 0u));
    } else {
        bx = q0;
        ax = q1;
    }
    return ((uint32_t)ax << 16) | bx;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_39B5_005C(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("39B5_005C", 4, run)
