/*
 * GIN3PS.EXE 321F:0B64  兩個圖示之間畫連線：以兩筆記錄 A、B 各自的圖示位置 (X, Y) 畫三條水平相鄰的線（y−1 色 6、y 色 4、y+1 色 6）
 *
 * 原版（211 bytes，far 呼叫，兩個 far 指標參數；序言 push bp／si／di、sub sp,0Eh、mov bp,sp，所以 A 的位移在 [bp+18h]、段在 [bp+1Ah]，B 的位移在 [bp+1Ch]、段在 [bp+1Eh]），
 *   無有意義的回傳值（描述檔 ret ＝ none）。區域變數：X1 ＝ [bp+2]、Y1 ＝ [bp+4]、X2 ＝ [bp+6]、Y2 ＝ [bp+8]。
 * 過程（已證實：反組譯逐指令；3870:07C9 由其差異測試證實）：
 *   A[+0x42]（word）＝ 0 或 B[+0x42]（word）＝ 0 → 返回，什麼都不做
 *   X1 ← A[+0x3E] ＋ 0x12；Y1 ← A[+0x40] − 2 − 4 × (((A[+0x08] ＋ 0x1F3) ÷ 500) − 1)      ; 全部 16 位元運算，÷ 是無號 div bx（bx ＝ 0x1F4），商最多 131，不會 #DE
 *   X2、Y2 對 B 同樣算
 *   3870:07C9(X1, Y1 − 1, X2, Y2 − 1, 0x0106, 0xFFFF)     ; 色 6（07C9 只用低位元組）
 *   3870:07C9(X1, Y1,     X2, Y2,     色 4,    0xFFFF)
 *   3870:07C9(X1, Y1 + 1, X2, Y2 + 1, 0xFF06, 0xFFFF)     ; 色 6
 *   07C9 的參數 6 個堆疊字依序是 x0、y0、x1、y1、色 word、最後 word；07C9 只取色 word 的低位元組（mov al,[bp+arg_8]；xor ah,ah）。
 *   第一條的色 word 是 bx ＝ 0x1F4 的 bl 改成 6，即 0x0106；第二條的色 word 是 `mov cl,4; push cx`，cx 的高位元組是 07C9 返回後的殘值，
 *   第三條是 `mov al,6; push ax`，ax 剛設成 0xFFFF，所以是 0xFF06。高位元組在 07C9 裡被遮掉，不影響行為；C 版第二條用 0x0004。
 * 語意（強推論）：呼叫者 321F:06FC（台帳）；docs 沒有 321F:0B64 本身的記載。A[+0x3E]／A[+0x40]／A[+0x08]／A[+0x42] 的名稱未知；
 *   `÷ 500 − 1` 乘 4 看起來是依 A[+0x08]（艘數？）每 500 單位把圖示上緣抬高 4 像素，名稱是假說，不是證據。
 * 副作用：全部來自 3870:07C9 → 3870:08FF（端點裁切，寫在堆疊上）與畫線替身 39D7:0647；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊。被呼叫的 f_3870_07C9 把端點放在 fcmd-base.state 進入點 SP ＝ 1F8A 的框架（f_3870_07C9.c 檔頭），原版從本函式呼叫時框架更深，
 *   位置不同，但兩處都在探針的堆疊帶內，比對時排除，C 版自寫自讀，不影響結果。畫線 39D7:0647 由描述檔的 stubs 取代，像素行為不在驗證範圍。
 * 輸入域：A、B 指到暫存區或任意位址的記錄（只讀 +0x08、+0x3E、+0x40、+0x42）；裁切矩形取記憶體 ds:8D6A..8D70；word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0
 *   （dosgolem 的 read16），用本檔的 w0B64_rd16s。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_3870_07C9(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);

static uint16_t w0B64_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* Y ＝ rec[+0x40] − 2 − 4 × ((rec[+0x08] ＋ 0x1F3) ÷ 500 − 1)，全 16 位元 */
static uint16_t w0B64_y(const Mem *m, uint16_t seg, uint16_t off)
{
    uint16_t cx = (uint16_t)(w0B64_rd16s(m, seg, (uint16_t)(off + 0x40u)) + 0xFFFEu);
    uint16_t ax = (uint16_t)(w0B64_rd16s(m, seg, (uint16_t)(off + 8u)) + 0x1F3u);

    ax = (uint16_t)(ax / 0x1F4u);
    ax = (uint16_t)(ax - 1u);
    ax = (uint16_t)(ax << 2);
    return (uint16_t)(cx - ax);
}

void f_321F_0B64(Mem *m, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg)
{
    uint16_t x1, y1, x2, y2;

    if (w0B64_rd16s(m, a_seg, (uint16_t)(a_off + 0x42u)) == 0)
        return;
    if (w0B64_rd16s(m, b_seg, (uint16_t)(b_off + 0x42u)) == 0)
        return;
    x1 = (uint16_t)(w0B64_rd16s(m, a_seg, (uint16_t)(a_off + 0x3Eu)) + 0x12u);
    y1 = w0B64_y(m, a_seg, a_off);
    x2 = (uint16_t)(w0B64_rd16s(m, b_seg, (uint16_t)(b_off + 0x3Eu)) + 0x12u);
    y2 = w0B64_y(m, b_seg, b_off);
    f_3870_07C9(m, x1, (uint16_t)(y1 - 1u), x2, (uint16_t)(y2 - 1u), 0x0106u, 0xFFFFu);
    f_3870_07C9(m, x1, y1, x2, y2, 0x0004u, 0xFFFFu);
    f_3870_07C9(m, x1, (uint16_t)(y1 + 1u), x2, (uint16_t)(y2 + 1u), 0xFF06u, 0xFFFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_321F_0B64(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("321F_0B64", 4, run)
