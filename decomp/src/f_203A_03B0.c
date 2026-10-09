/*
 * GIN3PS.EXE 203A:03B0  兩個物件的距離 ÷ (A × (1 + e))，e ＝ 3870:0643(A, B)：16.16 定點比值，回 AX:BX
 *
 * 原版（116 bytes，far 呼叫，六個 word 堆疊參數；序言 push bp／si／di、sub sp,0Ch、mov bp,sp，所以第一個參數在 [bp+16h]）：
 *   p1 ＝ far 指標（位移 [bp+16h]、段 [bp+18h]）、p2 ＝ far 指標（[bp+1Ah]、[bp+1Ch]）、A ＝ word [bp+1Eh]、B ＝ word [bp+20h]。
 *   回傳 AX ＝ 高字、BX ＝ 低字（描述檔 ret ＝ pack:ax,bx），就是 39B5:005C 的回傳，函式本身之後不再改 AX、BX。
 * 過程（已證實：反組譯逐指令；3870:0643、3CC9:0009、3310:0006、39B5:005C 各由其差異測試證實）：
 *   e ← 3870:0643(A, B)                                   ; AX:BX，16.16 的離心率（參數是兩個 word，呼叫後 mov sp,bp 清除）
 *   S ← (e × A) 的低 32 位元，A 以無號 32 位元（高字 0）參與乘法，乘法走 3CC9:0009
 *   S ← S ＋ (A << 16)                                    ; `adc ax, cx` 把 A（高字加 A，低字加 0）加進高字；只留 32 位元
 *   d ← 3310:0006(p1[+0x34], p1[+0x36], p2[+0x34], p2[+0x36])   ; 兩點 (x1, y1)、(x2, y2) 的 16.16 距離；每個 word 以 es:[si+34h]／[si+36h] 讀取
 *   回 39B5:005C(d, S)                                    ; (d << 16) ÷ S，有號 32 位元，除數 0 回 0
 *   讀取順序：先 p2 的 +0x36、+0x34，再 p1 的 +0x36、+0x34（push 順序），先 les p2 後 les p1。兩個指標指向同一位址時結果相同；
 *   p1／p2 不指向本函式會改寫的位置（本函式只寫 ds:8EDC..8EED，來自 39B5:005C），所以讀取順序沒有可觀察差異。
 * 語意（假說）：docs 沒有 203A:03B0 的記載（grep 只有 docs/internals/data/fleet-combat-damage.json 的觀測點清單與 docs/re/404 第 22 行提到呼叫者 203A:00EC）。
 *   +0x34／+0x36 在艦艇記錄裡是座標（docs/internals/64 與本專案 25DB 系列），所以 d 是兩艘艦艇的距離；A、B 看起來是橢圓的長短半軸
 *   （3870:0643 算離心率 e ＝ √(1 − B² ÷ A²)，docs/re/94 第 43 行）；S ＝ A × (1 ＋ e)。整體的物理意義未知。
 * 副作用：只有 39B5:005C 寫的 ds:8EDC..8EED 共 18 bytes。
 *
 * 機器模型：C 版沒有堆疊。被呼叫者的堆疊參數與區域變數都在探針的堆疊帶內，比對時排除。
 * 輸入域：p1、p2 指向艦艇池記錄（描述檔 table ship）或其他不在 A0000 到 BFFFF 的位址；A、B 全部 16 位元值。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint32_t f_3870_0643(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);

uint32_t f_203A_03B0(Mem *m, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off, uint16_t p2_seg, uint16_t a, uint16_t b)
{
    uint32_t e = f_3870_0643(m, a, b);
    uint32_t s = f_3CC9_0009((uint16_t)(e >> 16), (uint16_t)e, 0, a); /* CX ＝ 0，DX ＝ A */
    uint16_t x1, y1, x2, y2;
    uint32_t d;

    s = (uint32_t)(s + ((uint32_t)a << 16)); /* adc ax, cx：高字加 A */
    x2 = rd16(m, LIN(p2_seg, (uint16_t)(p2_off + 0x34u)));
    y2 = rd16(m, LIN(p2_seg, (uint16_t)(p2_off + 0x36u)));
    x1 = rd16(m, LIN(p1_seg, (uint16_t)(p1_off + 0x34u)));
    y1 = rd16(m, LIN(p1_seg, (uint16_t)(p1_off + 0x36u)));
    d = f_3310_0006(m, x1, y1, x2, y2);
    return f_39B5_005C(m, (uint16_t)d, (uint16_t)(d >> 16), (uint16_t)s, (uint16_t)(s >> 16));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_203A_03B0(m, a[0], a[1], a[2], a[3], a[4], a[5]);
}
GIN3_REGISTER("203A_03B0", 6, run)
