/*
 * GIN3PS.EXE 2E54:088F  戰略 AI 某一格的「指派滿足率」（far，一個參數，回 AL = 0..100 或 0xFF）
 *
 * 參數：[bp+6] 一個堆疊字，只有低位元組有用（mov bl,[bp+6]；xor bh,bh），高位元組被忽略。
 *   idx = 星域編號（docs/re/170：ds:D58E／ds:D5AE 都是 32 byte 的星域表，這裡不限制 idx < 32）。
 * 讀取：A = byte ds:D58E[idx]（asm 寫成 [bx-2A72h]），B = byte ds:D5AE[idx]（[bx-2A52h]）。不寫入記憶體。
 *   A > B 不成立（無號，ja）              → AL = 0xFF（夠了）
 *   idx 低位元組是 0 或 0x1F（首都）       → 0
 *   其餘                                   → B × 100 ÷ A（無號除法，商 0..99；A > B >= 0 故 A 不為 0，不會除以 0）
 *
 * 回傳 AX 的高位元組：「夠了」那條路徑只寫 AL（mov al,0FFh），AH 保持進入時的值；其餘兩條路徑
 *   AX 完整被覆寫。呼叫端 2E54:0842 只存 AL（mov [bp+3],al）。為了與原版 AX 逐位元一致，本版把
 *   進入時的 AX 當成第二個輸入 ax_in（spec 以 kind=reg 的 ax 輸入模擬），在該路徑保留它的高位元組。
 *
 * 推論等級
 *   已證實（差異測試）：上列算式與回傳形狀。
 *   強推論（docs/re/170、172）：D58E = 這一格「要派幾支」、D5AE = 「已指派幾支」；回傳是滿足率百分比。
 *   名稱 need／have 只是變數名，不是證據。
 * 輸入域：idx 低位元組 0..255 全域（原版呼叫端只會傳 0..31，但函式本身不限制）。
 */
#include "gin3_reg.h"

uint16_t f_2E54_088F(Mem *m, uint16_t idx_word, uint16_t ax_in)
{
    uint8_t idx = (uint8_t)idx_word;
    uint8_t need = rd8(m, DS((uint16_t)(0xD58Eu + idx)));
    uint8_t have = rd8(m, DS((uint16_t)(0xD5AEu + idx)));

    if (!(need > have))
        return (uint16_t)((ax_in & 0xFF00u) | 0xFFu);
    if (idx == 0 || idx == 0x1F)
        return 0;
    return (uint16_t)(((uint16_t)have * 100u) / need);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2E54_088F(m, a[0], a[1]); }
GIN3_REGISTER("2E54_088F", 2, run)
