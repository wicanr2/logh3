/*
 * GIN3PS.EXE 3982:0217  「參數 != ds:8F25」：陣營對座位 0 的陣營取反（0 或 1）
 *
 * 呼叫：far，一個參數 [bp+6]，原版只讀它的低位元組（mov cl,[bp+6]），與 ds:8F25 的位元組比較，回 AX（0 或 1）。
 * 讀 ds:8F25 一個 byte；無寫入。
 *
 * 語意：
 *   已證實（差異測試）：低位元組 == ds:8F25 → 0；否則 → 1。
 *   強推論：ds:8F25 是座位 0 的陣營（docs/internals/57、docs/internals/54、docs/re/332），
 *   所以回傳值是「這個陣營所在的座位」（座位 0 或 1），與 3982:022D 內部的座位換算同一個式子：
 *   座位 ＝ (ds:8F25 == 陣營) ? 0 : 1。docs 沒有專門記載 3982:0217，以上為本函式與
 *   3982:022D 反組譯的對照結果。唯一呼叫者 1ED2:0001。
 *   不特別處理陣營 2（中立）：2 也回 1（022D 才先排除 2）。
 *
 * 輸入域：參數字的高位元組原版不看，差異測試以全 16 位元隨機值餵入。
 */
#include "gin3_reg.h"

uint16_t f_3982_0217(Mem *m, uint8_t faction)
{
    return rd8(m, DS(0x8F25)) == faction ? 0 : 1;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3982_0217(m, (uint8_t)a[0]);
}
GIN3_REGISTER("3982_0217", 1, run)
