/*
 * GIN3PS.EXE 2264:0EC3  要塞砲射程橢圓的長短半軸
 *
 * 原版：far 呼叫，參數 = 行星記錄 far 指標（位移、段）。寫兩個全域字，沒有回傳值
 *   （離開時 AX 剛好是短半軸，但兩個呼叫端 2264:0049、2264:0248 呼叫後都不使用，描述檔 ret = none）：
 *   ds:A772 ← 行星[+0x16] ÷ 200 ＋ 0x50      ; 長半軸（外圈），無號 16 位元除，商不會溢位
 *   ds:A774 ← ds:A772 ÷ 4                    ; 短半軸（兩次邏輯右移）
 *
 * 語意（已證實：docs/re/134 "2264:0EC3(行星)" 偽碼；docs/internals/50 "長半軸 a ＝ 行星[+0x16] ÷ 200 ＋ 80，
 *   短半軸 b ＝ a ÷ 4"，伊謝爾倫 12800 → 144／36；docs/internals/29 +0x16 防衛力上限；
 *   docs/re/145 "ds:A772 長半軸、ds:A774 短半軸"；本檔與反組譯逐條對過）。
 *   長半軸最大 65535 ÷ 200 ＋ 80 ＝ 407，不會 16 位元回繞。
 *
 * 輸入域：行星指標落在行星表 81 筆內（ds:5336/5338 起，一筆 82 bytes）；+0x16 的全部 16 位元值。
 */
#include "gin3_reg.h"

void f_2264_0EC3(Mem *m, uint16_t planet_off, uint16_t planet_seg)
{
    uint16_t semi_major = (uint16_t)(rd16(m, LIN(planet_seg, planet_off + 0x16)) / 200u + 0x50u);

    wr16(m, DS(0xA772), semi_major);
    wr16(m, DS(0xA774), (uint16_t)(semi_major >> 2));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2264_0EC3(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2264_0EC3", 2, run)
