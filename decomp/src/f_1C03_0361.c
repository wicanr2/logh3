/*
 * GIN3PS.EXE 1C03:0361  戰場艦隊的短程平移實行常式：不轉身、每步直指目的地前進一步（最長 8），扣行進額度；抵達後才轉向 +0x4C
 *
 * 原版（451 bytes，far 呼叫，一個 far 指標堆疊參數 F（位移 [bp+26h]、段 [bp+28h]）；序言 push bp／si／di、sub sp,1Ch、mov bp,sp）。
 *   區域變數：[bp+0] ＝ 040C 的結果、[bp+2] ＝ 方位 byte（只存不讀）、[bp+3] ＝ 回傳 byte、[bp+4]／[bp+6] ＝ 新位置 x'、y'（026F 的輸出槽）、[bp+8]／[bp+0Ah] ＝ 步長 L（低字、高字）。
 *   回傳 AL（描述檔 ret ＝ pack:al；AH 在多數路徑是殘值，不比對）。
 * 過程（已證實：反組譯逐指令；3310:0006、3CC9:00B9、2BE1:011D、2BE1:026F、2583:040C、2583:0451、1C03:0524 由其各自的差異測試證實）：
 *   F[+49] &= 0xFE
 *   位置 (F[+5C], F[+5E]) ＝ 目的地 (F[+64], F[+66])（兩個 word 都相等）：
 *     F[+4A] ≠ F[+4C]（byte）→ 回 1C03:0524(F, CX 的字)，只取 AL（CX 的低位元組 ＝ F[+4C]，高位元組 ＝ F[+67] 的殘值，0524 只取目標的低位元組）
 *     否則回 0
 *   額度 F[+58]:F[+56] 與 8.0 做 3CC9:00B9，jge 不成立（有號 < 8.0）→ 回 0
 *   回傳 byte ← 1
 *   d ← 3310:0006(F[+5C], F[+5E], F[+64], F[+66])
 *   3CC9:00B9(d, 8.0) 後 jg（ZF ＝ 0 且 SF ＝ OF）：
 *     成立 → 方位 ← 2BE1:011D(F[+64] − F[+5C], F[+66] − F[+5E])；2BE1:026F(F[+5C], F[+5E], 方位, 8, &x', &y')（x'、y' 在堆疊槽）
 *     不成立（d ≤ 8.0，含高字 >= 0x8000 的負值）→ x' ← F[+64]、y' ← F[+66]
 *   L ← 3310:0006(F[+5C], F[+5E], x', y')；L ＝ 0 → L ← 1（低字 1、高字 0）
 *   F[+58]:F[+56] −= L（32 位元，低字借位進高字）
 *   r ← 2583:040C(x', y', word F[+1])
 *     r ＝ 0 → F[+5C] ← x'、F[+5E] ← y'
 *     r ＝ 0xFFFF → 2583:0451(F[+5C], F[+5E], &F[+5C], &F[+5E], 8, word F[+1])（x'、y' 不寫回，由 0451 避障改寫位置）
 *     其他（被佔）→ 不動位置
 *   回 回傳 byte（1）
 * 語意（強推論）：docs/internals/33-fleet-table.md 第 118 行（bit4 ＝ 短程平移：1C03:0361 不轉身、每一步重新瞄準目的地直線前進，朝向整段不變，到了才轉向 +0x4C）、
 *   第 117 行（+0x4C ＝ 抵達朝向，不同就 1C03:0524 轉一步）、docs/spec/027 第 87 行、docs/re/102 第 121 行。F[+56] ＝ 行進額度（16.16）、F[+1] ＝ 陣營（強推論）。
 * 副作用：F[+49]、F[+56..59]、F[+5C]、F[+5E]、0524／0451 的寫入、3310:0006 內 39B5:005C 的暫存（ds:8EDC..8EED）、026F 對堆疊槽的寫入（堆疊帶，比對排除）。
 *
 * 機器模型：C 版沒有堆疊。x'、y' 放在原版的實際位置（基底狀態 fcmd-standby.state：SS 固定 0x63EB，進入點 SP ＝ 1E76，探針推 2 個參數字 4 bytes、返回位址 4、
 *   push bp／si／di 6、sub sp,1Ch，所以 bp ＝ 1E4C，x' 在 SS:1E50、y' 在 SS:1E52），交給 f_2BE1_026F 用 far 指標改寫，再從那裡讀回。
 * 輸入域：F 指向艦隊表內（描述檔 tables）；其餘欄位與戰場地圖取自 fcmd-standby.state 與突變；步數超過預算的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off,
                            uint16_t oy_seg);
extern uint16_t f_2583_040C(Mem *m, uint16_t x, uint16_t y, uint16_t faction);
extern uint16_t f_2583_0451(Mem *m, uint16_t x, uint16_t y, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg, uint16_t dist,
                            uint16_t faction);
extern uint16_t f_1C03_0524(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t target);

#define F0361_SS 0x63EBu /* fcmd-standby.state 的 SS（diffprobe 無法設定） */
#define F0361_NX 0x1E50u /* 進入點 SP 1E76 − 4 − 4 − 6 − 0x1C ＝ bp(1E4C)，x' 在 bp+4 */
#define F0361_NY 0x1E52u /* bp+6 */
#define F0361_ZF 0x0040u
#define F0361_SF 0x0080u
#define F0361_OF 0x0800u

/* 3310:0006 的 C 版把 39B5 的堆疊參數放在 fcmd-base.state 的框架（SS:1F82..1F89），在 fcmd-standby.state（進入點 SP 1E76）下落在探針堆疊帶外。
 * 存下 SS:1F78..1F8F 呼叫前的內容並在呼叫後還原：這 24 bytes 只是 C 版被呼叫者的模型暫存，不是本函式的輸出（理由同 f_1588_0891.c）。 */
static uint32_t f0361_dist(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2)
{
    uint8_t save[24];
    uint32_t r;
    int k;

    for (k = 0; k < 24; k++)
        save[k] = rd8(m, LIN(F0361_SS, (uint16_t)(0x1F78u + k)));
    r = f_3310_0006(m, x1, y1, x2, y2);
    for (k = 0; k < 24; k++)
        wr8(m, LIN(F0361_SS, (uint16_t)(0x1F78u + k)), save[k]);
    return r;
}

static uint16_t f0361_f16(const Mem *m, uint16_t seg, uint16_t off, uint16_t k)
{
    return rd16(m, LIN(seg, (uint16_t)(off + k)));
}

uint8_t f_1C03_0361(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint8_t ret = 0;
    uint16_t fl, nx, ny, faction, r;
    uint32_t d, l;

    wr8(m, LIN(f_seg, (uint16_t)(f_off + 0x49u)), (uint8_t)(rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x49u))) & 0xFEu));
    if (f0361_f16(m, f_seg, f_off, 0x5Cu) == f0361_f16(m, f_seg, f_off, 0x64u) &&
        f0361_f16(m, f_seg, f_off, 0x5Eu) == f0361_f16(m, f_seg, f_off, 0x66u)) {
        uint8_t head = rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Au)));
        uint8_t tgt = rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Cu)));

        if (head == tgt)
            return 0;
        return (uint8_t)f_1C03_0524(m, f_off, f_seg, (uint16_t)((f0361_f16(m, f_seg, f_off, 0x66u) & 0xFF00u) | tgt));
    }
    fl = f_3CC9_00B9(f0361_f16(m, f_seg, f_off, 0x58u), f0361_f16(m, f_seg, f_off, 0x56u), 8, 0);
    if (((fl & F0361_SF) != 0) != ((fl & F0361_OF) != 0)) /* jge 不成立 */
        return 0;
    ret = 1;
    d = f0361_dist(m, f0361_f16(m, f_seg, f_off, 0x5Cu), f0361_f16(m, f_seg, f_off, 0x5Eu), f0361_f16(m, f_seg, f_off, 0x64u),
                    f0361_f16(m, f_seg, f_off, 0x66u));
    fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 8, 0);
    if ((fl & F0361_ZF) == 0 && ((fl & F0361_SF) != 0) == ((fl & F0361_OF) != 0)) { /* jg */
        uint8_t ang = f_2BE1_011D(m, (uint16_t)(f0361_f16(m, f_seg, f_off, 0x64u) - f0361_f16(m, f_seg, f_off, 0x5Cu)),
                                  (uint16_t)(f0361_f16(m, f_seg, f_off, 0x66u) - f0361_f16(m, f_seg, f_off, 0x5Eu)));

        (void)f_2BE1_026F(m, f0361_f16(m, f_seg, f_off, 0x5Cu), f0361_f16(m, f_seg, f_off, 0x5Eu), ang, 8, F0361_NX, F0361_SS, F0361_NY,
                          F0361_SS);
        nx = rd16(m, LIN(F0361_SS, F0361_NX));
        ny = rd16(m, LIN(F0361_SS, F0361_NY));
    } else {
        nx = f0361_f16(m, f_seg, f_off, 0x64u);
        ny = f0361_f16(m, f_seg, f_off, 0x66u);
    }
    l = f0361_dist(m, f0361_f16(m, f_seg, f_off, 0x5Cu), f0361_f16(m, f_seg, f_off, 0x5Eu), nx, ny);
    if (l == 0)
        l = 1;
    {
        uint32_t budget = ((uint32_t)f0361_f16(m, f_seg, f_off, 0x58u) << 16) | f0361_f16(m, f_seg, f_off, 0x56u);

        budget -= l;
        wr16(m, LIN(f_seg, (uint16_t)(f_off + 0x56u)), (uint16_t)budget);
        wr16(m, LIN(f_seg, (uint16_t)(f_off + 0x58u)), (uint16_t)(budget >> 16));
    }
    faction = f0361_f16(m, f_seg, f_off, 0x01u);
    r = f_2583_040C(m, nx, ny, faction);
    if (r == 0) {
        wr16(m, LIN(f_seg, (uint16_t)(f_off + 0x5Cu)), nx);
        wr16(m, LIN(f_seg, (uint16_t)(f_off + 0x5Eu)), ny);
    } else if (r == 0xFFFFu) {
        (void)f_2583_0451(m, f0361_f16(m, f_seg, f_off, 0x5Cu), f0361_f16(m, f_seg, f_off, 0x5Eu), (uint16_t)(f_off + 0x5Cu), f_seg,
                          (uint16_t)(f_off + 0x5Eu), f_seg, 8, f0361_f16(m, f_seg, f_off, 0x01u));
    }
    return ret;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_1C03_0361(m, a[0], a[1]);
}
GIN3_REGISTER("1C03_0361", 2, run)
