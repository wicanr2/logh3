/*
 * GIN3PS.EXE 321F:0C37  畫一個旋轉過的小圖樣（四條線段）：表 ds:6072 的四組線段兩端點先按 byte 角度旋轉，再做斜投影 (x − y ÷ 2, y ÷ 2)、平移，用 3870:07C9 畫
 *
 * 原版（218 bytes，far 呼叫，兩個堆疊字；序言 push bp／si、sub sp,0Ch、mov bp,sp，所以第一個參數 a 在 [bp+14h]、第二個參數 flag 在 [bp+16h]）。
 *   區域變數（全在堆疊上，旋轉時以 SS 指標傳給 3A92:0001）：[bp+0] ＝ x0、[bp+2] ＝ y0、[bp+4] ＝ x1、[bp+6] ＝ y1、[bp+8]／[bp+0Ah] ＝ 表指標（位移／段，段 ＝ DS）。
 *   沒有有意義的回傳（描述檔 ret ＝ none）。
 * 過程（已證實：反組譯逐指令；3A92:0001、3870:07C9 各由其差異測試證實）：
 *   表指標 ← ds:6072；對 si ＝ 0..3（無號 jb）：
 *     (x0, y0, x1, y1) ← 表指標處四個 word；表指標 ＋＝ 8                                       ; 基底值：(16,0,−16,−16)(16,0,−16,16)(−16,−16,0,0)(−16,16,0,0)
 *     角 ← byte a ＋ byte ds:D7EA（8 位元相加）                                                 ; 推入的 word 高位元組是 AH 殘值，3A92:0001 只取低位元組
 *     3A92:0001(far &x0, far &y0, 角)；3A92:0001(far &x1, far &y1, 角)                          ; 就地旋轉兩個端點
 *     y0 ← y0 ÷ 2（有號 idiv，向零取整）；x0 ← x0 − y0（16 位元）；y1 ← y1 ÷ 2；x1 ← x1 − y1       ; 斜投影
 *     byte flag ＝ 0（只看低位元組）→ x0 ＋＝ 0x30、y0 ＋＝ 0xC8、x1 ＋＝ 0x30、y1 ＋＝ 0xC8
 *                   否則        → x0 ＋＝ 0x250、y0 ＋＝ 0x170、x1 ＋＝ 0x250、y1 ＋＝ 0x170       ; 16 位元
 *     3870:07C9(x0, y0, x1, y1, 色 word ＝ 0xFF0D, 最後 word ＝ 0xFFFF)
 *       色 word 的 AH 是 `mov ax,0FFFFh; push ax; mov al,0Dh; push ax` 留下的 0xFF，07C9 只用低位元組（遮成 0x0D 再交給畫線），所以 0xFF0D 與 0x000D 對畫線無差別。
 * 語意（強推論）：沒有 docs 記載（grep docs 無 321F:0C37）。唯一呼叫者 321F:06FC、321F:0897。表值 ±16 像是一個小箭頭或方向標記；
 *   斜投影 (x − y ÷ 2, y ÷ 2) 與兩組平移 (0x30,0xC8)／(0x250,0x170) 的用意未知，不使用名稱。
 * 呼叫的函式：3870:07C9 內呼叫畫線 39D7:0647（IDA 沒切成函式，台帳 callees 看不到），差異測試用描述檔 stubs 取代，
 *   **畫線本身的像素行為不在驗證範圍**。
 * 副作用：只經由 3A92:0001（寫本函式堆疊上的端點）與 07C9（寫端點副本與替身）；本函式不寫全域記憶體。
 *
 * 機器模型：C 版沒有堆疊。四個端點 word 放在原版的實際位置（基底 fcmd-base.state：SS 固定 0x63EB，進入點 SP ＝ 1F8A；
 *   探針推 2 個參數字 4 bytes、返回位址 4、push bp／si 4、sub sp,0Ch，所以 bp ＝ 1F72，[bp+0] 在 SS:1F72、[bp+2] 在 1F74、[bp+4] 在 1F76、[bp+6] 在 1F78），
 *   交給 f_3A92_0001（它用 far 指標讀寫）就地旋轉，再從那裡讀回。這四個位址在探針的堆疊帶（SS:1B8A..1F89）內，寫入集合比對時雙方都排除，
 *   也不與 f_3870_07C9 寫死的端點槽（SS:1F7E..1F85）重疊。表指標是本函式的區域變數，C 版用區域變數。
 * 輸入域：a、flag 全部 16 位元值（a 與 flag 的高位元組不參與）；表與 ds:D7EA 取自記憶體；07C9 內 08FF 在原版不終止的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為未知。
 */
#include "gin3_reg.h"

extern uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word);
extern void f_3870_07C9(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);

#define W0C37_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define W0C37_X0 0x1F72u /* 進入點 SP 1F8A − 4 − 4 − 4 − 0Ch ＝ bp(1F72)，[bp+0] */
#define W0C37_Y0 0x1F74u /* [bp+2] */
#define W0C37_X1 0x1F76u /* [bp+4] */
#define W0C37_Y1 0x1F78u /* [bp+6] */

static uint16_t w0C37_ld(const Mem *m, uint16_t off) { return rd16(m, LIN(W0C37_SS, off)); }
static void w0C37_st(Mem *m, uint16_t off, uint16_t v) { wr16(m, LIN(W0C37_SS, off), v); }

void f_321F_0C37(Mem *m, uint16_t a_word, uint16_t flag_word)
{
    uint16_t tbl = 0x6072u;
    uint16_t si;

    for (si = 0; si < 4u; si++, tbl = (uint16_t)(tbl + 8u)) {
        uint16_t tx0 = rd16(m, DS(tbl));
        uint16_t ty0 = rd16(m, DS((uint16_t)(tbl + 2u)));
        uint16_t tx1 = rd16(m, DS((uint16_t)(tbl + 4u)));
        uint16_t ty1 = rd16(m, DS((uint16_t)(tbl + 6u)));
        uint8_t ang = (uint8_t)(a_word + rd8(m, DS(0xD7EAu)));
        uint16_t r;
        int16_t half;

        w0C37_st(m, W0C37_X0, tx0);
        w0C37_st(m, W0C37_Y0, ty0);
        w0C37_st(m, W0C37_X1, tx1);
        w0C37_st(m, W0C37_Y1, ty1);
        r = f_3A92_0001(m, W0C37_X0, W0C37_SS, W0C37_Y0, W0C37_SS, (uint16_t)((ty1 & 0xFF00u) | ang));
        (void)f_3A92_0001(m, W0C37_X1, W0C37_SS, W0C37_Y1, W0C37_SS, (uint16_t)((r & 0xFF00u) | ang));
        half = (int16_t)((int16_t)w0C37_ld(m, W0C37_Y0) / 2);
        w0C37_st(m, W0C37_Y0, (uint16_t)half);
        w0C37_st(m, W0C37_X0, (uint16_t)(w0C37_ld(m, W0C37_X0) - (uint16_t)half));
        half = (int16_t)((int16_t)w0C37_ld(m, W0C37_Y1) / 2);
        w0C37_st(m, W0C37_Y1, (uint16_t)half);
        w0C37_st(m, W0C37_X1, (uint16_t)(w0C37_ld(m, W0C37_X1) - (uint16_t)half));
        if ((uint8_t)flag_word == 0) {
            w0C37_st(m, W0C37_X0, (uint16_t)(w0C37_ld(m, W0C37_X0) + 0x30u));
            w0C37_st(m, W0C37_Y0, (uint16_t)(w0C37_ld(m, W0C37_Y0) + 0xC8u));
            w0C37_st(m, W0C37_X1, (uint16_t)(w0C37_ld(m, W0C37_X1) + 0x30u));
            w0C37_st(m, W0C37_Y1, (uint16_t)(w0C37_ld(m, W0C37_Y1) + 0xC8u));
        } else {
            w0C37_st(m, W0C37_X0, (uint16_t)(w0C37_ld(m, W0C37_X0) + 0x250u));
            w0C37_st(m, W0C37_Y0, (uint16_t)(w0C37_ld(m, W0C37_Y0) + 0x170u));
            w0C37_st(m, W0C37_X1, (uint16_t)(w0C37_ld(m, W0C37_X1) + 0x250u));
            w0C37_st(m, W0C37_Y1, (uint16_t)(w0C37_ld(m, W0C37_Y1) + 0x170u));
        }
        f_3870_07C9(m, w0C37_ld(m, W0C37_X0), w0C37_ld(m, W0C37_Y0), w0C37_ld(m, W0C37_X1), w0C37_ld(m, W0C37_Y1), 0xFF0Du, 0xFFFFu);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_321F_0C37(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("321F_0C37", 2, run)
