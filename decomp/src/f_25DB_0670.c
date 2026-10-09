/*
 * GIN3PS.EXE 25DB:0670  艦隊的繪製方向與圖示物件位置：設 艦隊[+0x6A]，把圖示物件旗標設成 0xFF，並寫物件座標 +6／+8
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄（122 bytes）的 far 指標（位移、段）。序言 push bp／si／di、sub sp,0Ch，
 *   所以指標在 [bp+16h]（位移）與 [bp+18h]（段）。無回傳值（描述檔 ret = none；AX 是殘值，三個呼叫者
 *   1C03:006A、25DB:00FB、25DB:0547 之後都不讀 AX 的語意）。用到的全域變數：鏡頭 ds:9BF4（X）、ds:9BF6（Y）、
 *   方向表 ds:4A3C 起（8 筆，每筆 4 bytes＝ X 位移 word、Y 位移 word）。
 *
 * 語意（位移與順序為已證實：與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   obj ← 艦隊[+0x6C] 的 far 指標（位移 +0x6C、段 +0x6E）。位移與段都是 0（`or`）：直接返回，什麼都不做。
 *   d ← 3982:0022(艦隊[+0x4A] 的低位元組)（0..7）。
 *   艦隊[+0x6A] 不等於 9 才改寫成 d（byte 比較）。
 *   obj[+1] ← 0xFF。                                    （obj 指標在這之前讀一次）
 *   x ← 艦隊[+0x5C] − ds:9BF4 ＋ word ds:(4A3C ＋ 4d)；obj'[+6] ← x。（obj' 重讀 艦隊[+0x6C..+0x6F]：寫了 obj[+1] 之後）
 *   y ← 艦隊[+0x5E] − ds:9BF6 ＋ word ds:(4A3C ＋ 4d ＋ 2)；obj''[+8] ← y。（obj'' 再重讀：寫了 obj'[+6] 之後；y 的
 *       各項也是在 obj'[+6] 寫入之後才讀）
 *   obj 若蓋到艦隊記錄、鏡頭、方向表或自己的指標，後面的讀取看得到前面的寫入，本版照原版的讀寫順序逐項重讀。
 *
 * 欄位語意（強推論）：艦隊[+0x4A] 是朝向（256 刻），艦隊[+0x5C]／[+0x5E] 是世界座標，艦隊[+0x6A] 是 8 方向繪製索引
 *   （9 表示「無方向」，與 25DB:04E7 對應），obj 是戰場上這支艦隊的圖示顯示物件（obj[+6]／[+8] 是畫面座標，
 *   方向表的 8 筆是 8 方向上的小位移，皆為 ±8／±16 的數值）。docs/internals/34-battle-screen.md 第 344 到 346 行
 *   描述 25DB:0E78 配置顯示物件；本函式沒有逐位元組記載。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000），w0670_put8 不動作；讀仍照基底傾印。
 *     基底狀態的大部分艦隊 +0x6C 是 FFFF:FFFF，寫入因此落在 ROM。
 *   - word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 行為，w0670_rd16／put16）。
 *
 * 輸入域：艦隊指標取艦隊表任一筆（32 筆）或種子指定的位址（含跨位移 0xFFFF）；物件指標由基底值、突變與種子供給，
 *   可以指到任意位址，包括艦隊記錄、鏡頭、方向表與指標自己。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3982_0022(uint8_t angle);

static uint16_t w0670_rd16(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0670_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w0670_put16(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w0670_put8(m, LIN(seg, off), (uint8_t)v);
    w0670_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_25DB_0670(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint16_t dir, ax, obj_off, obj_seg;

    if ((w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x6Cu)) |
         w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x6Eu))) == 0)
        return;
    dir = f_3982_0022(rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x4Au))));
    if (rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x6Au))) != 9)
        w0670_put8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x6Au)), (uint8_t)dir);

    obj_off = w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x6Cu));
    obj_seg = w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x6Eu));
    w0670_put8(m, LIN(obj_seg, (uint16_t)(obj_off + 1u)), 0xFF);

    ax = (uint16_t)(w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x5Cu)) - w0670_rd16(m, GIN3_DS, 0x9BF4u));
    ax = (uint16_t)(ax + w0670_rd16(m, GIN3_DS, (uint16_t)(0x4A3Cu + 4u * dir)));
    obj_off = w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x6Cu));
    obj_seg = w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x6Eu));
    w0670_put16(m, obj_seg, (uint16_t)(obj_off + 6u), ax);

    ax = (uint16_t)(w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x5Eu)) - w0670_rd16(m, GIN3_DS, 0x9BF6u));
    ax = (uint16_t)(ax + w0670_rd16(m, GIN3_DS, (uint16_t)(0x4A3Cu + 4u * dir + 2u)));
    obj_off = w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x6Cu));
    obj_seg = w0670_rd16(m, fleet_seg, (uint16_t)(fleet_off + 0x6Eu));
    w0670_put16(m, obj_seg, (uint16_t)(obj_off + 8u), ax);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_25DB_0670(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("25DB_0670", 2, run)
