/*
 * 2376:116C  字形緩衝區的網點變換（near，無參數，DS = DGROUP）
 *
 * 原版：對 ds:829C 起的 20 個 word 逐一 `and [si], ax`，ax 起始 0x5555，每個 word 後 `not ax`，
 *   所以偶數列（第 0、2、...）與 0x5555 相與，奇數列與 0xAAAA 相與。20 個 word 是偶數個，
 *   迴圈結束時 ax 又回到 0x5555。
 *   暫存器：CX 由 push／pop 保存；SI 結束時是 0x82C4、AX 是 0x5555，這兩個是暫用值。呼叫者 2376:1144
 *   之後只 `retn`，不使用它們，所以描述檔 ret = none，只比對記憶體寫入。
 * 輸出：ds:829C 起 40 bytes 的就地修改。不讀寫其他位置。
 *
 * 語意：docs/internals/25-text-rendering.md 第 14 行、31-text-plane.md 第 57 行、docs/re/138-window-frame-tiles.md
 *   第 202 行都把 2376:116C 記為「樣式旗標 bit 1 的網點，隔點挖空，畫成灰掉的樣子」，與本版相符（已證實：
 *   本檔與原版的差異測試逐 byte 相同）。25-text-rendering.md 第 14 行寫「每列一個 word，高位在前、bit15 在最左」，
 *   此函式只做 16 位元 `and`，位元順序不影響結果。
 * 呼叫者：2376:1144（只在顏色高位元組的 bit 1 設定時呼叫）。
 *
 * 輸入域：ds:829C 起 40 bytes 的全部值。不涉及指標，不會讀到視訊記憶體區。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；字形緩衝區的用途是已證實的文件結論。
 */
#include "gin3_reg.h"

void f_2376_116C(Mem *m)
{
    uint16_t mask = 0x5555u;
    uint16_t off = 0x829Cu;
    int i;

    for (i = 0; i < 20; i++) {
        wr16(m, DS(off), (uint16_t)(rd16(m, DS(off)) & mask));
        off = (uint16_t)(off + 2u);
        mask = (uint16_t)~mask;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2376_116C(m);
    return 0;
}
GIN3_REGISTER("2376_116C", 0, run)
