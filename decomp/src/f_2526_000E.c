/*
 * GIN3PS.EXE 2526:000E  戰場回合開頭：對「在場艦隊清單」的每一支艦隊套一次發現程度衰減（2526:005A）
 *
 * 原版：far 呼叫，無參數、無回傳值（描述檔 ret = none；AX 是殘值，唯一呼叫者 12AA:0245 之後不讀 AX）。序言 push bp／si／di、
 *   sub sp,8；區域變數 [bp+0]／[bp+2] ＝ 清單指標（位移從 0x9898 起、段 DS），[bp+4]／[bp+6] ＝ 艦隊指標暫存（寫了不再讀）。
 *   用到的全域變數：在場艦隊清單筆數 ds:9894（word，`cmp ax, ds:9894h` 比完整 word）、清單 ds:9898 起每筆一個 byte（艦隊編號）。
 *
 * 語意（已證實：與反組譯逐條對過；docs/internals/38-battle-units.md 第 304 行：12AA:0245 → 2526:000E → 2526:005A）：
 *   i ← 0，p ← 0x9898。當 i < ds:9894（無號 word 比較，每圈重讀）：
 *     id ← byte ds:p（`mov al, es:[bx]; xor ah,ah`）；艦隊 ← (ds:0xC074 ＋ id × 0x7A 的低 16 位元)，段 DS（艦隊表每筆 122 bytes）；
 *     call 2526:005A(艦隊)；i ← i + 1；p ← p + 1（位移 16 位元回繞）。
 *   2526:005A 的寫入可能蓋到 ds:9894 或清單本身，後面各圈重讀看得到，本版逐圈重讀。
 *
 * 呼叫 2526:005A 的方式：far 呼叫，兩個堆疊字：先推段（`push cx`，cx = DS）再推位移（`push si`），即 (位移, 段)。呼叫後 `add sp,4`。
 *
 * 比較是無號 word（jnb）；筆數 >= 0x8000 時要呼叫 005A 三萬多次，超出每案步數上限，有號與無號比較在可測範圍內無法區分（不涵蓋）。
 * 輸入域：ds:9894 的筆數（word）凍結在隨機突變之外，由種子給定 0 到 256（每圈要呼叫一次 005A，筆數太大會超出每案步數上限）；
 *   清單的各個 byte 與艦隊、艦艇資料由突變與種子涵蓋。基底狀態（戰略層）的 ds:9894 是 0，清單是空的。
 * 推論等級：已證實（反組譯逐指令＋差異測試，005A 本身另有差異測試）。
 */
#include "gin3_reg.h"

extern void f_2526_005A(Mem *m, uint16_t fleet_off, uint16_t fleet_seg);

void f_2526_000E(Mem *m)
{
    uint16_t i = 0;
    uint16_t p = 0x9898u;

    while (i < rd16(m, DS(0x9894u))) {
        uint16_t id = rd8(m, DS(p));
        uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(id * 0x7Au));

        f_2526_005A(m, fleet, GIN3_DS);
        i = (uint16_t)(i + 1u);
        p = (uint16_t)(p + 1u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2526_000E(m);
    return 0;
}
GIN3_REGISTER("2526_000E", 0, run)
