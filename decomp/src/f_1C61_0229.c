/*
 * GIN3PS.EXE 1C61:0229  艦艇的戰場世界座標與艦隊鏡頭座標不同步時，把艦艇對齊艦隊（回 AL：1 ＝ 有改、0 ＝ 沒改）
 *
 * 原版：far 呼叫，兩個 far 指標參數（序言 push bp／si／di、sub sp,8，所以第一個在 [bp+12h]／[bp+14h]，第二個在 [bp+16h]／[bp+18h]）：
 *   fleet（艦隊記錄，122 bytes 一筆）、ship（艦艇池記錄，78 bytes 一筆）。
 * 逐條對應反組譯：
 *   cx ← word 艦隊[+0x5C]；ax ← word 艦艇[+0x34]
 *   若 ax == cx：
 *       若 word 艦艇[+0x36] == word 艦隊[+0x5E] → 回 0（什麼都不寫）
 *   （否則，也就是座標不同）：
 *       word 艦艇[+0x34] ← cx
 *       word 艦艇[+0x28] ← cx
 *       ax ← word 艦隊[+0x5E]                 ; 重讀（艦隊與艦艇記錄若別名，會看到上面兩次寫入）
 *       word 艦艇[+0x36] ← ax
 *       word 艦艇[+0x2A] ← ax
 *       25DB:0714(ship)                       ; 把艦艇座標減鏡頭寫進它的顯示物件
 *       回 1
 * 艦艇 +0x34／+0x36 是現在座標、+0x28／+0x2A 是（強推論）同一對座標的另一份拷貝（目的地或前一格座標，名稱未驗），
 * 本函式把兩份都設成艦隊 +0x5C／+0x5E。
 *
 * 回傳：只比較 AL（描述檔 ret = pack:al）。AH 是殘值：沒改的路徑是艦艇 +0x37（Y 的高位元組），有改的路徑是 0714
 *   結束時的 AX（0714 沒有回傳值）。呼叫者 1C61:0153 只存 AL（mov [bp+4],al）。
 * 呼叫者：1C61:0153（docs/re/393-battle-work-control.md「依座位找首個對應陣營艦隊，以 FLEET+5C／+5E 計算鏡頭」的同族函式）。
 *
 * 推論等級
 *   已證實（差異測試）：欄位位移、比較與寫入的順序、艦隊 +0x5E 的重讀、呼叫 0714 的參數、回 1／0 的條件。
 *   強推論：艦隊 +0x5C／+0x5E 是艦隊的戰場座標（docs/re/393 稱 FLEET+5C／+5E）；艦艇 +0x34／+0x36 同 25DB:0714 的檔頭。
 *   未知：艦艇 +0x28／+0x2A 的語意。
 * 輸入域：兩個指標任意（位移 16 位元回繞，word 在位移 0xFFFF 時高位元組取同一段的位移 0）；本函式自己寫的 word 同樣處理。
 *   位址 0xF0000 以上（ROM）的寫入會被執行器丟棄；本版不涵蓋指標指到 ROM 區的情形（描述檔的指標取表內或種子指定的 RAM）。
 */
#include "gin3_reg.h"

extern void f_25DB_0714(Mem *m, uint16_t ship_off, uint16_t ship_seg);

static uint16_t f1C61_0229_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f1C61_0229_wr16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint8_t f_1C61_0229(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t ship_off, uint16_t ship_seg)
{
    uint16_t x_at = (uint16_t)(ship_off + 0x34u);
    uint16_t cx = f1C61_0229_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x5Cu));
    uint16_t ax = f1C61_0229_rd16(m, ship_seg, x_at);

    if (ax == cx) {
        uint16_t sy = f1C61_0229_rd16(m, ship_seg, (uint16_t)(ship_off + 0x36u));
        uint16_t fy = f1C61_0229_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x5Eu));

        if (sy == fy)
            return 0;
    }
    f1C61_0229_wr16(m, ship_seg, x_at, cx);
    f1C61_0229_wr16(m, ship_seg, (uint16_t)(ship_off + 0x28u), cx);
    ax = f1C61_0229_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x5Eu));
    f1C61_0229_wr16(m, ship_seg, (uint16_t)(ship_off + 0x36u), ax);
    f1C61_0229_wr16(m, ship_seg, (uint16_t)(ship_off + 0x2Au), ax);
    f_25DB_0714(m, ship_off, ship_seg);
    return 1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1C61_0229(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("1C61_0229", 4, run)
