/*
 * 2376:1144  字形緩衝區的樣式變換分派（near，暫存器傳參 CX，DS = DGROUP）
 *
 * 輸入：CX 的高位元組 CH 是樣式旗標（呼叫者 2376:10B7 在呼叫前準備；CL 不讀）。
 * 原版：
 *   test ch,4 → 不為 0 才 call 2376:1155（字形與右移一位的自己 OR，加粗）
 *   test ch,2 → 不為 0 才 call 2376:116C（隔點挖空，網點）
 *   兩者都做時先 1155 後 116C；兩者作用在同一塊緩衝區 ds:829C 起 40 bytes，順序不可交換
 *   （(w | w>>1) & mask 與 (w & mask) | ((w & mask)>>1) 不同）。
 *   1155、116C 開頭 push cx、結尾 pop cx，所以 CH 在第一次呼叫之後仍是進入時的值。
 * 輸出：只有 ds:829C 起 40 bytes 的就地修改。AX、SI 是被呼叫者留下的暫用值，呼叫者
 *   2376:10B7 在 call 之後只 `retn`（asm 讀得），描述檔 ret = none，只比對記憶體寫入。
 *
 * 語意：docs/internals/25-text-rendering.md 第 14 行：樣式旗標 bit 2 加粗、bit 1 網點（強推論，
 *   與 2376:1155、2376:116C 的檔頭一致）；CH 其餘位元本函式不看。
 * 輸入域：CX 的全部 16 位元值。緩衝區內容由突變與 seeds 涵蓋。不涉及指標。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern void f_2376_1155(Mem *m);
extern void f_2376_116C(Mem *m);

void f_2376_1144(Mem *m, uint16_t cx)
{
    uint8_t ch = (uint8_t)(cx >> 8);

    if ((ch & 4u) != 0)
        f_2376_1155(m);
    if ((ch & 2u) != 0)
        f_2376_116C(m);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2376_1144(m, a[0]);
    return 0;
}
GIN3_REGISTER("2376_1144", 1, run)
