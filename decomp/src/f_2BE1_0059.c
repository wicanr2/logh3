/*
 * GIN3PS.EXE 2BE1:0059  艦艇的隊形槽偏移依艦隊朝向旋轉、加上艦隊戰場座標，夾進 [0,0x7FF]×[0,0x7F7] 後存成艦艇 +0x28／+0x2A
 *
 * 原版：far 呼叫，兩個 far 指標參數（序言 push bp／si、sub sp,6、mov bp,sp，所以第一個在 [bp+0Eh]／[bp+10h]，第二個在 [bp+12h]／[bp+14h]），
 *   沒有有意義的回傳值（描述檔 ret ＝ none）。區域變數：[bp+0] ＝ x、[bp+2] ＝ y（兩個有號 word）、[bp+4] ＝ 艦艇 +5 的位址（位移）。
 *   args[0..1] ＝ F（艦隊記錄，122 bytes 一筆）  args[2..3] ＝ S（艦艇池記錄，78 bytes 一筆）
 * 過程（已證實：反組譯逐指令；3A92:0001 由其自己的差異測試證實）：
 *   slot ← byte S[+5]                                           ; 隊形槽索引（0–40）
 *   x ← word ds:(0x59BA + slot × 4)；y ← word ds:(0x59BA + slot × 4 + 2)      ; 隊形槽表，一筆 4 bytes（兩個有號 word）
 *     （x、y 都存進堆疊；S[+5] 讀兩次，中間沒有寫入）
 *   3A92:0001(&x, &y, word F[+0x4A])                            ; 以 byte 角度 F[+0x4A] 旋轉 (x, y)，結果寫回 x、y（在堆疊）
 *   x ← x + word F[+0x5C]；y ← y + word F[+0x5E]               ; 16 位元加法
 *   S[+0x28] ← 夾限(x)：x < 0（有號 jns）→ 0；x > 0x7FF（有號 jle）→ 0x7FF；否則 x
 *   S[+0x2A] ← 夾限(y)：y < 0 → 0；y > 0x7F7 → 0x7F7；否則 y
 *   兩次寫入都在所有讀取（S[+5]、隊形槽表、F[+0x4A]、F[+0x5C]、F[+0x5E]）之後，所以 S 與 F 重疊時，寫入不影響已算好的 x、y，只是
 *   把值寫進 F 的欄位上；S[+0x28] 的寫入先於 S[+0x2A]，兩者落在同一個 word 以外的位置時互不影響。
 * 語意（強推論）：docs/internals/38 第 121 行「`+0x05` 隊形槽索引（0–40）：目的地 ＝ 艦隊位置 ＋ ds:59BA 第 N 筆偏移依艦隊朝向旋轉（2BE1:0059、docs/re/126）」；
 *   docs/spec/025 第 17 行、docs/re/300 第 29 行、docs/internals/46 第 81 行：呼叫者 2BE1:0000、2BE1:0328、1C61:029C。
 *   S[+0x28]／S[+0x2A] 是艦艇的目的地座標（強推論）；0x7FF、0x7F7 是戰場世界座標的上限（強推論，兩者來源未驗）。
 * 副作用：S[+0x28]、S[+0x2A] 兩個 word。線性 0xF0000 以上（ROM）的寫入在 dosgolem 被安靜丟棄，本檔按 byte 模型化；
 *   word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（rd16s／wr16s）。
 *
 * 機器模型：C 版沒有堆疊。x、y 兩個 word 放在原版的實際位置 SS:1F52、SS:1F54（基底狀態（fcmd-s3-base，stack_hi 0x65E18）SP ＝ 1F68，SS 固定 0x63EB，diffprobe 無法設定；
 *   進入點 SP ＝ 1F68 − 參數 8 − 返回位址 4 − push bp／si 4 − sub sp,6 ＝ 1F52 ＝ bp），交給 f_3A92_0001 旋轉後讀回。這個位址在
 *   探針的堆疊帶（SS:1B68..1F67）內，寫入集合比對時雙方都排除，與原版把 x、y 放在堆疊一致。
 * 輸入域：F、S 指到艦隊表、艦艇池或 seeds 給的位址；不指向堆疊帶、A0000 到 BFFFF。基底狀態 fcmd-s3-base.state。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word);

#define F0059_SS 0x63EBu /* fcmd-s3-base.state 的 SS（diffprobe 無法設定） */
#define F0059_BP 0x1F52u /* 進入點 SP 1F68 − 8 − 4 − 4 − 6 */

static uint16_t f0059_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void f0059_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void f0059_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f0059_put8(m, LIN(seg, off), (uint8_t)v);
    f0059_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_2BE1_0059(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t s_off, uint16_t s_seg)
{
    uint16_t slot_at = (uint16_t)(s_off + 5u);
    uint16_t x, y;
    int16_t sx, sy;

    x = f0059_rd16s(m, GIN3_DS, (uint16_t)(0x59BAu + (uint16_t)(rd8(m, LIN(s_seg, slot_at)) << 2)));
    y = f0059_rd16s(m, GIN3_DS, (uint16_t)(0x59BAu + (uint16_t)(rd8(m, LIN(s_seg, slot_at)) << 2) + 2u));
    f0059_wr16s(m, F0059_SS, F0059_BP, x);
    f0059_wr16s(m, F0059_SS, (uint16_t)(F0059_BP + 2u), y);
    f_3A92_0001(m, F0059_BP, F0059_SS, (uint16_t)(F0059_BP + 2u), F0059_SS, f0059_rd16s(m, f_seg, (uint16_t)(f_off + 0x4Au)));
    x = (uint16_t)(f0059_rd16s(m, F0059_SS, F0059_BP) + f0059_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)));
    y = (uint16_t)(f0059_rd16s(m, F0059_SS, (uint16_t)(F0059_BP + 2u)) + f0059_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)));
    sx = (int16_t)x;
    sy = (int16_t)y;
    if (sx < 0)
        f0059_wr16s(m, s_seg, (uint16_t)(s_off + 0x28u), 0);
    else if (sx > 0x7FF)
        f0059_wr16s(m, s_seg, (uint16_t)(s_off + 0x28u), 0x7FF);
    else
        f0059_wr16s(m, s_seg, (uint16_t)(s_off + 0x28u), x);
    if (sy < 0)
        f0059_wr16s(m, s_seg, (uint16_t)(s_off + 0x2Au), 0);
    else if (sy > 0x7F7)
        f0059_wr16s(m, s_seg, (uint16_t)(s_off + 0x2Au), 0x7F7);
    else
        f0059_wr16s(m, s_seg, (uint16_t)(s_off + 0x2Au), y);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2BE1_0059(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("2BE1_0059", 4, run)
