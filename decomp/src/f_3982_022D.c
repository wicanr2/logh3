/*
 * GIN3PS.EXE 3982:022D  座位的操作模式：陣營 2 回 0xFF，其餘取 ds:8F21[座位]
 *
 * 呼叫：far，一個參數 [bp+6]，原版只讀它的低位元組（mov cl,[bp+6]），回 AL（位元組）。
 * 讀 ds:8F25、ds:8F21、ds:8F22（各一個 byte）；無寫入。39 個呼叫者。
 *
 * 語意：
 *   已證實（差異測試）：
 *     低位元組 == 2                 → AL = 0xFF（不讀記憶體）
 *     否則 座位 = (ds:8F25 == 低位元組) ? 0 : 1，AL = ds:(8F21 + 座位)
 *   強推論（docs/internals/57、docs/re/332、docs/re/295、docs/re/330）：
 *     ds:8F25 是座位 0 的陣營、ds:8F21／8F22 是兩個座位的操作模式（0 ＝ 人類座位，1 ＝ 電腦座位），
 *     陣營 2 是中立。呼叫端用「回傳值 == 0」判斷人類座位，「== 1」判斷電腦座位。
 *     兩格內容的完整值域（除 0、1 外是否有別的值）未知。
 *
 * 暫存器：每條出口都只寫 AL，AH 保留進入時的值（mov al,0FFh；mov al,ds:8F25；mov al,[bx-70DFh]
 * 都不碰 AH）。呼叫端只用 AL。本檔核心函式只回 AL；run() 以 ax_in 的高位元組補上 AH，
 * 讓整個 AX 都被差異測試比對。[bx-70DFh] 的 16 位元位移是 0x8F21（0x10000 - 0x70DF），bx 只會是 0 或 1。
 *
 * 輸入域：參數字的高位元組原版不看，差異測試以全 16 位元隨機值餵入。
 */
#include "gin3_reg.h"

uint8_t f_3982_022D(Mem *m, uint8_t faction)
{
    uint16_t seat;

    if (faction == 2)
        return 0xFF;
    seat = rd8(m, DS(0x8F25)) == faction ? 0 : 1;
    return rd8(m, DS(0x8F21 + seat));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    uint16_t ax_in = a[1];

    return (uint16_t)((ax_in & 0xFF00u) | f_3982_022D(m, (uint8_t)a[0]));
}
GIN3_REGISTER("3982_022D", 2, run)
