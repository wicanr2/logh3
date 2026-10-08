/*
 * 2376:0EF5  兩位元組字碼換算成「字型檔內位元組偏移」（near，暫存器傳參）
 *
 * 輸入：AH、AL、DX（DX 整個 16 位元都加進去）。
 * 計算（16 位元回繞）：b = AH * 157 + DX + AL
 * 輸出：DX:AX = 30 * b，b 先截成 16 位元再乘（最大 30 * 65535 = 1,966,050，DX 高位元組一定是 0）。
 *       原版用 shl 5 後連減兩次 b 並以 sbb 傳遞借位，等於 32b - b - b 的 21 位元結果。
 *       BX 以 push／pop 保存；CX 被當暫存（回來時 = AL，CH = 0），不納入比對；旗標也不納入。
 *
 * 語意（強推論）：157 是 Big5 第二位元組的字數（0x40..0x7E 共 63 個，加 0xA1..0xFE 共 94 個）；
 *   30 是一個雙位元組字的字型位元組數（唯一呼叫者 2376:0F2F 以 INT 21h AH=3Fh 讀 0x1E = 30 bytes，
 *   是 2376:0EE3 半形字 15 bytes 的兩倍）。0F2F 先把字碼換成連續編號（AH、AL 的意義由 0F2F 決定，
 *   其中有 sub al,0A1h、sub ah,40h 等 Big5 範圍的調整），呼叫本函式後以 add ax,基底／adc dx,0
 *   疊加，再用 INT 21h AH=42h 設檔案位置。證據：反組譯本身；docs 沒有逐位元組記載。
 *   對應的字型檔名稱與各基底值（6996h、0F00h、3ED0h）的來源：未知。
 * 輸入域：AX、DX 的全部 16 位元值；本函式不讀寫記憶體。
 */
#include "gin3_reg.h"

uint32_t f_2376_0EF5(uint16_t ax, uint16_t dx)
{
    uint16_t hi = (uint16_t)(ax >> 8);
    uint16_t lo = (uint16_t)(ax & 0xFFu);
    uint16_t b = (uint16_t)(hi * 157u + dx + lo);

    return 30u * (uint32_t)b; /* DX:AX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_2376_0EF5(a[0], a[1]);
}
GIN3_REGISTER("2376_0EF5", 2, run)
