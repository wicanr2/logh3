/*
 * GIN3PS.EXE 1C03:0185  戰場艦隊的一個移動步：清 +0x49 bit0；位置已等於目的地時只轉向（1C03:0524）；否則先朝目的地方位轉一步，
 *   行進額度 +0x56 >= 8.0 才往前走 8 像素（扣掉實際走的距離）、能站才更新位置，地形擋住時改用 2583:0451 找旁邊的落點
 *
 * 原版（476 bytes，far 呼叫，一個遠指標堆疊參數 F（位移 [bp+2Ah]、段 [bp+2Ch]）；序言 push bp／si／di、sub sp,20h、mov bp,sp）。
 *   回傳 AL（描述檔 ret ＝ pack:al；AH 是各路徑的殘值）。區域變數：[bp+0] ＝ 040C 的結果（word）、[bp+2] ＝ 方位角（byte）、[bp+3] ＝ 結果旗標（byte）、
 *   [bp+4]／[bp+6] ＝ 落點 x／y（026F 的輸出指標指向它們，SS 段）、[bp+14h]／[bp+16h] ＝ 實際走的距離（低字／高字）。
 * 過程（已證實：反組譯逐指令；1C03:0524、2BE1:011D、2BE1:026F、2583:040C、2583:0451、3310:0006、3CC9:00B9 各由其差異測試證實）：
 *   res ← 0；F[+0x49] &= 0xFE
 *   F[+0x5C] == F[+0x64] 且 F[+0x5E] == F[+0x66]（位置 ＝ 目的地，16 位元）：
 *     F[+0x4A] == F[+0x4C]（byte）→ 回 0
 *     否則回 1C03:0524(F, F[+0x4C])         ; 轉一步朝向（原版回傳 0524 的 AX 原樣，AH 是殘值；描述檔 ret ＝ pack:al，C 版只回 AL）
 *   否則：
 *     ang ← 2BE1:011D(a ＝ F[+0x64] − F[+0x5C], b ＝ F[+0x66] − F[+0x5E])
 *     ang ≠ F[+0x4A]（byte）：res ← AL of 1C03:0524(F, ang)；res == 1 時 F[+0x49] |= 1
 *   3CC9:00B9(dword F[+0x56], 8.0) 之後 jge 不成立（行進額度 < 8.0，有號 32 位元）→ 回 res
 *   res ← 1
 *   d ← 3310:0006(F[+0x5C], F[+0x5E], F[+0x64], F[+0x66])
 *   d <= 8.0（jg 不成立）：落點 (px, py) ← (F[+0x64], F[+0x66])
 *   否則：2BE1:026F(F[+0x5C], F[+0x5E], 角度 ＝ word F[+0x4A], dist ＝ 8, →px, →py)     ; 角度用「轉過之後」的 +0x4A，不是 ang
 *   moved ← 3310:0006(F[+0x5C], F[+0x5E], px, py)；moved ＝ 0 時改成 1（低字 1、高字 0）
 *   dword F[+0x56] −= moved（32 位元 sub／sbb，記憶體目的）
 *   r ← 2583:040C(px, py, word F[+1])
 *   r == 0：F[+0x5C] ← px；F[+0x5E] ← py
 *   r == 0xFFFF（地形擋住）：2583:0451(x ＝ F[+0x5C] 的值, y ＝ F[+0x5E] 的值, A ＝ →F[+0x5C], B ＝ →F[+0x5E], dist ＝ 8, faction ＝ word F[+1])
 *   其他（被別的分隊佔住）：不動
 *   回 res（＝ 1）
 * 語意（強推論）：F 是艦隊記錄；+0x56 是 16.16 行進額度（1C03:0524 的檔頭、docs/internals/38、spec/080）；+0x5C／+0x5E 現位置、+0x64／+0x66 目的地、+0x4A 朝向、
 *   +0x4C 抵達朝向、+0x49 bit0 是「這步轉向了」旗標（強推論，呼叫者 1C03:015C）。docs/re/294-short-move-strafe-and-arrival-facing.md、docs/playtest/011-fleet-short-move-strafe.md。
 * 副作用：F[+0x49]、F[+0x4A]（0524）、F[+0x56..0x59]、F[+0x5C]、F[+0x5E]（含 0451 的還原寫入）；被呼叫者的 ds:8EDC..8EED 暫存。
 *
 * 機器模型：C 版沒有堆疊。落點 px、py 這兩個「堆疊字」放在 SS:1F00、SS:1F02（基底 fcmd-base.state，SS 固定 0x63EB，探針堆疊帶 SS:1B8A..1F89 之內，
 *   寫入集合比對時雙方都排除；原版的位置 [bp+4]／[bp+6] 與被呼叫者 39B5 C 版借用的位置不同，所以不放原位）；2BE1:026F 的 C 版以遠指標寫它們，再讀回。
 * 輸入域：F 取艦隊表（ds:C074，32 筆 × 122）或 seeds 給定的 RAM 位置（輸出不在 ROM 區）；基底 fcmd-base.state，地圖、圖塊遮罩表、在場分隊清單與艦艇池由 seeds 造
 *   （fcmd-base 的戰場地圖指標是 0:0）。word 存取在位移 0xFFFF 時 dosgolem 的高位元組取同一段的位移 0（rd16s／wr16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_1C03_0524(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t target);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off, uint16_t oy_seg);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint16_t f_2583_040C(Mem *m, uint16_t x, uint16_t y, uint16_t faction);
extern uint16_t f_2583_0451(Mem *m, uint16_t x, uint16_t y, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg, uint16_t dist, uint16_t faction);

#define F0185_SS 0x63EBu
#define F0185_PX 0x1F00u
#define F0185_PY 0x1F02u

#define F0185_ZF 0x0040u
#define F0185_SF 0x0080u
#define F0185_OF 0x0800u

static uint16_t f1C03_0185_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f1C03_0185_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

uint16_t f_1C03_0185(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint8_t res = 0;
    uint16_t fl, px, py, mv_lo, mv_hi, lo_old, r;
    uint32_t d, mv;

    wr8(m, LIN(f_seg, (uint16_t)(f_off + 0x49u)), (uint8_t)(rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x49u))) & 0xFEu));
    if (f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)) == f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x64u))
        && f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)) == f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x66u))) {
        uint8_t cur = rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Au)));
        uint8_t arr = rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Cu)));

        if (cur == arr)
            return 0;
        return (uint8_t)f_1C03_0524(m, f_off, f_seg, arr); /* ret ＝ pack:al：只比 AL，0524 回傳的 AH 是殘值 */
    }
    {
        uint8_t ang = f_2BE1_011D(m,
                                  (uint16_t)(f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x64u)) - f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu))),
                                  (uint16_t)(f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x66u)) - f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu))));

        if (ang != rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Au)))) {
            res = (uint8_t)f_1C03_0524(m, f_off, f_seg, ang);
            if (res == 1) {
                res = 1;
                wr8(m, LIN(f_seg, (uint16_t)(f_off + 0x49u)), (uint8_t)(rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x49u))) | 1u));
            }
        }
    }
    fl = f_3CC9_00B9(f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x58u)), f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x56u)), 0x0008u, 0x0000u);
    if (((fl & F0185_SF) != 0) != ((fl & F0185_OF) != 0)) /* jge 不成立 */
        return res;
    res = 1;
    d = f_3310_0006(m,
                    f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)), f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)),
                    f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x64u)), f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x66u)));
    fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 0x0008u, 0x0000u);
    if (!(fl & F0185_ZF) && ((fl & F0185_SF) != 0) == ((fl & F0185_OF) != 0)) { /* jg：d > 8.0 */
        (void)f_2BE1_026F(m,
                          f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)), f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)),
                          f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x4Au)), 8,
                          F0185_PX, F0185_SS, F0185_PY, F0185_SS);
        px = rd16(m, LIN(F0185_SS, F0185_PX));
        py = rd16(m, LIN(F0185_SS, F0185_PY));
    } else {
        px = f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x64u));
        py = f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x66u));
    }
    mv = f_3310_0006(m,
                     f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)), f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)),
                     px, py);
    mv_lo = (uint16_t)mv;
    mv_hi = (uint16_t)(mv >> 16);
    if (mv_lo == 0 && mv_hi == 0) {
        mv_lo = 1;
        mv_hi = 0;
    }
    lo_old = f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x56u));
    f1C03_0185_wr16s(m, f_seg, (uint16_t)(f_off + 0x56u), (uint16_t)(lo_old - mv_lo));
    {
        unsigned borrow = lo_old < mv_lo ? 1u : 0u;
        uint16_t hi_old = f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x58u));

        f1C03_0185_wr16s(m, f_seg, (uint16_t)(f_off + 0x58u), (uint16_t)(hi_old - mv_hi - borrow));
    }
    r = f_2583_040C(m, px, py, f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 1u)));
    if (r == 0) {
        f1C03_0185_wr16s(m, f_seg, (uint16_t)(f_off + 0x5Cu), px);
        f1C03_0185_wr16s(m, f_seg, (uint16_t)(f_off + 0x5Eu), py);
    } else if (r == 0xFFFFu) {
        (void)f_2583_0451(m,
                          f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)), f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)),
                          (uint16_t)(f_off + 0x5Cu), f_seg, (uint16_t)(f_off + 0x5Eu), f_seg, 8,
                          f1C03_0185_rd16s(m, f_seg, (uint16_t)(f_off + 1u)));
    }
    return res;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1C03_0185(m, a[0], a[1]); }
GIN3_REGISTER("1C03_0185", 2, run)
