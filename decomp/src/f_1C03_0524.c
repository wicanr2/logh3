/*
 * GIN3PS.EXE 1C03:0524  戰場艦隊轉向一步：朝向 +0x4A 往目標方位靠近，每步最多轉「回轉速度」+0x4D（回 AL：1 ＝ 有轉、0 ＝ 沒轉）
 *
 * 原版：far 呼叫，三個堆疊字（序言 push bp／si／di、sub sp,10h，所以 fleet 在 [bp+1Ah]／[bp+1Ch]、target 在 [bp+1Eh]），回 AX（見下）。
 * 逐條對應反組譯：
 *   d ← 2BE1:021B(target, word 艦隊[+0x4A])          ; 兩個 byte 角度的有號差 ＝ 方位 − 朝向，約 −128 到 128（見 f_2BE1_021B.c）
 *   absd ← |d|（16 位元：d < 0 時 0 − d）
 *   rate ← byte 艦隊[+0x4D]（零延伸）
 *   absd <= rate（無號，jbe）：                         ; 一步轉得到
 *       byte 艦隊[+0x4A] += d 的低位元組；結果 1
 *   否則（要多步）：
 *       dword 艦隊[+0x56]（AX:BX ＝ 高字:低字）與 0x00080000 比（3CC9:00B9，有號 32 位元），小於 → 結果 0（不轉）
 *       否則：
 *         s ← d > 0 ? 1 : −1（d ＝ 0 的分支是死碼：absd > rate >= 0 蘊含 d ≠ 0）
 *         p ← s × (有號 byte)艦隊[+0x4D]（imul byte，16 位元結果）
 *         byte 艦隊[+0x4A] += p 的低位元組
 *         dword 艦隊[+0x56] −= 0x00080000（sub 低字 0、sbb 高字 8）
 *         結果 1
 *   「>= 8.0 才轉、轉一次扣 8.0」：+0x56 是 16.16 定點的行進額度（強推論，docs/internals/38、spec/080）。
 *   imul 把 +0x4D 當有號 byte，但走到這個分支時 rate < absd <= 128，所以 rate <= 127，有號與無號沒有差別。
 *   額度比較的常數 0x00080000 低字為 0，所以有號 32 位元比較等價於「高字（有號）< 8」；|d| 與 rate 的比較無號與有號等價（都 < 0x8000）。
 *   指令覆蓋 70/72：+0x7D、+0x7F 是 d ＝ 0 的死碼，見 decomp/specs/1C03_0524.notes.md。
 *
 * 回傳 AX（描述檔 ret = ax；呼叫者 1C03:0185、1C03:0361 只取 AL，AH 是 codegen 殘值，但差異測試照比）：
 *   AL ＝ 結果；AH ＝ 結果為 0 時是 word 艦隊[+0x58] 的高位元組（進入 3CC9:00B9 前載入 AX，00B9 結束 pop ax 還原）；
 *   一步轉得到時是 d 的高位元組；多步時是 p 的高位元組（符號延伸，d < 0 時 0xFF）。
 *
 * 呼叫者：1C03:0185、1C03:0361（兩支實行常式，docs/internals/33 第 117 行「抵達朝向：位置 ＝ 目的地時拿它與 +0x4A 比，不同就 1C03:0524 轉一步」、
 *   docs/spec/027 第 82 行「|角差| ≤ 回轉速度 …」）。與文件一致，補上：多步時另需 +0x56 的額度 >= 8.0、轉完扣 8.0；
 *   一步轉得到的情形不查也不扣額度；轉的量是有號的 ±rate。
 *
 * 推論等級
 *   已證實（差異測試）：欄位位移與寬度、021B 的參數次序（方位在前）、|d| 與 rate 的無號比較、額度門檻與扣除量、符號與乘法、
 *     byte 加法不進位到 +0x4B、回傳 AX 的組成。
 *   強推論（docs/internals/33、spec/027）：艦隊 +0x4A 是朝向、+0x4D 是回轉速度、+0x56 是行進額度（16.16）。
 * 輸入域：fleet 指標任意（位移 16 位元回繞，word 在位移 0xFFFF 時高位元組取同一段的位移 0）；target 任意 word（021B 只取低位元組）。
 *   本版不模型化寫入 ROM 區（0xF0000 以上）；描述檔的艦隊指標取表內與種子指定的 RAM。
 */
#include "gin3_reg.h"

extern uint16_t f_2BE1_021B(Mem *m, uint16_t a_word, uint16_t b_word);

static uint16_t f0524_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f0524_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_1C03_0524(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t target)
{
    uint16_t head = (uint16_t)(fleet_off + 0x4Au);
    uint16_t d = f_2BE1_021B(m, target, f0524_rd16(m, fleet_seg, head));
    uint16_t absd = (int16_t)d < 0 ? (uint16_t)(0u - d) : d;
    uint16_t rate = rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x4Du)));

    if (absd <= rate) {
        wr8(m, LIN(fleet_seg, head), (uint8_t)(rd8(m, LIN(fleet_seg, head)) + (uint8_t)d));
        return (uint16_t)((d & 0xFF00u) | 1u);
    } else {
        uint16_t pos = (uint16_t)(fleet_off + 0x56u);
        uint16_t hi = f0524_rd16(m, fleet_seg, (uint16_t)(pos + 2u));
        uint16_t lo = f0524_rd16(m, fleet_seg, pos);
        int8_t s;
        int16_t p;

        if ((int32_t)(((uint32_t)hi << 16) | lo) < (int32_t)0x00080000)
            return (uint16_t)(hi & 0xFF00u);
        s = (int16_t)d > 0 ? 1 : ((int16_t)d == 0 ? 0 : -1);
        p = (int16_t)((int16_t)s * (int16_t)(int8_t)rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x4Du))));
        wr8(m, LIN(fleet_seg, head), (uint8_t)(rd8(m, LIN(fleet_seg, head)) + (uint8_t)p));
        f0524_wr16(m, fleet_seg, pos, f0524_rd16(m, fleet_seg, pos));
        f0524_wr16(m, fleet_seg, (uint16_t)(pos + 2u), (uint16_t)(f0524_rd16(m, fleet_seg, (uint16_t)(pos + 2u)) - 8u));
        return (uint16_t)(((uint16_t)p & 0xFF00u) | 1u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1C03_0524(m, a[0], a[1], a[2]); }
GIN3_REGISTER("1C03_0524", 3, run)
