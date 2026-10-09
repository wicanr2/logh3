/*
 * GIN3PS.EXE 0374:0AF2  某星域內所有現役艦隊的行動設成 4，並把它們各組艦艇的 +0x17 歸零（far，一個參數，無回傳值）
 *
 * 原版：push bp、push si、sub sp,0Eh，所以唯一的參數在 [bp+16h]（word，只取低位元組：cmp al,[bp+16h]，byte 比較）。
 * 過程：對艦隊表 ds:C074 起 32 筆（每筆 122 bytes）逐筆：
 *   艦隊[+0x04] == 0xFF（沒在用）→ 跳過
 *   艦隊[+0x06]（所在星域）!= 參數低位元組 → 跳過
 *   艦隊[+0x48] ← 4（byte，整個 byte 重寫）
 *   對 j ＝ 0, 1, … 當 j < 艦隊[+0x0F]（現有組數，byte 無號比較；每一輪重讀）：
 *     idx ＝ word 艦隊[+0x10 + 4j]（該組艦艇池索引，位移 16 位元回繞）
 *     艦艇 ＝ (ds:CFB6 : ds:CFB4 ＋ idx × 0x4E)，位移是 16 位元（mul cx 取低 16 位元，再加上位移）
 *     艦艇[+0x17] ← 0
 * 內圈每一輪都重讀組數、組索引、ds:CFB4、ds:CFB6：艦艇池指標若指到 DS 內，艦艇寫入可能蓋掉這幾個值
 *   （艦隊自己的組數、組索引或池指標），後面的輪次照改過的值走。本版逐輪重讀，與原版順序一致；
 *   差異測試的 seeds 造出這三種別名。
 *
 * 欄位（docs/internals/33、64）：艦隊[+0x04] ＝ 提督索引（0xFF ＝ 沒在用）、[+0x06] ＝ 所在星域、[+0x0F] ＝ 現有組數、
 *   [+0x10] 起每組 4 bytes（[0..1] 艦艇池索引）、[+0x48] 低四位 ＝ 艦隊行動；艦艇記錄 78 bytes，
 *   ds:CFB4／CFB6 ＝ 池的遠指標（位移、段）、艦艇[+0x17] ＝ 索敵等級 0–3（被看到 +1、沒被看到 −1）。
 *   docs/internals/33 第 118 行：ds:4100 只收 +0x48 ＝ 00／10／01／11，其餘行動不進實行常式；4 屬於「其餘」，名稱未知。
 * 所以語意（強推論）：回合處理 0374:0556 的耗損與全滅（0374:0608）與 0374:0919 之後、打行星或回復之前，把該星域的現役艦隊行動設成 4，
 *   並清掉各組艦艇的索敵等級。docs/re/346 第 25 行、docs/internals/62 第 12 行把這支記為「未讀」，本檔解出其行為。
 *   為什麼是這個值與時機，docs 沒有記載（未知）。
 * 呼叫者：0374:0556（台帳 callers＝1）。
 *
 * 模型限制：位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000）。艦艇池的段取自 ds:CFB6，
 *   可以是任何值，所以艦艇寫入走 w0AF2_put8；艦隊 +0x48 在 DS 內，不會到 ROM。
 *   字的存取（組索引）逐位元組各自回繞位移；本檔不涵蓋 ds:C074 表以外的艦隊指標。
 * 輸入域：參數高位元組任意；艦隊表與艦艇池內容由突變與 seeds 決定；組數 0..255 全域（> 10 時讀到 +0x38 之後的欄位，原版同樣如此）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）：位移、常數、迴圈與別名行為；強推論：欄位語意如上。
 */
#include "gin3_reg.h"

static void w0AF2_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_0374_0AF2(Mem *m, uint16_t region_w)
{
    uint8_t region = (uint8_t)region_w;
    uint16_t fleet = 0xC074;
    uint8_t i;

    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS((uint16_t)(fleet + 4))) != 0xFF && rd8(m, DS((uint16_t)(fleet + 6))) == region) {
            uint16_t grp = (uint16_t)(fleet + 0x10);
            uint8_t j = 0;

            wr8(m, DS((uint16_t)(fleet + 0x48)), 4);
            while (j < rd8(m, DS((uint16_t)(fleet + 0x0F)))) {
                uint16_t idx = (uint16_t)(rd8(m, DS(grp)) | (rd8(m, DS((uint16_t)(grp + 1))) << 8));
                uint16_t prod = (uint16_t)(idx * 0x4E);
                uint16_t seg = rd16(m, DS(0xCFB6));
                uint16_t off = (uint16_t)(rd16(m, DS(0xCFB4)) + prod);

                w0AF2_put8(m, LIN(seg, (uint16_t)(off + 0x17)), 0);
                j++;
                grp = (uint16_t)(grp + 4);
            }
        }
        fleet = (uint16_t)(fleet + 0x7A);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_0374_0AF2(m, a[0]);
    return 0;
}
GIN3_REGISTER("0374_0AF2", 1, run)
