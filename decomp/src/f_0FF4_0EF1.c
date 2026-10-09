/*
 * GIN3PS.EXE 0FF4:0EF1  把格盤顯示物件掛上圖塊（無回傳值）
 *
 * 原版：far 呼叫，兩個參數 idx、tile（word，序言 push bp／si／di、sub sp,4，所以在 [bp+0Eh]、[bp+10h]），
 * 兩個都只讀低位元組。沒有回傳值（描述檔 ret=none）。
 *
 * 反組譯逐條對應：
 *   obj ← far 指標 ds:9850／ds:9852 ＋ idx × 0x28（位移 16 位元回繞；idx 零延伸；之後存在區域變數，不再重讀 ds:9850）
 *   tile == 0xFF（byte 比較）：
 *       圖塊 ← byte ds:984D
 *   否則：
 *       圖塊 ← byte ds:984E ＋ tile（16 位元）
 *   位址 ← word ds:F12C ＋ (圖塊 << 7)                    （16 位元；段 ← word ds:F12E 原樣，沒有進位）
 *   obj[+0x10] ← 段、obj[+0x0E] ← 位址（兩個 word，先寫 +0x10 再寫 +0x0E）
 *   obj[+1] ← 1（byte）
 * 所有輸入都在寫入之前讀完。
 *
 * 模型限制
 *   位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000），本版逐 byte 判斷。
 *   word 寫入的有效位移是 16 位元：位移 0xFFFF 的 word，低位元組在 seg:FFFF、高位元組回繞到同一段的 seg:0000
 *   （差異測試的 case 證實：obj ＝ E000:FFF1 時 +0x0E 的 word 高位元組落在 E000:0000，不是線性的 F0000）。
 *
 * 推論等級
 *   已證實（差異測試）：上列運算順序、常數、寫入位址與寬度、byte 參數、位移回繞。
 *   強推論（docs/internals/56-fleet-composition-panel.md 第 106 到 119 行）：ds:9850 是 41 × 40 bytes 的格盤顯示物件
 *     （idx ＝ 隊形槽）、ds:984D／984E 是空位與有部隊的圖塊基底（tile ＝ 艦種，0xFF ＝ 空位）、
 *     ds:F12C／F12E 是圖塊圖形緩衝區的 far 指標（每圖塊 128 bytes，所以 << 7）、obj[+0x0E]／[+0x10] 是物件的圖形指標、
 *     obj[+1] 是啟用旗標（名稱為假說，值 1 為已證實）。
 *
 * 輸入域：idx、tile 任意 word（只用低位元組）；idx 超過 40 時物件落在 41 筆之後的鄰近記憶體，本版照寫。
 */
#include "gin3_reg.h"

static void put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

/* word 寫入：高位元組的位移是 off + 1，16 位元回繞（off ＝ 0xFFFF 時落在同一段的位移 0，不是下一段） */
static void put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    put8(m, LIN(seg, off), (uint8_t)v);
    put8(m, LIN(seg, off + 1), (uint8_t)(v >> 8));
}

void f_0FF4_0EF1(Mem *m, uint16_t idx_word, uint16_t tile_word)
{
    uint16_t idx = (uint8_t)idx_word;
    uint8_t tile = (uint8_t)tile_word;
    uint16_t obj_off = (uint16_t)(rd16(m, DS(0x9850)) + (uint16_t)(idx * 0x28u));
    uint16_t obj_seg = rd16(m, DS(0x9852));
    uint16_t shifted, gfx_seg, gfx_off;

    if (tile == 0xFF)
        shifted = (uint16_t)((uint16_t)rd8(m, DS(0x984D)) << 7);
    else
        shifted = (uint16_t)((uint16_t)(rd8(m, DS(0x984E)) + tile) << 7);
    gfx_seg = rd16(m, DS(0xF12E));
    gfx_off = (uint16_t)(rd16(m, DS(0xF12C)) + shifted);
    put16(m, obj_seg, (uint16_t)(obj_off + 0x10), gfx_seg);
    put16(m, obj_seg, (uint16_t)(obj_off + 0x0E), gfx_off);
    put8(m, LIN(obj_seg, obj_off + 1), 1);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_0FF4_0EF1(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("0FF4_0EF1", 2, run)
