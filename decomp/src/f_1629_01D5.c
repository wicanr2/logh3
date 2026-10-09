/*
 * GIN3PS.EXE 1629:01D5  畫面座標 (x, y) 是否落在固定視窗 ds:9BF8 的內容格 1 或 2（回 1 或 0）
 *
 * 原版：far 呼叫，兩個 word 參數 x（[bp+0Eh]）、y（[bp+10h]）（序言 push bp／si／di、sub sp,4、mov bp,sp，所以參數在 [bp+0Eh]／[bp+10h]），
 *   回傳 AX ＝ 1 或 0（區域變數 [bp+2] 的 word，一開始寫 0）。逐指令：
 *   [bp+2] ← 0
 *   AX:BX ← 341B:039D(x, y)                           ; 游標下的視窗節點（seg:off），沒有命中為 0:0
 *   若 AX ≠ DS 或 BX ≠ 0x9BF8 → 回 0                  ; 命中的不是 DGROUP 內的固定視窗 ds:9BF8（段與位移分別比較，不比線性位址）
 *   341B:044A(win ＝ DS:9BF8, x, y, out_x ＝ SS:BP+0, out_y ＝ SS:BP+1)   ; 兩個輸出是區域變數 [bp+0]、[bp+1]（byte）
 *   al ← [bp+0]（即 out_x）；若 1 <= al <= 2（無號 byte 比較）→ [bp+2] ← 1
 *   回 [bp+2]
 *   只看 out_x，不看 out_y（out_y 寫了但未使用）。
 * 呼叫者（台帳）：1930:0815、1930:0F9C、1AAB:07C6、1AAB:0C8C（皆 hw-dependent）。ds:9BF8 在基底狀態全是 0（未使用的視窗槽）；
 *   用途判斷（假說）：9BF8 是某個選單／提示視窗的固定槽，out_x 為 1 或 2 表示游標在視窗內第 1 或第 2 欄項目上（044A 的有框扣除後）。
 *   名稱與用途：未知。
 *
 * 機器模型：C 版沒有堆疊。兩個輸出區域變數放在 SS:1F78／1F79，這是原版的實際位置（基底狀態 SP ＝ 1F8A，減返回位址 4、兩個參數 4、
 *   push bp／si／di 6、sub sp,4 → bp ＝ 1F78），SS 固定是基底狀態的 0x63EB。這個位址在探針的堆疊帶（SS:1B8A..1F89）內，
 *   寫入集合比對時雙方都排除，與原版把輸出寫進堆疊一致。
 * 輸入域：x、y 全部 16 位元值；視窗串列要是良構（會終止）。ds:9BF8 的視窗結構、ds:D804／D806 串列頭由探針突變，不涵蓋指向堆疊帶內的案例。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint32_t f_341B_039D(Mem *m, uint16_t x, uint16_t y);
extern void f_341B_044A(Mem *m, uint16_t woff, uint16_t wseg, uint16_t x, uint16_t y,
                        uint16_t xo, uint16_t xs, uint16_t yo, uint16_t ys);

#define F1629_01D5_SS 0x63EBu     /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F1629_01D5_BP 0x1F78u     /* 原版 mov bp,sp 之後的 BP（區域變數 [bp+0]、[bp+1]） */

uint16_t f_1629_01D5(Mem *m, uint16_t x, uint16_t y)
{
    uint32_t hit = f_341B_039D(m, x, y); /* AX:BX ＝ 段:位移 */
    uint16_t result = 0;

    if ((uint16_t)(hit >> 16) != GIN3_DS || (uint16_t)hit != 0x9BF8u)
        return 0;
    f_341B_044A(m, 0x9BF8u, GIN3_DS, x, y,
                F1629_01D5_BP, F1629_01D5_SS, (uint16_t)(F1629_01D5_BP + 1u), F1629_01D5_SS);
    {
        uint8_t al = rd8(m, LIN(F1629_01D5_SS, F1629_01D5_BP));

        if (al >= 1u && al <= 2u)
            result = 1;
    }
    return result;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1629_01D5(m, a[0], a[1]); }
GIN3_REGISTER("1629_01D5", 2, run)
