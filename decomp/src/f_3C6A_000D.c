/*
 * 3C6A:000D  qsort(base, n, size, cmp)（Datalight C runtime；實際演算法是梳狀排序 comb sort，不是快速排序）
 * 呼叫：far，堆疊傳參（由低位址到高位址）：base 位移、base 段、n、size、cmp 位移、cmp 段（cmp 是 far 函式指標）。
 * 無回傳值。
 * 演算法（逐步照做；佐證 docs/re/368 的結論）：
 *   n <= 1 直接返回。選擇交換常式：size 為奇數用「逐位元組交換」(3C6A:0158)，ds:F37C ＝ size；
 *   size 為偶數用「逐字交換」(3C6A:011E)，ds:F37C ＝ size >> 1。（ds:F37C 是交換常式的次數，這裡會寫入。）
 *   total ＝ n × size（低 16 位元）；gap ＝ n。重複：
 *     gap ← ((gap × 10) mod 64K) / 13（乘法只留低 16 位元）；gap ＝ 0 改 1；gap ＝ 9 或 10 改 11。
 *     step ＝ gap × size（低 16 位元）；rest ＝ total − step（16 位元）。rest ≠ 0 才做一趟：
 *       A ＝ base、B ＝ base ＋ step（位移 16 位元回繞）、limit ＝ rest ＋ base 位移（16 位元）。
 *       do { r ＝ cmp(A, B)；r 為有號正數（AX > 0）就交換 A、B 並計數；A、B 位移各加 size } while (A 位移 < limit，無號)。
 *     本趟有交換、或 gap > 1 就再來一趟；否則結束。
 *   cmp 以 4 個堆疊字呼叫：A 位移、A 段、B 位移、B 段。這個比較常式是呼叫者提供的間接呼叫目標，描述檔以替身取代，
 *   由 gin3_stub 重放（比對呼叫順序與參數）；回傳值只看 AX 的有號正負。
 *   交換常式 3C6A:011E／3C6A:0158 不在函式台帳內（IDA 沒有切成獨立函式），本檔內建它們的行為：
 *   對 A、B 各讀一個單位、互換寫回，單位是 word（011E）或 byte（0158），次數 ＝ ds:F37C（do-while，0 代表 65536 次）。
 *   其寫入效果由整個排序的寫入集合比對間接證實。
 * 輸入域：n × size 與位移不超過 64 KB 的一般情形；size ＝ 0 時 rest ＝ 0，不做任何比較。比較常式不一致導致不收斂時，
 * 原版不會結束，探針以步數上限丟棄。
 * 推論等級：已證實（差異測試）。
 */
#include <stdio.h>

#include "gin3_reg.h"

/* 段內的 word 存取：兩個位元組的位移各自 16 位元回繞（off＝0xFFFF 的 word 取同段 off＝0 的高位元組，與 dosgolem 的 cpu.go 一致） */
static uint16_t f_3C6A_000D_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1))) << 8));
}

static void f_3C6A_000D_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1)), (uint8_t)(v >> 8));
}

/* 3C6A:011E（word）／3C6A:0158（byte）：交換 A、B 各 count 個單位，位移 16 位元回繞 */
static void f_3C6A_000D_swap_elems(Mem *m, uint16_t seg, uint16_t aoff, uint16_t boff, int is_word)
{
    uint16_t cnt = rd16(m, DS(0xF37C));

    do {
        if (is_word) {
            uint16_t a = f_3C6A_000D_rd16s(m, seg, aoff);
            uint16_t b = f_3C6A_000D_rd16s(m, seg, boff);

            f_3C6A_000D_wr16s(m, seg, aoff, b);
            f_3C6A_000D_wr16s(m, seg, boff, a);
            aoff = (uint16_t)(aoff + 2);
            boff = (uint16_t)(boff + 2);
        } else {
            uint8_t a = rd8(m, LIN(seg, aoff));
            uint8_t b = rd8(m, LIN(seg, boff));

            wr8(m, LIN(seg, aoff), b);
            wr8(m, LIN(seg, boff), a);
            aoff = (uint16_t)(aoff + 1);
            boff = (uint16_t)(boff + 1);
        }
        cnt--;
    } while (cnt != 0);
}

void f_3C6A_000D(Mem *m, uint16_t base_off, uint16_t base_seg, uint16_t n, uint16_t size, uint16_t cmp_off,
                 uint16_t cmp_seg)
{
    uint16_t gap, swapped, total, step, rest, limit, a_off, b_off;
    int is_word = !(size & 1);
    char id[16];

    if (n <= 1)
        return;
    wr16(m, DS(0xF37C), is_word ? (uint16_t)(size >> 1) : size);
    total = (uint16_t)(n * size);
    gap = n;
    snprintf(id, sizeof id, "%04X_%04X", cmp_seg, cmp_off);
    for (;;) {
        gap = (uint16_t)((uint16_t)(gap * 10u) / 13u);
        if (gap == 0)
            gap = 1;
        else if (gap == 9 || gap == 10)
            gap = 11;
        swapped = 0;
        step = (uint16_t)(gap * size);
        rest = (uint16_t)(total - step);
        if (rest != 0) {
            a_off = base_off;
            b_off = (uint16_t)(base_off + step);
            limit = (uint16_t)(rest + base_off);
            do {
                uint16_t args[4];
                uint32_t r;

                args[0] = a_off;
                args[1] = base_seg;
                args[2] = b_off;
                args[3] = base_seg;
                r = gin3_stub(m, id, 4, args);
                if ((int16_t)(uint16_t)r > 0) {
                    f_3C6A_000D_swap_elems(m, base_seg, a_off, b_off, is_word);
                    swapped++;
                }
                a_off = (uint16_t)(a_off + size);
                b_off = (uint16_t)(b_off + size);
            } while (a_off < limit);
        }
        if (swapped != 0)
            continue;
        if (gap > 1)
            continue;
        return;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3C6A_000D(m, a[0], a[1], a[2], a[3], a[4], a[5]);
    return 0;
}
GIN3_REGISTER("3C6A_000D", 6, run)
