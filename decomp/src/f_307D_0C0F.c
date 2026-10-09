/*
 * GIN3PS.EXE 307D:0C0F  八個選項的權重依「各自門檻值」縮放：門檻越小越放大、越大越縮小
 *
 * 原版：far 呼叫，沒有參數。序言 push bp／si、sub sp,4、mov bp,sp；區域變數 [bp+0..3] 是 far 指標 ds:D6BC（每輪 +6）。
 *   AX 在結束時是殘值（最後一輪載入的高位元組與 8），沒有回傳值；描述檔 ret ＝ none。呼叫者 307D:0521（台帳 callers＝1）。
 *
 * 資料（強推論：docs/re/287 §八個選項「ds:605A／ds:6062／307D:0B41／ds:D6BC」；本檔只驗行為，語意名稱未知）：
 *   ds:606A 起 8 個 byte ＝ 位置對照表（i -> t，基底狀態 04 03 05 02 06 01 07 00）；
 *   ds:D69C 起 dword 陣列（每筆 4 bytes，第 t 筆位於 D69C ＋ 4t），本函式只讀每筆的低位 word（位移 D69C ＋ 4t），
 *   原版另載入高位 word（D69E ＋ 4t）到 AX，之後沒有使用；
 *   ds:D6BC 起 8 筆 × 6 bytes，本函式只改每筆開頭的 word。三塊在 DGROUP 緊鄰（D69C..D6BB、D6BC..D6EB），
 *   後面 D6EC 起 32 bytes 是 307D:0005 填 1 的區域。
 *
 * 語意（已證實，本檔與反組譯逐條對過）：
 *   for (i ＝ 0; i < 8; i++) {
 *       t ＝ byte ds:606A[i]；th ＝ word ds:(D69C ＋ 4t)               ; 位移 16 位元回繞；t 最大 255 -> 位移 D69C..DA98
 *       w ＝ word ds:(D6BC ＋ 6i)
 *       th ＝ 0       : w ←  w << 3            ; 16 位元截斷
 *       th <  250     : w ←  w << 1
 *       th <  500     : 不變
 *       th <  1000    : w ←  w >> 1            ; 邏輯右移
 *       th <  2000    : w ←  w >> 2
 *       th <  4000    : w ←  w >> 3
 *       其餘           : w ←  w >> 4
 *   }
 * 原版的細節（本版照做）：
 *   - 所有比較都是無號（jnb／jb）。門檻剛好等於 250／500／1000／2000／4000 時落在較大的那一檔。
 *   - 每一輪都從記憶體重讀 t、th 與 w。t >= 8 時 th 讀到 D6BC 之後的區域（本函式正在改的那塊），後面的輪次看得到前面的寫入；
 *     原版也不檢查 t 的範圍。
 *   - 沒有 ROM 或視訊寫入問題：所有位址都在 DGROUP（ds:D6BC 起 48 bytes）。
 *
 * 輸入域：全部可到達（無參數）；基底狀態 ds:D69C..D6EB 全 0（每輪 th ＝ 0 -> 左移 3 位），因此描述檔 seeds 另造各檔邊界。
 */
#include "gin3_reg.h"

void f_307D_0C0F(Mem *m)
{
    uint8_t i;
    uint16_t q = 0xD6BCu; /* 目前這筆的位移（far 指標 ds:D6BC，每輪 +6） */

    for (i = 0; i < 8; i++) {
        uint16_t t = rd8(m, DS((uint16_t)(0x606Au + i)));
        uint16_t bx = (uint16_t)(t << 2);
        uint16_t th = rd16(m, DS((uint16_t)(bx + 0xD69Cu)));
        uint16_t w = rd16(m, DS(q));

        if (th == 0)
            w = (uint16_t)(w << 3);
        else if (th < 0xFAu)
            w = (uint16_t)(w << 1);
        else if (th < 0x1F4u)
            w = w;
        else if (th < 0x3E8u)
            w = (uint16_t)(w >> 1);
        else if (th < 0x7D0u)
            w = (uint16_t)(w >> 2);
        else if (th < 0xFA0u)
            w = (uint16_t)(w >> 3);
        else
            w = (uint16_t)(w >> 4);
        wr16(m, DS(q), w);
        q = (uint16_t)(q + 6u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_307D_0C0F(m);
    return 0;
}
GIN3_REGISTER("307D_0C0F", 0, run)
