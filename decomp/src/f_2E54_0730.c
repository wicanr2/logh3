/*
 * GIN3PS.EXE 2E54:0730  戰略 AI 階段三：算每個星域「還缺幾支」ds:D58E（far，無參數，無回傳值；ret = none）
 *
 * 反組譯逐條對應（IDA sub_3B7E0）：
 *   3C9C:000A(ds:D58E, 0, 0x20)              ; memset 32 bytes 為 0（f_3C9C_000A，已通過差異測試）
 *   p ← ds:A79C（星域表，一筆 126 = 0x7E bytes）；對 i = 0 … 31（byte 計數，cmp al,20h；jnb）：
 *     星域[+0x12] != ds:8F1D（byte）→ 跳過
 *     front ← (byte)(word ds:D50E[i] >> 1)   ; 注意只取右移之後的低位元組（mov [bp+1],al）
 *     front >= 2（無號 byte，jb 反向）→ front ← 1
 *     have ← word ds:D5CE[i]；need ← word ds:D54E[i]
 *     have < need（無號 word，jb）→ ds:D58E[i] ← (byte)((word)(need − have) >> 2)
 *     否則                         → ds:D58E[i] ← 0
 *     ds:D58E[i] ← (byte)(ds:D58E[i] + front)   ; byte 加法，超過 255 回繞
 *   p ← p + 0x7E（16 位元）
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：上列算式、右移位數、byte 截斷與回繞、比較方向與常數。
 *   強推論（docs/re/165、172，docs/internals/54）：D50E = 優先度（敵方鄰居個數 × 4）、D54E = 需求（敵方下一步
 *     會在這一格的槽數）、D5CE = 現有（決定留守的支數）、D58E = 這一格要派幾支、front = 是不是前線。
 *     名稱不是證據，只有位移與算式是證據。
 *   與舊文件的差異：docs/re/172、internals/54 寫 front = min(D50E ÷ 2, 1)。原版是先把 word 右移一位，
 *     只留低位元組，再與 2 比較。D50E 的低位元組是 0（例如 0x0200）時 front 是 0，不是 1。
 *     遊戲內 D50E 的範圍（docs/re/171 的「鄰居個數 × 4」）達不到，屬輸入域邊緣。
 *
 * 輸入域：全部輸入都是記憶體（無參數）。ds:8F1D 取 0..255 全域；星域表 32 筆全掃，不涉及 16 位元回繞
 *   （p 最大 0xA79C + 31 × 0x7E + 0x12 = 0xB6F0）；D50E／D54E／D5CE 取 word 全域。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);

void f_2E54_0730(Mem *m)
{
    uint8_t i;
    uint16_t p = 0xA79Cu;

    (void)f_3C9C_000A(m, 0xD58Eu, GIN3_DS, 0, 0x20);
    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS((uint16_t)(p + 0x12u))) == rd8(m, DS(0x8F1D))) {
            uint8_t front = (uint8_t)(rd16(m, DS((uint16_t)(0xD50Eu + 2u * i))) >> 1);
            uint16_t have = rd16(m, DS((uint16_t)(0xD5CEu + 2u * i)));
            uint16_t need = rd16(m, DS((uint16_t)(0xD54Eu + 2u * i)));
            uint8_t gap = 0;

            if (front >= 2)
                front = 1;
            if (have < need)
                gap = (uint8_t)((uint16_t)(need - have) >> 2);
            wr8(m, DS((uint16_t)(0xD58Eu + i)), gap);
            wr8(m, DS((uint16_t)(0xD58Eu + i)), (uint8_t)(rd8(m, DS((uint16_t)(0xD58Eu + i))) + front));
        }
        p = (uint16_t)(p + 0x7Eu);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2E54_0730(m);
    return 0;
}
GIN3_REGISTER("2E54_0730", 0, run)
