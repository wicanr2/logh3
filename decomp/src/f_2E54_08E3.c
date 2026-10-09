/*
 * GIN3PS.EXE 2E54:08E3  戰略 AI 階段四：重算每個星域「要派幾支」ds:D58E（缺口 ÷ 5 ＋ 亂數 % 4）
 * （far，無參數，無回傳值；ret = none）
 *
 * 反組譯逐條對應（IDA sub_3B993）：
 *   3C9C:000A(ds:D58E, 0, 0x20)              ; memset 32 bytes 為 0（f_3C9C_000A）
 *   對 i = 0 … 31（byte 計數，cmp al,20h；jnb）：
 *     word ds:D50E[i] == 0 → 跳過（整個 word 比較，cmp word ptr [bx-2AF2h],0；不看星域擁有者）
 *     have ← word ds:D5CE[i]；need ← word ds:D54E[i]
 *     have < need（無號 word，jb） → ds:D58E[i] ← (byte)((word)(need − have) ÷ 5)   ; div cx，商 AL
 *     否則                         → ds:D58E[i] ← 0
 *     r ← 3CD5:000C()                          ; rand()，兩條路徑都會呼叫，各抽一次
 *     ds:D58E[i] ← (byte)(ds:D58E[i] + (r % 4))  ; cwd；idiv 4；餘數 DL。r 在 0..0x7FFF，餘數 0..3
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合、亂數種子 ds:94DC..94DF 的寫入）：上列算式、除數 5、餘數 4、
 *     byte 截斷與回繞、亂數在兩條路徑都呼叫。
 *   強推論（docs/re/165、174、176，docs/internals/54）：D50E = 階段四被 2D54:0000 換成的「攻擊目標分」，
 *     D54E／D5CE = 需求／現有，D58E = 要派幾支。名稱不是證據。
 *
 * 輸入域：全部輸入都是記憶體（無參數）。D50E／D54E／D5CE 取 word 全域；亂數種子 ds:94DC 取任意值。
 *   除法的商最大 0xFFFF ÷ 5 = 0x3333，不會 #DE；商只存低位元組。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);
extern uint16_t f_3CD5_000C(Mem *m);

void f_2E54_08E3(Mem *m)
{
    uint8_t i;

    (void)f_3C9C_000A(m, 0xD58Eu, GIN3_DS, 0, 0x20);
    for (i = 0; i < 0x20; i++) {
        uint16_t have, need;
        uint8_t base = 0;
        uint16_t r;

        if (rd16(m, DS((uint16_t)(0xD50Eu + 2u * i))) == 0)
            continue;
        have = rd16(m, DS((uint16_t)(0xD5CEu + 2u * i)));
        need = rd16(m, DS((uint16_t)(0xD54Eu + 2u * i)));
        if (have < need)
            base = (uint8_t)((uint16_t)(need - have) / 5u);
        wr8(m, DS((uint16_t)(0xD58Eu + i)), base);
        r = f_3CD5_000C(m);
        wr8(m, DS((uint16_t)(0xD58Eu + i)),
            (uint8_t)(rd8(m, DS((uint16_t)(0xD58Eu + i))) + (uint8_t)((int16_t)r % 4)));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2E54_08E3(m);
    return 0;
}
GIN3_REGISTER("2E54_08E3", 0, run)
