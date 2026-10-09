/*
 * GIN3PS.EXE 113A:0D2C  戰場入口表查詢：依「從哪個星域打進來」挑出開場的格欄與格列（回 1 找到、0 沒找到）
 *
 * 原版：far 呼叫，四個參數；序言 push bp／si／di、sub sp,4，所以
 *   key 在 [bp+0Eh]、idx 在 [bp+10h]（兩個都只讀低位元組）、out_col 在 [bp+12h]／[bp+14h]、out_row 在 [bp+16h]／[bp+18h]。
 * 回傳 AX（0 或 1，AH ＝ 0）。
 *
 * 反組譯逐條對應：
 *   ent ← far 指標 ds:[0x4B2E + idx × 4]（位移）／ds:[0x4B30 + idx × 4]（段）          idx 零延伸
 *   迴圈：k ← ent[+0]（byte）；k == key 或 k == 0xFF 就跳出；否則 ent += 4（16 位元位移）
 *   若 ent[+0] != 0xFF（重讀）：
 *       *out_col ← ent[+1]（byte）
 *       *out_row ← ent[+2]（byte；讀在上一行寫入之後，out_col 若指到 ent[+2] 就會讀到新值）
 *       回 1
 *   否則回 0（key ＝ 0xFF 時第一筆符合的就是結尾標記，也回 0）
 *
 * 推論等級
 *   已證實（差異測試）：上列結構、常數、byte 參數、每筆 4 bytes、0xFF 為結尾、重讀的位置。
 *   強推論（docs/internals/39-battle-entry-points.md、docs/re/84）：ds:4B2E 起是 32 個 far 指標，用戰場所在的
 *     星域編號索引；每筆 [0] ＝ 鍵（從哪個星域打進來）、[1] ＝ 格欄、[2] ＝ 格列；呼叫者 297B:0ABF 再換成
 *     艦隊座標（欄 × 16、列 × 20）。
 *
 * 輸入域：idx 取 0 到 31（表內）與其高位元組非 0 的變形（只取低位元組）；表與 key 任意。
 *   idx >= 32 時 far 指標讀到表之後的資料，不是設計的用法，本版照原版運算但描述檔不主動產生。
 *   找不到 key 也沒有結尾標記的表會繞過整個段而不終止，本版與原版一樣（探針以步數上限丟棄這類案例）。
 */
#include "gin3_reg.h"

uint16_t f_113A_0D2C(Mem *m, uint16_t key_word, uint16_t idx_word, uint16_t col_off, uint16_t col_seg,
                     uint16_t row_off, uint16_t row_seg)
{
    uint8_t key = (uint8_t)key_word;
    uint16_t idx = (uint8_t)idx_word;
    uint16_t ent_off = rd16(m, DS((uint16_t)(0x4B2Eu + idx * 4u)));
    uint16_t ent_seg = rd16(m, DS((uint16_t)(0x4B30u + idx * 4u)));

    for (;;) {
        uint8_t k = rd8(m, LIN(ent_seg, ent_off));

        if (k == key || k == 0xFF)
            break;
        ent_off = (uint16_t)(ent_off + 4);
    }
    if (rd8(m, LIN(ent_seg, ent_off)) == 0xFF)
        return 0;
    wr8(m, LIN(col_seg, col_off), rd8(m, LIN(ent_seg, ent_off + 1)));
    wr8(m, LIN(row_seg, row_off), rd8(m, LIN(ent_seg, ent_off + 2)));
    return 1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_113A_0D2C(m, a[0], a[1], a[2], a[3], a[4], a[5]); }
GIN3_REGISTER("113A_0D2C", 6, run)
