/*
 * GIN3PS.EXE 12AA:03BC  依星域重建「戰場艦隊清單」與「戰場艦艇池索引清單」
 *
 * 原版：far 呼叫，一個參數（push 一個字，函式只讀低位元組：mov dl,[bp+20h]）＝ 星域號。無回傳值（AX 是殘值，不比對）。
 * 呼叫者 12AA:026C、185C:001D（docs/re/339 §3.1：戰場回合結算最後一步「12AA:03BC(ds:8F17) 重建戰場表」）。
 *
 * 語意（已證實：docs/re/339 §3.1 的偽碼，與反組譯逐條對過）：
 *   ds:9894 ← 0；ds:9896 ← 0
 *   對艦隊表（ds:C074，一筆 122 bytes）i ＝ 0 … 31：
 *     f[+6] != 星域 → 下一筆；f[+4] == 0xFF → 下一筆
 *     ds:9898[ds:9894] ← i（byte，艦隊清單）；ds:9894 ＋1（word，每次對記憶體做讀改寫）
 *     對 j ＝ 0 … f[+0xF] − 1（f[+0xF] 每圈重讀）：
 *       ds:98BA[ds:9896] ← word f[+0x10 + 4j]（艦艇池索引清單，word）；ds:9896 ＋1
 *     ds:98BA[ds:9896] ← word f[+0x38]；ds:9896 ＋1     （索敵分遣隊那一筆，無條件加入）
 *   兩條寫入指標（位移從 0x9898、0x98BA 起）是 16 位元，各自累加，不依 ds:9894／9896 的內容決定；
 *   段是 DS。計數 ds:9894／9896 是 word，寫入指標與計數互不相干（原版用局部變數與全域計數兩套）。
 *
 * 輸入域：艦隊表 32 筆取基底狀態並由突變與種子改動；f[+0xF] 取 0 … 255（超過 DS 內艦艇池索引容量時，
 *   清單會寫到鄰近的全域變數與艦隊表本身，原版同樣如此，本版逐步照做：每一步都是對記憶體的讀與寫）。
 */
#include "gin3_reg.h"

void f_12AA_03BC(Mem *m, uint16_t side_word)
{
    uint8_t side = (uint8_t)side_word;
    uint16_t bytes_ptr = 0x9898u;
    uint16_t words_ptr = 0x98BAu;
    uint16_t fleet = 0xC074u;
    uint16_t idx;

    wr16(m, DS(0x9894), 0);
    wr16(m, DS(0x9896), 0);
    for (idx = 0; idx < 0x20u; idx++, fleet = (uint16_t)(fleet + 0x7Au)) {
        uint16_t j, slot;

        if (rd8(m, DS(fleet + 6u)) != side)
            continue;
        if (rd8(m, DS(fleet + 4u)) == 0xFFu)
            continue;
        wr8(m, DS(bytes_ptr), (uint8_t)idx);
        bytes_ptr = (uint16_t)(bytes_ptr + 1u);
        wr16(m, DS(0x9894), (uint16_t)(rd16(m, DS(0x9894)) + 1u));
        slot = (uint16_t)(fleet + 0x10u);
        for (j = 0; j < rd8(m, DS(fleet + 0x0Fu)); j++) {
            wr16(m, DS(words_ptr), rd16(m, DS(slot)));
            words_ptr = (uint16_t)(words_ptr + 2u);
            wr16(m, DS(0x9896), (uint16_t)(rd16(m, DS(0x9896)) + 1u));
            slot = (uint16_t)(slot + 4u);
        }
        wr16(m, DS(words_ptr), rd16(m, DS(fleet + 0x38u)));
        words_ptr = (uint16_t)(words_ptr + 2u);
        wr16(m, DS(0x9896), (uint16_t)(rd16(m, DS(0x9896)) + 1u));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_12AA_03BC(m, a[0]);
    return 0;
}
GIN3_REGISTER("12AA_03BC", 1, run)
