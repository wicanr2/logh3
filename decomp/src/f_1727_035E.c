/*
 * GIN3PS.EXE 1727:035E  戰場小地圖的視野框：擦掉舊框、畫新框，並記下新框的位置（回 AX = 參數）
 *
 * 原版：far 呼叫，一個堆疊字 arg（序言 push bp／si／di、沒有 sub sp，所以參數在 [bp+0Ah]），回 AX（見下）。
 * 逐條對應反組譯：
 *   si ← (word ds:9BF4 >> 4) ＋ 0x1E0（16 位元）          ; 小地圖框的 X：鏡頭 X 右移 4，加 DOS 小地圖原點 480
 *   di ← (word ds:9BF6 >> 4) ＋ 0x24（16 位元）           ; 小地圖框的 Y：鏡頭 Y 右移 4，加原點 36
 *   若 byte ds:A426 ≠ 0 且 byte ds:A427 ≠ 0：
 *       1727:03CF(x ＝ word ds:A428, y ＝ word ds:A42A, mask ＝ word ds:A42C)   ; 以舊位置重畫（擦掉舊框，mask 是舊的）
 *   若 byte ds:A426 ≠ 0 且 word arg ≠ 0：
 *       1727:03CF(si, di, 0xFFFF)                                              ; 畫新框（實線）
 *   word ds:A428 ← si；word ds:A42A ← di；word ds:A42C ← 0xFFFF；byte ds:A427 ← arg 的低位元組
 *   回 AX ← arg（最後一個寫進 AX 的是 `mov ax,[bp+arg_0]`）
 * 兩次比較是不同寬度：ds:A426、ds:A427 是 byte，arg 是完整 word（arg ＝ 0x0100 時第二個條件成立、A427 卻寫 0）。
 * 03CF 的畫線（39D7:0936，碰視訊記憶體）在差異測試以 stubs 取代；像素行為不在驗證範圍（見 f_1727_03CF.c）。
 *
 * 呼叫者：1727:002F、1727:0082、1727:0453 與另一處（共 4 處）。
 * 推論等級
 *   已證實（差異測試）：位移運算與常數、兩個條件的寬度與順序、03CF 的參數與次序、寫入的值與寬度、回傳值。
 *   強推論（docs/re/395-battle-data-scroll.md、docs/internals/34-battle-screen.md 第 97 行）：ds:9BF4／9BF6 是戰場鏡頭，
 *     框是 28 × 22 的小地圖視野框；ds:A426 是「小地圖啟用」、ds:A427 是「框已畫出」，ds:A428..A42C 是上一個框的 x、y 與 mask（名稱為假說）。
 * 輸入域：ds:9BF4／9BF6、ds:A426..A42C 與 arg 任意；03CF 的呼叫序列與參數由軌跡比對。
 */
#include "gin3_reg.h"

extern void f_1727_03CF(Mem *m, uint16_t x, uint16_t y, uint16_t mask);

uint16_t f_1727_035E(Mem *m, uint16_t arg)
{
    uint16_t si = (uint16_t)((uint16_t)(rd16(m, DS(0x9BF4)) >> 4) + 0x1E0u);
    uint16_t di = (uint16_t)((uint16_t)(rd16(m, DS(0x9BF6)) >> 4) + 0x24u);

    if (rd8(m, DS(0xA426)) != 0 && rd8(m, DS(0xA427)) != 0)
        f_1727_03CF(m, rd16(m, DS(0xA428)), rd16(m, DS(0xA42A)), rd16(m, DS(0xA42C)));
    if (rd8(m, DS(0xA426)) != 0 && arg != 0)
        f_1727_03CF(m, si, di, 0xFFFFu);
    wr16(m, DS(0xA428), si);
    wr16(m, DS(0xA42A), di);
    wr16(m, DS(0xA42C), 0xFFFFu);
    wr8(m, DS(0xA427), (uint8_t)arg);
    return arg;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1727_035E(m, a[0]); }
GIN3_REGISTER("1727_035E", 1, run)
