/*
 * GIN3PS.EXE 321F:0223  兩個記錄（各自 +0x10 的艦艇池索引）所指艦艇的座標抄進 ds:D7EE..D7F4，算出方位的負值與四個畫面座標 ds:D7F6..D7FC
 *
 * 原版（367 bytes，far 呼叫，兩個 far 指標參數 A、B；序言 push bp／si／di、sub sp,10h、mov bp,sp，所以 A 的位移在 [bp+1Ah]、段在 [bp+1Ch]，B 在 [bp+1Eh]、[bp+20h]），
 *   無回傳值（描述檔 ret ＝ none）。區域變數：[bp+0]／[bp+2] ＝ 中點 mx／my、[bp+4]／[bp+6] ＝ 距離（低字／高字）、[bp+8]／[bp+0Ah] ＝ 艦艇 A 的指標、[bp+0Ch]／[bp+0Eh] ＝ 艦艇 B 的指標。
 * 過程（已證實：反組譯逐指令；2BE1:011D、3310:0006、3CC9:00B9、3CC9:0009、3CC9:0070 各由其差異測試證實）：
 *   shipA ← (word ds:CFB4 ＋ low16(word A[+0x10] × 0x4E)) : word ds:CFB6      ; 先讀 A[+0x10]，再讀池段、池基底
 *   shipB ← 同樣算法用 B[+0x10]
 *   ds:D7EE ← shipA[+0x34]；ds:D7F0 ← shipA[+0x36]；ds:D7F2 ← shipB[+0x34]；ds:D7F4 ← shipB[+0x36]    ; 依序「讀、寫」，所以艦艇記錄與這四格別名時後面的讀取看得到前面的寫入
 *   byte ds:D7EA ← 0 − low8(2BE1:011D(D7F2 − D7EE, D7F4 − D7F0))                ; 方位（B 相對 A）的負值，位元組
 *   mx ← (D7EE ＋ D7F2) >> 1、my ← (D7F0 ＋ D7F4) >> 1                           ; 16 位元相加（進位丟掉）再邏輯右移
 *   d ← 3310:0006(D7EE, D7F0, D7F2, D7F4)                                       ; AB 兩點的 16.16 距離
 *   3CC9:00B9(d, 0x00C8:0000) 之後 jle（ZF ＝ 1 或 SF ≠ OF，有號 32 位元 d <= 200.0）：
 *     成立：d1 ← 3310:0006(mx, my, D7EE, D7F0)；h1 ← 高字(d1)；q1 ← (h1 × 3) 有號 ÷ 2（3CC9:0009、3CC9:0070，h1 以零延伸）；ds:D7F6 ← 0x140 − low16(q1)；ds:D7F8 ← 0x12A
 *           d2 ← 3310:0006(mx, my, D7F2, D7F4)；h2 ← 高字(d2)；q2 ← (h2 × 3) ÷ 2；ds:D7FA ← low16(q2) ＋ 0x140；ds:D7FC ← 0x12A
 *     不成立：ds:D7F6 ← 0xAA、ds:D7F8 ← 0x12A、ds:D7FA ← 0x1D6、ds:D7FC ← 0x12A
 *   （成立分支在第一次呼叫 3310:0006 之前多推一個 word 0x140，呼叫後 `add sp,8` 清 4 個參數，留下的 0x140 在第二次運算後由 `pop ax` 取回；這是堆疊上的暫存，沒有記憶體影響。）

 * 邊界觀察：d ＝ 200.0 時近距離分支的輸出（D7F6 ＝ 0x140 − 150 ＝ 0xAA、D7FA ＝ 150 ＋ 0x140 ＝ 0x1D6）與遠距離分支的常數相同，所以 jle 與 jl 在恰好 200.0 不可區分（等價）；
 *   d ＝ 200.11（高字 200、有小數）時近距離分支算出 D7F6 ＝ 172，與常數 170 不同，可區分「只比高字」的錯誤版本。
 * 語意（假說）：docs 沒有 321F:0223 的記載（grep 無命中；呼叫者 321F:000A）。0x140、0x12A、0xAA、0x1D6 看起來是畫面座標（320 寬畫面的中心 0x140 加減
 *   半徑 × 3 ÷ 2；0x12A、0xAA/0x1D6 為兩個固定位置），但缺證據，整體用途未知。D7EA 為位元組。A、B 的 +0x10 是艦艇池索引（艦隊記錄的第一個分隊槽也是；強推論）。
 * 副作用：ds:D7EA（1 byte）、ds:D7EE..D7FC（8 個 word 與其間）；3310:0006 內的 39B5:005C 在 ds:8EDC..8EED 的暫存。
 *
 * 機器模型：C 版沒有堆疊。word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 的 read16／write16），用本檔的 w0223_rd16s／w0223_wr16s。
 * 輸入域：A、B 指向艦隊表（描述檔 table fleet）或 seeds 給的位址；艦艇池取基底與 seeds，不指向 A0000 到 BFFFF。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0070(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F0223_ZF 0x0040u
#define F0223_SF 0x0080u
#define F0223_OF 0x0800u

static uint16_t w0223_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0223_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

static uint16_t w0223_dsw(const Mem *m, uint16_t off)
{
    return w0223_rd16s(m, GIN3_DS, off);
}

/* (h × 3) ÷ 2：3CC9:0009（高字 0 的乘數）接 3CC9:0070（有號，除數 0:2），回商的低字 */
static uint16_t w0223_scale(uint16_t h)
{
    uint32_t prod = f_3CC9_0009(0, h, 0, 3);
    uint64_t q = f_3CC9_0070((uint16_t)(prod >> 16), (uint16_t)prod, 0, 2);

    return (uint16_t)(q >> 32);
}

void f_321F_0223(Mem *m, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg)
{
    uint16_t a_idx = w0223_rd16s(m, a_seg, (uint16_t)(a_off + 0x10u));
    uint16_t a_pseg = w0223_dsw(m, 0xCFB6u);
    uint16_t a_rec = (uint16_t)(w0223_dsw(m, 0xCFB4u) + (uint16_t)(a_idx * 0x4Eu));
    uint16_t b_idx = w0223_rd16s(m, b_seg, (uint16_t)(b_off + 0x10u));
    uint16_t b_pseg = w0223_dsw(m, 0xCFB6u);
    uint16_t b_rec = (uint16_t)(w0223_dsw(m, 0xCFB4u) + (uint16_t)(b_idx * 0x4Eu));
    uint16_t angle, mx, my, h1, h2, d_hi, fl;
    uint32_t d;

    w0223_wr16s(m, GIN3_DS, 0xD7EEu, w0223_rd16s(m, a_pseg, (uint16_t)(a_rec + 0x34u)));
    w0223_wr16s(m, GIN3_DS, 0xD7F0u, w0223_rd16s(m, a_pseg, (uint16_t)(a_rec + 0x36u)));
    w0223_wr16s(m, GIN3_DS, 0xD7F2u, w0223_rd16s(m, b_pseg, (uint16_t)(b_rec + 0x34u)));
    w0223_wr16s(m, GIN3_DS, 0xD7F4u, w0223_rd16s(m, b_pseg, (uint16_t)(b_rec + 0x36u)));
    angle = f_2BE1_011D(m, (uint16_t)(w0223_dsw(m, 0xD7F2u) - w0223_dsw(m, 0xD7EEu)),
                        (uint16_t)(w0223_dsw(m, 0xD7F4u) - w0223_dsw(m, 0xD7F0u)));
    wr8(m, DS(0xD7EAu), (uint8_t)(0u - (uint8_t)angle));
    mx = (uint16_t)((uint16_t)(w0223_dsw(m, 0xD7EEu) + w0223_dsw(m, 0xD7F2u)) >> 1);
    my = (uint16_t)((uint16_t)(w0223_dsw(m, 0xD7F0u) + w0223_dsw(m, 0xD7F4u)) >> 1);
    d = f_3310_0006(m, w0223_dsw(m, 0xD7EEu), w0223_dsw(m, 0xD7F0u), w0223_dsw(m, 0xD7F2u), w0223_dsw(m, 0xD7F4u));
    d_hi = (uint16_t)(d >> 16);
    fl = f_3CC9_00B9(d_hi, (uint16_t)d, 0x00C8u, 0);
    if ((fl & F0223_ZF) || (((fl & F0223_SF) != 0) != ((fl & F0223_OF) != 0))) { /* jle */
        h1 = (uint16_t)(f_3310_0006(m, mx, my, w0223_dsw(m, 0xD7EEu), w0223_dsw(m, 0xD7F0u)) >> 16);
        w0223_wr16s(m, GIN3_DS, 0xD7F6u, (uint16_t)(0x0140u - w0223_scale(h1)));
        w0223_wr16s(m, GIN3_DS, 0xD7F8u, 0x012Au);
        h2 = (uint16_t)(f_3310_0006(m, mx, my, w0223_dsw(m, 0xD7F2u), w0223_dsw(m, 0xD7F4u)) >> 16);
        w0223_wr16s(m, GIN3_DS, 0xD7FAu, (uint16_t)(w0223_scale(h2) + 0x0140u));
        w0223_wr16s(m, GIN3_DS, 0xD7FCu, 0x012Au);
    } else {
        w0223_wr16s(m, GIN3_DS, 0xD7F6u, 0x00AAu);
        w0223_wr16s(m, GIN3_DS, 0xD7F8u, 0x012Au);
        w0223_wr16s(m, GIN3_DS, 0xD7FAu, 0x01D6u);
        w0223_wr16s(m, GIN3_DS, 0xD7FCu, 0x012Au);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_321F_0223(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("321F_0223", 4, run)
