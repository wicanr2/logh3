/*
 * GIN3PS.EXE 3870:042F  以 (xc, yc) 為中心、沿物件 obj 的朝向畫兩條交叉的線：一條沿朝向（從後方 −obj[+0A] 到前方 obj[+08]），一條垂直（−obj[+0C] 到 +obj[+0C]）
 *
 * 原版（301 bytes，far 呼叫，四個 word 堆疊字；序言 push bp／si／di、sub sp,12h、mov bp,sp，所以 xc 在 [bp+1Ch]、yc 在 [bp+1Eh]、obj 的位移在 [bp+20h]、段在 [bp+22h]），
 *   無有意義的回傳值（描述檔 ret ＝ none）。區域變數：四個 word 的暫存端點 L0..L3 ＝ [bp+0]、[bp+2]、[bp+4]、[bp+6]（兩個點的 x、y），
 *   [bp+8]／[bp+0Ah] ＝ obj ＋ 4 的位移／段、[bp+0Eh]／[bp+10h] ＝ obj ＋ 0Eh、[bp+0Ch] ＝ 段。
 * 過程（已證實：反組譯逐指令；3A92:0001、3870:0833 各由其差異測試證實）：
 *   (L0, L1) ← (−obj[+0A], 0)（word；neg 為 16 位元）；(L2, L3) ← (obj[+08], 0)
 *   3A92:0001(far &L0, far &L1, obj[+04])        ; 以 byte 朝向旋轉 (L0, L1)，就地
 *   3A92:0001(far &L2, far &L3, obj[+04])        ; 旋轉 (L2, L3)；朝向每次重讀
 *   3870:0833(xc + L0, yc + L1, xc + L2, yc + L3, word obj[+05], word obj[+0E])        ; 16 位元加；色 word ＝ obj[+5..+6]，最後 word ＝ obj[+0E..+0F]
 *   (L0, L1) ← (0, −obj[+0C])；(L2, L3) ← (0, obj[+0C])
 *   3A92:0001 兩次同上；3870:0833 同上
 *   3870:0833 是畫線前置（端點按 x 排序、裁切進矩形 ds:8D6A..8D70、再呼叫畫線 39D7:0936，由描述檔 stubs 取代）。
 * 語意（強推論）：docs/re/135-manoeuvre-boundary.md 第 225 行（「3870:042F：兩條交叉的線，[+8]／[+0A]／[+0C] 長度、[+4] 朝向」）、docs/re/122-arrival-facing-and-fleet-step.md 第 52 行
 *   （兩條交叉的線）、docs/solved-history.md 第 943 行（沿朝向從後方 32 拉到前方 64 一條、垂直 ±32 一條），與本檔一致；obj[+05] 當畫線色 word、obj[+0E] 當最後 word 是假說
 *   （07C9／0833 的最後 word，docs/internals/52 稱遮罩）。
 * 副作用：全部來自 3870:0833 → 3870:08FF（端點裁切，寫在堆疊上）與畫線替身 39D7:0936；本函式自己不寫全域記憶體（L0..L3 在堆疊上）。
 *
 * 機器模型：C 版沒有堆疊。四個暫存端點放在原版的實際位置（基底狀態 fcmd-base.state：SS 固定 0x63EB，進入點 SP ＝ 1F8A，探針推 4 個參數字 8 bytes、返回位址 4、
 *   push bp／si／di 6、sub sp,12h，所以 bp ＝ 1F66，L0..L3 在 SS:1F66、1F68、1F6A、1F6C），交給 f_3A92_0001（用 far 指標讀寫）就地旋轉，再從那裡讀回。
 *   這些位址在探針的堆疊帶（SS:1B8A..1F89）內，寫入集合比對時雙方都排除，與原版把端點放在堆疊一致。f_3870_0833 的端點另放在它自己的位置（SS:1F7E 起），互不重疊。
 * 輸入域：xc、yc 與 obj 欄位全部 16 位元值；obj 取暫存區（不指向堆疊帶與 A0000 到 BFFFF）；裁切矩形取記憶體 ds:8D6A..8D70；
 *   word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（dosgolem 的 read16），用本檔的 w042F_rd16s。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_3A92_0001(Mem *m, uint16_t x_off, uint16_t x_seg, uint16_t y_off, uint16_t y_seg, uint16_t angle_word);
extern void f_3870_0833(Mem *m, uint16_t x0, uint16_t y0, uint16_t x1, uint16_t y1, uint16_t color_word, uint16_t last_word);

#define F042F_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F042F_BP 0x1F66u /* 進入點 SP 1F8A − 參數 8 − 返回位址 4 − push bp／si／di 6 − sub sp,12h */

static uint16_t w042F_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w042F_put(Mem *m, uint16_t k, uint16_t v) { wr16(m, LIN(F042F_SS, (uint16_t)(F042F_BP + 2u * k)), v); }
static uint16_t w042F_get(const Mem *m, uint16_t k) { return rd16(m, LIN(F042F_SS, (uint16_t)(F042F_BP + 2u * k))); }

/* 旋轉 (L[a], L[b])：朝向每次重讀 obj[+04] */
static void w042F_rot(Mem *m, uint16_t obj_off, uint16_t obj_seg, uint16_t a, uint16_t b)
{
    uint16_t ang = rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 4u)));

    (void)f_3A92_0001(m, (uint16_t)(F042F_BP + 2u * a), F042F_SS, (uint16_t)(F042F_BP + 2u * b), F042F_SS, ang);
}

static void w042F_draw(Mem *m, uint16_t xc, uint16_t yc, uint16_t obj_off, uint16_t obj_seg)
{
    uint16_t last = w042F_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Eu));
    uint16_t color = w042F_rd16s(m, obj_seg, (uint16_t)(obj_off + 5u));

    f_3870_0833(m, (uint16_t)(xc + w042F_get(m, 0)), (uint16_t)(yc + w042F_get(m, 1)), (uint16_t)(xc + w042F_get(m, 2)),
                (uint16_t)(yc + w042F_get(m, 3)), color, last);
}

void f_3870_042F(Mem *m, uint16_t xc, uint16_t yc, uint16_t obj_off, uint16_t obj_seg)
{
    w042F_put(m, 0, (uint16_t)(0u - w042F_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Au))));
    w042F_put(m, 1, 0);
    w042F_put(m, 2, w042F_rd16s(m, obj_seg, (uint16_t)(obj_off + 8u)));
    w042F_put(m, 3, 0);
    w042F_rot(m, obj_off, obj_seg, 0, 1);
    w042F_rot(m, obj_off, obj_seg, 2, 3);
    w042F_draw(m, xc, yc, obj_off, obj_seg);

    w042F_put(m, 0, 0);
    w042F_put(m, 1, (uint16_t)(0u - w042F_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Cu))));
    w042F_put(m, 2, 0);
    w042F_put(m, 3, w042F_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Cu)));
    w042F_rot(m, obj_off, obj_seg, 0, 1);
    w042F_rot(m, obj_off, obj_seg, 2, 3);
    w042F_draw(m, xc, yc, obj_off, obj_seg);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3870_042F(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("3870_042F", 4, run)
