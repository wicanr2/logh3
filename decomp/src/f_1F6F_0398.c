/*
 * GIN3PS.EXE 1F6F:0398  攻方艦艇對目標艦艇的武器候選遮罩：依攻方艦種的三個武器類型，各問「目標在不在該武器的方向橢圓射程內」，位 1／2／4 各對應一種武器；
 *   第 1 種武器（bit0）另要 1F6F:091C 的射線通行
 *
 * 原版（468 bytes，far 呼叫，兩個 far 指標參數 A（攻方，位移 [bp+28h]、段 [bp+2Ah]）、B（目標，位移 [bp+2Ch]、段 [bp+2Eh]）；序言 push bp／si／di、sub sp,1Eh、mov bp,sp），
 *   回傳 AL ＝ 遮罩（[bp+5]；AH 是殘值，描述檔 ret ＝ pack:al）。區域變數：[bp+0]／[bp+2] ＝ E 的位移／段、[bp+4] ＝ Δ（byte）、[bp+5] ＝ 遮罩、
 *   [bp+6]／[bp+8] ＝ 距離 d 的低字／高字（3310:0006 的 BX／AX）、[bp+0Ah]／[bp+0Ch] ＝ W 的位移／段，其餘只存不讀。
 * 過程（已證實：反組譯逐指令；3310:0006、2BE1:011D、2BE1:024A、1F6F:056C、1F6F:091C 由其各自的差異測試證實）：
 *   遮罩 ← 0
 *   E ← DS:(0xC074 ＋ (byte A[+1]) × 0x7A)                        ; 艦隊表項（16 位元位移）；E[+0x4A] ＝ 朝向 word
 *   d ← 3310:0006(x1 ＝ A[+0x34], y1 ＝ A[+0x36], x2 ＝ B[+0x34], y2 ＝ B[+0x36])        ; 16.16 距離（AX:BX ＝ 高字:低字）
 *   Δ ← byte 2BE1:024A(2BE1:011D(B[+0x34] − A[+0x34], B[+0x36] − A[+0x36]), E[+0x4A])   ; 呼叫後重新讀欄位，差是 16 位元
 *   W ← DS:(0x4E3A ＋ (byte A[+0]) × 0xD0 ＋ (byte A[+4]) × 0x1A)  ; 攻方艦種的武器表項（16 位元位移，可回繞）
 *   第 1 種：t ← byte W[+0x0A]；t ≠ 0xFF → H ← DS:(0x4FDA ＋ t × 6)；1F6F:056C(d_lo, d_hi, word [bp+4]（Δ ＋ 遮罩 << 8）, H[+2], H[+4]) ≠ 0
 *             且 1F6F:091C(A, B) ≠ 0 → 遮罩 |= 1
 *   第 2 種：t ← byte W[+0x0E]；t ≠ 0xFF 且 (word A[+0x0E] | word A[+0x10]) ≠ 0 → 同樣問 056C（不問 091C）→ 遮罩 |= 2
 *   第 3 種：t ← byte W[+0x12]；t ≠ 0xFF 且 (word A[+0x12] | word A[+0x14]) ≠ 0 → 同樣問 056C → 遮罩 |= 4
 *   回 遮罩
 *   056C 只用 Δ 參數的低位元組（反組譯，見 f_1F6F_056C.c）；原版推入的是 word [bp+4]（高位元組 ＝ 目前遮罩），本版照傳。
 * 語意（強推論）：docs/re/403-weapon-candidates.md 第 32 行（依攻方艦種三個武器類型、距離與攻方艦隊 +4A 朝向生成 mask 1／2／4）、第 35 行（1F6F:056C 與既有定點方向橢圓相同，距離等於半徑仍可用）、
 *   第 36 行（1F6F:04D0／0520 飛彈／戰鬥機兩 word 庫存 OR 非零即可；0 不接受）、第 37 行（1F6F:04B3 只有光線砲續查 091C；飛彈與戰鬥機不檢查射線）、docs/spec/100-fleet-combat-candidates.md 第 32 行，
 *   與本檔一致（本函式內第 2、3 種武器的庫存 OR 檢查在 A[+0x0E..0x11]、A[+0x12..0x15]，第 1 種沒有庫存檢查而有射線檢查）。武器類型表 ds:4E3A、ds:4FDA 的欄位語意是假說。
 * 副作用：全部來自被呼叫者：3310:0006 與 056C 內 39B5:005C 的暫存（ds:8EDC..8EED）；1F6F:091C 的堆疊清單（堆疊帶，比對排除）。本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊。基底狀態 fcmd-standby.state（091C 需要戰場地圖）：SS 固定 0x63EB，進入點 SP ＝ 1E76，帶 SS:1A76..1E75。
 *   3310:0006 與 056C 的 C 版把 39B5 的堆疊參數放在 fcmd-base.state 的框架（SS:1F82..1F89），在 standby 下落在帶外；呼叫前後存下並還原 SS:1F78..1F8F
 *   （這 24 bytes 只是 C 版被呼叫者的模型暫存，不是本函式的輸出；理由同 f_1C03_0361.c、f_1588_0891.c，docs/internals/67 第 267 行）。
 *   091C 的 C 版把清單放在 SS:1B94 起（假設自己是從探針直接呼叫）；從本函式呼叫時原版清單在 SS:1B64 起（bp(0398) ＝ 1E76 − 8 − 4 − 6 − 1Eh ＝ 1E46，再減 8 − 4 − 6 − 2D0h），
 *   兩處都在帶內，比對時雙方排除，C 版自寫自讀，不影響結果。
 * 輸入域：A、B 取艦艇池（8 筆）內指標；池基底 ds:CFB4..CFB7、名冊筆數 ds:9896 列入 freeze（理由同 1F6F:091C）；武器表、艦隊表欄位取記憶體與突變。
 *   原版 091C 的迴圈沒有上限，超過探針步數上限的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論，武器表欄位為假說。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_024A(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_1F6F_056C(Mem *m, uint16_t d_lo, uint16_t d_hi, uint16_t delta, uint16_t a, uint16_t b);
extern uint16_t f_1F6F_091C(Mem *m, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg);

#define F0398_SS 0x63EBu /* fcmd-standby.state 的 SS（diffprobe 無法設定） */

static uint16_t w0398_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 3310:0006 的 C 版：呼叫前後存還原 SS:1F78..1F8F */
static uint32_t w0398_dist(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2)
{
    uint8_t save[24];
    uint32_t r;
    int k;

    for (k = 0; k < 24; k++)
        save[k] = rd8(m, LIN(F0398_SS, (uint16_t)(0x1F78u + k)));
    r = f_3310_0006(m, x1, y1, x2, y2);
    for (k = 0; k < 24; k++)
        wr8(m, LIN(F0398_SS, (uint16_t)(0x1F78u + k)), save[k]);
    return r;
}

/* 一個武器槽：H ＝ DS:(0x4FDA ＋ t × 6)，問 056C(d, Δ 與遮罩的 word, H[+2], H[+4])；同樣存還原 SS:1F78..1F8F */
static uint16_t w0398_ellipse(Mem *m, uint32_t d, uint16_t delta_word, uint8_t t)
{
    uint8_t save[24];
    uint16_t h = (uint16_t)(0x4FDAu + t * 6u);
    uint16_t a = w0398_rd16s(m, GIN3_DS, (uint16_t)(h + 2u));
    uint16_t b = w0398_rd16s(m, GIN3_DS, (uint16_t)(h + 4u));
    uint16_t r;
    int k;

    for (k = 0; k < 24; k++)
        save[k] = rd8(m, LIN(F0398_SS, (uint16_t)(0x1F78u + k)));
    r = f_1F6F_056C(m, (uint16_t)d, (uint16_t)(d >> 16), delta_word, a, b);
    for (k = 0; k < 24; k++)
        wr8(m, LIN(F0398_SS, (uint16_t)(0x1F78u + k)), save[k]);
    return r;
}

uint8_t f_1F6F_0398(Mem *m, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg)
{
    uint8_t mask = 0;
    uint8_t k = rd8(m, LIN(a_seg, (uint16_t)(a_off + 1u)));
    uint16_t e = (uint16_t)(0xC074u + k * 0x7Au);
    uint16_t ax1 = w0398_rd16s(m, a_seg, (uint16_t)(a_off + 0x34u));
    uint16_t ay1 = w0398_rd16s(m, a_seg, (uint16_t)(a_off + 0x36u));
    uint16_t bx1 = w0398_rd16s(m, b_seg, (uint16_t)(b_off + 0x34u));
    uint16_t by1 = w0398_rd16s(m, b_seg, (uint16_t)(b_off + 0x36u));
    uint32_t d = w0398_dist(m, ax1, ay1, bx1, by1);
    uint16_t face = w0398_rd16s(m, GIN3_DS, (uint16_t)(e + 0x4Au));
    uint16_t ang;
    uint8_t delta, t0;
    uint16_t w;

    by1 = w0398_rd16s(m, b_seg, (uint16_t)(b_off + 0x36u));
    ay1 = w0398_rd16s(m, a_seg, (uint16_t)(a_off + 0x36u));
    bx1 = w0398_rd16s(m, b_seg, (uint16_t)(b_off + 0x34u));
    ax1 = w0398_rd16s(m, a_seg, (uint16_t)(a_off + 0x34u));
    ang = f_2BE1_011D(m, (uint16_t)(bx1 - ax1), (uint16_t)(by1 - ay1));
    delta = (uint8_t)f_2BE1_024A(m, ang, face);
    w = (uint16_t)(0x4E3Au + rd8(m, LIN(a_seg, a_off)) * 0xD0u + rd8(m, LIN(a_seg, (uint16_t)(a_off + 4u))) * 0x1Au);

    t0 = rd8(m, LIN(GIN3_DS, (uint16_t)(w + 0x0Au)));
    if (t0 != 0xFFu) {
        if (w0398_ellipse(m, d, (uint16_t)(delta | ((uint16_t)mask << 8)), t0) != 0) {
            if (f_1F6F_091C(m, a_off, a_seg, b_off, b_seg) != 0)
                mask |= 1;
        }
    }

    t0 = rd8(m, LIN(GIN3_DS, (uint16_t)(w + 0x0Eu)));
    if (t0 != 0xFFu && (uint16_t)(w0398_rd16s(m, a_seg, (uint16_t)(a_off + 0x0Eu)) | w0398_rd16s(m, a_seg, (uint16_t)(a_off + 0x10u))) != 0) {
        if (w0398_ellipse(m, d, (uint16_t)(delta | ((uint16_t)mask << 8)), t0) != 0)
            mask |= 2;
    }

    t0 = rd8(m, LIN(GIN3_DS, (uint16_t)(w + 0x12u)));
    if (t0 != 0xFFu && (uint16_t)(w0398_rd16s(m, a_seg, (uint16_t)(a_off + 0x12u)) | w0398_rd16s(m, a_seg, (uint16_t)(a_off + 0x14u))) != 0) {
        if (w0398_ellipse(m, d, (uint16_t)(delta | ((uint16_t)mask << 8)), t0) != 0)
            mask |= 4;
    }
    return mask;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1F6F_0398(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("1F6F_0398", 4, run)
