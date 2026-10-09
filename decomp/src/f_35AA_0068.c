/*
 * 35AA:0068  文字視窗內容緩衝區清空：每一格寫成 (0x2020, 顏色, 0)（far）
 *
 * 原版：far 呼叫，一個 far 指標參數 obj（位移在前、段在後），無回傳值。
 *   堆疊：序言 push bp／si／di、sub sp,6、mov bp,sp，所以參數在 [bp+10h]（位移）與 [bp+12h]（段）。
 *   obj[+01] byte  顏色，obj[+0B] byte  內容格寬，obj[+0C] byte  內容格高，
 *   obj[+0E] word、obj[+10] word  內容緩衝區 far 指標（位移、段）。
 *   p ← 緩衝區指標；for (i = 0; i < obj[+0B] × obj[+0C]; i++)：
 *       p 處的 word ← 0x2020；p[+2] ← obj[+01]（寫了 word 之後才讀）；p[+3] ← 0；p 的位移 += 4。
 *   格數 obj[+0B] × obj[+0C] 與顏色 obj[+01] 每一圈都從 obj 重讀（`les bx,[bp+10h]`），
 *   迴圈條件是 16 位元無號比較 `jbe`（n ≤ i 就結束），n 最大 255 × 255 = 65025。
 *   緩衝區若與 obj 重疊（寫入改到 obj[+01]、[+0B]、[+0C]），後面的圈數與顏色隨之改變；本版照原版順序重讀。
 *   寫入的 word 位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 8086 語意）；p[+2]、p[+3] 的位移 16 位元回繞。
 * 語意：docs/re/137 第 157 行「35AA:0068 把那塊緩衝區每一格填成 0x2020（兩個半形空白）＋ 視窗的顏色 [+1]、
 *   旗標 0」，docs/re/141 第 75 行 `\x0c` 清空、docs/internals/49 第 45 行；一格 4 bytes（字碼 word、顏色 byte、旗標 byte）。
 *   與本檔一致（已證實：反組譯逐指令）。呼叫者 341B:0054（開視窗）、3617:03BE（換行超出最後一列、無捲動時整窗清空）。
 * 寫入落在線性 0xF0000 以上（BIOS ROM 區）時，dosgolem 安靜丟棄，本版以逐 byte 的 put8 表達（跨過 0xEFFFF／0xF0000 的 word，
 *   低位元組寫入、高位元組丟棄）。寫入或讀取落在 0xA0000 到 0xBFFFF 的案例由探針丟棄。
 *
 * 輸入域：obj 與緩衝區指標任意（含重疊）；步數 ≤ 描述檔 step_budget（1,000,000，約 n ≤ 55,000 格）。
 *   n 在 32768 以上時，有號比較與無號比較結果不同（描述檔種子涵蓋）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

/* 段內回繞的 word 讀取：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go）。 */
static uint16_t f35AA_0068_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 線性 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go）。 */
static void f35AA_0068_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void f35AA_0068_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f35AA_0068_put8(m, LIN(seg, off), (uint8_t)v);
    f35AA_0068_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_35AA_0068(Mem *m, uint16_t obj_off, uint16_t obj_seg)
{
    uint16_t p_seg = f35AA_0068_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x10u));
    uint16_t p_off = f35AA_0068_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Eu));
    uint16_t i = 0;
    uint16_t n;
    uint8_t al;

    for (;;) {
        n = (uint16_t)(rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Bu))) * rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Cu)))); /* mul cx，取 AX */
        if (n <= i) /* cmp ax,[bp+0]; jbe */
            break;
        f35AA_0068_wr16s(m, p_seg, p_off, 0x2020u);
        al = rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 1u)));
        f35AA_0068_put8(m, LIN(p_seg, (uint16_t)(p_off + 2u)), al);
        f35AA_0068_put8(m, LIN(p_seg, (uint16_t)(p_off + 3u)), 0);
        p_off = (uint16_t)(p_off + 4u);
        i++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_35AA_0068(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("35AA_0068", 2, run)
