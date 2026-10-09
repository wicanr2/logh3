/*
 * GIN3PS.EXE 0B20:003B  敵艦隊清單重建：各艦隊算出「索敵等級」寫進 +0x70，等級非 0 的編號收進清單，回傳筆數
 *
 * 原版：far 呼叫，兩個參數；序言 push bp／si、sub sp,14h，所以
 *   ids 在 [bp+1Ch]（位移）／[bp+1Eh]（段）：艦隊編號（byte）陣列；n 在 [bp+20h]（word，只用低位元組）：陣列長度。
 * 回傳 AX ＝ 清單筆數（AH ＝ 0；迴圈結束前最後一次寫 AX 的是 `xor ah,ah`，n ＝ 0 時 AX 從 `xor ax,ax` 起就是 0）。
 *
 * 反組譯逐條對應：
 *   out ← ds:95B6（清單的下一個寫入位置）、count ← 0、i ← 0
 *   每圈（i < n，byte 無號比較）：
 *     f ← ids[i]；艦隊 ← ds:C074 ＋ f × 0x7A（16 位元）
 *     max ← 0；對 j ← 0 起、j < 艦隊[+0x0F]（每圈重讀）：
 *         p ← word 艦隊[0x10 ＋ 4j]（艦艇池索引）
 *         艦艇 ← far 指標 ds:CFB4／ds:CFB6（每圈重讀）＋ p × 0x4E（16 位元）
 *         lv ← byte 艦艇[+0x17]；lv > max（無號）就 max ← lv
 *     艦隊[+0x70] ← word max（零延伸）
 *     max != 0：f2 ← ids[i]（寫入 +0x70 之後重讀）；ds:[out] ← f2；out++；count++
 *     ids 位移 + 1（16 位元）
 *   回 count
 *
 * 別名：ids 若指到艦隊的 +0x70 或清單 ds:95B6 起，重讀的 ids[i] 可能已被前面的寫入改掉，後面的 ids[k]
 *   也可能是清單剛寫進去的值。本版逐步照原版順序做。
 *
 * 推論等級
 *   已證實（差異測試）：上列結構、常數、讀寫寬度、重讀的位置。
 *   強推論（docs/internals/33-fleet-table.md 第 76 行、docs/re/325）：艦隊 +0x70 在戰略地圖上存的是「索敵等級」
 *     （各組艦艇 +0x17 取最大），右下角「敵艦隊」分頁重建清單時由此函式寫入，列印與點列都讀它；
 *     ds:95B6 是該分頁的艦隊編號清單。
 *   未知：艦艇 +0x17 的名稱（假說：索敵能力）；清單長度上限（此函式不檢查）。
 *
 * 輸入域：ids 指到任意位址、n 任意 byte；艦隊編號超過 31 時讀到艦隊表之外的資料（16 位元回繞），
 *   本版照原版運算，描述檔的隨機編號以 0 到 31 為主。
 *   字的存取跨越位移 0xFFFF 時，本版與探針都按線性位址連續處理，不涵蓋真機上回繞到段首的行為。
 */
#include "gin3_reg.h"

uint16_t f_0B20_003B(Mem *m, uint16_t ids_off, uint16_t ids_seg, uint16_t n_word)
{
    uint8_t n = (uint8_t)n_word;
    uint16_t out = 0x95B6;
    uint8_t count = 0;
    uint8_t i = 0;

    while (i < n) {
        uint8_t f = rd8(m, LIN(ids_seg, ids_off));
        uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(f * 0x7Au));
        uint16_t slot = (uint16_t)(fleet + 0x10u);
        uint8_t max = 0;
        uint8_t j = 0;

        while (j < rd8(m, DS(fleet + 0x0F))) {
            uint16_t p = rd16(m, DS(slot));
            uint16_t pool_seg = rd16(m, DS(0xCFB6));
            uint16_t pool_off = rd16(m, DS(0xCFB4));
            uint16_t ship = (uint16_t)(pool_off + (uint16_t)(p * 0x4Eu));
            uint8_t lv = rd8(m, LIN(pool_seg, ship + 0x17));

            if (lv > max)
                max = lv;
            j++;
            slot = (uint16_t)(slot + 4);
        }
        wr16(m, DS(fleet + 0x70), max);
        if (max != 0) {
            wr8(m, DS(out), rd8(m, LIN(ids_seg, ids_off)));
            out = (uint16_t)(out + 1);
            count++;
        }
        i++;
        ids_off = (uint16_t)(ids_off + 1);
    }
    return count;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0B20_003B(m, a[0], a[1], a[2]); }
GIN3_REGISTER("0B20_003B", 3, run)
