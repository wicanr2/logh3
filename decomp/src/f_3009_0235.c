/*
 * GIN3PS.EXE 3009:0235  艦隊朝目的地走一步：依狀態（艦隊 +0x48 低 4 位）把「下一個落點」寫進 +0x74／+0x76、朝向寫進 +0x78
 *
 * 原版（306 bytes，far 呼叫，一個遠指標堆疊參數 F；序言 push bp／si／di、sub sp,14h、mov bp,sp，所以 F 的位移在 [bp+1Eh]、段在 [bp+20h]）。
 *   沒有有意義的回傳值（AX 是各路徑的殘值，描述檔 ret ＝ none）。區域變數 [bp+0] ＝ 角度（byte）、[bp+2]／[bp+4] ＝ 距離（低字／高字）、
 *   [bp+8..12h] ＝ 暫存的段／位移（寫了再讀，沒有可觀察效果）。
 * 過程（已證實：反組譯逐指令；2BE1:011D、3310:0006、2BE1:026F、3CC9:00B9 各由其差異測試證實）：
 *   st ← F[+0x48] & 0x0F
 *   st == 4：F[+0x74] ← F[+0x5C]（word）；F[+0x76] ← F[+0x5E]（word）；F[+0x78] ← F[+0x4A]（byte）；結束
 *   st 不是 0 也不是 1：結束（不寫）
 *   st 是 0 或 1：
 *     ang ← 2BE1:011D(a ＝ F[+0x64] − F[+0x5C], b ＝ F[+0x66] − F[+0x5E])   ; 現位置 (+0x5C, +0x5E) 朝目的地 (+0x64, +0x66) 的方位角
 *     d ← 3310:0006(F[+0x5C], F[+0x5E], F[+0x64], F[+0x66])                  ; 16.16 距離
 *     S ← dword F[+0x52]（低字 +0x52、高字 +0x54）                          ; 16.16 速度（強推論，見下）
 *     3CC9:00B9(d, S) 之後 jge（有號 32 位元 d >= S，ZF ＝ 1 或 SF ＝ OF）：
 *       成立（走一步）：2BE1:026F(x ＝ F[+0x5C], y ＝ F[+0x5E], 角度 ＝ ang, dist ＝ F[+0x54]（S 的高字）, 輸出 x → F+0x74, 輸出 y → F+0x76)；F[+0x78] ← ang
 *       不成立（d < S，到達）：F[+0x74] ← F[+0x64]；F[+0x76] ← F[+0x66]；F[+0x78] ← F[+0x4C]（byte；每個讀取都在前一個寫入之後）
 * 語意（強推論）：F 是艦隊記錄（ds:C074 起 122 bytes 一筆）。+0x5C／+0x5E 現位置、+0x64／+0x66 目的地（docs/re/288-fleet-record-remaining-fields.md、
 *   docs/internals/33-fleet-table.md）；+0x52 的 dword 是每回合移動的 16.16 距離上限，+0x54 是它的整數部分，所以一步的 dist 取 +0x54（強推論，只由用法推得）；
 *   +0x74／+0x76／+0x78 是「下一步落點與朝向」（假說，名稱不是證據）；+0x48 低 4 位是艦隊狀態（4 ＝ 就地不動，0／1 ＝ 移動中，強推論）。
 * 副作用：F+0x74、+0x76、+0x78；2BE1:026F 寫的就是 F+0x74／+0x76；3310:0006 內部的 ds:8EDC..8EED 暫存由被呼叫者的 C 版寫入。
 *
 * 輸入域：F 取艦隊表（ds:C074，32 筆 × 122），或 seeds 給定的任意段:位移（不在 ROM 區）。走一步路徑的 2BE1:026F C 版按線性位址寫 +0x74／+0x76（f_2BE1_026F.c 檔頭），所以該路徑不涵蓋 +0x74／+0x76 的 word 落在位移 0xFFFF 的輸入；狀態 4 路徑用 wr16s，涵蓋回繞寫入（seeds 的 8000:FF8B、FF89）。word 讀取在位移回繞時，
 *   dosgolem 的高位元組取同一段的位移 0（rd16s／wr16s）。AX 殘值不比較。
 *   [bp+1]（角度 word 的高位元組）是堆疊殘值，2BE1:026F 只取低位元組，不影響結果。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off, uint16_t oy_seg);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

static uint16_t f3009_0235_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f3009_0235_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

#define F0235_ZF 0x0040u
#define F0235_SF 0x0080u
#define F0235_OF 0x0800u

void f_3009_0235(Mem *m, uint16_t f_off, uint16_t f_seg)
{
    uint8_t st = (uint8_t)(rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x48u))) & 0x0Fu);
    uint8_t ang;
    uint32_t d;
    uint16_t fl;
    uint16_t x1, y1, x2, y2;

    if (st == 4) {
        f3009_0235_wr16s(m, f_seg, (uint16_t)(f_off + 0x74u), f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)));
        f3009_0235_wr16s(m, f_seg, (uint16_t)(f_off + 0x76u), f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)));
        wr8(m, LIN(f_seg, (uint16_t)(f_off + 0x78u)), rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Au))));
        return;
    }
    st = (uint8_t)(rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x48u))) & 0x0Fu);
    if (st != 0 && st != 1)
        return;
    {
        uint16_t b = (uint16_t)(f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x66u)) - f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)));
        uint16_t a = (uint16_t)(f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x64u)) - f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)));

        ang = f_2BE1_011D(m, a, b);
    }
    x1 = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu));
    y1 = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu));
    x2 = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x64u));
    y2 = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x66u));
    d = f_3310_0006(m, x1, y1, x2, y2);
    {
        uint16_t s_hi = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x54u));
        uint16_t s_lo = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x52u));

        fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, s_hi, s_lo);
    }
    if (((fl & F0235_SF) != 0) == ((fl & F0235_OF) != 0)) { /* jge：d >= S */
        uint16_t dist = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x54u));
        uint16_t x = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu));
        uint16_t y = f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu));

        (void)f_2BE1_026F(m, x, y, ang, dist, (uint16_t)(f_off + 0x74u), f_seg, (uint16_t)(f_off + 0x76u), f_seg);
        wr8(m, LIN(f_seg, (uint16_t)(f_off + 0x78u)), ang);
    } else { /* 到達 */
        f3009_0235_wr16s(m, f_seg, (uint16_t)(f_off + 0x74u), f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x64u)));
        f3009_0235_wr16s(m, f_seg, (uint16_t)(f_off + 0x76u), f3009_0235_rd16s(m, f_seg, (uint16_t)(f_off + 0x66u)));
        wr8(m, LIN(f_seg, (uint16_t)(f_off + 0x78u)), rd8(m, LIN(f_seg, (uint16_t)(f_off + 0x4Cu))));
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3009_0235(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("3009_0235", 2, run)
