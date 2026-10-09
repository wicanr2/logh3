/*
 * GIN3PS.EXE 2B5F:0103  次月份造艦計劃：先算（2B5F:014D），參數非 0 時把各艦種的艘數抄進本座位的計劃表
 *
 * 原版：far 呼叫，一個堆疊字 flag（push bp／si／di、sub sp,4，所以參數在 [bp+0Eh]；`or di,di` 測整個 word，不是只測低位元組）。
 * 回傳：AX:BX ＝ 2B5F:014D 的回傳（實付總額，32 位元，AX 高字、BX 低字），中途暫存在 [bp+2]、[bp+0]，結尾還原，
 *   所以 3CA9:0002 的回傳值不影響結果。描述檔 ret = pack:ax,bx。
 *
 * 過程（已證實：反組譯逐指令；2B5F:014D、3CA9:0002 由其各自的差異測試證實）：
 *   total ← 2B5F:014D()
 *   flag != 0（整個 word）：
 *     F ← byte ds:8F1D（目前座位）；dst ← ds:(0x8F9C + F × 28)（16 位元乘法，F 最大 255 不會截斷）
 *     3CA9:0002(dst, ds:D01E, 28)  ← memmove：把 014D 剛算好的 ds:D01E..D039（七個艦種的艘數 × 10，每個 dword）抄進計劃表
 *     呼叫推入順序是 count、src 段、src 位移、dst 段、dst 位移（3CA9:0002 的參數由低位址到高位址是 dst、src、count）
 *   回 total
 * 副作用：014D 的寫入（ds:CFFE..D03A）；flag != 0 時再寫 ds:(8F9C + 28F) 起 28 bytes。
 * 欄位語意（強推論，docs/re/265、docs/internals/32）：ds:8F9C 起每座位 28 bytes 是造艦計劃表（七艦種 × dword 艘數）；
 *   名稱不是證據，只有位移與常數是證據。
 * 輸入域：同 014D（提督營運 [+0x0E] 不為 0；為 0 時原版 #DE，探針丟棄）。F 最大 255，目的位址最大 0x8F9C + 7140 = 0xAB80，
 *   小於來源 0xD01E，兩段不重疊，memmove 的方向（由位移大小決定）恆為由低往高。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint32_t f_2B5F_014D(Mem *m);
extern uint32_t f_3CA9_0002(Mem *m, uint16_t doff, uint16_t dseg, uint16_t soff, uint16_t sseg, uint16_t count);

uint32_t f_2B5F_0103(Mem *m, uint16_t flag)
{
    uint32_t total = f_2B5F_014D(m);

    if (flag != 0) {
        uint16_t f = rd8(m, DS(0x8F1Du));

        f_3CA9_0002(m, (uint16_t)(0x8F9Cu + (uint16_t)(f * 0x1Cu)), GIN3_DS, 0xD01Eu, GIN3_DS, 0x1Cu);
    }
    return total;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2B5F_0103(m, a[0]); }
GIN3_REGISTER("2B5F_0103", 1, run)
