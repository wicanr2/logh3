/*
 * GIN3PS.EXE 165F:0A71  要塞砲的橢圓中心：焦距 c ＝ 3870:0614(長半軸, 短半軸)，把 (c, 0) 依方位旋轉後加到 *x、*y
 *
 * 原版（75 bytes，far 呼叫；序言 push bp／si／di、sub sp,4、mov bp,sp，所以第一個參數在 [bp+0Eh]）：
 *   args[0] ＝ a（長半軸，[bp+0Eh]）、args[1] ＝ b（短半軸，[bp+10h]）、args[2] ＝ 方位（word，[bp+12h]，3A92:0001 只取低位元組）、
 *   args[3..4] ＝ x 的 far 指標（[bp+14h]、[bp+16h]）、args[5..6] ＝ y 的 far 指標（[bp+18h]、[bp+1Ah]）。
 *   回傳 AX ＝ 最後載入的 [bp+2]（旋轉後的 y 分量；描述檔 ret ＝ ax）。
 * 過程（已證實：反組譯逐指令；3870:0614、3A92:0001 由其各自的差異測試證實）：
 *   [bp+0] ← 3870:0614(a, b)（push b、push a：a 在低位址）；[bp+2] ← 0
 *   3A92:0001(far SS:bp+0, far SS:bp+2, 方位)         ; 旋轉堆疊上的 (c, 0)
 *   *x ← *x ＋ [bp+0]；*y ← *y ＋ [bp+2]               ; 先 x 後 y，各自是 word 的 read-modify-write（add es:[bx], ax）
 * 語意（強推論）：docs/re/145-fortress-gun-aim-and-fire.md 第 83 到 89 行（橢圓中心 ＝ 焦點 ＋ 沿瞄準方位走 c 個像素），與本檔一致。
 * 副作用：*x、*y 各一個 word；3870:0614 內 39B5:005C 的暫存（ds:8EDC..8EED）；堆疊上的 [bp+0]、[bp+2]（堆疊帶，比對排除）。
 *   word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 行為）；線性 0xF0000 以上（ROM）的寫入被安靜丟棄，本檔按 byte 模型化。
 *
 * 機器模型：C 版沒有堆疊。[bp+0]、[bp+2] 放在原版的實際位置（基底狀態 fcmd-base.state：SS 固定 0x63EB，進入點 SP ＝ 1F8A，探針推 7 個參數字 14 bytes、
 *   返回位址 4、push bp／si／di 6、sub sp,4，所以 bp ＝ 1F6E），交給 f_3A92_0001 用 far 指標讀寫，再讀回。
 *   被呼叫的 39B5 C 版把暫存參數放在 SS:1F82..1F89，同在堆疊帶內，比對時排除。
 * 輸入域：全部 16 位元參數與記憶體內容；x、y 指標指到堆疊或本函式程式碼的案例不在比較範圍。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3870_0614(Mem *m, uint16_t a, uint16_t b);
extern uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word);

#define F0A71_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F0A71_BP 0x1F6Eu /* 進入點 SP 1F8A − 14 − 4 − 6 − 4 */

static uint16_t w0A71_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void w0A71_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w0A71_add16s(Mem *m, uint16_t seg, uint16_t off, uint16_t add)
{
    uint16_t v = (uint16_t)(w0A71_rd16s(m, seg, off) + add);

    w0A71_put8(m, LIN(seg, off), (uint8_t)v);
    w0A71_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_165F_0A71(Mem *m, uint16_t a, uint16_t b, uint16_t angle_word, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg)
{
    uint16_t dx, dy;

    wr16(m, LIN(F0A71_SS, F0A71_BP), f_3870_0614(m, a, b));
    wr16(m, LIN(F0A71_SS, (uint16_t)(F0A71_BP + 2u)), 0);
    (void)f_3A92_0001(m, F0A71_BP, F0A71_SS, (uint16_t)(F0A71_BP + 2u), F0A71_SS, angle_word);
    dx = rd16(m, LIN(F0A71_SS, F0A71_BP));
    w0A71_add16s(m, x_seg, x_off, dx);
    dy = rd16(m, LIN(F0A71_SS, (uint16_t)(F0A71_BP + 2u)));
    w0A71_add16s(m, y_seg, y_off, dy);
    return dy;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_165F_0A71(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6]); }
GIN3_REGISTER("165F_0A71", 7, run)
