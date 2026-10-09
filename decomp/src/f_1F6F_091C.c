/*
 * GIN3PS.EXE 1F6F:091C  兩艘艦艇之間的直線路徑有沒有被擋：沿 Bresenham 式的線從艦艇 A 的位置 ＋7 走到艦艇 B 的位置 ＋7，每一步問 1F6F:0C11，被擋回 0，走完回 1
 *
 * 原版（494 bytes，far 呼叫，兩個 far 指標參數 A、B；序言 push bp／si／di、sub sp,2D0h、mov bp,sp，所以 A 的位移在 [bp+2DAh]、段在 [bp+2DCh]，B 的位移在 [bp+2DEh]、段在 [bp+2E0h]），
 *   回 AX（0 ＝ 被擋、1 ＝ 路徑通）。堆疊上 SS:bp 起 0x2D0 bytes 同時是清單緩衝區（1F6F:0B0A 的輸出）與尾端的區域變數：
 *   [bp+2C0] ＝ y0、[bp+2C2] ＝ x1、[bp+2C4] ＝ y1、[bp+2C6] ＝ adx、[bp+2C8] ＝ ady、[bp+2CA] ＝ err、[bp+2CC] ＝ 步進、[bp+2CE] ＝ k；di ＝ x0；si ＝ 清單筆數。
 * 過程（已證實：反組譯逐指令；1F6F:0B0A、1F6F:0C11、3C82:000F 各由其差異測試證實）：
 *   n ← 1F6F:0B0A(A, B, far SS:bp)                          ; 兩艦艇座標圍成的矩形內的在場分隊清單（word 艦艇池索引），回筆數
 *   x0 ← A[+34] ＋ 7；y0 ← A[+36] ＋ 7；x1 ← B[+34] ＋ 7；y1 ← B[+36] ＋ 7（16 位元）
 *   adx ← |x1 − x0|；ady ← |y1 − y0|                        ; 3C82:000F（abs；0x8000 取 abs 仍是 0x8000）
 *   (int16) adx > (int16) ady（X 軸為主）：
 *     x0 <= x1（有號）：步進 ← (y0 < y1) ? +1 : −1
 *     否則：步進 ← (y0 > y1) ? +1 : −1；兩端點對調（x0 ← x1、x1 ← 舊 x0、y0 ← y1；y1 不更新，之後用不到）
 *     err ← adx 算術右移 1；k ← 0
 *     重複：x0 ＋＝ 1；x0 > x1（有號）→ 回 1；err −＝ ady；err < 0（有號）→ err ＋＝ adx、y0 ＋＝ 步進；k ＋＝ 1；
 *           1F6F:0C11(x0, y0, k, far SS:bp, n) ≠ 0 → 回 0
 *   否則（Y 軸為主）：
 *     y0 <= y1（有號）：步進 ← (x0 < x1) ? +1 : −1
 *     否則：步進 ← (x0 > x1) ? +1 : −1；y0 ← y1、y1 ← 舊 y0、x0 ← x1（x1 不更新）
 *     err ← ady 算術右移 1；k ← 0
 *     重複：y0 ＋＝ 1；y0 > y1（有號）→ 回 1；err −＝ adx；err < 0 → err ＋＝ ady、x0 ＋＝ 步進；k ＋＝ 1；1F6F:0C11(x0, y0, k, far SS:bp, n) ≠ 0 → 回 0
 *   k 當 flags 傳給 0C11：0C11 只在 (k & 7) ＝ 0 時才真的檢查（見 f_1F6F_0C11.c），所以每走 8 步檢查一次。
 * 語意（強推論）：docs/ 沒有 1F6F:091C 的記載；呼叫者 1F6F:0398（台帳，只有一個）。0B0A 的「兩艘艦艇之間的候選分隊」與本函式合起來是「兩艦艇間視線是否被別的分隊擋住」，
 *   名稱是假說，不是證據。+0x34／+0x36 ＝ 艦艇現在位置 X／Y（docs/internals/38-battle-units.md，強推論）。
 * 副作用：0B0A 寫 SS:bp 起的清單（堆疊）；0C11 沒有寫入。本函式自己不寫全域記憶體。
 *
 * 機器模型：C 版沒有堆疊。清單緩衝區放在原版的實際位置（基底狀態 fcmd-standby.state：SS 固定 0x63EB，進入點 SP ＝ 1E76，探針推 4 個參數字 8 bytes、返回位址 4、
 *   push bp／si／di 6、sub sp,2D0h，所以 bp ＝ 1B94，清單從 SS:1B94 起），交給 f_1F6F_0B0A 寫、f_1F6F_0C11 讀。這些位址在探針的堆疊帶（SS:1A76..1E75）內，
 *   寫入集合比對時雙方都排除，與原版把清單放在堆疊一致。
 * 輸入域：A、B 取艦艇池；清單筆數 ds:9896 列入 freeze 由 seeds 給定（清單超過 352 筆時 0B0A 的輸出會蓋到原版的區域變數，C 版不涵蓋），池基底 ds:CFB4..CFB7 列入 freeze；
 *   地圖與遮罩取記憶體，基底 fcmd-standby.state（0C11 → 1588:0571 需要戰場地圖）。原版的迴圈沒有上限（座標差大時要走很多步），超過探針步數上限的案例由探針丟棄；
 *   x 軸為主時 x0 超過 0x7FFF 的有號回繞也在迴圈裡，照原版。C 版加 70000 圈的上界避免突變版本停不下來，收錄的案例遠小於此。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint16_t f_1F6F_0B0A(Mem *m, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg, uint16_t out_off, uint16_t out_seg);
extern uint16_t f_1F6F_0C11(Mem *m, uint16_t x, uint16_t y, uint16_t flags, uint16_t list_off, uint16_t list_seg, uint16_t n);
extern uint16_t f_3C82_000F(uint16_t x);

#define F091C_SS 0x63EBu /* fcmd-standby.state 的 SS（diffprobe 無法設定） */
#define F091C_BP 0x1B94u /* 進入點 SP 1E76 − 參數 8 − 返回位址 4 − push bp／si／di 6 − sub sp,2D0h */

static uint16_t w091C_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static uint16_t w091C_sar1(uint16_t v) { return (uint16_t)((v >> 1) | (v & 0x8000u)); }

uint16_t f_1F6F_091C(Mem *m, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg)
{
    uint16_t n = f_1F6F_0B0A(m, a_off, a_seg, b_off, b_seg, F091C_BP, F091C_SS);
    uint16_t x0 = (uint16_t)(w091C_rd16s(m, a_seg, (uint16_t)(a_off + 0x34u)) + 7u);
    uint16_t y0 = (uint16_t)(w091C_rd16s(m, a_seg, (uint16_t)(a_off + 0x36u)) + 7u);
    uint16_t x1 = (uint16_t)(w091C_rd16s(m, b_seg, (uint16_t)(b_off + 0x34u)) + 7u);
    uint16_t y1 = (uint16_t)(w091C_rd16s(m, b_seg, (uint16_t)(b_off + 0x36u)) + 7u);
    uint16_t adx = f_3C82_000F((uint16_t)(x1 - x0));
    uint16_t ady = f_3C82_000F((uint16_t)(y1 - y0));
    uint16_t err, step, k = 0;
    unsigned guard;

    if ((int16_t)adx > (int16_t)ady) {
        if ((int16_t)x0 <= (int16_t)x1) {
            step = ((int16_t)y0 < (int16_t)y1) ? 1u : 0xFFFFu;
        } else {
            uint16_t t = x0;

            step = ((int16_t)y0 > (int16_t)y1) ? 1u : 0xFFFFu;
            x0 = x1;
            x1 = t;
            y0 = y1;
        }
        err = w091C_sar1(adx);
        for (guard = 0; guard < 70000u; guard++) {
            x0 = (uint16_t)(x0 + 1u);
            if ((int16_t)x0 > (int16_t)x1)
                return 1;
            err = (uint16_t)(err - ady);
            if ((int16_t)err < 0) {
                err = (uint16_t)(err + adx);
                y0 = (uint16_t)(y0 + step);
            }
            k = (uint16_t)(k + 1u);
            if (f_1F6F_0C11(m, x0, y0, k, F091C_BP, F091C_SS, n) != 0)
                return 0;
        }
        return 1;
    }
    if ((int16_t)y0 <= (int16_t)y1) {
        step = ((int16_t)x0 < (int16_t)x1) ? 1u : 0xFFFFu;
    } else {
        uint16_t t = y0;

        step = ((int16_t)x0 > (int16_t)x1) ? 1u : 0xFFFFu;
        y0 = y1;
        y1 = t;
        x0 = x1;
    }
    err = w091C_sar1(ady);
    for (guard = 0; guard < 70000u; guard++) {
        y0 = (uint16_t)(y0 + 1u);
        if ((int16_t)y0 > (int16_t)y1)
            return 1;
        err = (uint16_t)(err - adx);
        if ((int16_t)err < 0) {
            err = (uint16_t)(err + ady);
            x0 = (uint16_t)(x0 + step);
        }
        k = (uint16_t)(k + 1u);
        if (f_1F6F_0C11(m, x0, y0, k, F091C_BP, F091C_SS, n) != 0)
            return 0;
    }
    return 1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1F6F_091C(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("1F6F_091C", 4, run)
