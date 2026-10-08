/*
 * GIN3PS.EXE 1AAB:13BE  從一串艦隊的編制槽裡，收下指定艦種的艦艇池索引，回傳收了幾筆
 *
 * 原版：far 呼叫，六個字的堆疊參數（由低位址到高位址）：
 *   out   ＝ 輸出 word 清單的 far 指標（位移在前、段在後）
 *   ids   ＝ 艦隊編號陣列的 far 指標（byte 陣列）
 *   n     ＝ 艦隊編號個數；原版只用低位元組（al 與 [bp+24h] 的 byte 比較），高位元組被忽略
 *   kind  ＝ 要收的艦種；原版只用低位元組（al 與 [bp+26h] 的 byte 比較）
 * 回 AX ＝ 收了幾筆（16 位元計數）。
 * 呼叫者 1AAB:0FAF 的實際推入順序（反組譯 sub_285CF）：kind、n（word 取自 [bp+1]，高位元組其實是 ids 陣列的第一個
 * byte）、ids、out；n 的高位元組確實是垃圾，只能靠函式自己忽略。
 *
 * 逐條對應反組譯（IDA sub_289DE）：
 *   found ← 0；i ← 0（byte）
 *   while i < (byte)n：
 *     艦隊 ← ds:C074 + low16(ids[0] × 0x7A)；ids 指標 += 1（16 位元回繞），在讀編號之後、內圈之前
 *     for j ← 0；j < 艦隊[+0x0F](byte，每圈重讀)；j++：
 *       idx ← 艦隊[+0x10 + 4j](word)
 *       艦艇 ← (ds:CFB6 : ds:CFB4 + low16(idx × 0x4E))        ; 池基底每圈重讀
 *       若 艦艇[+0x04](byte) == (byte)kind：out[0](word) ← idx；out 指標 += 2；found++
 *     i++
 *   艦隊記錄固定在 DS 段（原版把 DS 存進區域變數當段）。
 *
 * 推論等級
 *   已證實（差異測試的位元組、回傳值與寫入集合）：迴圈結構、欄位位移、n 與 kind 只看低位元組。
 *   強推論（docs/re/346-region-resupply.md §二的骨架、§五「1AAB:13BE 收候選時不篩掉已滿編的」）：
 *     艦隊 +0x0F 是編制組數、+0x10 起每 4 bytes 一槽（前 2 bytes 是池索引）、艦艇 +0x04 是艦種。
 *   與文件的差異：docs/re/346 骨架寫成 `1AAB:13BE(艦種, 清單)`，實際是 (out, ids, n, kind) 四個參數，
 *     清單由 ids/n 兩個參數描述。見回報。
 *
 * 輸入域：n 取 0..12 的低位元組與幾個高位元組不為 0 的值（描述檔）；艦隊編號任意 byte；
 *   艦隊編號超過 31 時讀艦隊表之後的記憶體，本版同樣照做。輸出指標的寫入範圍由描述檔的暫存緩衝區與其後的記憶體構成。
 */
#include "gin3_reg.h"

uint16_t f_1AAB_13BE(Mem *m, uint16_t out_off, uint16_t out_seg, uint16_t ids_off, uint16_t ids_seg, uint16_t n, uint16_t kind)
{
    uint16_t found = 0;
    uint8_t i = 0;

    while (i < (uint8_t)n) {
        uint16_t fleet = (uint16_t)(0xC074 + (uint16_t)(rd8(m, LIN(ids_seg, ids_off)) * 0x7Au));
        uint16_t slot = (uint16_t)(fleet + 0x10);
        uint16_t j = 0;

        ids_off = (uint16_t)(ids_off + 1);
        while (j < rd8(m, DS(fleet + 0x0F))) {
            uint16_t idx = rd16(m, DS(slot));
            uint16_t seg = rd16(m, DS(0xCFB6));
            uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));

            if (rd8(m, LIN(seg, ship + 0x04)) == (uint8_t)kind) {
                wr16(m, LIN(out_seg, out_off), idx);
                out_off = (uint16_t)(out_off + 2);
                found++;
            }
            j++;
            slot = (uint16_t)(slot + 4);
        }
        i++;
    }
    return found;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1AAB_13BE(m, a[0], a[1], a[2], a[3], a[4], a[5]); }
GIN3_REGISTER("1AAB_13BE", 6, run)
