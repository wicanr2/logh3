/*
 * GIN3PS.EXE 2BE1:0739  畫面座標 (x, y) 落在哪一個 16 × 16 的格（回格序，沒有命中回 0xFFFF）
 *
 * 原版：far 呼叫，兩個堆疊字 x、y（畫面座標；push bp／si／di、sub sp,0Ch，所以 x 在 [bp+16h]、y 在 [bp+18h]），回 AX。
 * 過程（已證實：反組譯逐指令；1588:0362 由其自己的差異測試證實）：
 *   結果 ← 0xFFFF
 *   1588:0362(x, y, &wx, &wy)：wx、wy 是本函式堆疊上的兩個 word（[bp+2]、[bp+4]），0362 把畫面座標換成世界座標
 *     （x − 0x10 + ds:9BF4、y − 0x14 + ds:9BF6；範圍外回 0 且不寫輸出）。本函式不看 0362 的回傳值。
 *   i ＝ 0..0x28（byte，i < 0x29 才進迴圈）且 結果 == 0xFFFF（每圈檢查，找到第一個就停）：
 *     表項 ＝ ds:(0xD072 + 10 × i)：+0 ＝ 格左緣 X（word）、+2 ＝ 格上緣 Y（word）、+4 ＝ 旗標（byte，0 才有效）
 *     旗標 == 0 且 wx >= X 且 (X + 0x10) > wx 且 wy >= Y 且 (Y + 0x10) > wy（全部無號 16 位元，X + 0x10 回繞）→ 結果 ← i
 *   回結果（AX = 0xFFFF 或 i）
 * 語意（強推論，docs/re/130「2BE1:0739(x, y) 先用 1588:0362 把畫面座標換成世界座標，然後依序…」）：表是「編隊」互動選擇時
 *   41 個槽位的格座標。名稱不是證據，位移與比較才是。
 * 副作用：無寫入（堆疊上的 wx、wy 不在比對範圍；C 版把它們放在暫存位置並於呼叫後還原，見下）。
 *
 * 輸入域與限制（重要）：
 *   原版把 wx、wy 放在自己的堆疊上，1588:0362 在 x、y 超出範圍（x < 0x10、y < 0x14、x >= 0x1D0、y >= 0x17C）時不寫輸出，
 *   迴圈讀到的 wx、wy 就是堆疊上的殘值（未初始化），那是框架無法重現的環境，輸入域排除這種情形：
 *   描述檔的 x 取 0x10..0x1CF、y 取 0x14..0x17B，0362 因此一定寫入兩個輸出。
 *   C 版用線性位址 0x9CFF0 起的 4 bytes 當 wx／wy，呼叫前存起來、取值後還原，所以對外的寫入集合與原版（堆疊不計）相同。
 *   word 存取在位移 0xFFFF 時…本函式的 word 讀取位移最大是 0xD072 + 410 + 2，不會到 0xFFFF。
 * 推論等級：已證實（反組譯逐指令＋差異測試，限上述輸入域）。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_0362(Mem *m, uint16_t x, uint16_t y, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off, uint16_t oy_seg);

#define W0739_TMP_SEG 0x9CFFu

uint16_t f_2BE1_0739(Mem *m, uint16_t x, uint16_t y)
{
    uint8_t save[4];
    uint16_t wx, wy, result = 0xFFFFu, t = 0xD072u;
    uint8_t i = 0;
    int k;

    for (k = 0; k < 4; k++)
        save[k] = rd8(m, LIN(W0739_TMP_SEG, (uint16_t)k));
    f_1588_0362(m, x, y, 0x0000u, W0739_TMP_SEG, 0x0002u, W0739_TMP_SEG);
    wx = rd16(m, LIN(W0739_TMP_SEG, 0x0000u));
    wy = rd16(m, LIN(W0739_TMP_SEG, 0x0002u));
    for (k = 0; k < 4; k++)
        wr8(m, LIN(W0739_TMP_SEG, (uint16_t)k), save[k]);

    while (i < 0x29u && result == 0xFFFFu) {
        if (rd8(m, DS((uint16_t)(t + 4u))) == 0) {
            uint16_t tx = rd16(m, DS(t));
            uint16_t ty = rd16(m, DS((uint16_t)(t + 2u)));

            if (wx >= tx && (uint16_t)(tx + 0x10u) > wx && wy >= ty && (uint16_t)(ty + 0x10u) > wy)
                result = i;
        }
        i++;
        t = (uint16_t)(t + 10u);
    }
    return result;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2BE1_0739(m, a[0], a[1]); }
GIN3_REGISTER("2BE1_0739", 2, run)
