/*
 * GIN3PS.EXE 203A:0424  戰略艦隊攻擊力（記憶體模型版）
 *
 * 語意見 docs/re/146、docs/internals/66。直接讀寫 1 MB 記憶體，指標以 (off, seg) 傳入，供通用差異測試使用。
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄的 far 指標，回 AX。
 * 艦艇池基底在 ds:CFB4（位移）與 ds:CFB6（段）；每次迴圈都重讀，因為原版就是這樣寫。
 */
#include "gin3_reg.h"

uint16_t f_203A_0424(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint16_t main_ships = 0;
    uint16_t slot = (uint16_t)(fleet_off + 0x10);
    uint8_t i = 0;

    while (i < rd8(m, LIN(fleet_seg, fleet_off + 0x0F))) {
        uint16_t index = rd16(m, LIN(fleet_seg, slot));
        uint16_t pool_off = rd16(m, DS(0xCFB4));
        uint16_t pool_seg = rd16(m, DS(0xCFB6));
        uint16_t ship = (uint16_t)(pool_off + (uint16_t)(index * 0x4Eu));
        uint8_t kind = rd8(m, LIN(pool_seg, ship + 4));

        if (kind == 0 || kind == 1) {
            main_ships = (uint16_t)(main_ships + rd16(m, LIN(pool_seg, ship + 8)));
        }
        i++;
        slot = (uint16_t)(slot + 4);
    }

    {
        uint32_t product = gin3_lmul(main_ships, rd8(m, LIN(fleet_seg, fleet_off + 0x40)));
        uint16_t low = (uint16_t)gin3_uldiv(product, 100);

        return (uint16_t)(low / 15u + 250u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_203A_0424(m, a[0], a[1]); }
GIN3_REGISTER("203A_0424", 2, run)
