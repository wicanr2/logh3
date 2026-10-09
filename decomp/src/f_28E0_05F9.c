/*
 * GIN3PS.EXE 28E0:05F9  提督功績夠門檻就升一級：扣功績、等級加一、等級分檔有變時記旗標、再抽亂數成長能力值，回旗標遮罩
 *
 * 原版：far 呼叫，一個 word 參數＝提督編號（只取低位元組：mov al,[bp+18h]；序言 push bp／si／di、sub sp,0Eh，所以參數在 [bp+18h]）。
 *   提督記錄 ds:(0xB77C + 編號 × 0x1C)（28 bytes）。區域變數：[bp+0／+2] ＝ 提督記錄位址（位移／段）、[bp+4] ＝ 夾限後的等級、
 *   [bp+6] ＝ 回傳遮罩、[bp+0Ah／+0Ch] ＝ 等級欄位（記錄 +6）的位址。回傳 AX ＝ 遮罩（0、1、3，或 | 28E0:06C3 的成長位元 0x004..0x800）。
 * 過程（已證實：反組譯逐指令；28E0:0096、28E0:06C3 由其各自的差異測試證實）：
 *   mask ← 0
 *   adm[+6]（等級，byte）＝ 0x63（99）→ 回 0                                  ; 99 級不再升
 *   L ← min(adm[+6], 0x27)（byte 無號比較 jb）
 *   V ← dword adm[+2]（低 word 在 +2、高 word 在 +4）；thr ← word ds:(0x4DBE + 2L)
 *   V < thr（32 位元無號：高 word 非 0 一律不小於）→ 回 0                       ; 功績不足
 *   mask |= 1
 *   V ← V − thr（sub 低 word、sbb 高 word，寫回 adm[+2..+5]）
 *   c1 ← AL of 28E0:0096(word (adm[+6] ＋ 1)（16 位元加法，0xFF → 0x100，被呼叫者只取低位元組）)
 *   c0 ← AL of 28E0:0096(word adm[+6]（連 +7 一起推，被呼叫者只取低位元組）)
 *   c1 ≠ c0 → mask |= 2                                                     ; 軍階等級（0096 的 12 檔）有變
 *   adm[+6] ← adm[+6] + 1（byte 回繞）
 *   mask |= 28E0:06C3(adm)                                                  ; 抽亂數決定哪幾項能力成長
 *   回 mask
 * 語意（強推論，docs 已寫）：docs/re/418 第 37 到 38 行「28E0:05F9 讀 ADMIRAL 的功績 u32 與等級 byte。僅等級恰為 99 直接拒絕。
 *   門檻取 DS:4DBE[min(level,39)]；功績不足回 0，否則扣門檻、加等級，回旗標 bit0」、docs/spec/062 第 81 行、docs/internals/08 第 7 行。
 *   mask 的 bit1 ＝ 軍階檔（0096）有變是本檔的新發現（docs/re/418 沒寫）；bit2 以上是 06C3 的成長位元。呼叫者 056A:0B0D、185C:060D。
 * 副作用：adm[+2..+5]（功績）、adm[+6]（等級）、06C3 寫的能力值與 ds:94DC 亂數種子。
 * 輸入域：編號低位元組任意（超過 81 時讀到提督表之外的 DGROUP 資料，照原版運算）；提督記錄與 DS 內其他資料的別名不涵蓋
 *   （記錄與門檻表 ds:4DBE 重疊的情形由突變自然產生時照原版運算）。基底狀態 fcmd-base.state。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint8_t f_28E0_0096(Mem *m, uint16_t v_word);
extern uint16_t f_28E0_06C3(Mem *m, uint16_t adm_off, uint16_t adm_seg);

uint16_t f_28E0_05F9(Mem *m, uint16_t idx_word)
{
    uint16_t adm = (uint16_t)(0xB77Cu + (uint16_t)((uint8_t)idx_word * 0x1Cu));
    uint16_t mask = 0;
    uint8_t level = rd8(m, DS((uint16_t)(adm + 6u)));
    uint8_t cap;
    uint16_t lo, hi, thr;
    uint8_t c0, c1;

    if (level == 0x63u)
        return 0;
    cap = (level < 0x27u) ? level : 0x27u;
    lo = rd16(m, DS((uint16_t)(adm + 2u)));
    hi = rd16(m, DS((uint16_t)(adm + 4u)));
    thr = rd16(m, DS((uint16_t)(0x4DBEu + 2u * cap)));
    if (hi == 0 && lo < thr)
        return 0;
    mask |= 1u;
    wr16(m, DS((uint16_t)(adm + 2u)), (uint16_t)(rd16(m, DS((uint16_t)(adm + 2u))) - thr));
    wr16(m, DS((uint16_t)(adm + 4u)), (uint16_t)(rd16(m, DS((uint16_t)(adm + 4u))) - (lo < thr ? 1u : 0u)));
    c1 = f_28E0_0096(m, (uint16_t)(rd8(m, DS((uint16_t)(adm + 6u))) + 1u));
    c0 = f_28E0_0096(m, rd16(m, DS((uint16_t)(adm + 6u))));
    if (c1 != c0)
        mask |= 2u;
    wr8(m, DS((uint16_t)(adm + 6u)), (uint8_t)(rd8(m, DS((uint16_t)(adm + 6u))) + 1u));
    mask |= f_28E0_06C3(m, adm, GIN3_DS);
    return mask;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_05F9(m, a[0]); }
GIN3_REGISTER("28E0_05F9", 1, run)
