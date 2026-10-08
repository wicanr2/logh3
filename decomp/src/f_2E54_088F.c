/*
 * GIN3PS.EXE 2E54:088F  戰略 AI 某一格的「指派滿足率」（far，一個參數，回 AL = 0..99 或 0xFF）
 *
 * 參數：[bp+6] 一個堆疊字，只有低位元組有用（mov bl,[bp+6]；xor bh,bh），高位元組被忽略。
 *   idx = 星域編號（docs/re/170：ds:D58E／ds:D5AE 都是 32 byte 的星域表，這裡不限制 idx < 32）。
 * 讀取：A = byte ds:D58E[idx]（asm 寫成 [bx-2A72h]），B = byte ds:D5AE[idx]（[bx-2A52h]）。不寫入記憶體。
 *   A > B 不成立（無號，ja）              → AL = 0xFF（夠了）
 *   idx 低位元組是 0 或 0x1F（首都）       → 0
 *   其餘                                   → B × 100 ÷ A（無號除法，商 0..99；A > B >= 0 故 A 不為 0，不會除以 0）
 *
 * 回傳只有 AL。「夠了」那條路徑只寫 AL（mov al,0FFh），AH 保持進入時的值，不是函式的輸出；
 * 呼叫端 2E54:0842 也只存 AL（mov [bp+3],al）。描述檔用 ret = pack:al 只比 AL。
 *
 * 推論等級
 *   已證實（差異測試）：上列算式與回傳形狀。
 *   強推論（docs/re/170、172）：D58E = 這一格「要派幾支」、D5AE = 「已指派幾支」；回傳是滿足率百分比。
 *   名稱 need／have 只是變數名，不是證據。
 * 輸入域：idx 低位元組 0..255 全域（原版呼叫端只會傳 0..31，但函式本身不限制）。
 *   控制流只依賴 idx 的低位元組是否為 0、0x1F，其餘值只改變表格索引；抽樣取這兩個值、鄰近值、
 *   表尾之外的值與高位元組有值的變體當等價類代表。
 */
#include "gin3_reg.h"

uint8_t f_2E54_088F(Mem *m, uint16_t idx_word)
{
    uint8_t idx = (uint8_t)idx_word;
    uint8_t need = rd8(m, DS((uint16_t)(0xD58Eu + idx)));
    uint8_t have = rd8(m, DS((uint16_t)(0xD5AEu + idx)));

    if (!(need > have))
        return 0xFFu;
    if (idx == 0 || idx == 0x1F)
        return 0;
    return (uint8_t)(((uint16_t)have * 100u) / need);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2E54_088F(m, a[0]); }
GIN3_REGISTER("2E54_088F", 1, run)
