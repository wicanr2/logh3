/*
 * GIN3PS.EXE 3310:0006  16.16 定點距離：兩點 (x1, y1)、(x2, y2) 的歐氏距離 ＝ 長邊 × √(1 ＋ (短邊 ÷ 長邊)²)，長短邊以「左移 16 位的差」做有號 32 位元比較
 *
 * 原版（298 bytes，far 呼叫，四個 word 堆疊參數；序言 push bp／si／di、sub sp,0Ch、mov bp,sp，所以 x1 在 [bp+16h]、y1 在 [bp+18h]、x2 在 [bp+1Ah]、y2 在 [bp+1Ch]），
 *   回傳 AX ＝ 高字、BX ＝ 低字（描述檔 ret ＝ pack:ax,bx）。區域變數：DX ＝ [bp+2]:[bp+0]、DY ＝ [bp+6]:[bp+4]（高字:低字）。
 * 過程（已證實：反組譯逐指令；39B5:000C、39B5:005C、3310:014B、3CC9:00B9 各由其差異測試證實）：
 *   adx ← |x1 − x2|、ady ← |y1 − y2|（無號 16 位元，jb 決定誰減誰）
 *   DX ← adx << 16、DY ← ady << 16（32 位元，低字為 0）
 *   adx ＝ 0 → 回 DY（AX ＝ ady、BX ＝ 0）
 *   ady ＝ 0 → 回 DX
 *   3CC9:00B9(DX, DY)，jl（32 位元有號：DX < DY）：
 *     DX >= DY：r ← 39B5:005C(DY, DX)                 ; (DY << 16) ÷ DX，16.16 的 短邊 ÷ 長邊
 *               t ← 39B5:000C(r, r)                    ; r × r >> 16
 *               t ← t ＋ 0x10000                        ; add bx,0；adc ax,1
 *               s ← 3310:014B(t)                        ; 16.16 開方
 *               回 39B5:000C(DX, s)
 *     DX <  DY：同上，角色互換：r ← 005C(DX, DY)，最後回 000C(DY, s)
 *   「DX >= DY」用有號比較：adx 或 ady 大於等於 0x8000 時，DX 或 DY 是負的，選到的「長邊」可能是較短的那一邊。
 * 語意（強推論）：docs/spec/019-planet-assault.md 第 50 行「距離 ＝ 3310:0006(行星 +0x46/+0x48, 艦隊 +0x5C/+0x5E) ≤ 150.0（16.16）」、
 *   docs/re/330-strategic-planet-assault.md 第 174 行、docs/re/102-battle-movement.md 第 32 行（距離，16.16），與本檔一致；
 *   docs/solved-history.md 第 1049 行「距離是歐氏距離（3310:0006）」。
 * 副作用：全部來自 39B5:005C（ds:8EDC..8EED 共 18 bytes，R／D／Q 暫存的最終值）；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊；被呼叫的 39B5:000C／005C 的堆疊參數放在 fcmd-base.state 進入點 SP ＝ 1F8A 的框架（f_39B5_000C.c／f_39B5_005C.c 檔頭），
 *   原版從本函式呼叫時框架更深，位置不同，但兩處都在探針的堆疊帶內，比對時排除，C 版自寫自讀，不影響結果。
 * 輸入域：四個參數全部 16 位元值。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_39B5_000C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint32_t f_3310_014B(Mem *m, uint16_t lo, uint16_t hi);

uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2)
{
    uint16_t adx = (x1 < x2) ? (uint16_t)(x2 - x1) : (uint16_t)(x1 - x2);
    uint16_t ady = (y1 < y2) ? (uint16_t)(y2 - y1) : (uint16_t)(y1 - y2);
    uint32_t dx = (uint32_t)adx << 16;
    uint32_t dy = (uint32_t)ady << 16;
    uint32_t num, den, big, r, t, s;

    if (adx == 0)
        return dy;
    if (ady == 0)
        return dx;
    if ((int32_t)dx >= (int32_t)dy) {
        num = dy;
        den = dx;
        big = dx;
    } else {
        num = dx;
        den = dy;
        big = dy;
    }
    r = f_39B5_005C(m, (uint16_t)num, (uint16_t)(num >> 16), (uint16_t)den, (uint16_t)(den >> 16));
    t = f_39B5_000C(m, (uint16_t)r, (uint16_t)(r >> 16), (uint16_t)r, (uint16_t)(r >> 16));
    t = t + 0x10000u;
    s = f_3310_014B(m, (uint16_t)t, (uint16_t)(t >> 16));
    return f_39B5_000C(m, (uint16_t)big, (uint16_t)(big >> 16), (uint16_t)s, (uint16_t)(s >> 16));
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3310_0006(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("3310_0006", 4, run)
