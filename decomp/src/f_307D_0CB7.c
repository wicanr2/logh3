/*
 * GIN3PS.EXE 307D:0CB7  艦隊 F 選一個目標點 ent（ds:D6BC 起 8 筆中第一個欄位最大的、同值取最後一筆），依剩餘行程決定 F 的目的地（+0x64／+0x66）與朝向（+0x4C）、速度（+0x52），再呼叫 2583:052C
 *
 * 原版（590 bytes，far 呼叫，兩個 far 指標堆疊參數 F、G；序言 push bp／si／di、sub sp,2Ah、mov bp,sp，所以 F 的位移在 [bp+34h]、段在 [bp+36h]，
 *   G 的位移在 [bp+38h]、段在 [bp+3Ah]）。沒有有意義的回傳值（AX 是 2583:052C 的殘值，描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ 剩餘行程 rem（word）、[bp+2] ＝ i（byte）、[bp+3] ＝ ang1（byte）、[bp+4] ＝ r1（byte）、[bp+5] ＝ 暫存（寫了不再讀）、[bp+6] ＝ 目前最大值、
 *   [bp+8] ＝ q1、[bp+0Ah] ＝ q2、[bp+0Ch..0Eh] ＝ 掃描指標、[bp+10h..12h] ＝ ent 指標、其餘是 F、ent 的欄位位址暫存（寫了再讀）。
 * 過程（已證實：反組譯逐指令；3310:0006、3CC9:0070、2BE1:011D、2BE1:024A、2BE1:026F、2583:052C 各由其差異測試證實）：
 *   選 ent：best ← word ds:D6E6、ent ← D6E6；i ← 0..7：w ← word ds:(D6BC ＋ 6 × i)；w < best（無號 jb）就跳過，否則 best ← w、ent ← D6BC ＋ 6 × i
 *     （所以同值取較後面的一筆；初值是第 8 筆）。ent 的三個 word：ent[0] ＝ 比較用的欄位、ent[2] ＝ x、ent[4] ＝ y（強推論，由用法）。
 *   rem ← byte F[+4F]
 *   d ← 3310:0006(x1 ＝ F[+5C], y1 ＝ F[+5E], x2 ＝ ent[2], y2 ＝ ent[4])        ; 16.16 距離
 *   rem ← rem − low16(3CC9:0070(d ÷ 8.0))                                         ; 有號 32 位元除法，商取低字
 *   ang1 ← 2BE1:011D(ent[2] − F[+5C], ent[4] − F[+5E])                            ; 讀取在 3310 之後重做（3310 內部會寫 ds:8EDC..8EED）
 *   r1 ← 2BE1:024A(ang1, F[+4A])；q1 ← r1 ÷ byte F[+4D]（無號 16 位元除法，除數 0 是 #DE，不在輸入域）；rem ← rem − q1
 *   ang2 ← 2BE1:011D(G[+5C] − ent[2], G[+5E] − ent[4])
 *   r2 ← 2BE1:024A(ang1 ＋ (r1 << 8), ang2)；q2 ← r2 ÷ byte F[+4D]；rem ← rem − q2
 *   rem > 0（有號 16 位元 jle 不成立）：F[+64] ← ent[2]；F[+66] ← ent[4]
 *   rem <= 0：s ← q1 ＋ q2（16 位元）；c ← byte F[+4F]
 *     s <= c（無號 ja 不成立）：dist ← (c − s) << 3（16 位元）
 *     s >  c：dist ← low16(3CC9:0070((c << 3) ÷ 3))
 *     2BE1:026F(x ＝ F[+5C], y ＝ F[+5E], 角度 ＝ ang1 ＋ (r1 << 8)（026F 只取低位元組）, dist, →F+0x64, →F+0x66)
 *   結尾（兩條路徑都做）：
 *     ang3 ← 2BE1:011D(G[+74] − F[+64], G[+76] − F[+66])；byte F[+4C] ← ang3；byte F[+48] ← 1
 *     dword F[+52] ← byte F[+4F] << 19（即 16.16 的 F[+4F] × 8.0；先寫高字 F[+54]、後寫低字 F[+52]）
 *     2583:052C(F)
 * 語意（假說）：docs 沒有 307D:0CB7 的記載（grep 無命中；呼叫者 307D:0521）。F[+64]／F[+66] 是目的地、F[+5C]／F[+5E] 是現位置、F[+48] 是狀態（1 ＝ 移動中，docs/internals/33-fleet-table.md）、
 *   F[+52] 起的 dword 是行程上限（16.16，2583:052C 檔頭）、F[+4C] 是朝向（docs/internals/33）；ds:D6BC 起的 8 筆是什麼表、F[+4F]／F[+4A]／F[+4D] 的意義未知。
 *   整體看起來是「在 8 個候選點中挑最大者，若走過去的距離、轉向成本在預算內就直接指定為目的地，否則朝它走一段」（假說）。
 * 副作用：F 的 +0x4C、+0x48、+0x52、+0x54，+0x64、+0x66（兩條路徑之一），+0x48 的 bit4（來自 2583:052C）；3310:0006 的 ds:8EDC..8EED 暫存。
 *
 * 機器模型：C 版沒有堆疊。基底 fcmd-base.state（被呼叫的 39B5 C 版寫死進入點 SP ＝ 1F8A 的堆疊位址，比對時雙方都排除）。
 * 輸入域：F、G 指向艦隊表（描述檔 tables）或 seeds 給定的位址，不在 A0000 到 BFFFF、不指向堆疊帶；byte F[+4D] ≠ 0；word 讀寫在位移 0xFFFF 時 dosgolem 的高位元組取同一段的位移 0（rd16s／wr16s）。
 *   限制：F 的 word 欄位 +0x52、+0x54、+0x5C、+0x5E、+0x64、+0x66 的位移不落在 0xFFFF，因為被呼叫的 2583:052C 的 C 版以線性位址讀這些 word（f_2583_052C.c 用 rd16，不模型化回繞），
 *   026F 的 C 版也按線性位址寫 +0x64／+0x66（f_2BE1_026F.c 檔頭）。這些位置的 word 讀寫在本函式自己是 rd16s／wr16s（照原版），但沒有測試能單獨驗證 F 欄位上的回繞；
 *   G 的 word 欄位（+0x5C、+0x5E、+0x74、+0x76）只被本函式讀，位移 0xFFFF 的回繞由 seeds（8000:FFA3、8000:FF89）驗證。F 的位移加欄位位移本身可以回繞（seeds 的 8000:FF9A，+0x66 → 0000）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_024A(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off,
                            uint16_t oy_seg);
extern void f_2583_052C(Mem *m, uint16_t off, uint16_t seg);

static uint16_t w0CB7_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0CB7_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* 無號 16 位元除法 DX:AX ÷ CX（DX ＝ 0）。除數 0 在原版是 #DE，不在輸入域，這裡回 0 只為了不讓 C 版當掉。 */
static uint16_t w0CB7_div(uint16_t a, uint8_t b) { return b == 0 ? 0 : (uint16_t)(a / b); }

void f_307D_0CB7(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t g_off, uint16_t g_seg)
{
    uint16_t best = w0CB7_rd16s(m, GIN3_DS, 0xD6E6u);
    uint16_t ent = 0xD6E6u;
    uint16_t scan = 0xD6BCu;
    uint8_t i;
    uint16_t rem, q1, q2, e2, e4, ang_word;
    uint8_t ang1, ang2, ang3, r1, r2;
    uint32_t d, q, r;

#define F16(k) w0CB7_rd16s(m, f_seg, (uint16_t)(f_off + (k)))
#define F8(k) rd8(m, LIN(f_seg, (uint16_t)(f_off + (k))))
#define G16(k) w0CB7_rd16s(m, g_seg, (uint16_t)(g_off + (k)))
#define E16(k) w0CB7_rd16s(m, GIN3_DS, (uint16_t)(ent + (k)))

    for (i = 0; i < 8; i++) {
        uint16_t w = w0CB7_rd16s(m, GIN3_DS, scan);

        if (!(w < best)) {
            best = w;
            ent = scan;
        }
        scan = (uint16_t)(scan + 6u);
    }
    rem = F8(0x4Fu);
    e4 = E16(4u);
    e2 = E16(2u);
    d = f_3310_0006(m, F16(0x5Cu), F16(0x5Eu), e2, e4);
    gin3_ldivmod(d, 0x80000u, &q, &r);
    rem = (uint16_t)(rem - (uint16_t)q);
    ang1 = f_2BE1_011D(m, (uint16_t)(E16(2u) - F16(0x5Cu)), (uint16_t)(E16(4u) - F16(0x5Eu)));
    r1 = (uint8_t)f_2BE1_024A(m, ang1, F16(0x4Au));
    q1 = w0CB7_div(r1, F8(0x4Du));
    rem = (uint16_t)(rem - q1);
    ang2 = f_2BE1_011D(m, (uint16_t)(G16(0x5Cu) - E16(2u)), (uint16_t)(G16(0x5Eu) - E16(4u)));
    ang_word = (uint16_t)(ang1 | ((uint16_t)r1 << 8));
    r2 = (uint8_t)f_2BE1_024A(m, ang_word, ang2);
    q2 = w0CB7_div(r2, F8(0x4Du));
    rem = (uint16_t)(rem - q2);
    if ((int16_t)rem > 0) {
        w0CB7_wr16s(m, f_seg, (uint16_t)(f_off + 0x64u), E16(2u));
        w0CB7_wr16s(m, f_seg, (uint16_t)(f_off + 0x66u), E16(4u));
    } else {
        uint16_t s = (uint16_t)(q1 + q2);
        uint16_t c = F8(0x4Fu);
        uint16_t dist;

        if (!(s > c)) {
            dist = (uint16_t)((uint16_t)(c - s) << 3);
        } else {
            gin3_ldivmod((uint32_t)F8(0x4Fu) << 3, 3u, &q, &r);
            dist = (uint16_t)q;
        }
        (void)f_2BE1_026F(m, F16(0x5Cu), F16(0x5Eu), ang_word, dist, (uint16_t)(f_off + 0x64u), f_seg, (uint16_t)(f_off + 0x66u), f_seg);
    }
    ang3 = f_2BE1_011D(m, (uint16_t)(G16(0x74u) - F16(0x64u)), (uint16_t)(G16(0x76u) - F16(0x66u)));
    wr8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Cu)), ang3);
    wr8(m, LIN(f_seg, (uint16_t)(f_off + 0x48u)), 1);
    {
        uint32_t v = (uint32_t)F8(0x4Fu) << 19;

        w0CB7_wr16s(m, f_seg, (uint16_t)(f_off + 0x54u), (uint16_t)(v >> 16));
        w0CB7_wr16s(m, f_seg, (uint16_t)(f_off + 0x52u), (uint16_t)v);
    }
    f_2583_052C(m, f_off, f_seg);
#undef F16
#undef F8
#undef G16
#undef E16
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_307D_0CB7(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("307D_0CB7", 4, run)
