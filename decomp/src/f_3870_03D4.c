/*
 * GIN3PS.EXE 3870:03D4  畫一條從 (x, y) 出發的射線：長度 obj[+8]、角度 obj[+4]（byte），終點 = (x, y) + 旋轉後的 (長度, 0)，經 3870:0833 畫（異或線）
 *
 * 原版（91 bytes，far 呼叫，四個堆疊字；序言 push bp／si／di、sub sp,4、mov bp,sp，所以 x 在 [bp+0Eh]、y 在 [bp+10h]、
 *   物件 far 指標的位移在 [bp+12h]、段在 [bp+14h]）。沒有有意義的回傳（描述檔 ret ＝ none）。
 * 過程（已證實：反組譯逐指令；3A92:0001、3870:0833 各由其差異測試證實）：
 *   di ← x
 *   [bp+0] ← word obj[+8]（長度）；[bp+2] ← 0                                  ; 兩個 word 是區域變數，下面當 (x, y) 讓 3A92:0001 就地旋轉
 *   3A92:0001(far &[bp+0], far &[bp+2], 角度 ← byte obj[+4] 零延伸成 word)       ; 旋轉後 [bp+0] ＝ x'、[bp+2] ＝ y'
 *   3870:0833(x0 ＝ x、y0 ＝ y、x1 ＝ x ＋ [bp+0]、y1 ＝ y ＋ [bp+2]（16 位元相加）、色 ← word obj[+5]、最後 word ← word obj[+0Ah])
 *     ; 物件在旋轉之後才重新載入（les bx,[bp+12h]），色與最後 word 都是 word 讀取：+5 與 +6 組成色 word，+0Ah 與 +0Bh 組成最後 word。
 *     ; 0833 只用色 word 的低位元組（+5），+6 的值不影響畫線，但仍讀；最後 word 原樣傳給畫線替身。
 * 語意（強推論）：docs/re/135-manoeuvre-boundary.md 第 224 行「3 ｜ 3870:03D4 ｜ 從中心射出的一條線 ｜ [+8] 長度、[+4] 角度、[+5] 色、[+0A] 遮罩」，與本檔一致。
 * 呼叫的函式：3870:0833 內呼叫畫線 39D7:0936（IDA 沒切成函式，台帳 callees 看不到），差異測試用描述檔 stubs 取代，
 *   **畫線本身的像素行為不在驗證範圍**。
 * 副作用：只經由 3A92:0001（寫本函式堆疊上的兩個 word）與 0833（寫本函式堆疊上的端點副本）；本函式不寫全域記憶體。
 *
 * 機器模型：C 版沒有堆疊。兩個區域變數放在原版的實際位置（基底 fcmd-base.state：SS 固定 0x63EB，進入點 SP ＝ 1F8A；
 *   探針推 4 個參數字 8 bytes、返回位址 4、push bp／si／di 6、sub sp,4，所以 bp ＝ 1F74，[bp+0] 在 SS:1F74、[bp+2] 在 SS:1F76），
 *   交給 f_3A92_0001（它用 far 指標讀寫）就地旋轉，再從那裡讀回。這兩個位址在探針的堆疊帶（SS:1B8A..1F89）內，寫入集合比對時雙方都排除；
 *   也不與 f_3870_0833 寫死的端點槽（SS:1F7E..1F85）重疊。
 * 輸入域：x、y 全部 16 位元值；物件欄位取自記憶體（描述檔放在 ds:F400）；物件不指向堆疊帶與 A0000 到 BFFFF；
 *   0833 內 08FF 在原版不終止（整數溢位推不進矩形）的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word);
extern void f_3870_0833(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);

#define W03D4_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define W03D4_L0 0x1F74u /* 進入點 SP 1F8A − 8 − 4 − 6 − 4 ＝ bp(1F74)，[bp+0] */
#define W03D4_L2 0x1F76u /* [bp+2] */

/* word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 read16） */
static uint16_t w03D4_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_3870_03D4(Mem *m, uint16_t x, uint16_t y, uint16_t obj_off, uint16_t obj_seg)
{
    uint16_t angle = rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 4u)));
    uint16_t dx, dy, color_word, last_word;

    wr16(m, LIN(W03D4_SS, W03D4_L0), w03D4_rd16s(m, obj_seg, (uint16_t)(obj_off + 8u)));
    wr16(m, LIN(W03D4_SS, W03D4_L2), 0);
    (void)f_3A92_0001(m, W03D4_L0, W03D4_SS, W03D4_L2, W03D4_SS, angle);
    dx = rd16(m, LIN(W03D4_SS, W03D4_L0));
    dy = rd16(m, LIN(W03D4_SS, W03D4_L2));
    last_word = w03D4_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Au));
    color_word = w03D4_rd16s(m, obj_seg, (uint16_t)(obj_off + 5u));
    f_3870_0833(m, x, y, (uint16_t)(x + dx), (uint16_t)(y + dy), color_word, last_word);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3870_03D4(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("3870_03D4", 4, run)
