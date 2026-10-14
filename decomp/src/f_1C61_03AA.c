/*
 * GIN3PS.EXE 1C61:03AA  戰場艦艇 S：還有目的地且「累積量」S[+24..+27]（16.16 有號）達 8.0 才呼叫 1C61:0444(S, 8) 走一步，否則把狀態 S[+1E] 設 3、累積量清 0
 *
 * 原版（154 bytes，far 呼叫，一個遠指標堆疊參數 S；序言 push bp／si、sub sp,2、mov bp,sp，所以 S 的位移在 [bp+0Ah]、段在 [bp+0Ch]）。
 *   區域變數：[bp+0] ＝ 回傳值 r（byte，初值 0）。回傳在 AL；AH 是殘值（重設路徑 xor ax,ax 回 0，其他路徑 mov al,[bp+0]），所以描述檔 ret ＝ pack:al。
 * 過程（已證實：反組譯逐指令；1C61:0444、3CC9:00B9 各由其差異測試證實）：
 *   有目的地 ← (word S[+34] ≠ word S[+28]) 或 (word S[+36] ≠ word S[+2A])
 *   沒有目的地：重設
 *   有目的地：
 *     byte S[+4] ≠ 7 且 dword S[+24]（低字 +24、高字 +26）< 8.0（3CC9:00B9 比 (S[+26], S[+24]) 與 (8, 0) 後 jge 不成立，有號 32 位元）：重設
 *     byte S[+4] ＝ 7 或 dword S[+24] >= 8.0：
 *       dword S[+24] < 8.0（jl，只可能發生在 S[+4] ＝ 7）：回傳 0，不重設
 *       否則：r ← 1；1C61:0444(S, 8)（堆疊字 8 的高位元組是殘值，0444 只用低位元組）；回傳 r（1）
 *   重設：byte S[+1E] ← 3；word S[+24] ← 0；word S[+26] ← 0；回傳 0
 * 語意：docs/re/127-squadron-step-and-slot-assignment.md 第 24 到 38 行（「下了移動命令的分隊怎麼走」）、docs/re/373-battle-interleaved-movement.md 第 32 行已記載本函式；呼叫者 1C61:0153。
 *   S[+24..+27] 在 1C61:0444 每走一步減 8.0，這裡以 8.0 為門檻，推測是「移動點數／行動力」之類的累積量（未知，名稱不是證據）；S[+4] ＝ 7 是索敵分遣隊（docs/re/127 同處、docs/re/372、f_2583_040C.c）。
 *   與 docs/re/127 第 26 到 33 行的差異（asm 與差異測試為準）：(1) 該文寫「已經到了 → 收工」，asm 是重設（S[+1E] ← 3、S[+24]／S[+26] ← 0、回 0）；
 *   (2) 該文寫 S[+4] ＝ 7 時「跳過」< 8.0 的檢查後直接走一步，asm 在 S[+4] ＝ 7 且累積量 < 8.0 時既不重設也不呼叫 0444，直接回 0（loc_2958C 再比一次）。
 * 副作用：重設路徑寫 S[+1E]、S[+24]、S[+26]；呼叫路徑的寫入全部來自 1C61:0444。
 *
 * 機器模型：C 版沒有堆疊。0444 C 版把落點槽 X、Y 放在 fcmd-standby.state 進入點 SP ＝ 1E76 的框架（SS:1E46／1E48），原版從本函式呼叫時框架更深，
 *   位置不同，但兩處都在探針的堆疊帶內，比對時排除，C 版自寫自讀，不影響結果。基底狀態與 1C61:0444 相同。
 * 輸入域：S 取艦艇池記錄（描述檔 table ship，同 0444）；S 不指向堆疊帶與 A0000 到 BFFFF。S 的位移加欄位時 16 位元回繞（word 讀取用 rd16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern void f_1C61_0444(Mem *m, uint16_t s_off, uint16_t s_seg, uint16_t step_word);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

static uint16_t w03AA_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w03AA_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

#define F03AA_SF 0x0080u
#define F03AA_OF 0x0800u

/* dword S[+24] 與 8.0 比較後的 SF ≠ OF（jl 成立，有號 32 位元 小於） */
static int w03AA_below8(const Mem *m, uint16_t s_off, uint16_t s_seg)
{
    uint16_t fl = f_3CC9_00B9(w03AA_rd16s(m, s_seg, (uint16_t)(s_off + 0x26u)), w03AA_rd16s(m, s_seg, (uint16_t)(s_off + 0x24u)), 8u, 0u);

    return ((fl & F03AA_SF) != 0) != ((fl & F03AA_OF) != 0);
}

uint8_t f_1C61_03AA(Mem *m, uint16_t s_off, uint16_t s_seg)
{
    int reset = 0;

    if (w03AA_rd16s(m, s_seg, (uint16_t)(s_off + 0x34u)) == w03AA_rd16s(m, s_seg, (uint16_t)(s_off + 0x28u))
        && w03AA_rd16s(m, s_seg, (uint16_t)(s_off + 0x36u)) == w03AA_rd16s(m, s_seg, (uint16_t)(s_off + 0x2Au))) {
        reset = 1;
    } else if (rd8(m, LIN(s_seg, (uint16_t)(s_off + 4u))) != 7 && w03AA_below8(m, s_off, s_seg)) {
        reset = 1;
    }
    if (reset) {
        wr8(m, LIN(s_seg, (uint16_t)(s_off + 0x1Eu)), 3);
        w03AA_wr16s(m, s_seg, (uint16_t)(s_off + 0x24u), 0);
        w03AA_wr16s(m, s_seg, (uint16_t)(s_off + 0x26u), 0);
        return 0;
    }
    if (w03AA_below8(m, s_off, s_seg))
        return 0;
    f_1C61_0444(m, s_off, s_seg, 8u);
    return 1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1C61_03AA(m, a[0], a[1]); }
GIN3_REGISTER("1C61_03AA", 2, run)
