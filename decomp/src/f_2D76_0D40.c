/*
 * GIN3PS.EXE 2D76:0D40  艦隊裡第一個「隊形槽不在樣板內」的編制組（回組序號，全部都在樣板內回 0xFF）
 *
 * 原版：far 呼叫，一個參數＝艦隊記錄的 far 指標（位移在前、段在後；push bp／si、sub sp,0Ah，所以參數在 [bp+12h]）。回 AL
 *   （AH 是殘值，描述檔 ret = pack:al；`mov al,0FFh` 與 `mov al,[bp+4]` 都只改 AL）。
 * 過程（已證實：反組譯逐指令；2D76:0DB2 由其自己的差異測試證實）：
 *   樣板 ＝ ds:(0x2BC6 + p[+9] × 0x14)（p[+9] 是 byte；樣板每列 0x14 bytes＝10 筆 × 2 bytes，筆的第 2 個 byte 是隊形槽）
 *   組指標 q ＝ p + 0x10（位移 16 位元回繞），i ＝ 0（byte）
 *   迴圈：n ＝ p[+0x0F]（byte，每圈重讀）；i >= n（無號 byte，jnb）→ 回 0xFF
 *     r ＝ 2D76:0DB2(樣板, 槽值 ＝ word(q + 3)，只有低位元組有用, n)    ; 槽值不在樣板前 n 筆內回 1，在內回 0
 *     r != 0 → 回 i
 *     否則 i++、q += 4
 *   0DB2 的第三、第四個參數推入的是 word（槽值取 q+3 的 word，n 取 CX，高位元組為殘值），0DB2 只取低位元組。
 * 語意（強推論，docs/re/229「2D76:0D40(艦隊) ＋ 2D76:0DB2(樣板, 隊形槽, n)」）：p[+9] ＝ 艦隊的隊形樣板編號、
 *   p[+0x0F] ＝ 現有編制組數、q[+3] ＝ 該組的隊形槽；本函式找出第一個隊形槽不在樣板內的組。名稱不是證據，位移與比較是證據。
 * 副作用：無寫入。輸入域：p 任意（取暫存區並由 seeds 造）；p[+9] 最大 255 時樣板位置 ds:(2BC6 + 5100) 仍在 DS 內。
 *   n 最大 255，最壞約 255 × 255 次 byte 比較，描述檔 step_budget 調高。word 存取在位移 0xFFFF 時高位元組取同一段位移 0（rd16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_2D76_0DB2(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t value_word, uint16_t n_word);

static uint16_t w0D40_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint8_t f_2D76_0D40(Mem *m, uint16_t p_off, uint16_t p_seg)
{
    uint16_t tmpl = (uint16_t)(0x2BC6u + (uint16_t)(rd8(m, LIN(p_seg, (uint16_t)(p_off + 9u))) * 0x14u));
    uint16_t q = (uint16_t)(p_off + 0x10u);
    uint8_t i = 0;

    for (;;) {
        uint8_t n = rd8(m, LIN(p_seg, (uint16_t)(p_off + 0x0Fu)));

        if (i >= n)
            return 0xFF;
        if (f_2D76_0DB2(m, tmpl, GIN3_DS, w0D40_rd16s(m, p_seg, (uint16_t)(q + 3u)), n) != 0)
            return i;
        i++;
        q = (uint16_t)(q + 4u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2D76_0D40(m, a[0], a[1]); }
GIN3_REGISTER("2D76_0D40", 2, run)
