/*
 * GIN3PS.EXE 2D76:0DB2  檢查「隊形槽」有沒有出現在清單裡：沒有回 1，有回 0
 *
 * 原版：far 呼叫，三個參數（push bp、sub sp,2，所以參數從 [bp+8] 起）：
 *   args[0..1] ＝ 清單的 far 指標（位移、段）；清單每筆 2 bytes，第 2 個 byte（+1）是被比較的值
 *   args[2]    ＝ value（word，只取低位元組：mov cl,[bp+0Ch]）
 *   args[3]    ＝ n（word，只取低位元組：mov dl,[bp+0Eh]）
 * 過程（i 是 byte 計數器，從 0 開始）：
 *   i >= n（無號 byte 比較）→ 回 1（沒找到）
 *   清單[i][+1] == value → 回 0（找到）
 *   否則 i++，指標 += 2（16 位元回繞）
 * 純函式，不寫記憶體。
 *
 * 輸入域：清單指標任意；value、n 全部 16 位元值（只有低位元組有效）。n 最大 255，最多讀到指標 + 510。
 * 證據：asm 逐指令；docs/re/229 第 44 到 46 行「2D76:0DB2 回『這個隊形槽沒有出現在…』」與本檔一致
 *   （呼叫者 2D76:0D40 傳入樣板、隊形槽與筆數 n）；清單與 value 的欄位語意為強推論，位移已證實。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2D76_0DB2(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t value_word, uint16_t n_word)
{
    uint8_t value = (uint8_t)value_word;
    uint8_t n = (uint8_t)n_word;
    uint8_t i = 0;

    while (i < n) {
        if (rd8(m, LIN(list_seg, list_off + 1)) == value)
            return 0;
        i++;
        list_off = (uint16_t)(list_off + 2);
    }
    return 1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2D76_0DB2(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("2D76_0DB2", 4, run)
