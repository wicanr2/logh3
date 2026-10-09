/*
 * GIN3PS.EXE 3870:033E  在描述槽表 ds:F254 起的 8 個槽（每槽 16 bytes）裡找第一個空槽，回傳槽編號，沒有空槽回 0xFF
 * 呼叫：far，無參數。回傳 AL：槽編號 0 到 7，或 0xFF（全滿）。AH 是進入時的 AH（函式只寫 AL），描述檔以 pack:al 只比 AL。
 * 語意：i 從 0 起，i < 8（無號 byte 比較），槽 i 的位址是 ds:F254 + 16 × i；槽的 [+7] 等於 0xFF 就是空槽，回傳 i。
 * 不寫記憶體。
 * 欄位語意：docs/re/135 第 79 行「代號 = 3870:033E()  配一個槽，0xFF ＝ 沒了」，呼叫者 3870:001A 會把 16 bytes 描述寫進
 *   ds:F254 + 代號 × 16。槽 [+7] ＝ 0xFF 為空閒（與 3870:0004 全填 0xFF 相符，強推論）。
 * 輸入域：全部。位移最大 0xF254 + 7 × 16 + 7 ＝ 0xF2CB，不跨 0xFFFF。推論等級：已證實（asm 與差異測試）。
 */
#include "gin3_reg.h"

uint8_t f_3870_033E(Mem *m)
{
    uint8_t i;

    for (i = 0; i < 8; i++) {
        if (rd8(m, DS(0xF254 + 16 * i + 7)) == 0xFF)
            return i;
    }
    return 0xFF;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_3870_033E(m);
}
GIN3_REGISTER("3870_033E", 0, run)
