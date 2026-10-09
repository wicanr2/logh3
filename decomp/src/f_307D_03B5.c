/*
 * GIN3PS.EXE 307D:03B5  戰場上一支艦隊的「編制完整度」是否低於門檻；結果寫進 ds:D61A（0 或 1）
 *
 * 原版：far 呼叫，序言 push bp／si／di、sub sp,0Eh、mov bp,sp。一個堆疊參數 fleet（far 指標，位移 [bp+18h]／段 [bp+1Ah]）。
 *   區域變數：[bp+0] byte ＝ 組序號 i；[bp+1] byte ＝ 完整度百分比；[bp+2] word ＝ Σ 艦艇[+0x0A]；[bp+4] word ＝ Σ 艦艇[+0x08]；
 *   [bp+6] ＝ 指向目前組的指標（fleet ＋ 0x10 起，每組 +4）。
 *   回傳 AX ＝ 結果（0 或 1，與寫進 ds:D61A 的值相同；呼叫者 307D:001B 之後讀 ds:D61A，是否用 AX 未驗證）。
 *
 * 資料（docs/re/287 §新解、docs/internals/08 第 120 行、docs/internals/33 第 175 行；本檔與反組譯逐條對過）：
 *   艦隊記錄 +0x0F byte ＝ 編制組數；+0x10 起每組 4 bytes，第一個 word ＝ 艦艇池索引；+0x44 byte ＝ 門檻（提督＋副官逐項取大的 +20）。
 *   艦艇池基底 ds:CFB4（位移）／ds:CFB6（段），一筆 78（0x4E）bytes；記錄 +0x08 word ＝ 現有艘數、+0x0A word ＝ 滿編艘數
 *   （欄位名稱採 docs/re/287；只有它們被加總與相除是已證實的）。
 *
 * 語意（已證實）：
 *   for (i ＝ 0; i < byte fleet[+0x0F]; i++):                        ; byte 無號比較，每次迭代重讀 +0x0F
 *       rec ＝ (ds:CFB6) : ((ds:CFB4) ＋ (idx_i × 0x4E 的低 16 位元))
 *       full ＋＝ word rec[+0x0A]；cur ＋＝ word rec[+0x08]            ; 各自 16 位元回繞
 *   pct ＝ 低位元組( (cur × 100) ÷ full )                              ; 32 位元無號乘除（3CC9:0009、3CC9:0020），full ＝ 0 時商 0xFFFFFFFF
 *   thr ＝ (byte fleet[+0x44] × 50) ÷ 100                              ; 16 位元，等於 +0x44 ÷ 2（無條件捨去）
 *   ds:D61A ← (thr > pct) ? 1 : 0                                      ; 無號 word 比較；thr 是 0..127，pct 已零延伸成 word
 * 與 docs 的出入：docs/re/287 寫「完整度 ＝ 現有 × 100 ÷ 滿編（百分比）」。實際 pct 只留商的低位元組：現有 > 滿編 × 2.55
 *   （商 >= 256）時回繞，滿編 ＝ 0 時 pct ＝ 0xFF（thr 最大 127，故結果必為 0）。
 *
 * 輸入域：組數最多 255；fleet 指向 A0000 到 BFFFF 的案例被探針丟棄。加總與 pct 的邊界由描述檔 seeds 造出。
 * 結果寫入只有 ds:D61A 一個 byte；基底狀態該值為 0，描述檔 seeds 另造 ds:D61A ＝ 1 的前態以驗證「寫 0」。
 */
#include "gin3_reg.h"

/* 段內 word 讀取：位移各自在段內回繞（dosgolem cpu.go read16），不是線性位址加 1。 */
static uint16_t f307D03B5_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_307D_03B5(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i = 0;
    uint16_t sum_full = 0; /* [bp+2]：Σ 艦艇[+0x0A] */
    uint16_t sum_cur = 0;  /* [bp+4]：Σ 艦艇[+0x08] */
    uint16_t p = (uint16_t)(fleet_off + 0x10u);
    uint32_t prod, quot;
    uint8_t pct;
    uint16_t thr, result;

    for (;;) {
        uint8_t cnt = rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0Fu)));
        uint16_t idx, rec, pool_seg;

        if (i >= cnt)
            break;
        idx = f307D03B5_rd16s(m, fleet_seg, p);
        rec = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));
        pool_seg = rd16(m, DS(0xCFB6));
        sum_full = (uint16_t)(sum_full + f307D03B5_rd16s(m, pool_seg, (uint16_t)(rec + 0x0Au)));
        sum_cur = (uint16_t)(sum_cur + f307D03B5_rd16s(m, pool_seg, (uint16_t)(rec + 0x08u)));
        i = (uint8_t)(i + 1u);
        p = (uint16_t)(p + 4u);
    }

    prod = gin3_lmul(sum_cur, 100u);   /* 3CC9:0009：(0:cur) × (0:100) */
    quot = gin3_uldiv(prod, sum_full); /* 3CC9:0020：除以 (0:full)，full ＝ 0 時 0xFFFFFFFF */
    pct = (uint8_t)quot;
    thr = (uint16_t)(((uint16_t)rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x44u))) * 50u) / 100u);
    result = (thr > pct) ? 1u : 0u;
    wr8(m, DS(0xD61A), (uint8_t)result);
    return result;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_307D_03B5(m, a[0], a[1]); }
GIN3_REGISTER("307D_03B5", 2, run)
