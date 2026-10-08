/*
 * 2376:02B0  矩形與物件圖塊範圍是否完全分離（near，暫存器傳參）
 *
 * 輸入：ES:SI = 顯示物件（讀 +1A、+1B、+1C、+1D 四個 signed byte）；
 *       BP = 呼叫者 2376:01E6 的框架（enter 8,0），讀 SS:[BP-2]、[BP-4]、[BP-6]、[BP-8] 四個有號字。
 * 輸出：AL = 1 表示完全分離，AL = 0 表示有重疊。
 *       四個 signed byte 與四個字的比較都是有號比較（cbw 後 cmp ax,[bp-n]）。
 *       AH 是最後一次 cbw 留下的符號延伸（0x00 或 0xFF），呼叫端只用 AL，
 *       但差異測試只能比整個 AX，所以這裡把 AH 也算出來；它不是函式的語意輸出。
 *   順序（短路）：
 *     [+1C] < [BP-2]  → 1
 *     [+1D] < [BP-4]  → 1
 *     [+1A] > [BP-6]  → 1
 *     [+1B] > [BP-8]  → 1
 *     其餘 → 0
 *
 * 語意：「以帶號比較圖塊範圍與物件 +1A..+1D 邊界，任一完全分離回 1，其餘回 0」，已證實，
 *   見 docs/internals/data/battle-spotting-display-semantics.json（original_location 2376:02B0，
 *   61 案原版完整雙跑）。欄位名稱（+1A..+1D 為 X 最小、Y 最小、X 最大、Y 最大圖塊，
 *   [BP-2..-8] 為矩形的 X 最小、Y 最小、X 最大、Y 最大）是對照 01E6 的用法推出的，強推論。
 *
 * 限制（差異測試的覆蓋範圍）：
 *   diffprobe 的 kind=reg 不支援 es 與 ss，ES 與 SS 都是基底狀態留下的值（實測兩者相同，
 *   都是 0x63EB），run 包裝函式把它們當常數傳入。所以：
 *   - 驗到：兩個基底位址的組合下，比較順序、符號、邊界、AH 殘值、位移 16 位元回繞（SI = 0xFFE5、BP = 4）。
 *   - 沒驗到：ES 與 SS 不同值的情形，以及「某筆讀取是走 ES 還是走 SS」的區別
 *     （ES 與 SS 相等，兩者對調不會被發現；走 DS 的錯誤會被發現，因為 DS 是 0x54B3）。
 *     這兩項靠反組譯：位元組讀取帶 es: 前綴，[bp-n] 預設走 SS。
 * 輸入域：SI、BP 各取幾個代表值（含回繞），記憶體內容由突變涵蓋全部 byte／字值。
 */
#include "gin3_reg.h"

/* cbw：signed byte 延伸成 16 位元 */
static uint16_t cbw(uint8_t b) { return (uint16_t)(int16_t)(int8_t)b; }

uint16_t f_2376_02B0(Mem *m, uint16_t es, uint16_t ss, uint16_t si, uint16_t bp)
{
    uint16_t ax;

    ax = cbw(rd8(m, LIN(es, si + 0x1C)));
    if ((int16_t)ax < (int16_t)rd16(m, LIN(ss, bp - 2)))
        return (uint16_t)((ax & 0xFF00u) | 1u);
    ax = cbw(rd8(m, LIN(es, si + 0x1D)));
    if ((int16_t)ax < (int16_t)rd16(m, LIN(ss, bp - 4)))
        return (uint16_t)((ax & 0xFF00u) | 1u);
    ax = cbw(rd8(m, LIN(es, si + 0x1A)));
    if ((int16_t)ax > (int16_t)rd16(m, LIN(ss, bp - 6)))
        return (uint16_t)((ax & 0xFF00u) | 1u);
    ax = cbw(rd8(m, LIN(es, si + 0x1B)));
    if ((int16_t)ax > (int16_t)rd16(m, LIN(ss, bp - 8)))
        return (uint16_t)((ax & 0xFF00u) | 1u);
    return (uint16_t)(ax & 0xFF00u);
}

/* fcmd-base.state 的 ES 與 SS（diffprobe 無法設定這兩個暫存器） */
#define BASE_ES 0x63EBu
#define BASE_SS 0x63EBu

static uint64_t run(Mem *m, const uint16_t *a) { return f_2376_02B0(m, BASE_ES, BASE_SS, a[0], a[1]); }
GIN3_REGISTER("2376_02B0", 2, run)
