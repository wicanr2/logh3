/*
 * GIN3PS.EXE 0FF4:131E  艦隊缺艦種時的替代：把「目前星域」設成艦隊所在星域，再呼叫 0FF4:1350 挑艦種（回 AL）
 *
 * 原版（50 bytes，far 呼叫，三個堆疊字：艦隊 far 指標（位移、段）與 a；push bp／si、沒有 sub sp，所以艦隊在 [bp+8]／[bp+0Ah]、a 在 [bp+0Ch]）：
 *   ES:BX ← 艦隊（les bx,[bp+8]）
 *   region_no ← byte 艦隊[+6]                       ; mov al,es:[bx+6]；xor ah,ah
 *   星域位移 ← 0xA79C ＋ region_no × 0x7E            ; mul cx（cx ＝ 0x7E），region_no <= 255 所以乘積不超過 16 位元
 *   ds:9874 ← DS（段，先寫）；ds:9872 ← 星域位移
 *   call 0FF4:1350(艦隊位移, 艦隊段, a)             ; 已 diff-tested，回 AL（挑到的艦種；列走完回 0xFF）
 *   回 AX 是 1350 的 AX（AL ＝ 結果，AH 是殘值）；描述檔 ret ＝ pack:al。
 * 與 0FF4:0E5A（先設 ds:9872／9874 再呼叫 0E84）同型，差別在星域編號取自艦隊[+6] 而不是參數，且呼叫的是 1350。
 * 副作用：ds:9874 ← 0x54B3、ds:9872 ← 星域位移（兩個 word，先段後位移），其餘寫入都在 1350（借用 0x9D000 的暫存區淨影響為零）。
 * 呼叫者（台帳）：2D76:0A44。docs/re/229-refit-chain-complete.md 第 27、95 行、docs/re/228-refit-callees.md 第 55、76 行：
 *   `0FF4:131E(艦隊, 艦種)`＝另一支替代；`0FF4:0E5A(星域, a) != 0 ? 艦種 = a : 艦種 = 0FF4:131E(艦隊, a)`。與本檔一致。
 *   艦隊[+6] 是所在星域（297B:131C 的檔頭與 docs/re/342 也以 +6 當星域）；名稱不是證據，位移與用途才是。
 * 輸入域：艦隊指標取艦隊表任一筆、8000:FFF0 這類跨界位置；a 任意 word（1350 取低位元組）；星域編號 0 到 255（>= 32 時指到星域表之後的記憶體，照原版運算）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint8_t f_0FF4_1350(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t a_word);

uint8_t f_0FF4_131E(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t a)
{
    uint8_t region_no = rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 6u)));
    uint16_t region = (uint16_t)(0xA79Cu + (uint16_t)(region_no * 0x7Eu));

    wr16(m, DS(0x9874u), GIN3_DS);
    wr16(m, DS(0x9872u), region);
    return f_0FF4_1350(m, fleet_off, fleet_seg, a);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0FF4_131E(m, a[0], a[1], a[2]); }
GIN3_REGISTER("0FF4_131E", 3, run)
