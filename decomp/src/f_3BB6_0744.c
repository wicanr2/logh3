/*
 * 3BB6:0744  讀模組資料段 seg130 的 ds:0BF0，零擴展後回 AX（far，無參數）
 *
 * 原版：push ds；mov ax,seg130；mov ds,ax；mov al,[0BF0]；xor ah,ah；pop ds；retf。
 *   與 3BB6:0232 同形，只差讀的位址（0232 讀 0BEF，本函式讀 0BF0）。AX ＝ byte ds:0BF0（高位元組清 0）。
 *   不寫記憶體（只動堆疊，堆疊不比對）。DS 在返回前還原。
 *
 * 段值：seg130 的重定位值，dosgolem 載入位置（映像在 0349）下是 0x4AFE（函式位元組 B8 FE 4A），
 *   載入位置相依，不是 DGROUP（0x54B3）。描述檔的 code_sha256_16 用重定位後的台帳 sha256_final_16。
 *
 * 語意（假說）：seg130 是 3BB6 模組的資料段（音樂驅動，見 f_3BB6_0232.c 檔頭）。ds:0BF0 是單一位元組：
 *   3BB6:0703 寫 0xFF（反組譯 `mov byte_58740, 0FFh`）。本函式是它的 getter；名稱與用途：未知。
 *   ds:0BF1 是 3BB6:0003 讀的另一個位元組，與本函式無關。
 * 推論等級：讀的位址與零擴展已證實（逐指令對過）；用途假說。
 * 輸入域：ds:0BF0 的全部 256 個值。
 */
#include "gin3_reg.h"

#define F3BB6_0744_SEG130 0x4AFEu

uint16_t f_3BB6_0744(Mem *m)
{
    return (uint16_t)rd8(m, LIN(F3BB6_0744_SEG130, 0x0BF0));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_3BB6_0744(m);
}
GIN3_REGISTER("3BB6_0744", 0, run)
