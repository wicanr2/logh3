/*
 * GIN3PS.EXE 0FF4:0C06  編成視窗格盤的點擊命中：螢幕座標 → 隊形槽編號（找不到回 0xFF）
 *
 * 原版：far 呼叫，兩個參數 px、py（word），序言 push bp／si／di、sub sp,6，
 * 所以 px 在 [bp+10h]、py 在 [bp+12h]。回傳 AX，呼叫端 0FF4:07D3、0FF4:0B85 只存 AL 並與 0xFF 比。
 *
 * 反組譯逐條對應：
 *   tx ← px − 0x58、ty ← py − 0x5C（16 位元）
 *   i 從 0 到 0x28（byte，共 41 格），讀 ds:28B8 起、每格 4 bytes 的 (X, Y) 兩個 word：
 *     tx <  X                 → 下一格            （有號 word 比較，jl）
 *     X + 0x10 <= tx          → 下一格            （X + 0x10 先做 16 位元加法，再有號比較，jle）
 *     ty <  Y                 → 下一格
 *     Y + 0x10 <= ty          → 下一格
 *     都過了                  → 回 AL = i
 *   41 格都不中                → 回 AL = 0xFF
 * 所以每格是 16 × 16 的半開區間 [X, X+0x10) × [Y, Y+0x10)，格序就是回傳的編號。
 *
 * AH：原版沒有清。每圈 `mov ax, ds` 把 AX 設成 DS（AH ＝ 0x54），走到第四個比較時 AX 再被
 *   `Y + 0x10` 蓋掉，所以 AH 是「最後一圈的 AX 高位元組」：命中時是該格 (Y + 0x10) 的高位元組，
 *   不中時是第 41 格（i ＝ 0x28）那一圈留下的值。本版照樣模擬，回傳完整 AX。
 *
 * 推論等級
 *   已證實（差異測試）：上列結構、常數 0x58／0x5C／0x10／0x29、比較的有號性與 16 位元回繞、AH 的來源。
 *   強推論（docs/internals/56-fleet-composition-panel.md）：ds:28B8 是 41 格格盤的左上角（相對於格盤原點，
 *     16 × 16 圖塊），0x58／0x5C 是格盤在螢幕上的原點；px、py 是滑鼠座標。
 *   未知：ds:28B8 表是靜態資料還是執行期填的（基底狀態裡是 41 組 (−0x40..0x40) 的六角格座標）。
 *
 * 輸入域：px、py 任意 word；表內容任意（突變會改它）。
 */
#include "gin3_reg.h"

uint16_t f_0FF4_0C06(Mem *m, uint16_t px, uint16_t py)
{
    uint16_t tx = (uint16_t)(px - 0x58u);
    uint16_t ty = (uint16_t)(py - 0x5Cu);
    uint16_t ax = 0;
    uint8_t i;

    for (i = 0; i < 0x29; i++) {
        uint16_t ent = (uint16_t)(0x28B8u + (uint16_t)i * 4u);
        uint16_t x = rd16(m, DS(ent));
        uint16_t y;

        ax = GIN3_DS;
        if ((int16_t)tx < (int16_t)x)
            continue;
        if ((int16_t)(uint16_t)(x + 0x10u) <= (int16_t)tx)
            continue;
        y = rd16(m, DS((uint16_t)(ent + 2)));
        if ((int16_t)ty < (int16_t)y)
            continue;
        ax = (uint16_t)(rd16(m, DS((uint16_t)(ent + 2))) + 0x10u);
        if ((int16_t)ax <= (int16_t)ty)
            continue;
        return (uint16_t)((ax & 0xFF00u) | i);
    }
    return (uint16_t)((ax & 0xFF00u) | 0xFFu);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0FF4_0C06(m, a[0], a[1]); }
GIN3_REGISTER("0FF4_0C06", 2, run)
