/*
 * GIN3PS.EXE 2703:0214  重數一個星域「被敵方看見的」帝國／同盟艦隊數（星域 +0x2A／+0x2B）
 *
 * 原版：far 呼叫，一個參數＝星域編號（word，只讀低位元組）；無有意義的回傳（AX 是迴圈殘值）。
 * 語意（已證實：反組譯逐指令；docs/re/398、docs/internals/22 一致）：
 *   星域 ＝ ds:A79C + 編號 × 126（位移 16 位元回繞）；星域[+0x2B] ＝ 星域[+0x2A] ＝ 0
 *   對艦隊表 ds:C074 起 32 筆（每筆 122 bytes）：
 *     艦隊[+0x04] == 0xFF（沒在用）或 艦隊[+0x06] != 編號 → 跳過
 *     旗標 ＝ 0；對編制的 艦隊[+0x0F] 組（j 與 byte 比較，無號）：
 *       艦艇 ＝ [ds:CFB4]（位移）＋ word(艦隊[+0x10 + 4j]) × 78（16 位元），段 ＝ [ds:CFB6]
 *       艦艇[+0x17] != 0 → 旗標 ＝ 1              （只標記，不累加：一支艦隊最多加一）
 *     旗標 → 星域[+0x2A + 艦隊[+0x01]] 加一（byte，位移 16 位元回繞）
 *   艦艇池基底每圈重讀（原版就是這樣寫）。
 * 欄位名稱：艦隊 +0x01＝陣營、+0x04＝使用旗標、+0x06＝所在星域、+0x0F＝編制組數、+0x10 起每組 4 bytes
 *   首 word 為艦艇池索引（docs/re/40、docs/internals/33，confirmed）；艦艇 +0x17 ＝ 索敵等級（docs/re/398，
 *   proven）；星域 +0x2A／+0x2B ＝ 被敵方看見的帝國／同盟艦隊數（docs/internals/22）。
 * 副作用：寫星域 +0x2A／+0x2B；陣營當 byte 零擴充，越界時落到 DS 其他位移（逐位元組照做）。
 * 輸入域：編號當 byte 使用（高位元組被忽略），測試涵蓋 0..31 與數個超出範圍的值（位址 16 位元回繞）。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

void f_2703_0214(Mem *m, uint16_t region_no)
{
    uint8_t no = (uint8_t)region_no;
    uint16_t region = (uint16_t)(0xA79C + (uint16_t)(no * 0x7E));
    uint16_t fleet = 0xC074;
    uint8_t i;

    wr8(m, DS((uint16_t)(region + 0x2B)), 0);
    wr8(m, DS((uint16_t)(region + 0x2A)), 0);
    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS((uint16_t)(fleet + 4))) != 0xFF && rd8(m, DS((uint16_t)(fleet + 6))) == no) {
            uint16_t seen = 0;
            uint16_t slot = (uint16_t)(fleet + 0x10);
            uint8_t j;

            for (j = 0; j < rd8(m, DS((uint16_t)(fleet + 0x0F))); j++) {
                uint16_t ship_off = (uint16_t)(rd16(m, DS(slot)) * 0x4Eu);
                uint16_t pool_seg = rd16(m, DS(0xCFB6));
                uint16_t pool_off = rd16(m, DS(0xCFB4));

                ship_off = (uint16_t)(pool_off + ship_off);
                if (rd8(m, LIN(pool_seg, (uint16_t)(ship_off + 0x17))) != 0)
                    seen = 1;
                slot = (uint16_t)(slot + 4);
            }
            if (seen) {
                uint32_t at = DS((uint16_t)((uint16_t)(region + 0x2A) + rd8(m, DS((uint16_t)(fleet + 1)))));

                wr8(m, at, (uint8_t)(rd8(m, at) + 1));
            }
        }
        fleet = (uint16_t)(fleet + 0x7A);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2703_0214(m, a[0]);
    return 0;
}
GIN3_REGISTER("2703_0214", 1, run)
