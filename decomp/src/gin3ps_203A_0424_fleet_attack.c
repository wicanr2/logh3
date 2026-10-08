/*
 * GIN3PS.EXE 203A:0424  戰略艦隊攻擊力（功能等價重寫，不是 byte-matching）
 *
 * 來源與證據：docs/re/146（公式）、docs/re/442（編譯器歸屬）、
 *   docs/spec/124-decomp-reference-oracle.md（本目錄的驗收契約）。
 * 原版呼叫慣例：far 呼叫，第一個參數是艦隊記錄的 far 指標（offset 在低位、segment 在高位），
 *   回傳 AX。艦艇池基底在 ds:CFB4（offset）與 ds:CFB6（segment），這裡改成參數。
 *
 * 資料版面（已證實，docs/internals/33、64）：
 *   艦隊記錄 122 bytes：+0x0F 編制槽數（byte）、+0x10 起每槽 4 bytes（前 2 bytes 是艦艇池索引）、
 *     +0x40 能力（byte）。
 *   艦艇記錄 78 bytes：+0x04 艦種（byte）、+0x08 艘數（word）。
 *
 * 保留的原版語意：
 *   1. 艘數合計是 16 位元累加，超過 65535 會回繞。
 *   2. 乘法與第一次除法用 32 位元（Datalight 的 long 輔助函式），
 *      但商只留低 16 位元（原版 mov [bp+4],bx），之後的 ÷15 與 +250 都是 16 位元。
 *   3. 迴圈計數是 byte，槽位指標的 offset 是 16 位元；本實作只涵蓋槽數 <= 12
 *      （記錄長度內）與池索引 < 352（池長度內）。超出這兩個範圍原版會讀到鄰近記憶體，
 *      本模型不涵蓋，測試向量也不產生。
 */
#include <stdint.h>

#include "gin3_model.h"

uint16_t gin3ps_fleet_attack(const uint8_t *fleet, const uint8_t *pool)
{
    uint16_t main_ships = 0;
    uint8_t i;

    for (i = 0; i < fleet[GIN3_FLEET_SLOT_COUNT]; i++) {
        const uint8_t *slot = fleet + GIN3_FLEET_SLOTS + (unsigned)i * 4u;
        uint16_t index = gin3_le16(slot);
        const uint8_t *ship = pool + (uint16_t)(index * (unsigned)GIN3_SHIP_SIZE);
        uint8_t kind = ship[GIN3_SHIP_KIND];

        if (kind == 0 || kind == 1) {
            main_ships = (uint16_t)(main_ships + gin3_le16(ship + GIN3_SHIP_COUNT));
        }
    }

    {
        uint32_t product = (uint32_t)main_ships * (uint32_t)fleet[GIN3_FLEET_ABILITY40];
        uint32_t quotient = product / 100u;
        uint16_t low = (uint16_t)quotient;

        return (uint16_t)(low / 15u + 250u);
    }
}
