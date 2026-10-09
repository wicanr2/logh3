/*
 * GIN3PS.EXE 307D:0F05  在場艦隊裡，列出「對方陣營、且至少有一艘艦艇 [+0x17] 不為 0」的艦隊（存進 ds:D61C 起的清單）
 *
 * 原版：far 呼叫，一個參數＝記錄 P 的 far 指標（push bp／si／di、sub sp,1Ah，所以參數在 [bp+24h]），回 AX（0 或 1）。
 * 過程（已證實：反組譯逐指令；3982:0250 由其自己的差異測試證實）：
 *   side ← 3982:0250(word P[+1])        ; P[+1] == 0（低位元組）→ 1，否則 0（「對方陣營」，docs/re/93）
 *   ds:D61B ← 0（byte，清個數）；di ← 0
 *   j ＝ 0..（byte，j < byte ds:9894 才進迴圈；每圈重讀 ds:9894），清單指標 L ← ds:9898（byte 清單，每圈 +1）：
 *     F ← ds:(0xC074 + 0x7A × byte [L])                       ; 艦隊（16 位元算術）
 *     F[+1] == side（byte）：
 *       flag ← 0；k ＝ 0..（byte，k < byte F[+0x0F] 才進迴圈；每圈重讀 F[+0x0F]），槽指標 S ← F + 0x10（每圈 +4）：
 *         艦艇 ← (ds:CFB6 段):(ds:CFB4 + 16 位元(word [S] × 0x4E))     ; 每圈重讀池指標
 *         艦艇[+0x17] != 0 → flag ← 1
 *       flag != 0：
 *         ds:(D61C + 4 × 個數) 的 word +2 ← DS、word +0 ← F 的位移（先寫段、後寫位移）；個數指標 +4
 *         ds:D61B（byte）++；di ← 1
 *   回 di
 * 語意（強推論，名稱不是證據）：ds:9894／9898 是「在場艦隊清單」（docs/internals/38 第 283 行、docs/re/145：筆數 byte、清單 byte），
 *   艦艇 +0x17 是索敵等級（docs/re/346、internals/67 的 0374:0AF2 一列）；本函式列出對方陣營「已被偵察到」的艦隊。ds:D61C 起的
 *   清單與 ds:D61B 個數是輸出。基底狀態在場艦隊清單是空的（3009:008D 的紀錄），由 seeds 造。
 * 副作用：寫 ds:D61B（先清零再遞增）與 ds:D61C 起每筆 4 bytes。
 * 輸入域：P 取艦隊表；清單筆數最多 255，輸出最多 255 × 4 = 1020 bytes；艦隊編號任意（16 位元回繞）；艦艇池指標指到
 *   A0000 到 BFFFF 的案例由探針丟棄。word 存取在位移 0xFFFF 時高位元組取同一段位移 0（槽位移恆為偶數，實際遇不到）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3982_0250(uint8_t faction);

static uint16_t w0F05_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0F05_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_307D_0F05(Mem *m, uint16_t p_off, uint16_t p_seg)
{
    uint8_t side = (uint8_t)f_3982_0250((uint8_t)w0F05_rd16s(m, p_seg, (uint16_t)(p_off + 1u)));
    uint16_t list = 0x9898u;
    uint16_t out = 0xD61Cu;
    uint16_t di = 0;
    uint8_t j = 0;

    wr8(m, DS(0xD61Bu), 0);
    while (j < rd8(m, DS(0x9894u))) {
        uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(rd8(m, DS(list)) * 0x7Au));

        if (rd8(m, DS((uint16_t)(fleet + 1u))) == side) {
            uint16_t flag = 0;
            uint16_t slot = (uint16_t)(fleet + 0x10u);
            uint8_t k = 0;

            while (k < rd8(m, DS((uint16_t)(fleet + 0x0Fu)))) {
                uint16_t pseg = rd16(m, DS(0xCFB6u));
                uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4u)) + (uint16_t)(w0F05_rd16s(m, GIN3_DS, slot) * 0x4Eu));

                if (rd8(m, LIN(pseg, (uint16_t)(ship + 0x17u))) != 0)
                    flag = 1;
                k++;
                slot = (uint16_t)(slot + 4u);
            }
            if (flag != 0) {
                w0F05_wr16s(m, GIN3_DS, (uint16_t)(out + 2u), GIN3_DS);
                w0F05_wr16s(m, GIN3_DS, out, fleet);
                out = (uint16_t)(out + 4u);
                wr8(m, DS(0xD61Bu), (uint8_t)(rd8(m, DS(0xD61Bu)) + 1u));
                di = 1;
            }
        }
        j++;
        list = (uint16_t)(list + 1u);
    }
    return di;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_307D_0F05(m, a[0], a[1]); }
GIN3_REGISTER("307D_0F05", 2, run)
