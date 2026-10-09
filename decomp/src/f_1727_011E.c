/*
 * GIN3PS.EXE 1727:011E  戰場艦艇清單內每一艘的 +0x19 清掉 bit6（小地圖標記「目前點著」）
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none）。序言 push bp／si／di、sub sp,8；沒有參數。
 *   清單由全域變數指定：ds:9896（word）＝ 筆數；ds:98BA 起每筆一個 word ＝ 艦艇池索引。
 *   艦艇池基底 ds:CFB4（位移）、ds:CFB6（段），一筆 78（0x4E）bytes。
 *
 * 語意（已證實：本檔與反組譯逐條對過；docs/internals/38 "+0x19 小地圖標記的狀態機：bit7 ＝ 要點、
 *   bit6 ＝ 目前點著、bit3–0 ＝ 點下去的顏色"；ds:9896／ds:98BA 的角色為強推論：
 *   docs/re/339 §3 "12AA:03BC … ds:9894 ← 0；ds:9896 ← 0" 一段重建戰場表，ds:9896 在其中被歸零後累計）：
 *   for i ＝ 0; i < word ds:9896（無號，每圈重讀）; i++：
 *       idx ＝ word ds:(98BA ＋ 2i)                           ; 位移 16 位元
 *       rec ＝ (ds:CFB4) ＋ (idx × 0x4E 的低 16 位元)          ; 池段 ds:CFB6
 *       byte rec[+0x19] ← rec[+0x19] AND 0xBF
 *   與 1727:0298（擦小地圖點）的關係未在此驗證。
 *
 * 原版的細節（本版照做）：
 *   - 筆數與清單位移每圈重讀、遞增；清單位移是 16 位元，不跨段。
 *   - 池位移的乘法只留低 16 位元，rec＋0x19 也是 16 位元回繞。
 *   - 池段被改到 0xF000 以上時，寫入落在 BIOS ROM 區（線性 0xF0000 起），機器忽略該寫入
 *     （dosgolem internal/machine/bios_rom.go）；本版用 wr8_skip_rom 表達。平面視訊記憶體的寫入由探針丟棄。
 *
 * 輸入域：基底狀態的 ds:9896 是 0（戰略層沒有戰場清單），所以清單由種子與突變構造：筆數 0 到約 11,000
 *   （超過探針的 200,000 步上限的案例被丟棄）。清單位移 98BA＋2i 在筆數上限內不會超過 0xFFFF，所以位移
 *   回繞在這個函式裡到不了，不在驗證範圍。池索引可到 0xFFFF。
 */
#include "gin3_reg.h"

/* 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go，真機同）。gin3_mem.h 的 wr8 沒有這個語意（回報給主工作階段）。 */
static void wr8_skip_rom(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_1727_011E(Mem *m)
{
    uint16_t i = 0;
    uint16_t p = 0x98BAu;

    while (i < rd16(m, DS(0x9896))) {
        uint16_t idx = rd16(m, DS(p));
        uint16_t rec = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));
        uint16_t seg = rd16(m, DS(0xCFB6));

        wr8_skip_rom(m, LIN(seg, rec + 0x19u), (uint8_t)(rd8(m, LIN(seg, rec + 0x19u)) & 0xBFu));
        i++;
        p = (uint16_t)(p + 2u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_1727_011E(m);
    return 0;
}
GIN3_REGISTER("1727_011E", 0, run)
