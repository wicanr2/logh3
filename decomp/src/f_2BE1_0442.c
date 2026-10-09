/*
 * GIN3PS.EXE 2BE1:0442  艦隊的 41 個隊形槽各自旋轉、加上艦隊戰場座標、夾進 [0,0x7FF]×[0,0x7F7]，寫成 ds:D072 起的 41 筆世界座標
 *
 * 原版：far 呼叫，一個 far 指標參數 F（艦隊記錄，122 bytes 一筆）（序言 push bp／si、sub sp,0Ah、mov bp,sp，所以參數在 [bp+12h]／[bp+14h]），
 *   沒有有意義的回傳值（描述檔 ret ＝ none）。區域變數：[bp+0] ＝ k（byte）、[bp+2] ＝ x、[bp+4] ＝ y（有號 word）、[bp+6] ＝ 目的筆的位址（位移，段 ＝ DS）。
 * 過程（已證實：反組譯逐指令；3A92:0001 由其自己的差異測試證實）：
 *   cell ← ds:D072；對 k ＝ 0..40（byte，jb 無號比較 < 0x29）：
 *     x ← word ds:(0x59BA + k × 4)；y ← word ds:(0x59BA + k × 4 + 2)       ; 隊形槽表一筆 4 bytes（兩個有號 word）
 *     3A92:0001(&x, &y, word F[+0x4A])                                      ; 以 byte 角度 F[+0x4A] 旋轉 (x, y)，結果寫回 x、y（在堆疊）
 *     x ← x + word F[+0x5C]；y ← y + word F[+0x5E]                          ; 16 位元加法
 *     cell[+0] ← 夾限(x)：x < 0（有號 jns）→ 0；x > 0x7FF（有號 jle）→ 0x7FF；否則 x（word 寫入）
 *     cell[+2] ← 夾限(y)：y < 0 → 0；y > 0x7F7 → 0x7F7；否則 y（word 寫入）
 *     cell[+4] ← 0（byte）
 *     k++；cell ← cell + 10
 *   F 的欄位每圈重讀（`les bx,[bp+12h]`）：F 指到 ds:D072..D21B 的筆時，前一圈的寫入影響後一圈的讀取（逐步照做）。
 * 語意（強推論）：2BE1:0059 的整艦隊版——對 41 個隊形槽預先算出世界座標，docs/re/130 第 19 行「2BE1:0442(艦隊) ; 算出 41 格的世界座標」、
 *   第 42 行「一筆 10 bytes、共 41 筆」。呼叫者 2BE1:0328。每筆 +4 的 byte 清 0（用途未驗）。
 * 副作用：ds:D072 起 41 × 10 ＝ 410 bytes 的前 5 個 byte（每筆 +0..+4）；其餘 5 個 byte 不動。DS 內寫入，不涉及 ROM 區。
 *
 * 機器模型：C 版沒有堆疊。x、y 兩個 word 放在原版的實際位置 SS:1F54、SS:1F56（基底狀態（fcmd-s3-base，stack_hi 0x65E18）SP ＝ 1F68，SS 固定 0x63EB，diffprobe 無法設定；
 *   進入點 SP ＝ 1F68 − 參數 4 − 返回位址 4 − push bp／si 4 − sub sp,0Ah ＝ 1F52 ＝ bp，x 在 bp+2、y 在 bp+4），交給 f_3A92_0001 旋轉後讀回。
 *   這個位址在探針的堆疊帶（SS:1B68..1F67）內，寫入集合比對時雙方都排除，與原版把 x、y 放在堆疊一致。
 * 輸入域：F 指到艦隊表或 seeds 給的位址；不指向堆疊帶、A0000 到 BFFFF。基底狀態 fcmd-s3-base.state。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word);

#define F0442_SS 0x63EBu /* fcmd-s3-base.state 的 SS（diffprobe 無法設定） */
#define F0442_BP 0x1F52u /* 進入點 SP 1F68 − 4 − 4 − 4 − 0x0A */

static uint16_t f0442_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f0442_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_2BE1_0442(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint16_t cell = 0xD072u;
    uint8_t k;

    for (k = 0; k < 0x29u; k++) {
        uint16_t slot = (uint16_t)(0x59BAu + (uint16_t)(k << 2));
        int16_t sx, sy;
        uint16_t x, y;

        f0442_wr16s(m, F0442_SS, (uint16_t)(F0442_BP + 2u), f0442_rd16s(m, GIN3_DS, slot));
        f0442_wr16s(m, F0442_SS, (uint16_t)(F0442_BP + 4u), f0442_rd16s(m, GIN3_DS, (uint16_t)(slot + 2u)));
        f_3A92_0001(m, (uint16_t)(F0442_BP + 2u), F0442_SS, (uint16_t)(F0442_BP + 4u), F0442_SS,
                    f0442_rd16s(m, f_seg, (uint16_t)(f_off + 0x4Au)));
        x = (uint16_t)(f0442_rd16s(m, F0442_SS, (uint16_t)(F0442_BP + 2u)) + f0442_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)));
        y = (uint16_t)(f0442_rd16s(m, F0442_SS, (uint16_t)(F0442_BP + 4u)) + f0442_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)));
        sx = (int16_t)x;
        sy = (int16_t)y;
        if (sx < 0)
            f0442_wr16s(m, GIN3_DS, cell, 0);
        else if (sx > 0x7FF)
            f0442_wr16s(m, GIN3_DS, cell, 0x7FF);
        else
            f0442_wr16s(m, GIN3_DS, cell, x);
        if (sy < 0)
            f0442_wr16s(m, GIN3_DS, (uint16_t)(cell + 2u), 0);
        else if (sy > 0x7F7)
            f0442_wr16s(m, GIN3_DS, (uint16_t)(cell + 2u), 0x7F7);
        else
            f0442_wr16s(m, GIN3_DS, (uint16_t)(cell + 2u), y);
        wr8(m, DS((uint16_t)(cell + 4u)), 0);
        cell = (uint16_t)(cell + 0x0Au);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2BE1_0442(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2BE1_0442", 2, run)
