/*
 * GIN3PS.EXE 341B:057F  把一個節點插進視窗串列（雙向鏈結）：prev 之後；prev 為空則插在最前面
 *
 * 原版：far 呼叫，序言 push bp／si／di、sub sp,4、mov bp,sp。堆疊參數（由低位址到高位址）：
 *   node（far 指標，位移 [bp+0Eh]／段 [bp+10h]）、prev（far 指標，位移 [bp+12h]／段 [bp+14h]）。
 *   區域變數 [bp+0]（prev 段暫存）與 [bp+2]（node 段暫存）只是暫存。回傳 AX:BX ＝ node 的 seg:off。
 *
 * 資料（與 341B:0353、039D、0677 共用）：節點 +0x12 word ＝ 前一個節點位移、+0x14 word ＝ 前一個節點段、
 *   +0x16 word ＝ 下一個節點位移、+0x18 word ＝ 下一個節點段。串列頭 ds:D804（位移）／ds:D806（段），尾 ds:D808／ds:D80A。
 *   「0:0」（兩個 word 同時為 0）是空指標。
 *
 * 語意（已證實，本檔與反組譯逐條對過；呼叫者 341B:0054 與 341B:02A3 在 docs/re/137 §掛進串列 傳 prev＝0:0，見該文件）：
 *   A) prev 不是空（插在 prev 之後）：
 *        node.prev ← prev
 *        old ← prev.next
 *        node.next ← old
 *        prev.next ← node
 *        next ← node.next（從記憶體重讀）
 *        若 next 不是空：next.prev ← node；否則：尾指標 ← node
 *   B) prev 是空：
 *        B1) 頭指標不是空：node.next ← 頭；頭指標 ← node；node.prev ← 0:0；頭.prev ← node（頭是從 node.next 重讀）
 *        B2) 頭指標是空（空串列）：頭、尾指標 ← node；node.prev ← 0:0；node.next ← 0:0
 *   以上都沒有檢查 node 是否已在串列裡。
 * 原版的細節（本版照做）：
 *   - 所有欄位都是記憶體讀寫，順序固定（見下面程式的逐步註解）。node、prev、prev 的下一個之間別名時
 *     （node ＝ prev、node 已經在 prev 之後、node 在串列頭…），後面的讀取看得到前面的寫入，本版逐步照做。
 *   - 分支 A 結尾的「next」是寫入之後從 node.next 重讀的值，不是 old；node＝prev 時兩者不同。
 *   - 分支 B1 的最後一步用 les：從 node.next（寫入後重讀）取出位移與段，兩個 word 各自在段內回繞。
 *   - word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem cpu.go read16／write16）；位移加法在 16 位元下回繞。
 *   - 寫入落在線性 0xF0000 以上（ROM 區）時 dosgolem 安靜丟棄；本版自己模型化。ds 內的頭尾指標寫入不受影響。
 *
 * 輸入域：node、prev、old、next 指向 A0000 到 BFFFF 的案例被探針丟棄；指標指到堆疊或函式自己的程式碼不在涵蓋範圍。
 * 呼叫者（台帳）：341B:0054、341B:02A3。用途「視窗掛進串列」為強推論。
 */
#include "gin3_reg.h"

/* 段內 word 讀取：位移各自在段內回繞（dosgolem cpu.go read16），不是線性位址加 1。 */
static uint16_t f341B057F_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。 */
static void f341B057F_wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void f341B057F_wr16s_rom(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f341B057F_wr8_skip_rom(m, LIN(seg, off), (uint8_t)v);
    f341B057F_wr8_skip_rom(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint32_t f_341B_057F(Mem *m, uint16_t node_off, uint16_t node_seg, uint16_t prev_off, uint16_t prev_seg)
{
    uint16_t ax, bx, dx, di, si, es, cx;

    if ((prev_seg | prev_off) != 0) {
        /* A) 插在 prev 之後 */
        ax = prev_seg;
        bx = prev_off;
        es = node_seg;
        si = node_off;
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 0x14u), ax); /* node.prev.seg ← prev.seg */
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 0x12u), bx); /* node.prev.off ← prev.off */
        es = ax;
        bx = (uint16_t)(bx + 0x16u);                            /* bx ＝ &prev.next.off */
        dx = f341B057F_rd16s(m, es, (uint16_t)(bx + 2u));       /* old.seg */
        di = f341B057F_rd16s(m, es, bx);                        /* old.off */
        /* [bp+0] ← es（prev 段）；les si,node；add si,16h */
        es = node_seg;
        si = (uint16_t)(node_off + 0x16u);
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 2u), dx);    /* node.next.seg ← old.seg */
        f341B057F_wr16s_rom(m, es, si, di);                     /* node.next.off ← old.off */
        dx = node_seg;
        di = node_off;
        /* [bp+2] ← es（node 段）；es ← [bp+0]（prev 段） */
        f341B057F_wr16s_rom(m, prev_seg, (uint16_t)(bx + 2u), dx); /* prev.next.seg ← node.seg */
        f341B057F_wr16s_rom(m, prev_seg, bx, di);                  /* prev.next.off ← node.off */
        es = node_seg;
        ax = f341B057F_rd16s(m, es, (uint16_t)(si + 2u));       /* next.seg ← node.next.seg（重讀） */
        bx = f341B057F_rd16s(m, es, si);                        /* next.off ← node.next.off（重讀） */
        cx = (uint16_t)(ax | bx);
        if (cx == 0) {
            wr16(m, DS(0xD80A), node_seg);                      /* 尾指標 ← node */
            wr16(m, DS(0xD808), node_off);
        } else {
            es = ax;
            f341B057F_wr16s_rom(m, es, (uint16_t)(bx + 0x14u), dx); /* next.prev.seg ← node.seg */
            f341B057F_wr16s_rom(m, es, (uint16_t)(bx + 0x12u), di); /* next.prev.off ← node.off */
        }
    } else if ((rd16(m, DS(0xD806)) | rd16(m, DS(0xD804))) != 0) {
        /* B1) 插在最前面（串列不空） */
        ax = rd16(m, DS(0xD806));
        bx = rd16(m, DS(0xD804));
        es = node_seg;
        si = node_off;
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 0x18u), ax); /* node.next.seg ← 舊頭.seg */
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 0x16u), bx); /* node.next.off ← 舊頭.off */
        ax = es;
        bx = si;
        wr16(m, DS(0xD806), ax);                                /* 頭指標 ← node */
        wr16(m, DS(0xD804), bx);
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 0x12u), 0);  /* node.prev ← 0:0 */
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 0x14u), 0);
        ax = es;
        bx = si;
        {
            /* les si, es:[si+16h]：先取位移，再取段（兩者都用舊的 es:si） */
            uint16_t t_off = f341B057F_rd16s(m, es, (uint16_t)(si + 0x16u));
            uint16_t t_seg = f341B057F_rd16s(m, es, (uint16_t)(si + 0x16u + 2u));
            si = t_off;
            es = t_seg;
        }
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 0x14u), ax); /* 舊頭.prev.seg ← node.seg */
        f341B057F_wr16s_rom(m, es, (uint16_t)(si + 0x12u), bx); /* 舊頭.prev.off ← node.off */
    } else {
        /* B2) 空串列 */
        ax = node_seg;
        bx = node_off;
        wr16(m, DS(0xD806), ax);
        wr16(m, DS(0xD804), bx);
        wr16(m, DS(0xD80A), ax);
        wr16(m, DS(0xD808), bx);
        cx = 0;
        si = 0;
        es = ax;
        f341B057F_wr16s_rom(m, es, (uint16_t)(bx + 0x14u), cx);
        f341B057F_wr16s_rom(m, es, (uint16_t)(bx + 0x12u), si);
        es = ax;
        f341B057F_wr16s_rom(m, es, (uint16_t)(bx + 0x18u), cx);
        f341B057F_wr16s_rom(m, es, (uint16_t)(bx + 0x16u), si);
    }

    return ((uint32_t)node_seg << 16) | node_off; /* AX:BX ＝ [bp+10h]:[bp+0Eh] */
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_341B_057F(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("341B_057F", 4, run)
