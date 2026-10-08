/*
 * GIN3PS.EXE 28E0:095C  兩次亂數餘數相加再除以 100（far，三個參數，回 AX）
 *
 * 原版：far 呼叫，三個堆疊字（args[0..2]），push bp、push si、push di、sub sp,8，參數從 [bp+12h] 起：
 *   args[0] ＝ 等級 v（word，只取低位元組：cmp byte ptr [bp+12h],...；高位元組被忽略）
 *   args[1] ＝ 跨度甲（word，載入 SI）
 *   args[2] ＝ 跨度乙（word，載入 DI）
 * 過程：
 *   v >= 0xC8（200，無號 byte）→ 回 0，不呼叫亂數。
 *   span ＝ (v < 0x64（100）) ? 跨度甲 : 跨度乙
 *   d ＝ span + 1（16 位元）
 *   r1 ＝ rand() % d；r2 ＝ rand() % d（兩次都是 3CD5:000C，先 r1 後 r2；除法是 DX:AX ÷ CX，DX 先清 0，
 *     rand() 最大 0x7FFF，商不會溢位）
 *   回 (r1 + r2) ÷ 100（16 位元無號除法；r1、r2 都 <= 0x7FFF，和不會回繞）
 * 副作用：v < 200 時呼叫 3CD5:000C 兩次，種子 ds:94DC 被寫兩次；v >= 200 時不碰亂數。
 * 輸入域：span + 1 回繞成 0（span == 0xFFFF）時原版 #DE（INT 0），差異測試會丟棄；C 版此時回 0 只為避免 SIGFPE，
 *   該路徑不在比較範圍內。v、span 其餘值（含高位元組）全域。
 * 語意：docs/re/418 第 45 行：「28E0:095C 在能力>=200時回0且不取RND。其餘取兩次RND，各以跨度+1取餘數，…」，
 *   與本檔一致；docs 中 v 稱「能力」、兩個跨度稱升級的加值範圍，是推論，本檔不使用。
 * 推論等級：已證實（反組譯逐指令＋差異測試；rand 由 3CD5:000C 的差異測試證實）。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD5_000C(Mem *m);

uint16_t f_28E0_095C(Mem *m, uint16_t level_word, uint16_t span_a, uint16_t span_b)
{
    uint8_t level = (uint8_t)level_word;
    uint16_t span, d, r1, r2;

    if (level >= 0xC8u)
        return 0;
    span = (level < 0x64u) ? span_a : span_b;
    d = (uint16_t)(span + 1u);
    if (d == 0u)
        return 0;
    r1 = (uint16_t)(f_3CD5_000C(m) % d);
    r2 = (uint16_t)(f_3CD5_000C(m) % d);
    return (uint16_t)((uint16_t)(r1 + r2) / 100u);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_095C(m, a[0], a[1], a[2]); }
GIN3_REGISTER("28E0_095C", 3, run)
