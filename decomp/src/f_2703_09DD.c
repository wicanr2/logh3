/*
 * GIN3PS.EXE 2703:09DD  各陣營「已失格數」（相對於新遊戲開局擁有者快照）
 *
 * 原版：far 呼叫，一個參數＝輸出緩衝區的 far 指標（位移在前、段在後），無有意義的回傳
 * （AX 是離開迴圈時的殘值，恆為 0x20，呼叫端 056A:0469、2D10:0319 都不用）。
 * 語意（已證實：反組譯逐指令；docs/re/292、245、spec/041 一致）：
 *   out[0] ＝ out[1] ＝ 0
 *   對星域 0..31（記錄在 ds:A79C，每筆 126 bytes）：
 *     擁有者 ＝ 星域[+0x12]；快照 ＝ ds:B75C[i]（byte，新遊戲開局當下的擁有者，docs/re/292）
 *     擁有者 == 2 跳過；快照 == 2 跳過；快照 == 擁有者 跳過；
 *     否則 out[快照] += 1（byte，位移 16 位元回繞）
 * 三道跳過的判斷順序照反組譯。快照 byte 照原樣零擴充當索引，不檢查範圍（遊戲裡只有 0／1／2）。
 *   名稱「已失格數」「快照」來自 docs/re/292（confirmed）。
 * 副作用：只寫 out 緩衝區。差異測試的緩衝區給 256 bytes，涵蓋任何 byte 索引。
 * 輸入域：全部可到達（沒有除法）。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

void f_2703_09DD(Mem *m, uint16_t out_off, uint16_t out_seg)
{
    uint16_t region = 0xA79C;
    uint8_t i;

    wr8(m, LIN(out_seg, (uint16_t)(out_off + 1)), 0);
    wr8(m, LIN(out_seg, out_off), 0);
    for (i = 0; i < 0x20; i++) {
        uint8_t owner = rd8(m, DS((uint16_t)(region + 0x12)));
        uint8_t snap = rd8(m, DS((uint16_t)(0xB75C + i)));

        if (owner != 2 && snap != 2 && owner != snap) {
            uint32_t at = LIN(out_seg, (uint16_t)(out_off + snap));

            wr8(m, at, (uint8_t)(rd8(m, at) + 1));
        }
        region = (uint16_t)(region + 0x7E);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2703_09DD(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2703_09DD", 2, run)
