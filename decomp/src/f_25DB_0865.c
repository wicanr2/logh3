/*
 * GIN3PS.EXE 25DB:0865  把艦隊的繪製順位推進它的圖示物件（有變才寫），回傳是否有寫
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄（122 bytes）的 far 指標，回 AX = 1 或 0。
 *   堆疊：序言 push bp／si、sub sp,2，所以第一個參數在 [bp+0Ah]（位移）與 [bp+0Ch]（段）。
 *
 * 語意（位移為已證實：本檔與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   p   ← 艦隊[+0x6B]（byte）
 *   obj ← 艦隊[+0x6C] 的 far 指標（位移 +0x6C、段 +0x6E）
 *   若 p == obj[+3]：回 0，什麼都不寫。
 *   否則：obj[+3] ← p；重讀 艦隊[+0x6C..0x6F] 得到 obj'（寫入 obj[+3] 之後才讀，obj 若蓋到艦隊記錄
 *         自己，obj' 可能不同於 obj）；obj'[+1] ← 0xFF；回 1。
 *   比較是 byte 對 byte（`cmp al,cl`），obj 的其他位元組不看。回傳是 AX = 0 或 1（AH = 0）。
 *
 * 欄位語意（強推論）：艦隊[+0x6B] 是戰場圖示的繪製順位，艦隊[+0x6C] 是這支艦隊的圖示顯示物件指標
 *   （docs/internals/33-fleet-table.md 第 122 行：`25DB:0865` 把 +0x6B 推進圖示物件的 [+3]；第 121 行：
 *   物件掛上時 [+3] ＝ 0F0h）；obj[+3] 是物件的優先值；obj[+1] 旗標的名稱為假說，值 0xFF 為已證實。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000）。基底狀態裡大部分艦隊的
 *     +0x6C 是 FFFF:FFFF，讀仍照基底傾印，寫入不動作，與執行器一致。
 *   - 字的存取跨越位移 0xFFFF 時，本版與探針都按線性位址連續處理，不涵蓋真機上回繞到段首的行為。
 *
 * 輸入域：艦隊指標取艦隊表任一筆（32 筆）；基底是戰略層（fcmd-base.state），沒有戰場顯示物件，
 *   物件指標全由基底值、突變與 seeds 供給，obj 可以指到任意位址，包括艦隊記錄自己。
 */
#include "gin3_reg.h"

static void put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

uint16_t f_25DB_0865(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint16_t ret = 0;
    uint8_t prio = rd8(m, LIN(fleet_seg, fleet_off + 0x6B));
    uint16_t a_off = rd16(m, LIN(fleet_seg, fleet_off + 0x6C));
    uint16_t a_seg = rd16(m, LIN(fleet_seg, fleet_off + 0x6E));
    uint16_t t_off = (uint16_t)(a_off + 3);
    uint8_t cur = rd8(m, LIN(a_seg, t_off));

    if (prio != cur) {
        uint16_t b_off, b_seg;

        put8(m, LIN(a_seg, t_off), prio);
        b_off = rd16(m, LIN(fleet_seg, fleet_off + 0x6C));
        b_seg = rd16(m, LIN(fleet_seg, fleet_off + 0x6E));
        put8(m, LIN(b_seg, b_off + 1), 0xFF);
        ret = 1;
    }
    return ret;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_25DB_0865(m, a[0], a[1]); }
GIN3_REGISTER("25DB_0865", 2, run)
