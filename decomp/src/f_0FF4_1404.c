/*
 * GIN3PS.EXE 0FF4:1404  數艦隊編制裡艦種 5 與艦種 6 的組數（結果寫進兩個 byte，無回傳值）
 *
 * 原版：far 呼叫，三個 far 指標參數；序言只有 push bp、sub sp,4，所以
 *   fleet 在 [bp+0Ah]（位移）／[bp+0Ch]（段）、out_a 在 [bp+0Eh]／[bp+10h]、out_b 在 [bp+12h]／[bp+14h]。
 * 沒有回傳值（描述檔 ret=none）。
 *
 * 反組譯逐條對應：
 *   *out_b ← 0、*out_a ← 0                                  （先 out_b 後 out_a，都是 byte）
 *   slot ← fleet + 0x10（16 位元位移）、j ← 0（byte）
 *   每圈先拿 j 與 byte 艦隊[+0x0F] 比（每圈重讀，無號，j >= 組數就結束）：
 *     slot[+2] == 5 → ++*out_a（byte 的 inc）
 *     slot[+2] == 6 → ++*out_b                               （slot[+2] 在上一個比較之後重讀）
 *     j++、slot += 4
 *
 * 別名：out_a、out_b 若指到艦隊自己的組欄位（或互相指到同一格），後面的圈與第二個比較會看到前面的 inc。
 *   本版逐步照原版順序做。
 *
 * 推論等級
 *   已證實（差異測試）：上列順序、常數、讀寫寬度、每圈重讀的位置。
 *   強推論（docs/internals/33-fleet-table.md 第 104 行、docs/internals/56-fleet-composition-panel.md）：
 *     艦隊 +0x10 起每組 4 bytes、[2] 是該組艦種、+0x0F 是組數。艦種 5 與 6 的語意（補給艦、修護艦等）
 *     見 docs/internals/64，此函式不依賴名稱。呼叫者 0FF4:1350 是編成時的艦種上限檢查（假說）。
 *   與推論無關的限制：組數在迴圈中不會改變（out 指到 +0x0F 時開頭就被清成 0，迴圈不會進入）。
 *
 * 輸入域：艦隊指標取艦隊表任一筆與 seeds 指定的位址；兩個輸出指標指到暫存緩衝區、艦隊自己的組欄位
 *   或互相指到同一格；組數任意 byte。
 */
#include "gin3_reg.h"

void f_0FF4_1404(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t a_off, uint16_t a_seg, uint16_t b_off,
                 uint16_t b_seg)
{
    uint16_t slot = (uint16_t)(fleet_off + 0x10u);
    uint8_t j = 0;

    wr8(m, LIN(b_seg, b_off), 0);
    wr8(m, LIN(a_seg, a_off), 0);
    while (j < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        if (rd8(m, LIN(fleet_seg, slot + 2)) == 5)
            wr8(m, LIN(a_seg, a_off), (uint8_t)(rd8(m, LIN(a_seg, a_off)) + 1));
        if (rd8(m, LIN(fleet_seg, slot + 2)) == 6)
            wr8(m, LIN(b_seg, b_off), (uint8_t)(rd8(m, LIN(b_seg, b_off)) + 1));
        j++;
        slot = (uint16_t)(slot + 4);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_0FF4_1404(m, a[0], a[1], a[2], a[3], a[4], a[5]);
    return 0;
}
GIN3_REGISTER("0FF4_1404", 6, run)
