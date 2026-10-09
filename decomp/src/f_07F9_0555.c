/*
 * GIN3PS.EXE 07F9:0555  把一個陣營的 16 個艦隊槽「是不是空位」寫進 ds:11AA 起的表（far，一個參數，無回傳值）
 *
 * 原版：push bp、push si、sub sp,0Ah，所以唯一的參數在 [bp+12h]（word，只取低位元組：cmp byte ptr [bp+12h],0）。
 * 過程：
 *   起點槽 ＝ 參數低位元組 != 0 ? 16 : 0              （同盟 16..31、帝國 0..15，docs/internals/33「槽是固定分陣營的」）
 *   艦隊指標 ＝ 0xC074 ＋ 起點槽 × 0x7A                （16 位元，艦隊表 32 筆 × 122 bytes）
 *   輸出指標 ＝ 0x11AA
 *   重複 16 次：
 *     ds:(輸出指標) ← (艦隊[+0x04] == 0xFF) ? 1 : 0     （byte；+0x04 ＝ 提督索引，0xFF ＝ 沒在用，docs/internals/33）
 *     輸出指標 += 6、艦隊指標 += 0x7A
 * 也就是 ds:11AA 起每 6 bytes 一筆、共 16 筆，每筆第 0 個 byte 寫「該陣營第 k 槽是空位」；每筆的其餘 5 個 byte 不動。
 * 未知：ds:11AA 這張 6 bytes 一筆的表是什麼、其餘 5 個 byte 的語意、被誰讀。docs 沒有記載。
 *   唯一的呼叫者 07F9:043F（艦隊編成視窗，docs/internals/58 第 154 行、docs/spec/013 第 135 行）在呼叫前把
 *   提督記錄[+1]（陣營）起的 word 推入堆疊，所以參數的高位元組其實是提督記錄[+2]（功績累計的低位元組），
 *   靠本函式只取低位元組才正確。
 * 回傳：沒有回傳值（離開時 AX ＝ 最後一筆寫入的 0 或 1 是殘值，呼叫者不使用，不比較）。
 * 副作用：只寫 ds:11AA、11B0、…、1204 共 16 個 byte。
 * 輸入域：參數高位元組任意；艦隊表內容由突變決定，+0x04 任意 byte。全部可到達。
 * 推論等級：寫入位址、值與條件已證實（反組譯逐指令＋差異測試）；「槽是否空位」的語意是強推論（+0x04 ＝ 0xFF ＝ 沒在用，
 *   docs/internals/33 已證實）；ds:11AA 這張表的身分是未知。
 */
#include "gin3_reg.h"

void f_07F9_0555(Mem *m, uint16_t faction_w)
{
    uint16_t out = 0x11AA;
    uint16_t slot = ((uint8_t)faction_w != 0) ? 0x10u : 0u;
    uint16_t fleet = (uint16_t)(0xC074 + (uint16_t)(slot * 0x7A));
    uint8_t i;

    for (i = 0; i < 0x10; i++) {
        uint8_t free_slot = (rd8(m, DS((uint16_t)(fleet + 4))) == 0xFF) ? 1u : 0u;
        wr8(m, DS(out), free_slot);
        out = (uint16_t)(out + 6);
        fleet = (uint16_t)(fleet + 0x7A);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_07F9_0555(m, a[0]);
    return 0;
}
GIN3_REGISTER("07F9_0555", 1, run)
