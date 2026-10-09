/*
 * GIN3PS.EXE 341B:039D  畫面座標 (x, y) 落在哪一個視窗：沿視窗串列找第一個「種類 2 且格子範圍涵蓋該點」的節點
 *
 * 原版：far 呼叫，兩個 word 參數 x（[bp+10h]）、y（[bp+12h]），序言 push bp／si／di、sub sp,6、mov bp,sp。
 *   回傳 AX:BX ＝ 節點的 seg:off，沒有命中回 0:0。沒有任何寫入。
 *
 * 資料（視窗串列，見 341B:0353）：頭指標 ds:D804（位移）／ds:D806（段）。節點欄位：
 *   +0x00 byte ＝ 種類（本函式只認 2）；+0x05 byte ＝ 左上角格 X；+0x06 byte ＝ 左上角格 Y；
 *   +0x07 byte ＝ 寬（格）；+0x08 byte ＝ 高（格）；+0x16 word ＝ 下一個節點位移、+0x18 word ＝ 下一個節點段。
 *   格子大小 16×20 像素（docs/re/16、docs/spec/023：「行星以完整 16×20 格命中」）。
 *   docs/re/136 稱種類 2 是「命中判定要找的」，種類 1 是訊息框（docs/re/137）。其餘欄位語意以上述文件為準，這裡只驗行為。
 *
 * 語意（已證實，本檔與反組譯逐條對過）：
 *   cell_x ＝ (x >> 4) 的低位元組；cell_y ＝ (y ÷ 20) 的低位元組（y ÷ 20 是 16 位元無號除法，商最大 3276）。
 *   p ＝ (ds:D806):(ds:D804)
 *   while (p 不是 0:0) {                        ; 先檢查，空串列直接回 0:0
 *       若 byte p[+0] == 2
 *        且 cell_x >= p[+5]                      ; byte 無號比較
 *        且 cell_y >= p[+6]
 *        且 p[+5] + p[+7] > cell_x               ; 兩個 byte 零延伸後做 16 位元加法，不截成 byte
 *        且 p[+6] + p[+8] > cell_y：回 p
 *       p ＝ p[+0x16/+0x18]
 *   }
 *   回 0:0
 * 原版的細節（本版照做）：
 *   - cell_x、cell_y 只留低位元組：x >= 0x1000 或 y >= 5120 時格號回繞（與全寬度比較不同）。
 *   - 加法 p[+5] + p[+7] 在 16 位元下做，最大 510，與 byte 的 cell_x 比較。
 *   - 空串列檢查在每次迭代前（與 341B:0353 的 do-while 不同）。
 *   - 下一個指標的段在 +0x18、位移在 +0x16；只有兩個 word 同時為 0 才視為串列結尾。
 *   - 位移加 5、6、7、8、16h、18h 都在段內 16 位元回繞；word 存取在位移 0xFFFF 時高位元組取同段位移 0。
 *
 * 輸入域：串列必須在探針步數內走完（有環的案例被探針丟棄）。呼叫者 17 個（台帳）；用途「全域命中測試：游標下的視窗」為強推論
 * （docs/re/342、docs/re/317、docs/re/157 都以這個意思呼叫它）。
 */
#include "gin3_reg.h"

/* 段內 word 讀取：位移各自在段內回繞（dosgolem cpu.go read16），不是線性位址加 1。 */
static uint16_t f341B039D_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint32_t f_341B_039D(Mem *m, uint16_t x, uint16_t y)
{
    uint8_t cell_x = (uint8_t)(x >> 4);
    uint8_t cell_y = (uint8_t)(y / 20u);
    uint16_t off = rd16(m, DS(0xD804));
    uint16_t seg = rd16(m, DS(0xD806));

    while ((seg | off) != 0) {
        uint16_t nseg, noff;

        if (rd8(m, LIN(seg, off)) == 2
            && !(cell_x < rd8(m, LIN(seg, (uint16_t)(off + 5u))))
            && !(cell_y < rd8(m, LIN(seg, (uint16_t)(off + 6u))))
            && !((uint16_t)(rd8(m, LIN(seg, (uint16_t)(off + 5u))) + rd8(m, LIN(seg, (uint16_t)(off + 7u)))) <= cell_x)
            && !((uint16_t)(rd8(m, LIN(seg, (uint16_t)(off + 6u))) + rd8(m, LIN(seg, (uint16_t)(off + 8u)))) <= cell_y))
            return ((uint32_t)seg << 16) | off; /* AX:BX */
        nseg = f341B039D_rd16s(m, seg, (uint16_t)(off + 0x18u));
        noff = f341B039D_rd16s(m, seg, (uint16_t)(off + 0x16u));
        seg = nseg;
        off = noff;
    }
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_341B_039D(m, a[0], a[1]); }
GIN3_REGISTER("341B_039D", 2, run)
