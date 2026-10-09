/*
 * GIN3PS.EXE 321F:0AF1  畫一條垂直的 8 像素短線（刻度）：位置由物件 P 的三個欄位決定，色 word 固定 0xFF04，經 3870:07C9 畫
 *
 * 原版（115 bytes，far 呼叫，一個 far 指標參數 P；序言 push bp／si／di、sub sp,8、mov bp,sp，所以 P 的位移在 [bp+12h]、段在 [bp+14h]）。
 *   沒有有意義的回傳（描述檔 ret ＝ none）。
 * 過程（已證實：反組譯逐指令；3870:07C9 由其差異測試證實）：
 *   word P[+42h] ＝ 0 → 返回（cmp word ptr es:[bx+42h],0 ／ jnz，整個 word）
 *   x ← word P[+3Eh] ＋ 0x12                                   ; 16 位元
 *   y ← word P[+40h]
 *   q ← (word P[+8] ＋ 0x1F3 的低 16 位元) ÷ 0x1F4             ; `add ax,1F3h` 在 16 位元，DX 清 0 的無號除法 `div bx`
 *   s ← ((q − 1) 的低 16 位元) << 2 的低 16 位元                ; q ＝ 0 時 dec 成 0xFFFF，再左移 2 位成 0xFFFC
 *   y0 ← y ＋ 0xFFFE − s（16 位元）＝ y − 2 − s；y1 ← y ＋ 6 − s
 *   3870:07C9(x, y0, x, y1, 色 word ＝ 0xFF04, 最後 word ＝ 0xFFFF)        ; 垂直線
 *     色 word 的 AH 是 `mov ax,0FFFFh; push ax; mov al,4; push ax` 留下的 0xFF，07C9 只用低位元組（遮成 4 再交給畫線），
 *     所以 0xFF04 與 0x0004 對畫線無差別（已證實：07C9 的 C 版與差異測試）。
 * 語意（強推論）：沒有 docs 記載（grep docs 無 321F:0AF1）。唯一呼叫者 321F:06FC（hw-dependent）連呼叫兩次。
 *   P[+8] 每 500 為一格，使刻度上移 4 像素，推測為某個量的指示標；欄位語意未知，不使用名稱。
 * 呼叫的函式：3870:07C9 內呼叫畫線 39D7:0647（IDA 沒切成函式，台帳 callees 看不到），差異測試用描述檔 stubs 取代，
 *   **畫線本身的像素行為不在驗證範圍**。
 * 副作用：只經由 07C9（寫本函式堆疊上的端點副本與替身）；本函式不寫全域記憶體。
 *
 * 機器模型：C 版沒有堆疊（本函式的區域變數 [bp+0..7] 只存不傳指標，C 版省略）。f_3870_07C9 把端點放在原版的 SS:1F7E..1F85（見其檔頭），
 *   落在探針的堆疊帶內，寫入集合比對時雙方都排除。
 * 輸入域：P 全部 16 位元位址（描述檔放在 ds:F400）；P 欄位的 word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（rd16s）；
 *   07C9 內 08FF 在原版不終止的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_3870_07C9(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);

/* word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 read16） */
static uint16_t w0AF1_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_321F_0AF1(Mem *m, uint16_t p_off, uint16_t p_seg)
{
    uint16_t x, y, q, s, y0, y1;

    if (w0AF1_rd16s(m, p_seg, (uint16_t)(p_off + 0x42u)) == 0)
        return;
    x = (uint16_t)(w0AF1_rd16s(m, p_seg, (uint16_t)(p_off + 0x3Eu)) + 0x12u);
    y = w0AF1_rd16s(m, p_seg, (uint16_t)(p_off + 0x40u));
    q = (uint16_t)((uint16_t)(w0AF1_rd16s(m, p_seg, (uint16_t)(p_off + 8u)) + 0x1F3u) / 0x1F4u);
    s = (uint16_t)((uint16_t)(q - 1u) << 2);
    y0 = (uint16_t)(y + 0xFFFEu - s);
    y1 = (uint16_t)(y + 6u - s);
    f_3870_07C9(m, x, y0, x, y1, 0xFF04u, 0xFFFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_321F_0AF1(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("321F_0AF1", 2, run)
