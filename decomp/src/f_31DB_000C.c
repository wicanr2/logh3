/*
 * GIN3PS.EXE 31DB:000C  收集戰場地圖上圖塊 1 的位置，存到 ds:D70C 起的陣列，並把個數存進 ds:D78C
 *
 * 原版：far 呼叫，沒有參數。序言沒有 C 的 push bp，指令只有：
 *   call 1588:05A7(tile = 1, dst = ds:D70C（位移 0xD70C、段 = DS）, 第三個參數 = 0x0020)
 *   add sp,8；ds:D78C（byte）← AL（個數的低位元組）；retf。
 * 回傳：AX = 1588:05A7 的回傳值（個數，word），沒有被別的指令改動，所以 ret = ax 比對整個 AX。
 *
 * 語意（已證實：反組譯逐指令；1588:05A7 由其自己的差異測試證實）：
 *   - 1588:05A7 掃整張戰場地圖（ds:3C1A／3C1C 指向的 128 × 102 格），每個等於 1 的格寫兩個 word（欄 × 16、列 × 20）。
 *     第三個參數 0x20 被 1588:05A7 載入 DI 後完全沒用到（docs/internals/67 的 1588:05A7 一列），迴圈沒有數量上限。
 *   - 個數超過 32 時輸出會越過 ds:D78C（D70C + 32 × 4 = D78C），之後 ds:D78C 才被個數的低位元組蓋掉（先呼叫、後存）。
 *   - 個數 >= 256 時存進 ds:D78C 的只有低位元組，AX 仍是完整個數。
 * 欄位語意（強推論，名稱不是證據）：ds:D70C 起 32 筆 × 4 bytes 是「圖塊 1」（待命點？）的座標清單，ds:D78C 是個數。
 *   docs 是否有記載：grep 31DB:000C 在 docs/ 沒有命中（只有台帳的呼叫關係）。
 * 副作用：寫 ds:D70C 起 4 × 個數 bytes、ds:D78C。
 * 輸入域：地圖指標取基底狀態（fcmd-standby.state，戰場地圖已載入）與其突變；ds:D70C 起的輸出區在 DS 內，
 *   不涉及 ROM 區。讀到 A0000 到 BFFFF（視訊記憶體）的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_05A7(Mem *m, uint16_t tile, uint16_t dst_off, uint16_t dst_seg, uint16_t max_unused);

uint16_t f_31DB_000C(Mem *m)
{
    uint16_t count = f_1588_05A7(m, 0x0001u, 0xD70Cu, GIN3_DS, 0x0020u);

    wr8(m, DS(0xD78Cu), (uint8_t)count);
    return count;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_31DB_000C(m);
}
GIN3_REGISTER("31DB_000C", 0, run)
