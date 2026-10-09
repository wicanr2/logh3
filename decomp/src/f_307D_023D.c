/*
 * GIN3PS.EXE 307D:023D  以艦隊 P 為中心，300 以內的敵方艦隊依方位（8 向）把「各分隊艘數合計」累加進 ds:D69C 起的 8 個 dword，並從對向扣掉合計的一半；回 1（有算）或 0
 *
 * 原版（376 bytes，far 呼叫，一個 far 指標參數 P；序言 push bp／si／di、sub sp,18h、mov bp,sp，所以 P 的位移在 [bp+22h]、段在 [bp+24h]）：
 *   回傳 AX ＝ [bp+0]（word）＝ 有沒有任何一支艦隊通過距離檢查（1 或 0），描述檔 ret ＝ ax。
 *   區域變數：[bp+0] ＝ 結果、[bp+2] ＝ i（byte）、[bp+3] ＝ 方位 sel（byte）、[bp+4]／[bp+6] ＝ 名冊指標（位移／段，段 ＝ DS）、[bp+8] ＝ j、
 *   [bp+0Ah] ＝ 合計 S（word）、[bp+0Ch]／[bp+0Eh] ＝ 艦隊記錄指標 F、[bp+10h]／[bp+12h] ＝ 分隊槽指標、[bp+14h]／[bp+16h] ＝ 艦艇記錄指標（只存不讀）。
 * 過程（已證實：反組譯逐指令；3C9C:000A、3310:0006、3CC9:00B9、2BE1:011D、3982:0022 各由其差異測試證實）：
 *   結果 ← 0；3C9C:000A(ds:D69C, 0, 0x20)                    ; 32 bytes 清 0（8 個 dword）
 *   i ← 0；名冊指標 ← DS:0x9898
 *   重複：i >= byte ds:9894（無號 jb 的反面，每圈重讀）→ 結束
 *     idx ← byte 名冊[0]；F ← 0xC074 ＋ low16(idx × 0x7A)（DS 內）
 *     byte F[+1] ＝ byte P[+1]（同陣營）→ 跳過
 *     byte F[0] 的 bit0 有設 → 跳過
 *     d ← 3310:0006(P[+0x5C], P[+0x5E], F[+0x5C], F[+0x5E])    ; 16.16 距離（push 順序：F[+5E]、F[+5C]、P[+5E]、P[+5C]；兩點對稱）
 *     3CC9:00B9(d, 0x012C:0000) 之後 jle（ZF ＝ 1 或 SF ≠ OF，有號 32 位元 d <= 300.0）不成立 → 跳過
 *     結果 ← 1
 *     dx ← F[+0x5C] − P[+0x5C]、dy ← F[+0x5E] − P[+0x5E]（16 位元，這時才重新讀）
 *     sel ← 3982:0022(低位元組 of 2BE1:011D(dx, dy))           ; 8 向索引 0 到 7
 *     j ← 0；槽 ← DS:(F 位移 ＋ 0x10)
 *     j < byte F[+0x0F]（無號，每圈重讀）時：
 *       s ← word 槽[0]；艦艇 ← (word ds:CFB4 ＋ low16(s × 0x4E)) : word ds:CFB6    ; 池基底每圈重讀（先讀段、後讀位移）
 *       S ← S ＋ word 艦艇[+8]（16 位元）；j ＋= 1；槽 ＋= 4
 *     dword ds:(D69C ＋ 4 × sel) ＋= S（低字 add、高字 adc 0）
 *     r ← (sel ＋ 4) mod 8；dword ds:(D69C ＋ 4 × r) −= (S >> 1)（低字 sub、高字 sbb 0；S >> 1 是邏輯右移）
 *     i ＋= 1；名冊指標 ＋= 1
 *   回 結果
 * 原版的細節（本版照做）：
 *   - S 沒有初始化：[bp+0Ah] 在函式內只被 `add [bp+0Ah], ax` 寫過，初值是進入時堆疊上的殘值，而且 S 在各支艦隊之間不歸零，會一路累加。
 *     初值是進入點框架的這個位置：fcmd-base.state 的 SS ＝ 0x63EB、進入點 SP ＝ 0x1F8A；探針先推入 P 的 4 bytes 與 4 bytes 的返回位址
 *     （SP ＝ 0x1F82），再 push bp／si／di（0x1F7C）、sub sp,18h（bp ＝ 0x1F64），所以 [bp+0Ah] ＝ SS:0x1F6E（線性 0x65E1E），位於探針堆疊帶內。
 *     C 版從該位址讀初值（與 f_2D76_011C.c 把清單放在堆疊位址同一種做法）。基底狀態一換，這兩個常數要重算。
 *   - P、F 的座標在呼叫 3310:0006 前後各讀一次（第二次用於方位的減法），中間 39B5:005C 會寫 ds:8EDC..8EED，P 指到那裡時兩次讀到的值不同。
 *   - 方位 sel 只可能是 0 到 7（3982:0022 的 and al,7），所以對向的索引不會超出 8 個 dword。
 * 語意（假說）：docs 沒有 307D:023D 的記載（grep 無命中；呼叫者 307D:00CD）。+0x5C／+0x5E 是艦隊座標（docs/re/129 等），F[+0x10 起每 4 bytes] 的 word 是艦艇池索引、
 *   F[+0x0F] 是分隊數（docs/internals/33、本專案 25DB:0221），艦艇 +8 是艘數（docs/re/330 的 25DB 系列）。整體看起來是「各方位的敵方兵力威脅」的累計，
 *   對向扣一半；為什麼 S 跨艦隊累加、初值不歸零，未知（可能是原版 bug）。
 * 副作用：ds:D69C 起 32 bytes（清 0 與累加）、39B5:005C 在 ds:8EDC..8EED 的暫存。
 *
 * 機器模型：C 版沒有堆疊；S 的初值從 SS:0x1F6E 讀（見上）。word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 read16／write16），用本檔的 w023D_rd16s／w023D_wr16s。
 * 輸入域：基底狀態 fcmd-base.state；艦隊名冊 ds:9894／ds:9898 起、艦隊記錄、艦艇池取 seeds 與突變。P 指向艦隊表或 seeds 給的位址，不指向堆疊帶與 A0000 到 BFFFF。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_3982_0022(uint8_t angle);

#define F023D_SS 0x63EBu    /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F023D_SUM 0x1F6Eu   /* [bp+0Ah]：進入點 SP 1F8A − 參數 4 − 返回位址 4 − push bp／si／di 6 − sub sp,18h ＝ bp(1F64)，再加 0Ah */
#define F023D_ZF 0x0040u
#define F023D_SF 0x0080u
#define F023D_OF 0x0800u

static uint16_t w023D_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w023D_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_307D_023D(Mem *m, uint16_t p_off, uint16_t p_seg)
{
    uint16_t result = 0;
    uint8_t i = 0;
    uint16_t lp = 0x9898u;
    uint16_t sum = w023D_rd16s(m, F023D_SS, F023D_SUM); /* 未初始化的區域變數 [bp+0Ah]：進入時堆疊上的殘值 */

    f_3C9C_000A(m, 0xD69Cu, GIN3_DS, 0, 0x20);
    while (i < rd8(m, DS(0x9894u))) {
        uint8_t idx = rd8(m, DS(lp));
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(idx * 0x7Au));

        if (rd8(m, DS((uint16_t)(f + 1u))) != rd8(m, LIN(p_seg, (uint16_t)(p_off + 1u)))
            && !(rd8(m, DS(f)) & 1u)) {
            uint16_t fx = w023D_rd16s(m, GIN3_DS, (uint16_t)(f + 0x5Cu));
            uint16_t fy = w023D_rd16s(m, GIN3_DS, (uint16_t)(f + 0x5Eu));
            uint16_t px = w023D_rd16s(m, p_seg, (uint16_t)(p_off + 0x5Cu));
            uint16_t py = w023D_rd16s(m, p_seg, (uint16_t)(p_off + 0x5Eu));
            uint32_t d = f_3310_0006(m, px, py, fx, fy);
            uint16_t fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 0x012Cu, 0);

            if ((fl & F023D_ZF) || (((fl & F023D_SF) != 0) != ((fl & F023D_OF) != 0))) { /* jle */
                uint16_t dy, dx, slot, j = 0, si, lo, hi, carry;
                uint8_t sel;

                result = 1;
                dy = (uint16_t)(w023D_rd16s(m, GIN3_DS, (uint16_t)(f + 0x5Eu)) - w023D_rd16s(m, p_seg, (uint16_t)(p_off + 0x5Eu)));
                dx = (uint16_t)(w023D_rd16s(m, GIN3_DS, (uint16_t)(f + 0x5Cu)) - w023D_rd16s(m, p_seg, (uint16_t)(p_off + 0x5Cu)));
                sel = (uint8_t)f_3982_0022(f_2BE1_011D(m, dx, dy));
                slot = (uint16_t)(f + 0x10u);
                while (j < rd8(m, DS((uint16_t)(f + 0x0Fu)))) {
                    uint16_t pseg = w023D_rd16s(m, GIN3_DS, 0xCFB6u);
                    uint16_t rec = (uint16_t)(w023D_rd16s(m, GIN3_DS, 0xCFB4u)
                                              + (uint16_t)(w023D_rd16s(m, GIN3_DS, slot) * 0x4Eu));

                    sum = (uint16_t)(sum + w023D_rd16s(m, pseg, (uint16_t)(rec + 8u)));
                    j = (uint16_t)(j + 1u);
                    slot = (uint16_t)(slot + 4u);
                }
                /* dword ds:(D69C + 4 × sel) += S（add／adc 0） */
                si = (uint16_t)(sel << 2);
                lo = w023D_rd16s(m, GIN3_DS, (uint16_t)(0xD69Cu + si));
                carry = (uint16_t)(((uint32_t)lo + sum) >> 16);
                w023D_wr16s(m, GIN3_DS, (uint16_t)(0xD69Cu + si), (uint16_t)(lo + sum));
                hi = w023D_rd16s(m, GIN3_DS, (uint16_t)(0xD69Eu + si));
                w023D_wr16s(m, GIN3_DS, (uint16_t)(0xD69Eu + si), (uint16_t)(hi + carry));
                /* dword ds:(D69C + 4 × ((sel + 4) mod 8)) −= S >> 1（sub／sbb 0） */
                si = (uint16_t)(((sel + 4u) % 8u) << 2);
                lo = w023D_rd16s(m, GIN3_DS, (uint16_t)(0xD69Cu + si));
                carry = (uint16_t)(lo < (uint16_t)(sum >> 1) ? 1u : 0u);
                w023D_wr16s(m, GIN3_DS, (uint16_t)(0xD69Cu + si), (uint16_t)(lo - (uint16_t)(sum >> 1)));
                hi = w023D_rd16s(m, GIN3_DS, (uint16_t)(0xD69Eu + si));
                w023D_wr16s(m, GIN3_DS, (uint16_t)(0xD69Eu + si), (uint16_t)(hi - carry));
            }
        }
        i = (uint8_t)(i + 1u);
        lp = (uint16_t)(lp + 1u);
    }
    return result;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_307D_023D(m, a[0], a[1]);
}
GIN3_REGISTER("307D_023D", 2, run)
