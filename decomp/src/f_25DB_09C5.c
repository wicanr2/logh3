/*
 * GIN3PS.EXE 25DB:09C5  把戰場在場艦隊與其分隊的圖示物件旗標歸零
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none）。序言 push bp／si／di、sub sp,12h，
 *   沒有堆疊參數。用到的全域變數（DS 內位移）：
 *     ds:9894（word）  在場艦隊清單的筆數
 *     ds:9898（byte[]）在場艦隊清單，每筆是艦隊表索引（byte）
 *     ds:C074          艦隊表，每筆 122（0x7A）bytes
 *     ds:CFB4／CFB6    艦艇池的 far 指標（位移／段），每筆 78（0x4E）bytes
 *
 * 語意（位移為已證實：本檔與反組譯逐條對過；欄位名稱為強推論，見下）：
 *   對清單的 di = 0 .. ds:9894 − 1（每圈重讀 ds:9894，無號比較）：
 *     f ← ds:C074 + 清單[di] × 0x7A                    ; 位移 16 位元回繞，清單[di] 是無號 byte
 *     f[+0x6C] 那個 far 指標指到的物件：obj[+1] ← 0      ; 艦隊的圖示物件
 *     對 k = 0 .. f[+0x0F] − 1（每圈重讀 f[+0x0F]，byte 無號比較）：
 *       s ← (ds:CFB6):(ds:CFB4 + f[+0x10 + 4k] 的 word × 0x4E)   ; 乘積只取低 16 位元，每圈重讀 CFB4／CFB6
 *       s[+0x38] 那個 far 指標指到的物件：obj[+1] ← 0   ; 艦艇的圖示物件
 *   讀寫順序與原版一致，因為物件若蓋到艦隊記錄、清單或這些全域變數，後面的讀取會看到前面的寫入
 *   （艦隊物件寫完才讀艦隊筆數 f[+0x0F]；每一圈的 ds:9894、f[+0x0F]、f[+0x10+4k]、ds:CFB4／CFB6 都是當圈才讀）。
 *
 * 欄位語意（強推論）：清單 ds:9898／筆數 ds:9894 是戰場在場艦隊（docs/internals/38-battle-units.md 第 283 行、
 *   docs/spec/019 第 154 行；其他函式如 185C:09D1 只讀 ds:9894 的低 byte，本函式用 16 位元比較 `cmp ax,ds:9894h`）；
 *   艦隊[+0x6C] 是艦隊圖示的顯示物件指標（docs/internals/33-fleet-table.md 第 121 行）；艦艇[+0x38] 是艦艇圖示
 *   的顯示物件指標；艦隊[+0x0F] 是分隊數、[+0x10 + 4k] 是分隊的艦艇池索引（docs/internals/33 第 104 行）。
 *   obj[+1] 的名稱為假說（25DB:0714／0865／0915 把它設成 0xFF，本函式把它清成 0，兩者成對：看起來是「需要重畫」
 *   旗標；docs/internals/34-battle-screen.md 第 369 行說 1588:0432 呼叫 25DB:09C5，0437 是該次完整鏡頭重畫返回後
 *   邊界，與此一致，但未獨立證實）。呼叫者是 1588:02CE、1588:03A9、25DB:108E。
 *
 * 模型限制：
 *   - 位址 0xF0000 以上是 BIOS ROM，寫入無效（dosgolem 的 ROMBase ＝ 0xF0000）。基底狀態裡艦隊的 +0x6C 是
 *     FFFF:FFFF，寫入落在 ROM，put8 不動作，與執行器一致。
 *   - 字的存取跨越位移 0xFFFF 時，本版與探針都按線性位址連續處理，不涵蓋真機上回繞到段首的行為。
 *
 * 輸入域：基底是戰略層（fcmd-base.state），ds:9894 ＝ 0（清單是空的），沒有戰場顯示物件；
 *   清單、艦隊、艦艇池與物件指標全由突變與 seeds 供給。ds:9894 變得很大時（無號 word 最多 65535 圈）
 *   原版超過步數上限，探針丟棄該案，不在比較範圍內。
 */
#include "gin3_reg.h"

static void put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

/* 艦隊 f（DS 內位移）的每個分隊：艦艇的圖示物件旗標歸零 */
static void zero_ship_objs(Mem *m, uint16_t f)
{
    uint8_t k = 0;
    uint16_t slot = (uint16_t)(f + 0x10);

    while (k < rd8(m, DS(f + 0x0F))) {
        uint16_t index = rd16(m, DS(slot));
        uint16_t s_off = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(index * 0x4Eu));
        uint16_t s_seg = rd16(m, DS(0xCFB6));
        uint16_t p_off = rd16(m, LIN(s_seg, s_off + 0x38));
        uint16_t p_seg = rd16(m, LIN(s_seg, s_off + 0x3A));

        put8(m, LIN(p_seg, p_off + 1), 0);
        k++;
        slot = (uint16_t)(slot + 4);
    }
}

void f_25DB_09C5(Mem *m)
{
    uint16_t di = 0;
    uint16_t list_off = 0x9898;

    while (di < rd16(m, DS(0x9894))) {
        uint8_t fi = rd8(m, DS(list_off));
        uint16_t f = (uint16_t)(0xC074u + (uint16_t)(fi * 0x7Au));
        uint16_t o_off = rd16(m, DS(f + 0x6C));
        uint16_t o_seg = rd16(m, DS(f + 0x6E));

        put8(m, LIN(o_seg, o_off + 1), 0);
        zero_ship_objs(m, f);
        di++;
        list_off = (uint16_t)(list_off + 1);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_25DB_09C5(m);
    return 0;
}
GIN3_REGISTER("25DB_09C5", 0, run)
