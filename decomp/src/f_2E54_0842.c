/*
 * GIN3PS.EXE 2E54:0842  戰略 AI：在 32 格星域裡挑「指派滿足率」最低的一格（回該格編號，沒有候選回 0xFF）
 *
 * 原版：far 呼叫，沒有參數（push bp 的小函式），回 AL（AH 是殘值，呼叫者 2E54:07D8 只存 AL，描述檔 ret = pack:al）。
 * 區域變數（byte）：[bp+0] 目前格 j、[bp+1] 目前最佳格（起點 0xFF）、[bp+2] 目前最佳值（起點 0xFF）、[bp+3] 本格值。
 * 過程（已證實：反組譯逐指令；2E54:088F 由其自己的差異測試證實）：
 *   j ＝ 0 .. 0x1F（無號 byte 比較 jnb）：
 *     若 ds:D58E[j]（byte，asm 寫成 [bx-2A72h]）== 0 → 略過（這一格沒有需求）
 *     v ＝ 2E54:088F(j)（回 AL）
 *     v < 最佳值（無號 byte，jnb 跳過 >=）→ 最佳格 ← j、最佳值 ← v   ; 同值保留先出現的
 *   回最佳格。
 * 088F 的參數是 push ax（AH 為殘值），088F 只用低位元組（mov bl,[bp+6]），C 版傳 j。
 * 語意（強推論，docs/re/170、docs/internals/54）：D58E[j]＝這一格「要派幾支」、088F 回「已指派幾支 ÷ 要派幾支」的百分比
 *   （0xFF ＝ 夠了），本函式挑百分比最低的一格。名稱不是證據，只有位移與比較方向是證據。
 * 副作用：無寫入（088F 也不寫）。輸入域：ds:D58E、D5AE 表任意內容；基底狀態兩張表全 0（回 0xFF），由 seeds 與突變造出候選。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint8_t f_2E54_088F(Mem *m, uint16_t idx_word);

uint8_t f_2E54_0842(Mem *m)
{
    uint8_t best_idx = 0xFF;
    uint8_t best_val = 0xFF;
    uint8_t j;

    for (j = 0; j < 0x20u; j++) {
        if (rd8(m, DS((uint16_t)(0xD58Eu + j))) != 0) {
            uint8_t v = f_2E54_088F(m, j);

            if (v < best_val) {
                best_idx = j;
                best_val = v;
            }
        }
    }
    return best_idx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2E54_0842(m);
}
GIN3_REGISTER("2E54_0842", 0, run)
