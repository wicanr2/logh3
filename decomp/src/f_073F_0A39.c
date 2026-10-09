/*
 * GIN3PS.EXE 073F:0A39  星系情報視窗：畫小圖外框，再為目前星域切片的每顆行星畫一條「陰影 + 彩色」引線
 *
 * 原版（196 bytes，far 呼叫，沒有參數；序言 push bp／si／di、sub sp,0Ch、mov bp,sp）。無有意義的回傳值（描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ 迴圈計數 i（byte）、[bp+2]／[bp+4] ＝ 行星記錄指標的位移／段、[bp+6] ＝ y 螢幕座標、
 *   [bp+8] ＝ 行星 X 縮放後的終點 x、[bp+0Ah] ＝ 行星 Y 縮放後的終點 y。
 * 過程（已證實：反組譯逐指令；3870:0723、3870:07C9 各由其差異測試證實）：
 *   3870:0723(23, 44, 152, 173, 0xFF07, 0xFFFF)                  ; 小圖外框。色 word 的高位元組 FF 來自先前的 mov ax,0FFFFh，07C9 只用低位元組
 *   p ← (word ds:CFF8) : (word ds:CFFA)                          ; 行星表基底（目前星域切片）
 *   for i ← 0; i < byte ds:CFFC（無號 byte 比較，jnb 結束）; i++, p += 0x52：
 *     y  ← (i + 3) × 20 + 10                                     ; 16 位元（imul 之後只用 AX）
 *     x2 ← (word p[+0x46] >> 4) + 0x18                           ; 邏輯右移，16 位元
 *     y2 ← (word p[+0x48] >> 4) + 0x2D
 *     3870:07C9(0x9F, y − 1, x2 − 1, y2 − 1, 色 word ＝ (y 的高位元組 << 8) | 3, 0xFFFF)   ; 影子（mov al,3 只改 AX 低位元組）
 *     3870:07C9(0xA0, y,     x2,     y2,     word ds:(6092 + byte p[+0x11]), 0xFFFF)       ; 彩色引線，色表以擁有者 byte 直接當位移（沒有 ×2）
 *   色 word 的高位元組：影子線是 y 的高位元組，彩色線是 ds:6092 表相鄰兩 byte 的高 byte；07C9 只傳低位元組給畫線，所以不可觀察，本版仍照原樣組出。
 * 語意（強推論）：docs/internals/data/region-info-panel.json 第 180 行「073F:0A39：從 (160, 110＋20i) 拉到 (24＋X>>4, 85＋Y>>4)，顏色 ds:6092[擁有者]，
 *   先畫 (−1,−1) 的色 3 影子」、docs/re/320-region-info-window.md 第 47 行。本版與之一致；座標是滑鼠座標（螢幕 y 再加 40 之類的轉換不在此函式）。
 *   [+0x11] 擁有者、[+0x46]／[+0x48] 行星 X／Y 的欄位名來自 docs/internals，未再驗證（假說到強推論）。
 * 呼叫的函式：3870:0723（外框）、3870:07C9（畫線前置）。07C9 內呼叫畫線 39D7:0647（IDA 沒切成函式，台帳 callees 看不到），
 *   差異測試用描述檔 stubs 取代，**畫線本身的像素行為不在驗證範圍**。
 * 副作用：只經由 0723／07C9（寫本函式堆疊上的端點副本與替身）；本函式不寫全域記憶體。
 *
 * 機器模型：C 版沒有堆疊（區域變數用 C 變數）。f_3870_07C9 把端點放在原版的 SS:1F7E..1F85（見其檔頭），落在探針的堆疊帶內，寫入集合比對時雙方都排除。
 * 輸入域：基底 fcmd-base.state（行星表 4 筆）；行星指標、筆數 byte、色表與裁切矩形都可被突變改動。
 *   行星記錄內的 word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（rd16s，dosgolem 的 8086 語意）。
 *   08FF 在原版不終止的案例由探針丟棄；讀到視訊記憶體的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern void f_3870_0723(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);
extern void f_3870_07C9(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);

/* seg:FFFF 的 word 高位元組在 seg:0000（8086 位移各自回繞）。 */
static uint16_t w0A39_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_073F_0A39(Mem *m)
{
    uint8_t i;
    uint16_t p_off, p_seg;

    f_3870_0723(m, 0x17, 0x2C, 0x98, 0xAD, 0xFF07u, 0xFFFFu);
    p_seg = rd16(m, DS(0xCFFAu));
    p_off = rd16(m, DS(0xCFF8u));
    for (i = 0; i < rd8(m, DS(0xCFFCu)); i++) {
        uint16_t y = (uint16_t)((uint16_t)(i + 3u) * 20u + 0x0Au);
        uint16_t x2 = (uint16_t)((w0A39_rd16s(m, p_seg, (uint16_t)(p_off + 0x46u)) >> 4) + 0x18u);
        uint16_t y2 = (uint16_t)((w0A39_rd16s(m, p_seg, (uint16_t)(p_off + 0x48u)) >> 4) + 0x2Du);
        uint8_t owner;

        f_3870_07C9(m, 0x9F, (uint16_t)(y - 1u), (uint16_t)(x2 - 1u), (uint16_t)(y2 - 1u),
                    (uint16_t)((y & 0xFF00u) | 3u), 0xFFFFu);
        owner = rd8(m, LIN(p_seg, (uint16_t)(p_off + 0x11u)));
        f_3870_07C9(m, 0xA0, y, x2, y2, rd16(m, DS((uint16_t)(0x6092u + owner))), 0xFFFFu);
        p_off = (uint16_t)(p_off + 0x52u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_073F_0A39(m);
    return 0;
}
GIN3_REGISTER("073F_0A39", 0, run)
