/*
 * GIN3PS.EXE 2D76:0C4F  在艦隊的編制組清單裡找「隊形槽」相同的組（從第 1 組找起）
 *
 * 原版：far 呼叫，三個參數（push bp、sub sp,6，所以參數從 [bp+0Ch] 起）：
 *   args[0..1] ＝ 艦隊記錄的 far 指標（位移、段）
 *   args[2]    ＝ kind（word，只取低位元組：mov dl,[bp+10h]）
 *   args[3]    ＝ slot（word，只取低位元組：cmp al,[bp+12h]）
 * 艦隊記錄 +0x0F ＝ 編制組數（byte），編制組從 +0x10 起每組 4 bytes；組內 +2 ＝ 艦種、+3 ＝ 隊形槽（docs/re/229 第 34 行
 * 的參數名稱：2D76:0C4F(艦隊, 艦種, 隊形槽)；欄位語意為強推論，位移已證實）。
 *
 * 過程（i 是 byte 計數器，從 1 開始，指標從 +0x14 開始，也就是跳過第 0 組）：
 *   i >= 艦隊[+0x0F]（無號 byte 比較，每圈重讀）→ 回 0
 *   組[+3] == slot：
 *     組[+2] == kind → 回 0xFFFF（槽相同、艦種也相同）
 *     否則 → 回 i（槽相同、艦種不同；零延伸）
 *   否則 i++，指標 += 4（16 位元回繞）
 * 組指標的位移用 16 位元運算（超過 0xFFFF 在同一段內回繞）。純函式，不寫記憶體。
 *
 * 輸入域：艦隊指標取艦隊表任一筆或任意位址（描述檔 seeds 另造）；kind、slot 全部 16 位元值（只有低位元組有效）。
 * 證據：asm 逐指令；docs/re/229、231、232 描述的三種回傳（0＝新增、非 0＝覆蓋同槽、−1＝槽與艦種都相同跳過）與本檔一致。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2D76_0C4F(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t kind_word, uint16_t slot_word)
{
    uint8_t kind = (uint8_t)kind_word;
    uint8_t slot = (uint8_t)slot_word;
    uint8_t i = 1;
    uint16_t p = (uint16_t)(fleet_off + 0x14);

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        if (rd8(m, LIN(fleet_seg, p + 3)) == slot) {
            if (rd8(m, LIN(fleet_seg, p + 2)) == kind)
                return 0xFFFFu;
            return i;
        }
        i++;
        p = (uint16_t)(p + 4);
    }
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2D76_0C4F(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("2D76_0C4F", 4, run)
