/*
 * GIN3PS.EXE 17A0:01BE  戰場上點 (x, y) 命中了哪些艦艇與行星：把命中的編號寫進輸出陣列，回命中個數
 *
 * 原版：far 呼叫，四個 word 參數（x, y, out_off, out_seg），回 AX ＝ 命中個數。
 *   序言 push bp／si／di、sub sp,12h，所以 x 在 [bp+1Ch]、y 在 [bp+1Eh]、輸出陣列的 far 指標在 [bp+20h]（位移）與
 *   [bp+22h]（段）。輸出指標是堆疊上的區域複本，每寫一筆就加 2。
 *   艦艇清單：ds:9896（word）＝ 筆數、ds:98BA 起每筆一個 word ＝ 艦艇池索引（同 1727:011E）。
 *   艦艇池基底 ds:CFB4／CFB6，一筆 78（0x4E）bytes。行星切片：ds:CFF8／CFFA 起、ds:CFFC（byte）＝ 筆數，
 *   一筆 82（0x52）bytes。
 *
 * 語意（已證實：本檔與反組譯逐條對過；欄位語意：docs/internals/38 艦艇 +0x1E bit7 ＝ Out、+0x34／+0x36 ＝ 現在位置 X／Y；
 *   docs/internals/29 行星 +0x46／+0x48 ＝ 星域地圖上的位置 X／Y。函式用途「滑鼠點擊命中測試」為強推論，
 *   呼叫者 1629:0005 未在此驗證）：
 *   階段 1，對艦艇清單的每一筆（i ＝ 0 起，i < ds:9896，無號）：
 *       idx ＝ word ds:(98BA ＋ 2i)；rec ＝ (ds:CFB4) ＋ (idx × 0x4E 的低 16 位元)
 *       命中條件（全部成立）：
 *           (byte rec[+0x1E] & 0x80) == 0
 *           word rec[+0x34] <= x <  word rec[+0x34] ＋ 0x10      ; 無號，＋0x10 是 16 位元，會繞回
 *           word rec[+0x36] <= y <  word rec[+0x36] ＋ 0x10
 *       命中則 word out[count] ← idx，count++
 *   階段 2，對行星切片的每一筆（n ＝ 0 起，n < byte ds:CFFC）：
 *       命中條件：word p[+0x46] <= x < p[+0x46] ＋ 0x10 且 word p[+0x48] <= y < p[+0x48] ＋ 0x10
 *       命中則 word out[count] ← (n | 0x8000)，count++
 *   結尾 word out[count] ← 0xFFFF，回 count。
 *
 * 原版的細節（本版照做）：
 *   - 比較全部無號（jb／jbe）。「＋0x10 > 座標」用 16 位元加法：欄位 >= 0xFFF0 時加法繞回，結果小於座標，不命中。
 *   - 輸出指標位移 16 位元遞增，段不變。
 *   - 階段 1 只看艦艇池，與艦艇所屬陣營、艦種無關；階段 2 的編號帶 0x8000 旗標以便與艦艇編號區分。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（rd16s／wr16s_rom），不是下一段的第一個 byte。
 *   - 輸出段被改到 0xF000 以上時，寫入落在 BIOS ROM 區（線性 0xF0000 起），機器忽略該寫入
 *     （dosgolem internal/machine/bios_rom.go）；本版用 wr16s_rom 表達。平面視訊記憶體的寫入由探針丟棄。
 *
 * 輸入域：基底狀態的 ds:9896 是 0，艦艇清單由種子與突變構造（筆數到約 5,000，超過探針步數上限的被丟棄）；
 *   清單位移在這個範圍內不會超過 0xFFFF。行星切片取基底狀態的 4 顆（座標都是 0）並由突變與種子改動。
 *   輸出陣列的段取暫存緩衝區（0x9D00）或種子指定位置。
 */
#include "gin3_reg.h"

/*
 * 段內存取的兩個輔助。gin3_mem.h 的 rd16／wr16 吃線性位址，高位元組取「線性位址＋1」；
 * 真正的 CPU 位移各自回繞，seg:FFFF 的 word 高位元組在 seg:0000（dosgolem internal/cpu/cpu.go read16／write16，
 * 8086 語意）。位移可能是 0xFFFF 時用這組。寫入另外表達 ROM：0xF0000 以上寫入無效
 * （dosgolem internal/machine/bios_rom.go），用本檔的 wr8_skip_rom。這兩點 gin3_mem.h 都沒有（回報給主工作階段）。
 */
static uint16_t rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意（回報給主工作階段）。 */
static void wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void wr16s_rom(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8_skip_rom(m, LIN(seg, off), (uint8_t)v);
    wr8_skip_rom(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_17A0_01BE(Mem *m, uint16_t x, uint16_t y, uint16_t out_off, uint16_t out_seg)
{
    uint16_t count = 0;
    uint16_t i = 0;
    uint16_t p = 0x98BAu;
    uint16_t off, seg;
    uint16_t n;

    while (i < rd16(m, DS(0x9896))) {
        uint16_t idx = rd16(m, DS(p));
        uint16_t rec = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));
        uint16_t pool_seg = rd16(m, DS(0xCFB6));

        if ((rd8(m, LIN(pool_seg, rec + 0x1Eu)) & 0x80u) == 0
            && x >= rd16s(m, pool_seg, (uint16_t)(rec + 0x34u))
            && (uint16_t)(rd16s(m, pool_seg, (uint16_t)(rec + 0x34u)) + 0x10u) > x
            && y >= rd16s(m, pool_seg, (uint16_t)(rec + 0x36u))
            && (uint16_t)(rd16s(m, pool_seg, (uint16_t)(rec + 0x36u)) + 0x10u) > y) {
            wr16s_rom(m, out_seg, out_off, idx);
            out_off = (uint16_t)(out_off + 2u);
            count++;
        }
        i++;
        p = (uint16_t)(p + 2u);
    }

    off = rd16(m, DS(0xCFF8));
    seg = rd16(m, DS(0xCFFA));
    n = 0;
    while (n < rd8(m, DS(0xCFFC))) {
        if (x >= rd16s(m, seg, (uint16_t)(off + 0x46u))
            && (uint16_t)(rd16s(m, seg, (uint16_t)(off + 0x46u)) + 0x10u) > x
            && y >= rd16s(m, seg, (uint16_t)(off + 0x48u))
            && (uint16_t)(rd16s(m, seg, (uint16_t)(off + 0x48u)) + 0x10u) > y) {
            wr16s_rom(m, out_seg, out_off, (uint16_t)(n | 0x8000u));
            out_off = (uint16_t)(out_off + 2u);
            count = (uint16_t)(count + 1u);
        }
        n++;
        off = (uint16_t)(off + 0x52u);
    }

    wr16s_rom(m, out_seg, out_off, 0xFFFFu);
    return count;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_17A0_01BE(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("17A0_01BE", 4, run)
