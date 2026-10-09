/*
 * 3BB6:0A4B  回傳 seg130 資料段中一個固定位址的 far 指標（AX＝段、BX＝位移）（far，無堆疊參數）
 *
 * 原版：mov ax,seg130；mov bx,13F2h；retf。AX ＝ 段（重定位值）、BX ＝ 位移 0x13F2；不讀不寫記憶體，
 *   與進入時的暫存器與記憶體都無關。回傳打包成 (AX << 16) | BX（描述檔 ret ＝ pack:ax,bx）。
 *   AX:BX 的順序與配置器 3982:0001、3ABF:0000 一樣是「AX＝段、BX＝位移」，不是 DX:AX。
 *
 * 段值：seg130 的重定位值，dosgolem 載入位置（映像在 0349）下是 0x4AFE（函式位元組 B8 FE 4A）。載入位置相依；
 *   描述檔的 code_sha256_16 用重定位後的台帳 sha256_final_16。
 *
 * 語意（假說）：seg130:13F2 是 3BB6:0003 讀的區塊起點（`mov si,13F2h`；基底狀態裡 13F2 起有 "M1E_BOLEOR"
 *   字串，其後是歌曲標頭）。呼叫者 3B74:0137、3B74:0155 取得此指標的用途：未知。本函式回傳固定常數。
 * 推論等級：回傳常數已證實；用途假說。
 * 輸入域：無輸入。描述檔加 ax、bx 兩個 reg 參數（隨機值），驗證回傳不依賴進入時的暫存器。
 */
#include "gin3_reg.h"

#define F3BB6_0A4B_SEG130 0x4AFEu

uint64_t f_3BB6_0A4B(Mem *m, uint16_t ax_in, uint16_t bx_in)
{
    (void)m;
    (void)ax_in;
    (void)bx_in;
    return ((uint64_t)F3BB6_0A4B_SEG130 << 16) | 0x13F2u;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3BB6_0A4B(m, a[0], a[1]); }
GIN3_REGISTER("3BB6_0A4B", 2, run)
