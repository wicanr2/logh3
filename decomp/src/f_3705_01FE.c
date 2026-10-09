/*
 * GIN3PS.EXE 3705:01FE  1 位元遮罩圖上的像素測試：點 (x, y) 在遮罩裡是不是 1，回傳「該位元組 AND 遮罩位元」
 * 呼叫：far，六個 word 參數（序言 push bp／si／di、sub sp,8，第一個參數在 [bp+12h]）：
 *   x    [bp+12h]  像素 x（相對遮罩區原點）
 *   y    [bp+14h]  像素 y
 *   w    [bp+16h]  遮罩區的寬，單位是 16 像素的欄數（用來算每列圖塊的位元組數）
 *   h    [bp+18h]  不使用（呼叫者 3705:00DC 推入熱區記錄的 +3，函式沒讀）
 *   off  [bp+1Ah]、seg [bp+1Ch]  遮罩資料的 far 指標
 * 回傳 AX：byte & 遮罩位元（0 或遮罩位元本身；AH ＝ 0）。
 *
 * 算法（asm 逐條，所有算術是 16 位元回繞，除法無號、除數是常數所以不會 #DE）：
 *   遮罩資料分成 16 × 20 像素的圖塊，每塊 40 bytes（20 列 × 2 bytes，每列 16 位元，左邊 8 像素在第一個 byte）。
 *   圖塊按列優先排列，每列圖塊 w 塊：
 *   tile  ＝ (y / 20) × w + (x >> 4)         ；乘法只留低 16 位元
 *   addr  ＝ off + tile × 2 × 20 ＋ ((x % 16) >> 3) ＋ 2 × (y % 20)
 *   byte  ＝ 記憶體 seg:addr（addr 是 16 位元位移，在段內回繞）
 *   mask  ＝ ds:8BCA[x % 8]                   ；基底狀態該表是 80 40 20 10 08 04 02 01（bit 7 ＝ 最左像素）
 *   回傳  ＝ byte & mask
 * 用途：3705:00DC（hw-dependent）逐一檢查熱區記錄，記錄 +4 是遮罩資料相對 ds:F140 + 0xC0 的位移（強推論，見該函式 asm），
 *   用遮罩判斷滑鼠是否點在形狀裡。「熱區遮罩」是強推論，欄位語意未在本函式內驗證。
 * 輸入域：全部 16 位元的 x、y、w、off、seg。讀取落在視訊記憶體區（A0000 到 BFFFF）的案例由探針丟棄。
 * 推論等級：算術與位移已證實（asm 與差異測試）；遮罩圖塊格式為強推論。
 */
#include "gin3_reg.h"

uint16_t f_3705_01FE(Mem *m, uint16_t x, uint16_t y, uint16_t w, uint16_t unused, uint16_t off, uint16_t seg)
{
    uint16_t tile_col = (uint16_t)(x >> 4);
    uint16_t tile_row = (uint16_t)(y / 20u);
    uint16_t row_in_tile = (uint16_t)(y % 20u);
    uint16_t half = (uint16_t)((x % 16u) >> 3);
    uint16_t tile = (uint16_t)((uint16_t)(tile_row * w) + tile_col);
    uint16_t tile_bytes = (uint16_t)((uint16_t)(tile << 1) * 20u);
    uint16_t addr = (uint16_t)(off + tile_bytes);
    uint8_t byte, mask;

    (void)unused;
    addr = (uint16_t)(addr + half);
    addr = (uint16_t)(addr + (uint16_t)(row_in_tile << 1));
    byte = rd8(m, LIN(seg, addr));
    mask = rd8(m, DS(0x8BCA + (x % 8u)));
    return (uint16_t)(byte & mask);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3705_01FE(m, a[0], a[1], a[2], a[3], a[4], a[5]);
}
GIN3_REGISTER("3705_01FE", 6, run)
