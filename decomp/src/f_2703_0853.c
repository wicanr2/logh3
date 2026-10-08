/*
 * GIN3PS.EXE 2703:0853  重數各星域的艦隊數（駐軍計數）
 *
 * 原版：far 呼叫，無參數，無有意義的回傳（AX 只是最後一次 xor ah,ah 留下的殘值）。
 * 語意（已證實：反組譯逐指令；docs/re/40 手工解碼、docs/re/297、docs/internals/22 一致）：
 *   第一圈：對星域表 ds:A79C 起 32 筆（每筆 126 bytes），把 +0x29、+0x28 兩個 byte 清成 0。
 *   第二圈：對艦隊表 ds:C074 起 32 筆（每筆 122 bytes）逐筆：
 *     艦隊 +0x04 == 0xFF（沒在用）  → 跳過
 *     否則 星域 ＝ ds:A79C + 艦隊[+0x06] × 126（16 位元乘積、16 位元回繞）
 *          ds:(星域 + 0x28 + 艦隊[+0x01]) 的 byte 加一（8 位元回繞，位移 16 位元回繞）
 * 欄位名稱：+0x01 ＝ 陣營（0 帝國、1 同盟）、+0x06 ＝ 所在星域，來自 docs/re/40（confirmed）。
 *   陣營與星域編號都按原樣當 byte 零擴充使用，不檢查範圍；越界時寫入落到 DS 內的其他位移，
 *   C 版逐位元組照做（寫入後再讀艦隊欄位也照原順序重讀）。
 * 副作用：寫星域表 +0x28／+0x29，不寫其他地方。
 * 輸入域：全部可到達（沒有除法、沒有迴圈上限來自資料）。
 * 推論等級：已證實（差異測試）。
 * 注意：docs/re/40 的手工解碼把 `mov byte [bp+00],0 / mov word [bp+06],C074` 排在 loop2 標籤之後，
 *   實際反組譯它們在迴圈之前（只做一次）；語意是我們這裡寫的那樣。
 */
#include "gin3_reg.h"

void f_2703_0853(Mem *m)
{
    uint16_t region = 0xA79C;
    uint16_t fleet = 0xC074;
    uint8_t i;

    for (i = 0; i < 0x20; i++) {
        wr8(m, DS((uint16_t)(region + 0x29)), 0);
        wr8(m, DS((uint16_t)(region + 0x28)), 0);
        region = (uint16_t)(region + 0x7E);
    }

    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS((uint16_t)(fleet + 4))) != 0xFF) {
            uint16_t target = (uint16_t)(0xA79C + (uint16_t)(rd8(m, DS((uint16_t)(fleet + 6))) * 0x7E));
            uint32_t at;

            target = (uint16_t)(target + 0x28);
            target = (uint16_t)(target + rd8(m, DS((uint16_t)(fleet + 1))));
            at = DS(target);
            wr8(m, at, (uint8_t)(rd8(m, at) + 1));
        }
        fleet = (uint16_t)(fleet + 0x7A);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2703_0853(m);
    return 0;
}
GIN3_REGISTER("2703_0853", 0, run)
