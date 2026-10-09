/*
 * GIN3PS.EXE 297B:10DE  某陣營全部現役艦隊的維持費合計（回傳）與移動經費合計（寫進 ds:8F82 起的表）
 *
 * 原版：far 呼叫，一個堆疊字 faction（push bp／si、sub sp,0Eh，所以參數在 [bp+16h]；只用低位元組：
 *   cmp al,[bp+16h]、mov cl,[bp+16h]），回 AX:BX（32 位元，AX 高字、BX 低字，描述檔 ret = pack:ax,bx）。
 * 過程（已證實：反組譯逐指令；297B:116C、297B:12FB 由其各自的差異測試證實）：
 *   U ← 0；M ← 0（32 位元）
 *   對 i ＝ 0..31（byte，無號比較 jnb）：艦隊 ＝ ds:(0xC074 + 0x7A × i)
 *     艦隊[+4] != 0xFF 且 艦隊[+1] == faction（低位元組）：
 *       U ← U + 297B:116C(艦隊)        ; 回 AX:BX，add [..],bx／adc [..],ax（32 位元回繞相加）
 *       M ← M + 297B:12FB(艦隊)
 *   表[faction] ← M：ds:(0x8F84 + 4 × faction) ← M 高字，ds:(0x8F82 + 4 × faction) ← M 低字（先高字後低字；
 *     asm 寫成 [si-707Ch]、[si-707Eh]，si ＝ 4 × faction，16 位元回繞）
 *   回 U
 * 順序很重要：116C 會寫艦隊 +0x0A、+0x0C，而 12FB 讀 +0x0C（除非 +0x08 是 2 或 5），所以 12FB 讀到的是 116C 剛算好的移動經費。
 * 語意（強推論，docs/re/239、docs/internals/27）：回傳值＝該陣營艦隊維持費合計；ds:8F82 起每陣營 4 bytes 是移動經費合計。
 *   名稱不是證據，位移與呼叫順序才是。
 * 副作用：116C 對每支符合的艦隊寫 +0x0A、+0x0C；寫 ds:8F82／8F84 + 4 × faction。
 * 輸入域：faction 全 16 位元（只取低位元組）；艦隊表、艦艇池與艦種表取自記憶體。艦艇池指標指到 A0000 到 BFFFF 的案例
 *   由探針丟棄（同 297B:116C 的限制）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint32_t f_297B_116C(Mem *m, uint16_t fleet_off, uint16_t fleet_seg);
extern uint32_t f_297B_12FB(Mem *m, uint16_t fleet_off, uint16_t fleet_seg);

uint32_t f_297B_10DE(Mem *m, uint16_t faction_word)
{
    uint8_t faction = (uint8_t)faction_word;
    uint32_t upkeep = 0;
    uint32_t move = 0;
    uint16_t slot;
    uint8_t i;

    for (i = 0; i < 0x20u; i++) {
        uint16_t fleet = (uint16_t)(0xC074u + (uint16_t)(i * 0x7Au));

        if (rd8(m, DS((uint16_t)(fleet + 4u))) != 0xFF && rd8(m, DS((uint16_t)(fleet + 1u))) == faction) {
            upkeep += f_297B_116C(m, fleet, GIN3_DS);
            move += f_297B_12FB(m, fleet, GIN3_DS);
        }
    }
    slot = (uint16_t)(faction << 2);
    wr16(m, DS((uint16_t)(slot + 0x8F84u)), (uint16_t)(move >> 16));
    wr16(m, DS((uint16_t)(slot + 0x8F82u)), (uint16_t)move);
    return upkeep;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_10DE(m, a[0]); }
GIN3_REGISTER("297B_10DE", 1, run)
