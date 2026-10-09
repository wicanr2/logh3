/*
 * GIN3PS.EXE 04C8:02D3  數有幾個星域的擁有者等於指定陣營（far，一個參數，回 AX＝個數 0..32）
 *
 * 原版：push bp、sub sp,6，所以唯一的參數在 [bp+0Ch]（word，只取低位元組：mov cl,[bp+0Ch]）。
 * 過程：對星域表 ds:A79C 起 32 筆（每筆 126 bytes）逐筆，星域[+0x12] 與參數的低位元組相等（byte 比較，jnz）就計數加一。
 *   計數器是 byte（inc byte ptr [bp+0]），最多 32，不會回繞。
 * 回傳：AL ＝ 計數，AH ＝ 0（函式一開始 xor ax,ax，迴圈只動 AL）。即 AX ＝ 個數。
 * 副作用：沒有（stores＝0，只讀）。
 * 欄位：星域[+0x12] ＝ 擁有者（0 帝國、1 同盟、2 無主），docs/internals/22、docs/re/295 第 62 到 66 行
 *   （「擁有者讀星域記錄的 +0x12」）。值 2 以外的 byte 一樣照原樣比較，不檢查範圍。
 * 呼叫者：04C8:00F3（台帳 callers＝1）；呼叫端怎麼使用個數沒有讀（未知）。
 *
 * 輸入域：參數高位元組任意（原版忽略）；星域表內容由突變決定，+0x12 任意 byte。全部可到達。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；「擁有者」的稱呼來自 docs（強推論，欄位位移已證實）。
 */
#include "gin3_reg.h"

uint16_t f_04C8_02D3(Mem *m, uint16_t owner_w)
{
    uint8_t want = (uint8_t)owner_w;
    uint8_t count = 0;
    uint16_t region = 0xA79C;
    uint8_t i;

    for (i = 0; i < 0x20; i++) {
        if (rd8(m, DS((uint16_t)(region + 0x12))) == want)
            count++;
        region = (uint16_t)(region + 0x7E);
    }
    return count; /* AH ＝ 0 */
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_04C8_02D3(m, a[0]); }
GIN3_REGISTER("04C8_02D3", 1, run)
