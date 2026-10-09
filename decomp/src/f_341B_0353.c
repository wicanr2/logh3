/*
 * GIN3PS.EXE 341B:0353  在視窗串列裡依「編號」找第一個符合的視窗節點
 *
 * 原版：far 呼叫，一個堆疊參數 key（word，只用低位元組 CL）。回傳 AX:BX ＝ 節點的 seg:off，找不到回 0:0。
 *   push bp；sub sp,4；mov bp,sp，所以區域變數在 [bp+0..3]（目前節點 off／seg），參數在 [bp+0Ah]。
 *
 * 資料：視窗串列是雙向鏈結串列，頭指標 ds:D804（位移）／ds:D806（段），尾指標 ds:D808／ds:D80A。
 *   節點：+0x04 byte ＝ 視窗編號（本函式比對的欄位）、+0x16 word ＝ 下一個節點的位移、+0x18 word ＝ 下一個節點的段。
 *   串列的建立與拆除是 341B:057F（插入）與 341B:0677（拆除）。節點其餘欄位的語意見 341B:039D 與 341B:044A。
 *
 * 語意（已證實，本檔與反組譯逐條對過）：
 *   p ＝ (ds:D806):(ds:D804)
 *   do { if (byte p[+4] == key 的低位元組) return p;  p ＝ p[+0x16/+0x18]; } while (p 不是 0:0);
 *   return 0:0
 * 原版的細節（本版照做）：
 *   - 這是 do-while：頭指標是 0:0 時**不先檢查**，先讀 0000:0004（中斷向量表）比對一次；
 *     比對符合就回 0:0，不符合再從 0000:0016／0018 取「下一個」繼續走。呼叫者是否可能遇到空串列未知。
 *   - 比對只看 key 的低位元組；節點指標的位移加 4、16h、18h 都在段內 16 位元回繞。
 *   - 下一個指標的段在 +0x18、位移在 +0x16；只有兩個 word 同時為 0 才視為串列結尾。
 *   - 沒有任何寫入。
 *
 * 輸入域：串列必須在探針步數內走完（有環或太長的案例被探針丟棄）。位移 0xFFFF 的 word 存取在段內回繞（見 rd16s）。
 * 呼叫者：35D1:005D（台帳 callers＝1）；用途「依編號找視窗」為強推論。
 */
#include "gin3_reg.h"

/* 段內 word 讀取：位移各自在段內回繞（dosgolem cpu.go read16），不是線性位址加 1。 */
static uint16_t f341B0353_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint32_t f_341B_0353(Mem *m, uint16_t key)
{
    uint16_t off = rd16(m, DS(0xD804));
    uint16_t seg = rd16(m, DS(0xD806));
    uint8_t k = (uint8_t)key;

    for (;;) {
        uint16_t nseg, noff;

        if (rd8(m, LIN(seg, (uint16_t)(off + 4u))) == k)
            return ((uint32_t)seg << 16) | off; /* AX:BX */
        nseg = f341B0353_rd16s(m, seg, (uint16_t)(off + 0x18u));
        noff = f341B0353_rd16s(m, seg, (uint16_t)(off + 0x16u));
        seg = nseg;
        off = noff;
        if ((seg | off) == 0)
            return 0;
    }
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_341B_0353(m, a[0]); }
GIN3_REGISTER("341B_0353", 1, run)
