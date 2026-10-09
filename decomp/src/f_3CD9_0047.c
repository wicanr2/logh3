/*
 * 3CD9:0047  malloc(size)（Datalight C runtime 的堆積配置器，首次適合、環狀自由串列）
 * 呼叫：far，堆疊傳參：size（word）。回傳 AX ＝ 區塊段、BX ＝ 使用者指標位移（= 區塊表頭位移 + 2）；失敗回 AX＝0、BX＝0。
 * 資料結構與 3CD9:01FC（free）相同：自由區塊 6 位元組（+0 next 位移、+2 next 段、+4 大小，大小含 2 位元組表頭）；
 * 已配置區塊的表頭 2 位元組存大小，使用者指標 ＝ 表頭 + 2。串列起點哨兵在 ds:94E0（next 位移）、94E2（next 段）、
 * 94E4（大小，恆 0），ds:94E6／94E8 是 rover 位移／段；94E6 ＝ 0xFFFF 代表尚未初始化。
 * 演算法（逐步照做）：
 *   0. ds:94E6 ＝ 0xFFFF：把 94E2:94E0 當種子，呼叫 3D17:0003 正規化後存回 94E2:94E0 與 94E8:94E6。
 *   1. 要求大小 r ＝ (size + 3) & 0xFFFE（16 位元）。r < 4 回失敗（含 size ＝ 0 與 size ≥ 0xFFFD）；r < 6 改成 6。
 *   2. 串列的起點是 94E2:94E0 所指節點 head（不是 rover）。cur ＝ head；
 *      重複：nxt ＝ cur.next；nxt.size ≥ r 就找到（前一節點 ＝ cur）；否則 cur ＝ nxt；cur 回到 head 就結束（沒找到）。
 *      （所以第一個被檢查的是 head.next，head 本身最後才檢查。）
 *   3. 找到：rover ← cur；若 r + 6 溢位或 r + 6 > nxt.size 就整塊取走：cur.next ← nxt.next，nxt 表頭（+0）← nxt.size；
 *      否則切開：餘塊起點 ＝ nxt + r，經 3D17:0003 正規化；餘塊.size ＝ nxt.size − r、餘塊.next ＝ nxt.next，
 *      cur.next ← 餘塊，nxt 表頭（+0）← r。回傳 nxt 段、nxt 位移 + 2。
 *   4. 沒找到：向系統要記憶體 3AC4:000D(max(r, 0x200))，回傳 AX:BX ＝ 區塊段:位移（BX ＝ 0xFFFF 代表失敗）；
 *      成功就 free(AX:BX+2)（3CD9:01FC）把它併進串列，free 回 0xFFFF 就失敗；否則回到步驟 1 重試（不再做初始化）。
 * 3AC4:000D 是 DOS 記憶體呼叫（int 21h AH=4Ah／48h），描述檔以替身取代（3AC4:000D 回傳 AX:BX 並在 AX:0 寫入區塊大小；
 * 替身沒有記憶體副作用，所以區塊大小由種子預先寫入，不在驗證範圍）。替身內部行為（int 21h）在這個框架下沒有驗證。
 * 步驟 4 的 BX＝0xFFFF 檢查與 free 對奇數指標的處理重疊（0xFFFF＋2 ＝ 1 是奇數，free 會回 0xFFFF 且不寫任何東西），
 * 所以拿掉這個檢查，輸出也相同（都回 0:0），差別只在少呼叫一次 free；差異測試無法區分，本版照原版保留。
 * 輸入域：全部 16 位元 size；自由串列要是良構；損壞的串列原版可能無窮迴圈，探針以步數上限丟棄。
 * 機器模型：word 存取的兩個位元組各自在段內回繞（off＝0xFFFF 的 word 取同段 off＝0 的高位元組，與 dosgolem 一致）；
 * 寫入落在 F0000 以上（ROM）時無效。這兩點是機器的性質，不是原版函式的邏輯，本檔以 f_3CD9_0047_rd16s／f_3CD9_0047_wr16s／f_3CD9_0047_wr16m 重現。
 * 推論等級：已證實（差異測試；替身行為除外）。
 */
#include "gin3_reg.h"

/*
 * 機器模型：F0000–FFFFF 是 BIOS ROM，寫入無效（真機如此，dosgolem 的 machine.go 也忽略這段的寫入）。
 * gin3_mem.h 的 wr16 是平坦 RAM，這裡本地補上這個語意，只作用於本檔的寫入；讀取兩邊一致，照 RAM 讀。
 * 這是機器的性質，不是原版函式的行為：原版函式只是照指標寫，指標落在 ROM 時由機器吃掉寫入。
 */
static void f_3CD9_0047_wr16m(Mem *m, uint32_t lin, uint16_t v)
{
    lin &= 0xFFFFFu;
    if (lin < 0xF0000u)
        m->b[lin] = (uint8_t)v;
    lin = (lin + 1) & 0xFFFFFu;
    if (lin < 0xF0000u)
        m->b[lin] = (uint8_t)(v >> 8);
}

/* 段內存取：兩個位元組的位移各自 16 位元回繞（8086 的 word 存取在 off＝0xFFFF 時第二個位元組取同段的 off＝0，dosgolem 的 cpu.go 同） */
static uint16_t f_3CD9_0047_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1))) << 8));
}

static void f_3CD9_0047_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    uint32_t a = LIN(seg, off), b = LIN(seg, (uint16_t)(off + 1));

    if (a < 0xF0000u)
        m->b[a] = (uint8_t)v;
    if (b < 0xF0000u)
        m->b[b] = (uint8_t)(v >> 8);
}

extern uint64_t f_3D17_0003(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint16_t f_3CD9_01FC(Mem *m, uint16_t off, uint16_t seg);

uint32_t f_3CD9_0047(Mem *m, uint16_t size)
{
    uint16_t ax, bx, cx, dx, si, di, bp;
    uint64_t r;

    if (rd16(m, DS(0x94E6)) == 0xFFFF) {
        ax = rd16(m, DS(0x94E2));
        bx = rd16(m, DS(0x94E0));
        r = f_3D17_0003(ax, bx, 0, 0);
        ax = (uint16_t)(r >> 48);
        bx = (uint16_t)(r >> 32);
        f_3CD9_0047_wr16m(m, DS(0x94E2), ax);
        f_3CD9_0047_wr16m(m, DS(0x94E0), bx);
        f_3CD9_0047_wr16m(m, DS(0x94E8), ax);
        f_3CD9_0047_wr16m(m, DS(0x94E6), bx);
    }
    for (;;) {
        bp = (uint16_t)((uint16_t)(size + 3) & 0xFFFEu);
        if (bp < 4)
            return 0;
        if (bp < 6)
            bp = 6;
        dx = rd16(m, DS(0x94E2));
        di = rd16(m, DS(0x94E0));
        cx = dx;
        si = di;
        for (;;) {
            ax = f_3CD9_0047_rd16s(m, cx, si + 2);
            bx = f_3CD9_0047_rd16s(m, cx, si);
            if (bp <= f_3CD9_0047_rd16s(m, ax, bx + 4))
                goto found;
            cx = ax;
            si = bx;
            if (cx == dx && si == di)
                break;
        }
        if (bp < 0x200)
            bp = 0x200;
        {
            uint16_t a[1];
            uint32_t sr;

            a[0] = bp;
            sr = gin3_stub(m, "3AC4_000D", 1, a);
            ax = (uint16_t)(sr >> 16);
            bx = (uint16_t)sr;
        }
        if (bx == 0xFFFF)
            return 0;
        bx = (uint16_t)(bx + 2);
        if (f_3CD9_01FC(m, bx, ax) == 0xFFFF)
            return 0;
    }
found:
    f_3CD9_0047_wr16m(m, DS(0x94E8), cx);
    f_3CD9_0047_wr16m(m, DS(0x94E6), si);
    {
        uint32_t t = (uint32_t)bp + 6u;

        if (t > 0xFFFFu || (uint16_t)t > f_3CD9_0047_rd16s(m, ax, bx + 4)) {
            dx = f_3CD9_0047_rd16s(m, ax, bx + 2);
            di = f_3CD9_0047_rd16s(m, ax, bx);
            f_3CD9_0047_wr16s(m, cx, si + 2, dx);
            f_3CD9_0047_wr16s(m, cx, si, di);
            dx = f_3CD9_0047_rd16s(m, ax, bx + 4);
            f_3CD9_0047_wr16s(m, ax, bx, dx);
        } else {
            uint16_t rem_seg, rem_off, rest;

            dx = ax;
            di = bx;
            r = f_3D17_0003(ax, (uint16_t)(bx + bp), cx, dx);
            rem_seg = (uint16_t)(r >> 48);
            rem_off = (uint16_t)(r >> 32);
            /* xchg ax,dx；xchg bx,di：ax:bx ＝ 找到的區塊，dx:di ＝ 正規化後的餘塊 */
            dx = rem_seg;
            di = rem_off;
            rest = (uint16_t)(f_3CD9_0047_rd16s(m, ax, bx + 4) - bp);
            f_3CD9_0047_wr16s(m, dx, di + 4, rest);
            f_3CD9_0047_wr16s(m, dx, di, f_3CD9_0047_rd16s(m, ax, bx));
            f_3CD9_0047_wr16s(m, dx, di + 2, f_3CD9_0047_rd16s(m, ax, bx + 2));
            f_3CD9_0047_wr16s(m, cx, si + 2, dx);
            f_3CD9_0047_wr16s(m, cx, si, di);
            f_3CD9_0047_wr16s(m, ax, bx, bp);
        }
    }
    bx = (uint16_t)(bx + 2);
    return ((uint32_t)ax << 16) | bx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3CD9_0047(m, a[0]);
}
GIN3_REGISTER("3CD9_0047", 1, run)
