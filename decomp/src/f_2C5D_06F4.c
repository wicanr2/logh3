/*
 * GIN3PS.EXE 2C5D:06F4  重設戰略層的顯示指標：32 支艦隊的 +0x6C/+0x6E 清 0，艦艇池 352 筆的 +0x19 bit6 與 +0x38/+0x3A 清 0
 *
 * 原版：far 呼叫，無參數，無回傳值（AX 是殘留值，呼叫者 2C5D:061A 不用；描述檔 ret=none）。
 * 過程：
 *   1. 艦隊表 ds:C074 起，32 筆，每筆 0x7A（122）bytes：
 *        艦隊[+0x6C] ← 0（word）、艦隊[+0x6E] ← 0（word）
 *   2. 艦艇池：指標在 ds:CFB4（位移）與 ds:CFB6（段），在第 1 步結束之後才讀（一次，複製到堆疊）；352（0x160）筆，每筆 0x4E（78）bytes：
 *        池[+0x19] ← 池[+0x19] & 0xBF（byte，清 bit6）
 *        池[+0x38] ← 0（word）、池[+0x3A] ← 0（word）
 *      池指標的位移用 16 位元運算（超過 0xFFFF 在同一段內回繞）；字的存取是兩次 byte 存取、位移各自 16 位元回繞（DS:FFFF 的 word 取 FFFF 與 0000，不跨到下一段；dosgolem cpu.read16／write16）。
 *      池指標指到線性 0xF0000 以上（dosgolem 的 ROM 區，寫入被安靜丟棄）時，寫入無效，唯讀取照基底內容。
 * 欄位語意（強推論）：艦隊 +0x6C/+0x6E ＝ 圖示顯示物件的 far 指標（docs/internals/33 第 75、121 行）；
 *   艦艇 +0x19 bit6 與 +0x38/+0x3A 的名稱未知。位移與常數已證實。
 * 輸入域：池指標任意（描述檔 seeds 另造基底池、段內回繞、ROM 區、ROM 邊界跨字寫入）；讀 A0000 到 BFFFF 的案例由探針丟棄。
 *   寫入落在本函式自己的程式碼或堆疊上的案例（突變把池指標改到那裡）不在比較範圍。
 * 證據：asm 逐指令。推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

static void w06f4_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w06f4_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w06f4_put8(m, LIN(seg, off), (uint8_t)v);
    w06f4_put8(m, LIN(seg, off + 1), (uint8_t)(v >> 8));
}

void f_2C5D_06F4(Mem *m)
{
    uint16_t fleet = 0xC074;
    uint16_t i;
    uint16_t pool_off;
    uint16_t pool_seg;

    for (i = 0; i < 0x20; i++) {
        w06f4_put16(m, GIN3_DS, (uint16_t)(fleet + 0x6C), 0);
        w06f4_put16(m, GIN3_DS, (uint16_t)(fleet + 0x6E), 0);
        fleet = (uint16_t)(fleet + 0x7A);
    }
    pool_seg = rd16(m, DS(0xCFB6));
    pool_off = rd16(m, DS(0xCFB4));
    for (i = 0; i < 0x160; i++) {
        uint32_t flags = LIN(pool_seg, pool_off + 0x19);

        w06f4_put8(m, flags, (uint8_t)(rd8(m, flags) & 0xBFu));
        w06f4_put16(m, pool_seg, (uint16_t)(pool_off + 0x38), 0);
        w06f4_put16(m, pool_seg, (uint16_t)(pool_off + 0x3A), 0);
        pool_off = (uint16_t)(pool_off + 0x4E);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_2C5D_06F4(m);
    return 0;
}
GIN3_REGISTER("2C5D_06F4", 0, run)
