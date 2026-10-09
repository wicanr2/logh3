/*
 * GIN3PS.EXE 0FF4:0B38  依艦種順序表挑第一個編得出至少一組的艦種（回 AL；沒有就回 0xFF）
 *
 * 原版：far 呼叫，一個堆疊字（序言 push bp／si、sub sp,4，所以參數在 [bp+0Ch]），只讀低位元組，回 AL。
 *   p ← ds:(0x2986 ＋ low16(idx(byte) × 7))      ; 順序表一列 7 bytes，以 0xFF 結尾；列內是艦種編號
 *   迴圈：
 *     k ← byte ds:p；k ＝ 0xFF → 回 0xFF
 *     0FF4:0E84(k) ≠ 0（存量 ÷ 每組艘數 的低 16 位元非 0，語意見 f_0FF4_0E84.c）→ 回「重讀的 byte ds:p」（＝ k）
 *     p++（16 位元），回到迴圈
 * 0E84 沒有寫入，所以本函式不寫記憶體。p 是 16 位元位移，列內沒有 0xFF 時會一路讀過 DS 段尾回繞到段首。
 *
 * 回傳：只比較 AL（描述檔 ret = pack:al）。AH 是 codegen 殘值：回 k 的路徑上是 0E84 回傳值的高位元組，
 * 回 0xFF 的路徑上是「idx × 7 的高位元組」（一次都沒呼叫 0E84 時）或 0（呼叫過 0E84 且回 0 時）。呼叫者 0FF4:0A4B
 * 只存 AL（mov [bp+2],al；cmp al,0FFh），所以 AH 不屬於語意，本版不重現。
 *
 * 呼叫者：0FF4:0A4B（編成視窗的模板選擇，docs/internals/data/fleet-editor.json「ds:2986，一個艦種 7 bytes、0xFF 結尾；
 *   第一個編得出至少一組的就用，都沒有回 0xFF」，與本函式一致）。
 *
 * 推論等級
 *   已證實（差異測試）：表基底 0x2986、列寬 7、0xFF 結尾、第一個非 0 就回、idx 只取低位元組。
 *   強推論（docs/internals/data/fleet-editor.json）：表的內容是各種編成模板的艦種優先順序。
 * 輸入域：idx 任意 word（取低位元組）；表內容與 0E84 讀的星域存量、艦種表任意。列沒有 0xFF 又沒有任何艦種編得出時，
 *   原版會繞過整個 64 KB 段，超過探針步數上限的案例由探針丟棄。
 */
#include "gin3_reg.h"

extern uint16_t f_0FF4_0E84(Mem *m, uint16_t kind_word);

uint8_t f_0FF4_0B38(Mem *m, uint16_t idx_word)
{
    uint16_t p = (uint16_t)(0x2986u + (uint16_t)((idx_word & 0xFFu) * 7u));

    for (;;) {
        uint8_t k = rd8(m, DS(p));

        if (k == 0xFFu)
            return 0xFFu;
        if (f_0FF4_0E84(m, k) != 0)
            return rd8(m, DS(p));
        p = (uint16_t)(p + 1u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0FF4_0B38(m, a[0]); }
GIN3_REGISTER("0FF4_0B38", 1, run)
