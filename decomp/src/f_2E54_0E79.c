/*
 * GIN3PS.EXE 2E54:0E79  戰略 AI 的「清掉已到達／在途決定」：對本方每支現役艦隊，把 +0x79 是 4 或 6 的改成 0xFF
 * （far，無參數，無回傳值；描述檔 ret = none。AX 在 retf 時是迴圈殘值，不是輸出）
 *
 * 反組譯逐條對應（IDA sub_3BF29）：
 *   p ← ds:C074（艦隊表，一筆 122 = 0x7A bytes）；對 i = 0 … 31（byte 計數，cmp dl,20h；jnb）：
 *     艦隊[+0x04] == 0xFF                 → 跳過（空位）
 *     艦隊[+0x01] != ds:8F1D（byte）      → 跳過（不是本方）
 *     艦隊[+0x79] ∈ {4, 6}                → 艦隊[+0x79] ← 0xFF
 *     p ← p + 0x7A（16 位元）
 *   只在 +0x79 是 4 或 6 時寫入，其他值（含 5、7、0xFF）不動。
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：迴圈結構、欄位位移、比較值與寫入值。
 *   強推論（docs/re/170、docs/internals/54）：+0x01 是陣營、ds:8F1D 是目前執行 AI 的陣營、+0x79 是
 *     「本回合決定」狀態碼（4、6 與 docs/internals/54 的「已到達」「待命」相符）、+0x04 == 0xFF 是空位。
 *     名稱不是證據，只有位移與比較值是證據。
 *   未知：4 與 6 各自的確切語意（docs 對 6 寫「已經到了」，對 4 沒有獨立說明）。
 *
 * 輸入域：全部輸入都是記憶體（無參數）。艦隊表 32 筆全掃；ds:8F1D 取 0..255 全域。
 *   艦隊指標 p 的位移最大 0xC074 + 31 × 0x7A + 0x79 = 0xCFB3，不會超過 0xFFFF，沒有回繞問題。
 */
#include "gin3_reg.h"

void f_2E54_0E79(Mem *m)
{
    uint8_t i;
    uint16_t p = 0xC074u;

    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS((uint16_t)(p + 4u))) != 0xFFu && rd8(m, DS((uint16_t)(p + 1u))) == rd8(m, DS(0x8F1D))) {
            uint8_t st = rd8(m, DS((uint16_t)(p + 0x79u)));

            if (st == 4 || st == 6)
                wr8(m, DS((uint16_t)(p + 0x79u)), 0xFFu);
        }
        p = (uint16_t)(p + 0x7Au);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2E54_0E79(m);
    return 0;
}
GIN3_REGISTER("2E54_0E79", 0, run)
