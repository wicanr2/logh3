/*
 * GIN3PS.EXE 297B:12FB  艦隊「移動經費」單項（297B:10DE 逐艦隊加總的一項）
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄的 far 指標（位移在前、段在後）。
 * 回傳：32 位元放在 AX:BX（AX＝高字、BX＝低字），不是 DX:AX。DX 同樣被寫成高字，
 *   唯一的呼叫者（297B:10DE）只用 BX 與 AX：add [..],bx / adc [..],ax。所以 ret 取 pack:ax,bx。
 * 語意（已證實：反組譯逐指令；docs/re/239、docs/internals/27 的描述一致）：
 *   艦隊 +0x08（移動狀態）== 2 或 == 5 → 0
 *   其餘                                → 艦隊 +0x0C 的 word，以「有號」sign-extend 成 32 位元（cwd）
 * 注意（反組譯直接可見）：+0x0C 是 word 且被 cwd 符號延伸，值 >= 0x8000 時回負數。
 *   docs 只寫「回 艦隊[+0x0C]」，沒記這個符號延伸；實際資料是否出現 >= 0x8000 未驗。
 * 艦隊記錄 ds:C074 起 32 筆 × 122 bytes（docs/internals/33）。
 * 副作用：無。輸入域：全部可到達。
 */
#include "gin3_reg.h"

uint32_t f_297B_12FB(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t state = rd8(m, LIN(fleet_seg, fleet_off + 0x08));
    uint16_t cost = 0;

    if (state != 2 && state != 5)
        cost = rd16(m, LIN(fleet_seg, fleet_off + 0x0C));
    return (uint32_t)(int32_t)(int16_t)cost;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_12FB(m, a[0], a[1]); }
GIN3_REGISTER("297B_12FB", 2, run)
