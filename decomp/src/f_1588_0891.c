/*
 * GIN3PS.EXE 1588:0891  直線通行判定：從 (x, y) 朝 (x1, y1) 每次前進 8 單位，每一步用 1588:04EA 檢查落點矩形是否通行；距離不足 8.0 時直接跳到終點
 *
 * 原版（159 bytes，far 呼叫，四個 word 堆疊參數；序言 push bp／si、sub sp,2、mov bp,sp，所以 x 在 [bp+0Ah]、y 在 [bp+0Ch]、x1 在 [bp+0Eh]、y1 在 [bp+10h]；
 *   [bp+0] ＝ 方位 byte，只存不讀）。回傳 AX ＝ 1（到達）或 0（某一步不通行）。x、y 就是堆疊上的參數槽，每步被 2BE1:026F 經 far 指標（SS:bp+0Ah、SS:bp+0Ch）就地改寫。
 * 過程（已證實：反組譯逐指令；3310:0006、3CC9:00B9、2BE1:011D、2BE1:026F、1588:04EA 由其各自的差異測試證實）：
 *   迴圈：
 *     x ＝ x1 且 y ＝ y1 → 回 1
 *     d ← 3310:0006(x, y, x1, y1)                          ; 16.16 距離
 *     3CC9:00B9(d, 8.0) 後 jge（SF ＝ OF，有號 32 位元 d >= 0x80000）不成立 → x ← x1、y ← y1，回迴圈頂端（下一輪回 1，不呼叫 04EA）
 *       （d 的高字 >= 0x8000 時有號為負，也走這條）
 *     方位 ← 2BE1:011D(x1 − x, y1 − y)                      ; 16 位元相減
 *     2BE1:026F(x, y, 方位, 8, &x, &y)                      ; 朝該方位前進 8，夾進戰場範圍，寫回 x、y
 *     1588:04EA(x, y) ＝ 0 → 回 0；否則繼續迴圈
 * 語意（強推論）：docs/re/123-fleet-move-engine.md 第 60 行（internal/rules/terrain.go 的 LineOpen ＝ 原版 1588:0891 的直線判定）、docs/re/102-battle-movement.md 第 21 行
 *   （每一步長度 √65 ＝ 8.06 和直線通行判定相同尺度）、docs/re/294 第 137 行（1588:0891 驗過直線通行的點）。名稱不是證據。
 * 副作用：026F 對 x、y 槽的寫入（堆疊帶內，比對排除）、3310:0006 內 39B5:005C 的暫存（ds:8EDC..8EED）；04EA／0621 不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊。x、y 放在原版的實際位置（基底狀態 fcmd-standby.state：SS 固定 0x63EB，進入點 SP ＝ 1E76，探針推 4 個參數字 8 bytes、返回位址 4、
 *   push bp／si 4、sub sp,2，所以 bp ＝ 1E64，x 在 SS:1E6E、y 在 SS:1E70），交給 f_2BE1_026F 用 far 指標改寫，再從那裡讀回。這個位址在探針的堆疊帶內，比對時雙方都排除。
 *   迴圈保險：原版在目標超出戰場夾限（x1 > 0x7EF 等）時不終止，探針以步數上限丟棄那些案例；C 版只用於探針收錄的終止案例（每輪至少數百步，
 *   步數預算內不可能超過 4000 輪），迴圈上限 20000 輪僅為避免突變版無窮迴圈，超過回 0xDEAD。
 * 輸入域：x、y、x1、y1 全部 16 位元值，但探針收錄的只有在步數預算內終止的案例；戰場地圖與遮罩取自 fcmd-standby.state 與突變。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint16_t f_2BE1_026F(Mem *m, uint16_t x, uint16_t y, uint16_t angle_word, uint16_t dist, uint16_t ox_off, uint16_t ox_seg, uint16_t oy_off,
                            uint16_t oy_seg);
extern uint16_t f_1588_04EA(Mem *m, uint16_t x, uint16_t y);

#define F0891_SS 0x63EBu /* fcmd-standby.state 的 SS（diffprobe 無法設定） */
#define F0891_X 0x1E6Eu  /* 進入點 SP 1E76 − 8 − 4 − 4 − 2 ＝ bp(1E64)，x 在 bp+0Ah */
#define F0891_Y 0x1E70u  /* bp+0Ch */
#define F0891_SF 0x0080u
#define F0891_OF 0x0800u

/*
 * 3310:0006 的 C 版把 39B5:000C／005C 的堆疊參數放在 fcmd-base.state 的框架（SS:1F82..1F89，見 f_39B5_005C.c），
 * 在 fcmd-base.state 下那裡是探針的堆疊帶，比對時排除。fcmd-standby.state 的進入點 SP 是 1E76，帶是 SS:1A76..1E75，
 * 該框架落在帶外，C 版的暫存寫入會被誤判成多寫。原版在此狀態下把被呼叫者的框架放在更深的帶內。
 * 這裡存下 SS:1F78..1F8F 呼叫前的內容並在呼叫後還原：這 24 bytes 不是本函式的輸出，只是 C 版被呼叫者的模型暫存。
 */
static uint32_t f0891_dist(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2)
{
    uint8_t save[24];
    uint32_t r;
    int k;

    for (k = 0; k < 24; k++)
        save[k] = rd8(m, LIN(F0891_SS, (uint16_t)(0x1F78u + k)));
    r = f_3310_0006(m, x1, y1, x2, y2);
    for (k = 0; k < 24; k++)
        wr8(m, LIN(F0891_SS, (uint16_t)(0x1F78u + k)), save[k]);
    return r;
}

uint16_t f_1588_0891(Mem *m, uint16_t x, uint16_t y, uint16_t x1, uint16_t y1)
{
    int guard;

    wr16(m, LIN(F0891_SS, F0891_X), x);
    wr16(m, LIN(F0891_SS, F0891_Y), y);
    for (guard = 0; guard < 20000; guard++) {
        uint32_t d;
        uint16_t fl, cx, cy;
        uint8_t ang;

        cx = rd16(m, LIN(F0891_SS, F0891_X));
        cy = rd16(m, LIN(F0891_SS, F0891_Y));
        if (cx == x1 && cy == y1)
            return 1;
        d = f0891_dist(m, cx, cy, x1, y1);
        fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 8, 0);
        if (((fl & F0891_SF) != 0) != ((fl & F0891_OF) != 0)) { /* jge 不成立 */
            wr16(m, LIN(F0891_SS, F0891_X), x1);
            wr16(m, LIN(F0891_SS, F0891_Y), y1);
            continue;
        }
        ang = f_2BE1_011D(m, (uint16_t)(x1 - cx), (uint16_t)(y1 - cy));
        (void)f_2BE1_026F(m, cx, cy, ang, 8, F0891_X, F0891_SS, F0891_Y, F0891_SS);
        cx = rd16(m, LIN(F0891_SS, F0891_X));
        cy = rd16(m, LIN(F0891_SS, F0891_Y));
        if (f_1588_04EA(m, cx, cy) == 0)
            return 0;
    }
    return 0xDEADu;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_1588_0891(m, a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("1588_0891", 4, run)
