/*
 * 2376:1585  讀 DS:9208 的 byte 回傳（far，無參數）
 *
 * 原版：push ds；mov ax,seg；mov ds,ax；mov al,[9208]；xor ah,ah；pop ds；retf。
 * 輸出：AX = DS:9208 的 byte（零延伸）；DS 回來時不變。
 * 設定 DS 的 seg 立即值是重定位項（relocs=1），載入後等於 DGROUP（GIN3_DS）：
 *   向量檔裡被突變的線性位址是 DGROUP 段基底加 9208，由差異測試實際確認，不是由 IDA 名稱推出來。
 *
 * 語意（已證實）：docs/internals/data/fleet-sound-source.json 與 docs/re/415-fleet-sound.md 記為
 *   「DS:9208 busy byte」，唯一呼叫者 3B74:031A 忙等這個值到 0 才送下一個音效。
 *   該 byte 由誰寫入、值的意義（是否為 DSP 忙碌旗標）docs 標「未深入」，本檔不另行斷言：未知。
 * 輸入域：DS:9208 的任意 byte 值。
 */
#include "gin3_reg.h"

uint16_t f_2376_1585(Mem *m)
{
    return rd8(m, DS(0x9208));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2376_1585(m);
}
GIN3_REGISTER("2376_1585", 0, run)
