/*
 * GIN3PS.EXE 1C61:0444  戰場艦艇朝目的地 (S[+28], S[+2A]) 前進一步（最多 step），落點被擋時依艦種與狀態處理，最後寫回位置並更新顯示物件
 *
 * 原版（517 bytes，far 呼叫，兩個參數：S ＝ 艦艇記錄 far 指標、step ＝ word；序言 push bp／si／di、sub sp,22h、mov bp,sp，
 *   所以 S 的位移在 [bp+2Ch]、段在 [bp+2Eh]，step 在 [bp+30h]，只用低位元組），無回傳值（描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ 方位（只存不讀）、[bp+2]／[bp+4] ＝ 落點 X／Y（word，以 far 指標 SS:bp+2、SS:bp+4 交給 2BE1:026F 與 2583:0451 改寫）、
 *   [bp+6] ＝ 2583:040C 的結果、[bp+8]／[bp+0Ah] ＝ 落點距離（只存不讀）、[bp+0Ch]／[bp+0Eh] ＝ step << 16。
 * 過程（已證實：反組譯逐指令；3310:0006、3CC9:00B9、2BE1:011D、2BE1:026F、2583:040C、2583:0451、25DB:0714 各由其差異測試證實）：
 *   lim ← low8(step) << 16
 *   d ← 3310:0006(S[+34], S[+36], S[+28], S[+2A])               ; 16.16 距離（位置 到 目的地）
 *   3CC9:00B9(d, lim) 之後 jg（ZF ＝ 0 且 SF ＝ OF，有號 32 位元 d > lim）：
 *     不成立：X ← S[+28]、Y ← S[+2A]                              ; 一步到
 *     成立  ：ang ← 2BE1:011D(S[+28] − S[+34], S[+2A] − S[+36])
 *             2BE1:026F(S[+34], S[+36], ang, low8(step) 零延伸成 word, &X, &Y)   ; 朝目的地走 step，夾進戰場範圍，結果寫進 X、Y
 *   3310:0006(S[+34], S[+36], X, Y)                              ; 只有 ds:8EDC..8EED 的暫存副作用，結果不用（原版把 0 改成 1，也不用）
 *   r ← 2583:040C(X, Y, word S[0])                               ; 0 能站；0xFFFF 地形擋；其他 ＝ 池索引 ＋ 1 被佔
 *   r ≠ 0 時：
 *     byte S[+4] ＝ 7：
 *       r ＝ 0xFFFF：byte S[+1E] ← 3；word S[+24] ← 0；word S[+26] ← 8；X ← S[+34]、Y ← S[+36]
 *       否則 byte S[+1E] ＝ 3：2583:0451(S[+34], S[+36], &X, &Y, low8(step), word S[0])（避障搜尋，改寫 X、Y）
 *       否則（S[+1E] ≠ 3）：byte ds:(0xC074 ＋ low16(S[+1] × 0x7A) ＋ 0x3B) ← 4；word S[+24] ← 0；word S[+26] ← 8；X ← S[+34]、Y ← S[+36]
 *     byte S[+4] ≠ 7：2583:0451(S[+34], S[+36], &X, &Y, low8(step), word S[0])
 *   S[+34] ← X、S[+36] ← Y；word S[+24] 減 0（原樣）、word S[+26] 減 8（`sub …,0；sbb …,8`，進位恆 0）
 *   25DB:0714(S)                                                 ; 艦艇世界座標減鏡頭寫進顯示物件
 * 語意（假說）：docs 沒有 1C61:0444 的記載（grep 無命中；呼叫者 1C61:029C、1C61:03AA）。S[+4] ＝ 7 是 docs/re/372 與 f_2583_040C.c 所說的特殊艦種（索敵分遣隊，不擋人也不被擋）；
 *   S[+28]／[+2A] 是目的地、S[+34]／[+36] 是目前座標（f_25DB_0714.c）；S[+1E] 的值 3、S[+24]／[+26]（dword，高字被減 8）、艦隊 +0x3B ← 4 的語意未知。
 * 副作用：S 的 +0x1E、+0x24、+0x26、+0x34、+0x36；艦隊 +0x3B；0451 的寫入（只有 X、Y 槽）；25DB:0714 的物件寫入；39B5:005C 在 ds:8EDC..8EED 的暫存；026F 對 X、Y 槽的寫入（堆疊帶內，比對排除）。
 *
 * 機器模型：C 版沒有堆疊。X、Y 放在原版的實際位置（基底狀態 fcmd-standby.state：SS 固定 0x63EB，進入點 SP ＝ 1E76，探針推 3 個參數字 6 bytes、返回位址 4、
 *   push bp／si／di 6、sub sp,22h，所以 bp ＝ 1E44，X 在 SS:1E46、Y 在 SS:1E48），交給 2BE1:026F、2583:0451 用 far 指標改寫，再從那裡讀回。基底狀態一換，這兩個常數要重算。
 *   3310:0006 的 C 版把 39B5 的堆疊參數放在 fcmd-base.state 的框架（SS:1F82..1F89），在 fcmd-standby.state 下落在探針堆疊帶外；
 *   存下 SS:1F78..1F8F 呼叫前的內容並在呼叫後還原（理由同 f_1C03_0361.c、f_1588_0891.c）。
 * 輸入域：S 指向艦艇池記錄（描述檔 table ship）；戰場地圖與遮罩取自 fcmd-standby.state 與突變；池基底 ds:CFB4..CFB7 列入 freeze；S 不指向堆疊帶與 A0000 到 BFFFF。
 *   S 的顯示物件指標（+0x38）可指到 ROM 區（寫入被丟棄，由 25DB:0714 的 C 版模型化）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off,
                            uint16_t oy_seg);
extern uint16_t f_2583_040C(Mem *m, uint16_t x, uint16_t y, uint16_t faction);
extern uint16_t f_2583_0451(Mem *m, uint16_t x, uint16_t y, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg, uint16_t dist,
                            uint16_t faction);
extern void f_25DB_0714(Mem *m, uint16_t ship_off, uint16_t ship_seg);

#define F0444_SS 0x63EBu /* fcmd-standby.state 的 SS（diffprobe 無法設定） */
#define F0444_X 0x1E46u  /* 進入點 SP 1E76 − 6 − 4 − 6 − 0x22 ＝ bp(1E44)，X 在 bp+2 */
#define F0444_Y 0x1E48u  /* bp+4 */
#define F0444_ZF 0x0040u
#define F0444_SF 0x0080u
#define F0444_OF 0x0800u

static uint16_t w0444_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void w0444_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* 3310:0006 的 C 版在 fcmd-standby.state 下會寫探針堆疊帶外的 SS:1F82..1F89；存下 SS:1F78..1F8F 並在呼叫後還原 */
static uint32_t w0444_dist(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2)
{
    uint8_t save[24];
    uint32_t r;
    int k;

    for (k = 0; k < 24; k++)
        save[k] = rd8(m, LIN(F0444_SS, (uint16_t)(0x1F78u + k)));
    r = f_3310_0006(m, x1, y1, x2, y2);
    for (k = 0; k < 24; k++)
        wr8(m, LIN(F0444_SS, (uint16_t)(0x1F78u + k)), save[k]);
    return r;
}

void f_1C61_0444(Mem *m, uint16_t s_off, uint16_t s_seg, uint16_t step_word)
{
    uint16_t step = (uint16_t)(step_word & 0xFFu);
    uint32_t d;
    uint16_t fl, r;

#define S16(k) w0444_rd16s(m, s_seg, (uint16_t)(s_off + (k)))
#define S8(k) rd8(m, LIN(s_seg, (uint16_t)(s_off + (k))))
    d = w0444_dist(m, S16(0x34u), S16(0x36u), S16(0x28u), S16(0x2Au));
    fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, step, 0);
    if (!(fl & F0444_ZF) && (((fl & F0444_SF) != 0) == ((fl & F0444_OF) != 0))) { /* jg */
        uint16_t dy = (uint16_t)(S16(0x2Au) - S16(0x36u));
        uint16_t dx = (uint16_t)(S16(0x28u) - S16(0x34u));
        uint8_t ang = f_2BE1_011D(m, dx, dy);

        (void)f_2BE1_026F(m, S16(0x34u), S16(0x36u), ang, step, F0444_X, F0444_SS, F0444_Y, F0444_SS);
    } else {
        w0444_wr16s(m, F0444_SS, F0444_X, S16(0x28u));
        w0444_wr16s(m, F0444_SS, F0444_Y, S16(0x2Au));
    }
    (void)w0444_dist(m, S16(0x34u), S16(0x36u), w0444_rd16s(m, F0444_SS, F0444_X), w0444_rd16s(m, F0444_SS, F0444_Y));
    r = f_2583_040C(m, w0444_rd16s(m, F0444_SS, F0444_X), w0444_rd16s(m, F0444_SS, F0444_Y), S16(0x00u));
    if (r != 0) {
        int do451 = 1;

        if (S8(0x04u) == 7) {
            if (r == 0xFFFFu) {
                do451 = 0;
                wr8(m, LIN(s_seg, (uint16_t)(s_off + 0x1Eu)), 3);
                w0444_wr16s(m, s_seg, (uint16_t)(s_off + 0x24u), 0);
                w0444_wr16s(m, s_seg, (uint16_t)(s_off + 0x26u), 8);
                w0444_wr16s(m, F0444_SS, F0444_X, S16(0x34u));
                w0444_wr16s(m, F0444_SS, F0444_Y, S16(0x36u));
            } else if (S8(0x1Eu) != 3) {
                uint16_t f = (uint16_t)(0xC074u + (uint16_t)(S8(0x01u) * 0x7Au));

                do451 = 0;
                wr8(m, DS((uint16_t)(f + 0x3Bu)), 4);
                w0444_wr16s(m, s_seg, (uint16_t)(s_off + 0x24u), 0);
                w0444_wr16s(m, s_seg, (uint16_t)(s_off + 0x26u), 8);
                w0444_wr16s(m, F0444_SS, F0444_X, S16(0x34u));
                w0444_wr16s(m, F0444_SS, F0444_Y, S16(0x36u));
            }
        }
        if (do451)
            (void)f_2583_0451(m, S16(0x34u), S16(0x36u), F0444_X, F0444_SS, F0444_Y, F0444_SS, step, S16(0x00u));
    }
    w0444_wr16s(m, s_seg, (uint16_t)(s_off + 0x34u), w0444_rd16s(m, F0444_SS, F0444_X));
    w0444_wr16s(m, s_seg, (uint16_t)(s_off + 0x36u), w0444_rd16s(m, F0444_SS, F0444_Y));
    w0444_wr16s(m, s_seg, (uint16_t)(s_off + 0x26u), (uint16_t)(S16(0x26u) - 8u)); /* sub [+24],0；sbb [+26],8 */
    f_25DB_0714(m, s_off, s_seg);
#undef S16
#undef S8
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1C61_0444(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("1C61_0444", 3, run)
