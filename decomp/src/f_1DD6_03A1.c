/*
 * GIN3PS.EXE 1DD6:03A1  把一串艦隊的編制槽（艦艇池索引）依序收進一張 word 清單，回傳收了幾筆
 *
 * 原版：far 呼叫，五個字的堆疊參數（由低位址到高位址）：
 *   out  ＝ 輸出清單的 far 指標（位移在前、段在後）
 *   ids  ＝ 艦隊編號陣列的 far 指標（byte 陣列）
 *   n    ＝ 艦隊編號個數（完整 word，無號比較）
 * 回 AX ＝ 收了幾筆（16 位元計數）。
 *
 * 逐條對應反組譯（IDA sub_2AC71）：
 *   total ← 0；i ← 0
 *   while i < n：
 *     艦隊 ← ds:C074 + low16(ids[0] × 0x7A)；ids 指標 += 1 在外圈尾端（16 位元回繞）
 *     for j ← 0；j < 艦隊[+0x0F](byte，每圈重讀)；j++：
 *       out[0](word) ← 艦隊[+0x10 + 4j](word)；out 指標 += 2；total++
 *     i++
 *   艦隊編號是 byte，乘 122 不會超過 16 位元（最大 31110），艦隊記錄固定在 DS 段（原版把 DS 存進區域變數當段）。
 *
 * 推論等級
 *   已證實（差異測試的位元組、回傳值與寫入集合）：迴圈結構、欄位位移、n 是完整 word。
 *   強推論（docs/internals/38-battle-units.md 第 400 行「1DD6:030F→03A1 收參戰名冊中全部編制一般分隊」、
 *     docs/internals/64-ship-pool.md）：艦隊 +0x0F 是編制組數，+0x10 起每 4 bytes 一槽、前 2 bytes 是艦艇池索引。
 *   未知：槽的後 2 bytes 的語意（本函式不讀）。
 *
 * 輸入域：n 取 0..12（描述檔；n 到 65535 時多半超過 200,000 步被探針丟棄）；艦隊編號任意 byte，
 *   編號超過 31 時讀艦隊表之後的記憶體，本版同樣照做。輸出指標的寫入範圍由描述檔的暫存緩衝區與其後的記憶體構成。
 */
#include "gin3_reg.h"

uint16_t f_1DD6_03A1(Mem *m, uint16_t out_off, uint16_t out_seg, uint16_t ids_off, uint16_t ids_seg, uint16_t n)
{
    uint16_t total = 0;
    uint16_t i = 0;

    while (i < n) {
        uint16_t fleet = (uint16_t)(0xC074 + (uint16_t)(rd8(m, LIN(ids_seg, ids_off)) * 0x7Au));
        uint16_t slot = (uint16_t)(fleet + 0x10);
        uint16_t j = 0;

        while (j < rd8(m, DS(fleet + 0x0F))) {
            wr16(m, LIN(out_seg, out_off), rd16(m, DS(slot)));
            out_off = (uint16_t)(out_off + 2);
            total++;
            j++;
            slot = (uint16_t)(slot + 4);
        }
        i++;
        ids_off = (uint16_t)(ids_off + 1);
    }
    return total;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1DD6_03A1(m, a[0], a[1], a[2], a[3], a[4]); }
GIN3_REGISTER("1DD6_03A1", 5, run)
