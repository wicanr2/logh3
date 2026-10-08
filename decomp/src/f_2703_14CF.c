/*
 * GIN3PS.EXE 2703:14CF  艦艇記錄的飛彈與艦載機存量 ＝ 艦種表係數 × 艘數
 *
 * 原版：far 呼叫，兩個參數：艦艇記錄 far 指標（位移在前、段在後）、n（word）。
 * 無有意義的回傳（AX 是第二個乘積的高字，兩個呼叫端 2AD2:0029、1AAB:0FAF 都隨即用 mov ax 蓋掉）。
 * 語意（已證實：反組譯逐指令；docs/re/345 一致）：
 *   列 ＝ ds:4E3A + 記錄[+0x00] × 208 + 記錄[+0x04] × 26            （艦種表，位移 16 位元回繞）
 *   dword(記錄 +0x0E) ← 列[+0x10]（byte，零擴充）× n                 （指定，不是累加）
 *   dword(記錄 +0x12) ← 列[+0x14]（byte，零擴充）× n
 * 乘法走 3CC9:0009（取低 32 位元）；byte × word 不會溢位 32 位元，本版用 gin3_lmul 表達同一個結果。
 * 欄位名稱：記錄[+0x00]＝陣營、[+0x04]＝艦種；列[+0x10]＝每艘飛彈、[+0x14]＝每艘艦載機；
 *   記錄 +0x0E／+0x12 ＝ 飛彈／艦載機存量，來自 docs/re/345、docs/re/112（proven）。
 * 副作用：只寫記錄 +0x0E..+0x15 八個 bytes。第二個係數在第一組寫入之後才讀（asm 順序）。
 * 輸入域：全部可到達；陣營、艦種當 byte 零擴充，不檢查範圍。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

void f_2703_14CF(Mem *m, uint16_t rec_off, uint16_t rec_seg, uint16_t n)
{
    uint16_t row = (uint16_t)(0x4E3A + (uint16_t)(rd8(m, LIN(rec_seg, (uint16_t)(rec_off + 4))) * 0x1A));
    uint32_t missiles, fighters;

    row = (uint16_t)(row + (uint16_t)(rd8(m, LIN(rec_seg, rec_off)) * 0xD0));
    missiles = gin3_lmul(rd8(m, DS((uint16_t)(row + 0x10))), n);
    wr16(m, LIN(rec_seg, (uint16_t)(rec_off + 0x10)), (uint16_t)(missiles >> 16));
    wr16(m, LIN(rec_seg, (uint16_t)(rec_off + 0x0E)), (uint16_t)missiles);
    fighters = gin3_lmul(rd8(m, DS((uint16_t)(row + 0x14))), n);
    wr16(m, LIN(rec_seg, (uint16_t)(rec_off + 0x14)), (uint16_t)(fighters >> 16));
    wr16(m, LIN(rec_seg, (uint16_t)(rec_off + 0x12)), (uint16_t)fighters);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2703_14CF(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("2703_14CF", 3, run)
