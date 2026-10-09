/*
 * 3BB6:0232  讀模組資料段 seg130 的 ds:0BEF，零擴展後回 AX（far，無參數）
 *
 * 原版：push ds；mov ax,seg130；mov ds,ax；mov al,[0BEF]；xor ah,ah；pop ds；retf。
 *   AX ＝ byte ds:0BEF（高位元組清 0）。不寫記憶體（只動堆疊，堆疊不比對）。
 *   唯一的 ret 是 AX；DS 在返回前還原，所以呼叫者的 DS 與本函式無關。
 *
 * 段值：seg130 是 IDA 的段名，函式裡的 `mov ax,seg130` 是重定位項；dosgolem 載入位置（映像在 0349）下的值是
 *   0x4AFE（函式位元組 B8 FE 4A）。與 gin3_mem.h 的 GIN3_DS 相同，是載入位置相依的常數，不是 DGROUP（0x54B3）。
 *   台帳 sha256_16（未重定位）與 sha256_final_16（重定位後）不同，描述檔用後者。
 *
 * 語意（假說）：seg130 是 3BB6 模組的資料段（音樂驅動：ds:13F2 起有 "M1E_BOLEOR" 歌曲標頭、聲道記錄在 ds:6 起、
 *   每筆 0x14 bytes；來自基底狀態傾印，docs/ 沒有 3BB6 的記載）。ds:0BEF 是單一位元組：3BB6:0212 寫 0、
 *   3BB6:01FB 寫 0xFF（兩者的反組譯）。本函式是它的 getter；名稱與用途：未知。
 * 推論等級：讀的位址與零擴展已證實（逐指令對過）；用途假說。
 * 輸入域：ds:0BEF 的全部 256 個值。
 */
#include "gin3_reg.h"

#define F3BB6_0232_SEG130 0x4AFEu

uint16_t f_3BB6_0232(Mem *m)
{
    return (uint16_t)rd8(m, LIN(F3BB6_0232_SEG130, 0x0BEF));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_3BB6_0232(m);
}
GIN3_REGISTER("3BB6_0232", 0, run)
