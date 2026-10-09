/*
 * GIN3PS.EXE 2E54:1B17  對一支艦隊下命令（far，三個參數，無回傳值；描述檔 ret = none，AL 是殘值）
 *
 * 參數（堆疊字，由低位址到高位址；序言只有 push bp，所以第一個參數在 [bp+6]）：
 *   fleet    far 指標（位移、段）：艦隊記錄
 *   action   一個字，只用低位元組（mov cl,[bp+0Ah]）：行動碼
 *   region   一個字，只用低位元組（mov dl,[bp+0Ch]）：目標星域編號
 *
 * 反組譯逐條對應（IDA sub_3CBC7）：
 *   艦隊[+0x06] == region（byte 比較）：
 *     艦隊[+0x08] == 5 → 什麼都不做
 *     否則             → 艦隊[+0x08] ← 5；艦隊[+0x07] ← 0xFF
 *   否則：
 *     艦隊[+0x08] ← action；艦隊[+0x07] ← region
 *   位移一律 16 位元（add bx,8 與 [bx+7] 各自在段內回繞）。
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：兩條分支、欄位位移、常數 5 與 0xFF。
 *   強推論（docs/internals/54「挑格與指派」、docs/re/172）：+0x06 是艦隊目前所在星域，+0x07 是目標星域，
 *     +0x08 是行動碼（5 = 就地待命，docs/re/79），0xFF = 沒有目標。
 *   未知：action 的其餘值各自的意義（呼叫端傳 0、1、2、3，docs/re/177）。
 *
 * 輸入域：艦隊指標取艦隊表任一筆或任意位址（含位移靠近 0xFFFF 的回繞、位於 BIOS ROM 區）。
 *   線性 0xF0000 以上是 ROM，寫入在 dosgolem 被安靜丟棄（internal/machine/bios_rom.go）；本版用 w1B17_put8 表達。
 *   +0x07 在這個函式裡只被寫、不被讀，探針不會突變它；描述檔 seeds 先 poke 它的前態才能看出「少寫」。
 */
#include "gin3_reg.h"

/* 0xF0000 以上是 ROM：寫入無效。 */
static void w1B17_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

void f_2E54_1B17(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t action_w, uint16_t region_w)
{
    uint8_t action = (uint8_t)action_w;
    uint8_t region = (uint8_t)region_w;

    if (rd8(m, LIN(f_seg, f_off + 6u)) == region) {
        uint16_t st = (uint16_t)(f_off + 8u);

        if (rd8(m, LIN(f_seg, st)) != 5) {
            w1B17_put8(m, LIN(f_seg, st), 5);
            w1B17_put8(m, LIN(f_seg, f_off + 7u), 0xFF);
        }
    } else {
        w1B17_put8(m, LIN(f_seg, f_off + 8u), action);
        w1B17_put8(m, LIN(f_seg, f_off + 7u), region);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2E54_1B17(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("2E54_1B17", 4, run)
