/*
 * GIN3PS.EXE 25DB:0952  艦隊的圖示物件換圖塊：艦隊[+0x2F] 與物件[+0x0C] 不同時，改取圖塊池裡新的圖塊並更新物件，回是否有改
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄（122 bytes）的 far 指標（位移、段）。序言 push bp／si／di、sub sp,0Ah，
 *   所以指標在 [bp+14h]（位移）與 [bp+16h]（段）。回 AX = 1（有改）或 0；AH = 0（回傳走 `mov ax,[bp+0]`）。
 *
 * 語意（位移與順序為已證實：與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   cur ← 艦隊[+0x2F]（byte）；obj ← 艦隊[+0x38] 的 far 指標（位移 +0x38、段 +0x3A）；cl ← obj[+0x0C]（byte）。
 *   cur == cl：回 0，什麼都不寫。
 *   否則：(seg, off) ← 3982:0001(cur)（圖塊池第 cur 塊的遠指標：AX = 段、BX = 位移，只讀 cur 的低位元組）；
 *         obj'  ← 重讀 艦隊[+0x38..+0x3B]；obj'[+0x10] ← seg（word）；obj'[+0x0E] ← off（word）；
 *         cur'  ← 重讀 艦隊[+0x2F]（byte）；obj'' ← 再重讀；obj''[+0x0C] ← cur'；
 *         obj''' ← 再重讀；obj'''[+1] ← 0xFF；回 1。
 *   物件若蓋到艦隊記錄自己（例如 obj + 0x10 ＝ 艦隊 + 0x38），後面的重讀看得到前面的寫入，本版照原版的讀寫順序逐項重讀。
 *
 * 欄位語意（假說）：與 25DB:08A2 是同一個樣子（逐指令結構相同，只換了艦隊欄位：艦隊[+0x2F] 取代 [+0x6A]、
 *   指標 +0x38／+0x3A 取代 +0x6C／+0x6E；物件欄位 +0x0C、+0x0E、+0x10、+1 不變）。艦隊[+0x2F] 與指標 +0x38 各自
 *   是什麼，文件沒有記載（未知）；08A2 與 0952 的呼叫者都含 25DB:0547 與 25DB:0759（台帳）。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000），w0952_put8 不動作；讀仍照基底傾印。
 *     物件指標若是 FFFF:FFFF（艦隊記錄沒掛物件時的值，假說），obj[+0x0C] 讀的是 ROM 裡的位元組。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 行為，w0952_rd16／put16）。
 *
 * 輸入域：艦隊指標取艦隊表任一筆（32 筆）或種子指定的位址（含跨位移 0xFFFF）；物件指標由基底值、突變與種子供給，
 *   可以指到任意位址，包括艦隊記錄自己。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint32_t f_3982_0001(Mem *m, uint8_t index);

static uint16_t w0952_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0952_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w0952_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w0952_put8(m, LIN(seg, off), (uint8_t)v);
    w0952_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_25DB_0952(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint8_t cur = rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x2Fu)));
    uint16_t obj_off = w0952_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x38u));
    uint16_t obj_seg = w0952_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x3Au));
    uint32_t tile;

    if (cur == rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Cu))))
        return 0;
    tile = f_3982_0001(m, cur); /* AX = 段（高 16 位元）、BX = 位移（低 16 位元） */
    obj_off = w0952_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x38u));
    obj_seg = w0952_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x3Au));
    w0952_put16(m, obj_seg, (uint16_t)(obj_off + 0x10u), (uint16_t)(tile >> 16));
    w0952_put16(m, obj_seg, (uint16_t)(obj_off + 0x0Eu), (uint16_t)tile);
    cur = rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x2Fu)));
    obj_off = w0952_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x38u));
    obj_seg = w0952_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x3Au));
    w0952_put8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Cu)), cur);
    obj_off = w0952_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x38u));
    obj_seg = w0952_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x3Au));
    w0952_put8(m, LIN(obj_seg, (uint16_t)(obj_off + 1u)), 0xFF);
    return 1;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_25DB_0952(m, a[0], a[1]); }
GIN3_REGISTER("25DB_0952", 2, run)
