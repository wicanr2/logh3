/*
 * 2376:1155  字形緩衝區的加粗變換（near，無參數，DS = DGROUP）
 *
 * 原版：對 ds:829C 起的 20 個 word 逐一做
 *     ax = [si]; xchg ah,al; shr ax,1; xchg ah,al; or [si], ax
 *   也就是把 word 當成「第一個 byte 在左、第二個 byte 在右」的 16 位元圖樣，整列右移 1 位
 *   （第一個 byte 的 bit 0 流進第二個 byte 的 bit 7，最左邊補 0），再與原圖樣 OR 起來。
 *   本版直接寫成 xchg／shr／xchg 的對應運算，不另外化簡成 byte 級公式。
 * 輸出：ds:829C 起 40 bytes 的就地修改。不讀寫其他位置。
 *   暫存器：CX 由 push／pop 保存；SI 結束時是 0x82C4，AX 是最後一列的暫用值，呼叫者 2376:1144 之後只
 *   `retn`，不使用它們，所以描述檔 ret = none，只比對記憶體寫入。
 *
 * 語意：docs/internals/25-text-rendering.md 第 14 與 124 行、31-text-plane.md 第 58 行、docs/re/138 第 203 行
 *   都把 2376:1155 記為「樣式旗標 bit 2 的加粗：字形每列與右移一位 OR 起來」，與本版相符（已證實）。
 *   25-text-rendering.md 第 14 行寫「每列一個 word，高位在前、bit15 在最左」：這與本函式的 xchg 一致的讀法是
 *   「word 的第一個 byte 是左半，byte 內 bit7 在最左」，文件的「word」指的是兩個 byte 依序排成的 16 位元圖樣。
 * 呼叫者：2376:1144（只在顏色高位元組的 bit 2 設定時呼叫）。
 *
 * 輸入域：ds:829C 起 40 bytes 的全部值。不涉及指標，不會讀到視訊記憶體區。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

void f_2376_1155(Mem *m)
{
    uint16_t off = 0x829Cu;
    int i;

    for (i = 0; i < 20; i++) {
        uint16_t w = rd16(m, DS(off));
        uint16_t sw = (uint16_t)((w << 8) | (w >> 8)); /* xchg ah,al */

        sw = (uint16_t)(sw >> 1);                       /* shr ax,1 */
        sw = (uint16_t)((sw << 8) | (sw >> 8));         /* xchg ah,al */
        wr16(m, DS(off), (uint16_t)(w | sw));           /* or [si], ax */
        off = (uint16_t)(off + 2u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2376_1155(m);
    return 0;
}
GIN3_REGISTER("2376_1155", 0, run)
