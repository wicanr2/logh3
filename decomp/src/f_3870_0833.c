/*
 * GIN3PS.EXE 3870:0833  畫一段線（異或線）：兩端點按 x 排好（有號）、經 3870:08FF 裁切進矩形 ds:8D6A..8D70，可見才呼叫畫線 39D7:0936
 *
 * 原版（106 bytes，far 呼叫，六個堆疊字；序言 push bp／si、mov bp,sp，沒有 sub sp，所以參數從 [bp+8] 起）：
 *   arg_0 ＝ x0（[bp+8]）  arg_2 ＝ y0（[bp+0Ah]）  arg_4 ＝ x1（[bp+0Ch]）  arg_6 ＝ y1（[bp+0Eh]）
 *   arg_8 ＝ 色（[bp+10h]，只用低位元組：mov al,[bp+arg_8]; xor ah,ah）  arg_A ＝ 最後一個 word（[bp+12h]，原樣傳給畫線，docs/internals/52 稱遮罩）
 *   無有意義的回傳值（描述檔 ret ＝ none；AX 是 08FF 的 0 或色的零延伸，且畫線替身不碰 AX）。
 * 過程（已證實：反組譯逐指令；3870:08FF 由其差異測試證實）：
 *   x0 > x1（有號 jle 反向）→ x0 與 x1 對調、y0 與 y1 對調（兩者都在堆疊上對調）
 *   3870:08FF(far &x0, far &y0, far &x1, far &y1)       ; 指標都指到本函式自己的堆疊參數（SS:bp+8、+0Ah、+0Ch、+0Eh），就地裁切
 *   回傳 0 → 結束；否則 39D7:0936(x0, y0, x1, y1, 色低位元組, arg_A)（六個堆疊字，由低位址到高位址）
 * 呼叫的函式：39D7:0936 是畫線（寫 VGA 與 I/O 埠，自己維護虛線圖樣 ds:8F0E 並用自我修改程式碼，見 f_1727_03CF.c 檔頭），IDA 把目標標成
 *   loc_47216 而不是函式（呼叫位元組 9A 36 09 D7 39），台帳 callees 只有 08FF。差異測試用描述檔的 stubs 取代它：探針不執行它，
 *   只記錄參數與呼叫順序，C 版以 gin3_stub 做同樣的呼叫。**畫線本身的像素行為不在驗證範圍。**
 * 語意（強推論）：docs/re/155-manoeuvre-boundary-exact.md 第 66 行「呼叫端 3870:0833 先把兩端按 x 排好」、docs/re/135 第 110 行
 *   「折線每一段走 3870:0833(x0,y0,x1,y1,色,遮罩)：先把兩點按 x 排好，…」、docs/re/314 第 165 行「XOR 線（3870:0833 → 39D7:0936）」、
 *   docs/internals/52 第 66 行，與本檔一致。與 3870:07C9 逐位元組相同（位元組序列只差最後的 call 目標），畫線目標是 39D7:0936 而不是 39D7:0647。
 * 副作用：只經由 08FF（0A30 寫端點的 x、y，即本函式堆疊上的參數）與替身；本函式不寫全域記憶體。
 *
 * 機器模型：C 版沒有堆疊。四個端點 word 放在原版的實際位置（基底狀態 fcmd-base.state：SS 固定 0x63EB，進入點 SP ＝ 1F8A，
 *   探針推 6 個參數 12 bytes、返回位址 4、push bp／si 4，所以 bp ＝ 1F76，x0 在 SS:1F7E、y0 在 1F80、x1 在 1F82、y1 在 1F84），
 *   交給 f_3870_08FF（它用 far 指標讀寫）就地裁切，再從那裡讀回。這個位址在探針的堆疊帶（SS:1B8A..1F89）內，寫入集合比對時雙方都排除，
 *   與原版把端點放在堆疊一致。
 * 輸入域：座標、色、最後一個 word 全部 16 位元值；裁切矩形取記憶體 ds:8D6A..8D70；08FF 在原版不終止（整數溢位推不進矩形）的案例
 *   由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3870_08FF(Mem *m, uint16_t xa_off, uint16_t xa_seg, uint16_t ya_off, uint16_t ya_seg, uint16_t xb_off,
                            uint16_t xb_seg, uint16_t yb_off, uint16_t yb_seg);

#define W0833_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define W0833_X0 0x1F7Eu /* 進入點 SP 1F8A − 12 − 4 − 4 ＝ bp(1F76)，x0 在 bp+8 */
#define W0833_Y0 0x1F80u
#define W0833_X1 0x1F82u
#define W0833_Y1 0x1F84u

static void w0833_wr(Mem *m, uint16_t off, uint16_t v) { wr16(m, LIN(W0833_SS, off), v); }
static uint16_t w0833_rd(const Mem *m, uint16_t off) { return rd16(m, LIN(W0833_SS, off)); }

void f_3870_0833(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word)
{
    uint16_t a[6];

    if ((int16_t)x0 > (int16_t)x1) {
        uint16_t t = x0;

        x0 = x1;
        x1 = t;
        t = y0;
        y0 = y1;
        y1 = t;
    }
    w0833_wr(m, W0833_X0, x0);
    w0833_wr(m, W0833_Y0, y0);
    w0833_wr(m, W0833_X1, x1);
    w0833_wr(m, W0833_Y1, y1);
    if (f_3870_08FF(m, W0833_X0, W0833_SS, W0833_Y0, W0833_SS, W0833_X1, W0833_SS, W0833_Y1, W0833_SS) != 0) {
        a[0] = w0833_rd(m, W0833_X0);
        a[1] = w0833_rd(m, W0833_Y0);
        a[2] = w0833_rd(m, W0833_X1);
        a[3] = w0833_rd(m, W0833_Y1);
        a[4] = (uint16_t)(color_word & 0xFFu);
        a[5] = last_word;
        (void)gin3_stub(m, "39D7_0936", 6, a);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3870_0833(m, a[0], a[1], a[2], a[3], a[4], a[5]);
    return 0;
}
GIN3_REGISTER("3870_0833", 6, run)
