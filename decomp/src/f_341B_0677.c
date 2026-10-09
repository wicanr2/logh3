/*
 * GIN3PS.EXE 341B:0677  把一個節點從視窗串列（雙向鏈結）拆下來，補上前後鄰居與頭尾指標
 *
 * 原版：far 呼叫，序言 push bp／si、mov bp,sp（沒有區域變數）。一個堆疊參數 node（far 指標，位移 [bp+8]／段 [bp+0Ah]）。
 *   回傳 AX:BX ＝ node 的 seg:off（原樣，不管節點拆下來之後它還在不在串列裡）。
 *
 * 資料（與 341B:0353、039D、057F 共用）：節點 +0x12 word ＝ 前一個節點位移、+0x14 word ＝ 前一個節點段、
 *   +0x16 word ＝ 下一個節點位移、+0x18 word ＝ 下一個節點段。串列頭 ds:D804（位移）／ds:D806（段），尾 ds:D808／ds:D80A。
 *   「0:0」（兩個 word 同時為 0）是空指標。
 *
 * 語意（已證實，本檔與反組譯逐條對過）：
 *   prev ＝ node[+0x14]:node[+0x12]；next ＝ node[+0x18]:node[+0x16]
 *   若 prev 不是空：prev[+0x18]:prev[+0x16] ← next          ; 前一個的「下一個」指向 next
 *   否則：ds:D806:D804 ← next                              ; node 是頭，頭指標指向 next
 *   （重新讀 node 的 +0x14／+0x12／+0x18／+0x16；第一段若寫到 node 自己的欄位，第二段看得到）
 *   若 next 不是空：next[+0x14]:next[+0x12] ← node 的 prev   ; 後一個的「前一個」指向 prev
 *   否則：ds:D80A:D808 ← node 的 prev                       ; node 是尾，尾指標指向 prev
 *   node 自己的 +0x12..+0x19 不清除。
 * 原版的細節（本版照做）：
 *   - 所有欄位都從記憶體讀寫，順序固定：第一段讀 prev 的兩個 word（段先於位移）、讀 next（段先於位移）、
 *     寫 prev[+0x18]、prev[+0x16]；第二段重讀 next、重讀 prev（段先於位移）、寫 next[+0x14]、next[+0x12]。
 *     鄰居與 node 別名（例如 prev 就是 node 自己、兩個鄰居相同）時，後面的讀取看得到前面的寫入。
 *   - word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem cpu.go read16／write16）；加 0x12..0x18 的位移加法在 16 位元下回繞。
 *   - 寫入落在線性 0xF0000 以上（ROM 區）時 dosgolem 安靜丟棄；本版自己模型化。ds 內的頭尾指標寫入不受影響。
 *   - 不檢查 node 是否真的在串列裡；node 的鄰居欄位為垃圾時照寫。
 *
 * 輸入域：node 與兩個鄰居指向 A0000 到 BFFFF 的案例被探針丟棄。呼叫者（台帳）：341B:00D9、341B:02A3。
 * 用途「視窗從串列移除」為強推論（與 057F 的插入對偶）。
 */
#include "gin3_reg.h"

/* 段內 word 讀取：位移各自在段內回繞（dosgolem cpu.go read16），不是線性位址加 1。 */
static uint16_t f341B0677_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。 */
static void f341B0677_wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void f341B0677_wr16s_rom(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f341B0677_wr8_skip_rom(m, LIN(seg, off), (uint8_t)v);
    f341B0677_wr8_skip_rom(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint32_t f_341B_0677(Mem *m, uint16_t noff, uint16_t nseg)
{
    uint16_t cx, si, ax, bx;

    /* 第一段：前一個節點（或頭指標）改指向 node 的下一個 */
    cx = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x14u)); /* prev 段 */
    si = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x12u)); /* prev 位移 */
    if ((cx | si) != 0) {
        ax = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x18u)); /* next 段 */
        bx = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x16u)); /* next 位移 */
        f341B0677_wr16s_rom(m, cx, (uint16_t)(si + 0x18u), ax);
        f341B0677_wr16s_rom(m, cx, (uint16_t)(si + 0x16u), bx);
    } else {
        ax = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x18u));
        bx = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x16u));
        wr16(m, DS(0xD806), ax);
        wr16(m, DS(0xD804), bx);
    }

    /* 第二段：後一個節點（或尾指標）改指向 node 的前一個 */
    cx = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x18u)); /* next 段 */
    si = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x16u)); /* next 位移 */
    if ((cx | si) != 0) {
        ax = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x14u)); /* prev 段 */
        bx = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x12u)); /* prev 位移 */
        f341B0677_wr16s_rom(m, cx, (uint16_t)(si + 0x14u), ax);
        f341B0677_wr16s_rom(m, cx, (uint16_t)(si + 0x12u), bx);
    } else {
        ax = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x14u));
        bx = f341B0677_rd16s(m, nseg, (uint16_t)(noff + 0x12u));
        wr16(m, DS(0xD80A), ax);
        wr16(m, DS(0xD808), bx);
    }

    return ((uint32_t)nseg << 16) | noff; /* AX:BX */
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_341B_0677(m, a[0], a[1]); }
GIN3_REGISTER("341B_0677", 2, run)
