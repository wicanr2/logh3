/*
 * GIN3PS.EXE 20B4:000C  行星攻擊的開場：把在場艦隊清單裡每支艦隊的 +0x70 清 0、+0x5A 設 0xFF，再對目前星域切片的每顆行星 i 呼叫 20B4:0091(i, P_i)
 *
 * 原版（133 bytes，far 呼叫，沒有堆疊參數；序言 push bp／si、sub sp,0Eh、mov bp,sp）。沒有有意義的回傳值（AX 是殘值，描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ i（byte）、[bp+2]／[bp+4] ＝ 清單指標（位移從 ds:9898 起）、[bp+6]／[bp+8] ＝ 行星指標（位移、段）、[bp+0Ah]／[bp+0Ch] ＝ 艦隊指標（寫了不再讀）。
 * 過程（已證實：反組譯逐指令；20B4:0091 由其差異測試證實）：
 *   第一圈：i ← 0；當 i < byte ds:9894（無號，每圈重讀）：
 *     f ← 0xC074 + byte ds:(9898 + i) × 0x7A；word f[+0x70] ← 0；byte f[+0x5A] ← 0xFF；i 加一
 *   第二圈：P ← (word ds:CFFA):(word ds:CFF8)（只讀一次）；i ← 0；當 i < byte ds:CFFC（無號，每圈重讀）：
 *     20B4:0091(side ＝ i, P)    ; push 的順序是段、位移、AX（AL ＝ i，AH 是殘值，0091 只用低位元組）；呼叫後 add sp,6
 *     i 加一、P 的位移加 0x52（16 位元回繞）
 * 語意（強推論）：docs/spec/019-planet-assault.md 第 152 行「開場由 20B4:000C 把 +0x70 清 0、+0x5A 設」、docs/spec/062-battle-turn-eight-steps.md 第 64 行
 *   「20B4:000C 標攻擊者」；ds:CFF8／CFFA／CFFC 是目前星域的行星切片遠指標與顆數（docs/internals/29-planets-and-economy.md 第 20 行，一筆 82 bytes ＝ 0x52）。
 *   +0x70 的 bit0／bit1 與 +0x5A（正在攻打哪一顆行星的索引，0xFF ＝ 沒有）見 f_20B4_0091.c 檔頭。
 * 副作用：第一圈寫 f[+0x70]（word）、f[+0x5A]；第二圈全部來自 20B4:0091（f[+0x70]、f[+0x5A] 與 3310:0006 的 ds:8EDC..8EED 暫存）。
 *
 * 輸入域：同 f_20B4_0091.c（基底 fcmd-base.state，被呼叫的 39B5 C 版寫死進入點 SP ＝ 1F8A 的堆疊位址；P 的位移不落在 0xFFEB，2AF5:01F7 的 C 版用線性位址讀 P[+0x14] 的 word）。
 *   艦隊指標 f 加欄位時 16 位元回繞，word 寫入在位移 0xFFFF 時 dosgolem 的高位元組取同一段的位移 0（wr16s）。被 0091 可以改寫的位置（f + 0x70、f + 0x5A）不含
 *   ds:9894、ds:9898..9997、ds:CFF8..CFFC，所以「每圈重讀」與「只讀一次」在輸入域內等價；本版仍照原版每圈重讀。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_20B4_0091(Mem *m, uint16_t side, uint16_t p_off, uint16_t p_seg);

static void w000C_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

static uint16_t w000C_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w000C_clear(Mem *m)
{
    uint8_t i = 0;
    uint16_t list = 0x9898u;

    while (i < rd8(m, DS(0x9894u))) {
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(list)) * 0x7Au));

        w000C_wr16s(m, GIN3_DS, (uint16_t)(f + 0x70u), 0);
        wr8(m, DS((uint16_t)(f + 0x5Au)), 0xFF);
        i = (uint8_t)(i + 1u);
        list = (uint16_t)(list + 1u);
    }
}

static void w000C_mark(Mem *m)
{
    uint8_t i = 0;
    uint16_t p_seg = w000C_rd16s(m, GIN3_DS, 0xCFFAu);
    uint16_t p_off = w000C_rd16s(m, GIN3_DS, 0xCFF8u);

    while (i < rd8(m, DS(0xCFFCu))) {
        f_20B4_0091(m, i, p_off, p_seg);
        i = (uint8_t)(i + 1u);
        p_off = (uint16_t)(p_off + 0x52u);
    }
}

void f_20B4_000C(Mem *m)
{
    w000C_clear(m);
    w000C_mark(m);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_20B4_000C(m);
    return 0;
}
GIN3_REGISTER("20B4_000C", 0, run)
