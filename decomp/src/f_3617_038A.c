/*
 * 3617:038A  文字引擎的跳脫碼參數收字元：數字累加進 ds:F125，回傳字元是否結束參數
 *
 * 原版：far 呼叫，一個參數 c（word，只讀低位元組 cl），回 AX ＝ 0 或 1。
 *   堆疊：序言只有 push bp、mov bp,sp，參數在 [bp+6]。
 *   bx ＝ cl + 1；若 ds:[bx − 0x6C2A]（＝ ds:[0x93D7 + cl]，位移 16 位元回繞）的 bit7 設定：
 *       這個位元組是數字（表與 3CA8:0002 同一張，基底狀態只有 '0'..'9' 的表項有 bit7）：
 *       al ＝ 10 × ds:F125 的低位元組（mul byte ptr，只存 AL）；ds:F125 ← al；
 *       al ＝ cl + 0xD0（＝ cl − 0x30，8 位元）；ds:F125 ← ds:F125 + al（8 位元加）；回 0（繼續收參數）。
 *     否則若 cl == 0x5B（'['）：回 0（跳過，不累加）。
 *     否則回 1（結束符，呼叫端把這個字元吃掉、不印出）。
 *   數字判斷看表的 bit7，不看字元範圍；表項被改成有 bit7 的非數字字元，一樣走累加並用 cl − 0x30 當位值。
 * 副作用：只寫 ds:F125（byte）。F125 是 byte，累加以 8 位元回繞（docs/re/141：參數是十進位，逐位收）。
 * 呼叫者：seg085（3617 的跳脫碼處理）的六處呼叫；docs/re/141 第 63 行「由 3617:038A 逐位收：[ 跳過、
 *   數字累加、其他任何字元結束並提交」與本檔一致（已證實）。
 *
 * 輸入域：c 的全部 16 位元值（高位元組不讀）；表項與 ds:F125 的全部位元組值（由探針突變）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；ds:F125 為「目前收到的十進位參數」是文件推論。
 */
#include "gin3_reg.h"

uint16_t f_3617_038A(Mem *m, uint16_t c)
{
    uint8_t cl = (uint8_t)c;
    uint16_t bx = (uint16_t)(cl + 1u);
    uint16_t dx = 0;

    if (rd8(m, DS((uint16_t)(bx + 0x93D6u))) & 0x80u) { /* test byte [bx-6C2Ah],80h */
        uint8_t al = (uint8_t)(10u * rd8(m, DS(0xF125u))); /* mul byte ptr ds:F125，只存 AL */
        wr8(m, DS(0xF125u), al);
        al = (uint8_t)(cl + 0xD0u);
        wr8(m, DS(0xF125u), (uint8_t)(rd8(m, DS(0xF125u)) + al));
    } else if (cl != 0x5Bu) {
        dx = 1;
    }
    return dx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3617_038A(m, a[0]);
}
GIN3_REGISTER("3617_038A", 1, run)
