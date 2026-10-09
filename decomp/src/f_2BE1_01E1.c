/*
 * GIN3PS.EXE 2BE1:01E1  反正切表查找：第一個「比值 <= 表[k]」的 k，最多回 0x40
 *
 * 原版：far 呼叫，一個堆疊字 ratio（push bp、sub sp,6，所以參數在 [bp+0Ch]），回 AX = k（0 到 0x40）。
 *   表是 ds:593A 起的 word 陣列（0x41 筆有效，k = 0 到 0x40，docs/re/330 第 313 行稱它為反正切表）。
 *   迴圈（每圈重讀表項）：
 *     entry ← 表[k]
 *     ratio <= entry（無號 word 比較，jbe）→ 結束
 *     k >= 0x40（無號 byte 比較，jnb）→ 結束
 *     否則 k++，表指標 += 2
 *   k 是 byte 計數器，從 0 開始；回傳 AH = 0。
 *
 * 輸入域：ratio 全部 16 位元值。表項取自記憶體（基底狀態與突變），沒有假設遞增。k 最大 0x40，
 *   所以最多讀到 ds:593A + 0x80（含）。
 * 證據：asm 逐指令（2BE1:01E1 的反組譯），docs/re/330 第 313 行的虛擬碼「第一個 比值 <= ds:593A[k] 的 k，上限 64」
 *   與本檔一致。推論等級：已證實（反組譯逐指令＋差異測試）。ratio 與表項的名稱（正切 ×1024）是強推論，本檔不使用。
 */
#include "gin3_reg.h"

uint16_t f_2BE1_01E1(Mem *m, uint16_t ratio)
{
    uint8_t k = 0;
    uint16_t p = 0x593A;

    for (;;) {
        uint16_t entry = rd16(m, DS(p));

        if (ratio <= entry)
            break;
        if (k >= 0x40)
            break;
        p = (uint16_t)(p + 2);
        k++;
    }
    return k;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2BE1_01E1(m, a[0]); }
GIN3_REGISTER("2BE1_01E1", 1, run)
