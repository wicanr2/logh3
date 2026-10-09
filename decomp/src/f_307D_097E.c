/*
 * GIN3PS.EXE 307D:097E  抽籤決定艦隊的目標位置：原地（50%）或 沿隨機方向偏移一小段（50%）
 *
 * 原版：far 呼叫，四個 far 指標參數（push bp／si／di、sub sp,0Ah，所以第一個參數在 [bp+14h]），無回傳值（AX 為殘值）：
 *   args[0..1] ＝ F（艦隊記錄）    args[2..3] ＝ P1（word 輸出，X）    args[4..5] ＝ P2（word 輸出，Y）    args[6..7] ＝ P3（byte 輸出）
 * 過程（已證實：反組譯逐指令；3CD5:000C（rand）、2BE1:026F 由其各自的差異測試證實）：
 *   r0 ← rand()；sel ← r0 & 1（`cwd; xor ax,dx; and ax,1; add dx,ax`，r0 ≥ 0 所以就是 r0 & 1）
 *   sel == 1：  *P1（word）← F[+0x5C]；*P2（word）← F[+0x5E]；*P3（byte）← F[+0x4A]        （每一步寫完才讀下一個欄位）
 *   sel == 0：
 *     rand()                                        ; 只為了推進亂數，其結果進 3CC9:0070（有號 32 位元除法，被除數是這次 rand、除數是 word F[+0x54]
 *                                                   ; 零延伸）；商與餘數存進區域變數後完全沒被讀取（死碼，已確認 [bp+4]／[bp+6]／[bp+8] 之外無讀取），
 *                                                   ; 3CC9:0070 沒有副作用（除數 0 回商 0xFFFFFFFF，不引發 #DE），所以 C 版略過除法。
 *                                                   ; 這段會讀 word F[+0x52] 與 F[+0x54]（結果不用）。
 *     r2 ← rand()；angle ← (F[+0x4A] + r2 % 32 + 0xF0) 的低位元組      ; F[+0x4A] 在 *P3 寫入之前讀
 *     *P3（byte）← F[+0x4A]
 *     dist ← word F[+0x54]（重讀；*P3 的寫入或亂數種子的寫入可能改到它）；讀 word F[+0x52]（結果不用）
 *     2BE1:026F(x ＝ word F[+0x5C], y ＝ word F[+0x5E], angle, dist, P1, P2)   ; 推入順序先 F[+0x5E] 後 F[+0x5C]；讀 F[+0x5E] 與 F[+0x5C] 都在 *P3 寫入之後
 * 欄位語意（強推論，名稱不是證據）：F[+0x5C]／[+0x5E] 是艦隊的目標座標？F[+0x4A] 是方位？F[+0x54] 是距離？—— docs 沒有記載
 *   （grep 307D:097E 在 docs/ 只有台帳呼叫關係）；位移與運算才是證據。
 * 副作用：ds:94DC（亂數種子，一次或三次 rand）、P1／P2／P3 的寫入、026F 的寫入（P1、P2）。
 * 輸入域：F 取艦隊表；P1、P2、P3 取暫存區與 seeds。2BE1:026F 的 C 版用線性位址寫 word，所以 P1、P2 的位移不涵蓋 0xFFFF；
 *   本檔 sel == 1 分支的 word 寫入與 *P3 有 ROM 區（線性 0xF0000 以上）模型，F 欄位的 word 讀取在位移 0xFFFF 時高位元組取同一段位移 0。
 *   讀到 A0000 到 BFFFF（視訊記憶體）的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD5_000C(Mem *m);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off,
                            uint16_t ox_seg, uint16_t oy_off, uint16_t oy_seg);

static uint16_t w097E_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void w097E_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w097E_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w097E_put8(m, LIN(seg, off), (uint8_t)v);
    w097E_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_307D_097E(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off,
                 uint16_t p2_seg, uint16_t p3_off, uint16_t p3_seg)
{
    uint16_t r0 = f_3CD5_000C(m);

    if ((r0 & 1u) == 0) {
        uint16_t r2, x, y, dist;
        uint8_t f4a, angle;

        (void)f_3CD5_000C(m);
        (void)w097E_rd16s(m, f_seg, (uint16_t)(f_off + 0x52u));
        (void)w097E_rd16s(m, f_seg, (uint16_t)(f_off + 0x54u));
        r2 = f_3CD5_000C(m);
        f4a = rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Au)));
        angle = (uint8_t)(f4a + (r2 % 32u) + 0xF0u);
        w097E_put8(m, LIN(p3_seg, p3_off), f4a);
        (void)w097E_rd16s(m, f_seg, (uint16_t)(f_off + 0x52u));
        dist = w097E_rd16s(m, f_seg, (uint16_t)(f_off + 0x54u));
        y = w097E_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu));
        x = w097E_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu));
        f_2BE1_026F(m, x, y, angle, dist, p1_off, p1_seg, p2_off, p2_seg);
    } else {
        w097E_wr16s(m, p1_seg, p1_off, w097E_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)));
        w097E_wr16s(m, p2_seg, p2_off, w097E_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)));
        w097E_put8(m, LIN(p3_seg, p3_off), rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Au))));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_307D_097E(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7]);
    return 0;
}
GIN3_REGISTER("307D_097E", 8, run)
