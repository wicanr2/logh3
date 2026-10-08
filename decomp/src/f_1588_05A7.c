/*
 * GIN3PS.EXE 1588:05A7  掃整張戰場地圖，收集指定圖塊的位置
 *
 * 原版：far 呼叫，四個字的參數（由低位址起）：圖塊（word，只比較低位元組：cmp cl,[bp+14h]）、
 * 輸出陣列指標（far，位移在前、段在後）、第三個參數（word）。回 AX ＝ 找到的格數。
 *
 * 第三個參數：呼叫端 2AF5:00CB 推的是「最多幾格」（5 與 5 減已收的數量），原版把它載入 DI（mov di,[bp+1Ah]），
 * 之後 DI 完全沒用到，迴圈也不檢查數量上限，所以這個參數沒有效果（已證實：反組譯）。本版接受並忽略。
 * 結果是呼叫端的緩衝區可能被寫超過「最多」的格數；輸入域在檔尾註明。
 *
 * 語意（已證實：與反組譯逐條對過；用途見 docs/re/320 §5「行星位置的來源：1588:05A7(6, …, 5) 收圖塊 6，再補圖塊 22」）：
 *   地圖緩衝區遠指標在 ds:3C1A（位移）／ds:3C1C（段），128 欄 × 0x66（102）列，由第 0 列第 0 欄起逐格往後讀，
 *   位移每格加 1（16 位元回繞，段不變）。
 *   每讀到一格等於圖塊：寫 word (欄 << 4)，輸出指標 ＋2；再寫 word (列 × 20)，輸出指標 ＋2；計數 ＋1。
 *   （欄 × 16 與 列 × 20 是格的左上像素座標，格是 16 × 20 像素，docs/re/121。）
 *   回計數（word）。
 *
 * 輸入域：地圖指標取基底狀態（戰場地圖已載入的局面）。輸出陣列要容得下所有符合的格（每格 4 bytes）；
 *   差異測試的輸出緩衝區是 8192 bytes，所以圖塊選在符合數遠小於 2048 的值，不含 0。
 */
#include "gin3_reg.h"

uint16_t f_1588_05A7(Mem *m, uint16_t tile, uint16_t dst_off, uint16_t dst_seg, uint16_t max_unused)
{
    uint16_t src_off = rd16(m, DS(0x3C1A));
    uint16_t src_seg = rd16(m, DS(0x3C1C));
    uint16_t count = 0;
    uint16_t row, col;

    (void)max_unused;
    for (row = 0; row < 0x66u; row++) {
        for (col = 0; col < 0x80u; col++) {
            uint8_t cell = rd8(m, LIN(src_seg, src_off));

            src_off = (uint16_t)(src_off + 1u);
            if (cell == (uint8_t)tile) {
                wr16(m, LIN(dst_seg, dst_off), (uint16_t)(col << 4));
                dst_off = (uint16_t)(dst_off + 2u);
                wr16(m, LIN(dst_seg, dst_off), (uint16_t)(row * 20u));
                dst_off = (uint16_t)(dst_off + 2u);
                count++;
            }
        }
    }
    return count;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1588_05A7(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("1588_05A7", 4, run)
