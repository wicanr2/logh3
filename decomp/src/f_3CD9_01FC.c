/*
 * 3CD9:01FC  free(p)（Datalight C runtime 的堆積釋放，把區塊併回自由串列）
 * 呼叫：far，堆疊傳參：p 位移、p 段（兩個 word，位移在低位址）。回傳 AX：0 成功（含 p＝0:0 的空操作）、
 * 0xFFFF 失敗。BX、CX、DX、SI、DI、BP 不保證，DS 由 push／pop 保留。
 * 資料結構（docs/internals/03 §5，本版重新量過）：
 *   自由區塊（6 位元組）：+0 下一個自由區塊的位移、+2 下一個的段、+4 這個區塊的大小（位元組，含 2 位元組表頭）。
 *   已配置區塊：表頭 2 位元組＝大小，使用者指標 p＝表頭＋2。所以 free 取 p−2 當表頭 hdr。
 *   串列起點（哨兵）是 DGROUP 內的 ds:94E0（next 位移）、94E2（next 段）、94E4（大小，恆 0），
 *   正規化後的位址是 5E01:0000；ds:94E6／94E8 是 rover（上次插入處的前一節點）的位移／段。
 * 演算法：
 *   1. p＝0:0 回 0；p 位移為奇數回 0xFFFF；hdr＝p−2。
 *   2. 起點 cur：rover < hdr 就從 rover 開始，否則從 54B3:94E0（DGROUP:94E0，沒有正規化的哨兵寫法）開始。
 *   3. 沿串列找插入處：cur >= hdr 回 0xFFFF（重複釋放或位址在 cur 之前）；next > hdr 或 next <= cur（繞回）
 *      就在 cur 之後插入，否則 cur ← next。
 *   4. 插入：rover ← cur；hdr 的大小搬到 +4，hdr.next ← cur.next，cur.next ← hdr。
 *   5. 與前一塊併：cur 的結尾（正規化 cur＋cur.size，呼叫 3D17:0003）等於 hdr 就以 cur 為基準，否則以 hdr 為基準。
 *   6. 迴圈：基準 b 的結尾等於 b.next 就併入（大小相加、b.next ← b.next.next）；
 *      大小相加進位時改成：在 b 之後 0xFFF 段處（b 段＋0xFFF、同位移）建新節點，其大小為
 *      (相加結果 mod 64K) − 0xFFF0（即加 0x10），b.size ← 0xFFF0，b.next ← 新節點，基準 ← 新節點。
 *      結尾不等於 next 就回 0。
 * 所有位址比較是 (段, 位移) 字典序的無號比較，不是線性位址；段先比，段相同才比位移。
 * 記憶體讀寫的順序照原版逐步照做（區塊與串列節點可能重疊，順序會影響結果）。
 * 輸入域：全部 16 位元值；串列要是良構（會終止）；損壞的串列原版會無窮迴圈，本版不涵蓋（探針以步數上限丟棄）。
 * 機器模型：word 存取的兩個位元組各自在段內回繞（off＝0xFFFF 的 word 取同段 off＝0 的高位元組，與 dosgolem 一致）；
 * 寫入落在 F0000 以上（ROM）時無效。這兩點是機器的性質，不是原版函式的邏輯，本檔以 f_3CD9_01FC_rd16s／f_3CD9_01FC_wr16s／f_3CD9_01FC_wr16m 重現。
 * 推論等級：已證實（差異測試，串列結構由 docs/internals/03 與反組譯交叉確認）。
 */
#include "gin3_reg.h"

/*
 * 機器模型：F0000–FFFFF 是 BIOS ROM，寫入無效（真機如此，dosgolem 的 machine.go 也忽略這段的寫入）。
 * gin3_mem.h 的 wr16 是平坦 RAM，這裡本地補上這個語意，只作用於本檔的寫入；讀取兩邊一致，照 RAM 讀。
 * 這是機器的性質，不是原版函式的行為：原版函式只是照指標寫，指標落在 ROM 時由機器吃掉寫入。
 */
static void f_3CD9_01FC_wr16m(Mem *m, uint32_t lin, uint16_t v)
{
    lin &= 0xFFFFFu;
    if (lin < 0xF0000u)
        m->b[lin] = (uint8_t)v;
    lin = (lin + 1) & 0xFFFFFu;
    if (lin < 0xF0000u)
        m->b[lin] = (uint8_t)(v >> 8);
}

/* 段內存取：兩個位元組的位移各自 16 位元回繞（8086 的 word 存取在 off＝0xFFFF 時第二個位元組取同段的 off＝0，dosgolem 的 cpu.go 同） */
static uint16_t f_3CD9_01FC_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1))) << 8));
}

static void f_3CD9_01FC_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    uint32_t a = LIN(seg, off), b = LIN(seg, (uint16_t)(off + 1));

    if (a < 0xF0000u)
        m->b[a] = (uint8_t)v;
    if (b < 0xF0000u)
        m->b[b] = (uint8_t)(v >> 8);
}

extern uint64_t f_3D17_0003(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

/* (aseg:aoff) 與 (bseg:boff) 的字典序比較（無號）：負＝a 在前，0＝相等，正＝a 在後 */
static int f_3CD9_01FC_cmp_so(uint16_t aseg, uint16_t aoff, uint16_t bseg, uint16_t boff)
{
    if (aseg != bseg)
        return aseg < bseg ? -1 : 1;
    if (aoff != boff)
        return aoff < boff ? -1 : 1;
    return 0;
}

uint16_t f_3CD9_01FC(Mem *m, uint16_t off, uint16_t seg)
{
    uint16_t dx = seg, di = off, cx, si, ax, bx, bp;
    uint64_t r;

    if ((uint16_t)(dx | di) == 0)
        return 0;
    if (di & 1)
        return 0xFFFF;
    di = (uint16_t)(di - 2);
    cx = rd16(m, DS(0x94E8));
    si = rd16(m, DS(0x94E6));
    if (!(f_3CD9_01FC_cmp_so(cx, si, dx, di) < 0)) {
        cx = GIN3_DS;
        si = 0x94E0;
    }
    for (;;) {
        ax = f_3CD9_01FC_rd16s(m, cx, si + 2);
        bx = f_3CD9_01FC_rd16s(m, cx, si);
        if (f_3CD9_01FC_cmp_so(cx, si, dx, di) >= 0)
            return 0xFFFF;
        if (f_3CD9_01FC_cmp_so(ax, bx, dx, di) > 0)
            break;
        if (f_3CD9_01FC_cmp_so(ax, bx, cx, si) <= 0)
            break;
        cx = ax;
        si = bx;
    }
    f_3CD9_01FC_wr16m(m, DS(0x94E8), cx);
    f_3CD9_01FC_wr16m(m, DS(0x94E6), si);
    bp = f_3CD9_01FC_rd16s(m, dx, di);
    f_3CD9_01FC_wr16s(m, dx, di + 4, bp);
    f_3CD9_01FC_wr16s(m, dx, di + 2, ax);
    f_3CD9_01FC_wr16s(m, dx, di, bx);
    f_3CD9_01FC_wr16s(m, cx, si + 2, dx);
    f_3CD9_01FC_wr16s(m, cx, si, di);
    ax = cx;
    bx = (uint16_t)(si + f_3CD9_01FC_rd16s(m, cx, si + 4));
    r = f_3D17_0003(ax, bx, cx, dx);
    ax = (uint16_t)(r >> 48);
    bx = (uint16_t)(r >> 32);
    if (!(ax == dx && bx == di)) {
        cx = dx;
        si = di;
    }
    for (;;) {
        uint32_t sum;

        ax = cx;
        bx = (uint16_t)(si + f_3CD9_01FC_rd16s(m, cx, si + 4));
        r = f_3D17_0003(ax, bx, cx, dx);
        ax = (uint16_t)(r >> 48);
        bx = (uint16_t)(r >> 32);
        dx = f_3CD9_01FC_rd16s(m, cx, si + 2);
        di = f_3CD9_01FC_rd16s(m, cx, si);
        if (!(ax == dx && bx == di))
            return 0;
        bp = f_3CD9_01FC_rd16s(m, cx, si + 4);
        sum = (uint32_t)bp + f_3CD9_01FC_rd16s(m, dx, di + 4);
        bp = (uint16_t)sum;
        if (sum <= 0xFFFFu) {
            f_3CD9_01FC_wr16s(m, cx, si + 4, bp);
            ax = f_3CD9_01FC_rd16s(m, dx, di);
            f_3CD9_01FC_wr16s(m, cx, si, ax);
            ax = f_3CD9_01FC_rd16s(m, dx, di + 2);
            f_3CD9_01FC_wr16s(m, cx, si + 2, ax);
        } else {
            ax = (uint16_t)(cx + 0xFFF);
            bx = f_3CD9_01FC_rd16s(m, dx, di);
            f_3CD9_01FC_wr16s(m, ax, si, bx);
            bx = f_3CD9_01FC_rd16s(m, dx, di + 2);
            f_3CD9_01FC_wr16s(m, ax, si + 2, bx);
            bp = (uint16_t)(bp - 0xFFF0);
            f_3CD9_01FC_wr16s(m, ax, si + 4, bp);
            f_3CD9_01FC_wr16s(m, cx, si + 2, ax);
            f_3CD9_01FC_wr16s(m, cx, si, si);
            f_3CD9_01FC_wr16s(m, cx, si + 4, 0xFFF0);
            cx = ax;
        }
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3CD9_01FC(m, a[0], a[1]);
}
GIN3_REGISTER("3CD9_01FC", 2, run)
