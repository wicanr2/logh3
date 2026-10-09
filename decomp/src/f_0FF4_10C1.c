/*
 * GIN3PS.EXE 0FF4:10C1  編成視窗按「決定」：把工作表寫回艦隊記錄（無回傳值）
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄（122 bytes）的 far 指標；序言 push bp／si／di、sub sp,2，
 * 參數在 [bp+0Ch]（位移）與 [bp+0Eh]（段）。沒有回傳值（描述檔 ret=none）。
 *
 * 反組譯逐條對應：
 *   i 從 0 起（byte），每圈先拿 i 與 byte ds:984B 比（無號，i >= ds:984B 就結束；每圈都重讀 ds:984B）：
 *     bx ← i × 4
 *     艦隊[0x10 + bx + 2] ← byte ds:[981E + bx + 2]
 *     艦隊[0x10 + bx + 3] ← byte ds:[981E + bx + 3]      （讀在上一行寫入之後）
 *     艦隊[0x10 + bx + 0] ← word ds:[981E + bx]           （讀在上兩行寫入之後）
 *   艦隊[+0x0F] ← byte ds:984B                            （迴圈結束後重讀）
 * 位移都是 16 位元運算，艦隊段原樣使用。
 *
 * 別名：艦隊記錄若蓋到 ds:981E 起的工作表或 ds:984B 本身，後面的讀取會看到前面的寫入，
 *   迴圈上界也可能在迴圈中途改變。本版逐步照原版順序做。
 *
 * 推論等級
 *   已證實（差異測試）：上列順序、常數、讀寫寬度。
 *   強推論（docs/internals/56-fleet-composition-panel.md 第 181 行、docs/internals/64-ship-pool.md 第 54 行）：
 *     ds:981E 是壓實排序後的 10 組 × 4 bytes 工作表（[0..1] 艦艇池索引、[2] 艦種、[3] 隊形槽）、
 *     ds:984B 是已編成的組數、艦隊 +0x10 起每組 4 bytes、+0x0F 是組數。「決定」才把工作表寫回艦隊記錄。
 *
 * 輸入域：艦隊指標取艦隊表任一筆與 seeds 指定的位址；ds:984B 任意 byte（超過 10 時讀到工作表之後的資料、
 *   寫過艦隊記錄之後的記憶體，本版照原版運算）。
 */
#include "gin3_reg.h"

/* word 寫入：高位元組的位移是 off + 1，16 位元回繞（off ＝ 0xFFFF 時落在同一段的位移 0；見 0FF4:0EF1 的差異測試） */
static void w10C1_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, off + 1), (uint8_t)(v >> 8));
}

void f_0FF4_10C1(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t i = 0;

    while (i < rd8(m, DS(0x984B))) {
        uint16_t bx = (uint16_t)(i << 2);
        uint16_t dst = (uint16_t)(fleet_off + 0x10u + bx);

        wr8(m, LIN(fleet_seg, dst + 2), rd8(m, DS(0x981Eu + bx + 2)));
        wr8(m, LIN(fleet_seg, dst + 3), rd8(m, DS(0x981Eu + bx + 3)));
        w10C1_wr16(m, fleet_seg, dst, rd16(m, DS(0x981Eu + bx)));
        i++;
    }
    wr8(m, LIN(fleet_seg, fleet_off + 0x0F), rd8(m, DS(0x984B)));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_0FF4_10C1(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("0FF4_10C1", 2, run)
