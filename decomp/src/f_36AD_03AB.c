/*
 * 36AD:03AB  圖塊矩形的螢幕裁切，算出左右裁切量、來源與目的位移、列數與位移量（near，DS = DGROUP）
 *
 * 原版（逐指令，全部 16 位元回繞）：
 *   ds:8AF8 ← 0；ds:8AFA ← 0。
 *   x ← ds:8ADE（有號）；ax ← x；dx ← x + 0xF；ax ← ax >> 3、dx ← dx >> 3（兩者都是 sar，算術右移）；
 *   cx ← 2 − (dx − ax)；
 *   若 ax < 0（有號）：ds:8AF8 ← −ax。
 *   否則若 dx ≥ 0x50（有號）：ds:8AFA ← (dx − 0x4F) + cx。
 *   si ← ds:8AF2；ax ← ds:8AE0（Y，有號）；bp ← 0x10。
 *   若 ax < 0（有號）：a ← −ax；bp ← 0x10 − a；si ← si + 2a；di ← di + 80a（dx 暫存 16a）。
 *   否則若 ax ≥ 0x180（有號）：bp ← 0x10 − (ax − 0x180)。
 *   ax ← ds:8ADE；al ← al & 7；cl ← al。
 * 輸入暫存器：DI（只有 Y < 0 的分支用到，其餘分支原樣保留）。
 * 輸出暫存器：SI、DI、BP、CL（CH 保留 cx = 2 − (dx − ax) 的高位元組）。AX、DX 是暫用值。
 *   唯一呼叫者 36AD:0279 緊接著 call 36AD:041F；041F 用 SI（來源位移）、DI（目的位移）、CL（位元移位數）、
 *   BP（列數迴圈計數）、ds:8AF8、ds:8AFA（左右裁切量），一進入就改寫 DX 與 AX，不看 CH。
 *   描述檔 ret = pack:cx,si,di,bp（框架一個 ret 最多 4 個暫存器）；AX、DX 的最終值不在比對範圍內
 *   （是死值：呼叫者之後立刻覆蓋）。DX 在 ax ≥ 0 且 dx ≥ 0x50 的分支會寫進 ds:8AFA，這一部分由寫入集合驗證。
 * 寫入：ds:8AF8、ds:8AFA（兩個 word）。DS 是 DGROUP 0x54B3（呼叫者 36AD:00D2 的 mov ds 為證，見 36AD:0211）。
 *
 * 語意（假說）：與 2376:09AE（圖塊矩形的螢幕裁切）同型；X 是像素、Y 是畫面列，ds:8AF8／8AFA 是左右需裁掉的
 *   位元組欄數，si 起於 ds:8AF2（來源圖樣位移）加上上緣裁掉的列，bp 是可見列數。這是游標／小圖的位元平面繪製準備，
 *   不屬已證實的語意。
 * 輸入域：ds:8ADE、8AE0、8AF2 的全部 16 位元值，DI 的全部 16 位元值；沒有指標，不讀視訊記憶體區。
 *   邊界值（x = −1／0／7／8／0x270／0x271、x + 0xF 溢位的 0x7FF1、y = −1／0／0x17F／0x180／0x181、
 *   0x8000）由描述檔 seeds 打到。
 * 推論等級：算式已證實（反組譯逐指令＋差異測試）；欄位語意為假說。
 */
#include "gin3_reg.h"

/* 回傳 pack:cx,si,di,bp（CX 在最高位）。di_in 是進入時的 DI。 */
uint64_t f_36AD_03AB(Mem *m, uint16_t di_in)
{
    uint16_t x, ax, dx, cx, si, di = di_in, bp, t;

    wr16(m, DS(0x8AF8u), 0);
    wr16(m, DS(0x8AFAu), 0);
    x = rd16(m, DS(0x8ADEu));
    ax = x;
    dx = (uint16_t)(ax + 0xFu);
    ax = (uint16_t)((int16_t)ax >> 3); /* sar ax,3 */
    dx = (uint16_t)((int16_t)dx >> 3); /* sar dx,3 */
    t = dx; /* push dx */
    dx = (uint16_t)(dx - ax);
    cx = (uint16_t)(2u - dx);
    dx = t; /* pop dx */
    if ((int16_t)ax >= 0) {
        if ((int16_t)dx >= 0x50) {
            dx = (uint16_t)(dx - 0x4Fu);
            dx = (uint16_t)(dx + cx);
            wr16(m, DS(0x8AFAu), dx);
        }
    } else {
        ax = (uint16_t)(0u - ax); /* neg ax */
        wr16(m, DS(0x8AF8u), ax);
    }

    si = rd16(m, DS(0x8AF2u));
    ax = rd16(m, DS(0x8AE0u));
    bp = 0x10;
    if ((int16_t)ax >= 0) {
        if ((int16_t)ax >= 0x180) {
            ax = (uint16_t)(ax - 0x180u);
            bp = (uint16_t)(bp - ax);
        }
    } else {
        ax = (uint16_t)(0u - ax); /* neg ax */
        bp = (uint16_t)(bp - ax);
        ax = (uint16_t)(ax + ax); /* 2a */
        si = (uint16_t)(si + ax);
        ax = (uint16_t)(ax << 3); /* 三次 add ax,ax：16a */
        dx = ax;
        ax = (uint16_t)(ax << 2); /* 兩次 add ax,ax：64a */
        ax = (uint16_t)(ax + dx); /* 80a */
        di = (uint16_t)(di + ax);
    }

    ax = rd16(m, DS(0x8ADEu));
    cx = (uint16_t)((cx & 0xFF00u) | (ax & 7u)); /* and al,7; mov cl,al */
    return ((uint64_t)cx << 48) | ((uint64_t)si << 32) | ((uint64_t)di << 16) | bp;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_36AD_03AB(m, a[0]);
}
GIN3_REGISTER("36AD_03AB", 1, run)
