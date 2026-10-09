/*
 * 3667:034A  把文字視窗內容緩衝區中第 row 列的每一格的顏色位元組設成 color（far）
 *
 * 原版：far 呼叫，三個參數：obj（far 指標，位移在前、段在後）、row（word，只讀低位元組）、color（word，只讀低位元組）。
 *   堆疊：序言 push bp／si、sub sp,6、mov bp,sp，所以 obj 在 [bp+0Eh]／[bp+10h]，row 在 [bp+12h]，color 在 [bp+14h]。
 *   obj[+0B] byte  內容的格寬 w，obj[+0E] word、obj[+10] word  內容緩衝區 far 指標（位移、段）。
 *   起點 ＝ (obj[+0E] + (w × (row 的低位元組) × 4 取低 16 位元)，obj[+10])；
 *   for (col = 0; col < obj[+0B]; col++)：起點[+2] ← color 低位元組；起點位移 += 4。
 *   w 在迴圈條件每一圈從 obj[+0B] 重讀（`les bx,[bp+0Eh]` 再取 `es:[bx+0Bh]`），不是只讀一次；
 *   起點位移是 16 位元回繞；寫入的有效位址 `es:[bx+2]` 的位移也 16 位元回繞（0xFFFE＋2 ＝ 0）。
 *   obj 的位移欄位（+0B、+0E、+10）各自在段內 16 位元回繞，+0E 的 word 在位移 0xFFFF 時高位元組取同段位移 0。
 * 語意（強推論）：一格 4 bytes（字碼 word、顏色 byte、旗標 byte，docs/re/137、docs/internals/49），
 *   所以 +2 是格的顏色位元組；本函式把一整列的顏色統一設成 color（呼叫者 3667:00F2、023E）。
 *   「列」「顏色」是由格式推得的名稱，欄位位移與運算已證實。
 * 寫入落在線性 0xF0000 以上（BIOS ROM 區）時，dosgolem 安靜丟棄；本版用 put8 表達。
 *   寫入或讀取落在 0xA0000 到 0xBFFFF 的案例由探針丟棄。
 *
 * 輸入域：obj 任意；緩衝區可指到任何位址（包括物件自己，寫入改到 obj[+0B] 時迴圈次數隨之改變，本版照原版重讀）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為強推論。
 */
#include "gin3_reg.h"

/* 段內回繞的 word 讀取：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go）。 */
static uint16_t f3667_034A_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 線性 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go）。 */
static void f3667_034A_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_3667_034A(Mem *m, uint16_t obj_off, uint16_t obj_seg, uint16_t row, uint16_t color)
{
    uint16_t ax, cx, si, ptr_off, ptr_seg;
    uint8_t col;

    ax = (uint16_t)(rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Bu))) * (uint8_t)row); /* mul cx（cx ＝ row 低位元組） */
    ax = (uint16_t)(ax << 2); /* 兩次 shl ax,1 */
    cx = f3667_034A_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x10u));
    si = f3667_034A_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Eu));
    si = (uint16_t)(si + ax);
    ptr_seg = cx;
    ptr_off = si;
    col = 0;
    while (col < rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Bu)))) {
        f3667_034A_put8(m, LIN(ptr_seg, (uint16_t)(ptr_off + 2u)), (uint8_t)color);
        col++;
        ptr_off = (uint16_t)(ptr_off + 4u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3667_034A(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("3667_034A", 4, run)
