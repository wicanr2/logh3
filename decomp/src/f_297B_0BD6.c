/*
 * GIN3PS.EXE 297B:0BD6  戰場像素落點排擠：艦隊 F 目前的落點 (F[+5C], F[+5E]) 不能站時，繞著原落點以 32 像素的倍數找一個能站的點，改寫 F 的落點
 *
 * 原版（158 bytes，far 呼叫，一個 far 指標參數 F；序言 push bp／si／di、sub sp,0Ah、mov bp,sp，所以 F 的位移在 [bp+14h]、段在 [bp+16h]）。
 *   區域變數：[bp+0] ＝ k（只存不讀）、[bp+2] ＝ n（word 計數）、[bp+4]／[bp+6] ＝ 候選 x／y、[bp+8] ＝ DS 的暫存。
 *   回傳 AX：第一次檢查就能站時是 297B:0C74 的回傳（1）；否則是候選 y（寫進 F[+5E] 的值）。呼叫者 297B:0ABF 不用它（描述檔 ret ＝ ax 逐位元比對）。
 * 過程（已證實：反組譯逐指令；297B:0C74 由其差異測試證實）：
 *   r ← 297B:0C74(F, F[+5C], F[+5E])                    ; 先讀 y 再讀 x（沒有寫入，順序無關）
 *   r ≠ 0 → 回 r（F 目前的落點可以站）
 *   n ← 0；重複：
 *     idx ← n & 7；k ← ((int16)n 算術右移 3) ＋ 1                              ; 16 位元
 *     (dx, dy) ← 兩個 word，位址 ds:4E0E ＋ idx × 4（每筆 4 bytes）
 *     x' ← low16(dx × k) ＋ F[+5C]；y' ← low16(dy × k) ＋ F[+5E]                 ; `imul` 只留低字，F 的欄位在乘法之後才讀
 *     n ← n ＋ 1
 *     0C74(F, x', y') ＝ 0 → 繼續（n 沒有上限，是 word，65536 圈後回繞；原版沒有終止保證）
 *     否則 F[+5C] ← x'；F[+5E] ← y'；回 AX ＝ y'
 *   表 ds:4E0E 的基底值是八個方向 (0,−32)(32,−32)(32,0)(32,32)(0,32)(−32,32)(−32,0)(−32,−32)，所以第 k 圈（n 從 8(k−1) 到 8k−1）的候選是原落點加 32k 的八方向。
 * 語意（強推論）：docs/re/335-ai-order-command-kind.md 第 81 行（`297B:0BD6(f)：0C74(f, f[+5C], f[+5E]) != 0 → 回；否則 k = 0,1,…`）、
 *   docs/re/330-strategic-planet-assault.md 第 271 行與 docs/spec/058-strategic-planet-assault.md 第 23 行（`ds:4E0E` 八方向 × 32 的倍數）、
 *   docs/internals/53-strategic-planet-assault.md 第 48 行，與本檔一致。同一星域兩支以上攻方時，後定位的那支被擠開（docs/re/330 第 180 行，未有收據）。
 * 呼叫的函式：297B:0C74（已 diff-tested，其內部只讀記憶體、不碰硬體）。
 * 副作用：只在找到時寫 F[+5C]、F[+5E] 兩個 word。
 *
 * 輸入域：基底 fcmd-standby.state（0C74 的 1588:04EA 需要戰場地圖）；F 指到艦隊表（ds:C074 起每筆 0x7A）。
 *   原版迴圈沒有上界，找不到能站的點時不會結束（落點全被擋或全出界）；探針的步數上限（描述檔 step_budget）會丟棄這類案例。
 *   C 版為了讓突變版本停得下來，迴圈在 2048 圈後放棄並回 0xFFFF（原版在步數上限內到不了這麼多圈；每圈約數百到上千步）。
 *   n 的算術右移只在 n >= 0x8000 時與邏輯右移不同，同樣在步數上限內到不了，不涵蓋。
 *   F 欄位的 word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem 的 read16），用本檔的 rd16s；F 指在 DGROUP，寫入不會落在 ROM 區。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_297B_0C74(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t x, uint16_t y);

static uint16_t w0BD6_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0BD6_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_297B_0BD6(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint16_t r, n, nx, ny;
    unsigned guard;

    r = f_297B_0C74(m, f_off, f_seg, w0BD6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)), w0BD6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)));
    if (r != 0)
        return r;
    n = 0;
    for (guard = 0; guard < 2048u; guard++) {
        uint16_t idx = (uint16_t)(n & 7u);
        uint16_t k = (uint16_t)((uint16_t)(((int32_t)(int16_t)n) >> 3) + 1u);
        uint16_t dx = rd16(m, DS((uint16_t)(0x4E0Eu + idx * 4u)));
        uint16_t dy = rd16(m, DS((uint16_t)(0x4E0Eu + idx * 4u + 2u)));

        nx = (uint16_t)((uint16_t)((int32_t)(int16_t)dx * (int32_t)(int16_t)k) + w0BD6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)));
        ny = (uint16_t)((uint16_t)((int32_t)(int16_t)dy * (int32_t)(int16_t)k) + w0BD6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)));
        n = (uint16_t)(n + 1u);
        if (f_297B_0C74(m, f_off, f_seg, nx, ny) != 0) {
            w0BD6_wr16s(m, f_seg, (uint16_t)(f_off + 0x5Cu), nx);
            w0BD6_wr16s(m, f_seg, (uint16_t)(f_off + 0x5Eu), ny);
            return ny;
        }
    }
    return 0xFFFFu;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_0BD6(m, a[0], a[1]); }
GIN3_REGISTER("297B_0BD6", 2, run)
