/*
 * GIN3PS.EXE 203A:00EC  武器的距離曲線：兩個物件 (P1, P2) 的距離比值 T 落在曲線的哪一段，線性內插出 y，回 y × power ÷ 100 的低 16 位元
 *
 * 原版（296 bytes，far 呼叫，四個 word 堆疊參數 ＝ 兩個 far 指標 P1、P2；序言 push bp／si／di、sub sp,24h、mov bp,sp，所以 P1 的位移在 [bp+2Eh]、段在 [bp+30h]，
 *   P2 的位移在 [bp+32h]、段在 [bp+34h]）。回傳 AX（描述檔 ret ＝ ax）。
 *   區域變數：[bp+0]／[bp+2] ＝ T（低字／高字）、[bp+14h]／[bp+16h] ＝ 前一項曲線指標（位移／段）、[bp+18h]／[bp+1Ah] ＝ 目前曲線指標、[bp+1Ch]／[bp+1Eh] ＝ 類型表項指標。
 * 過程（已證實：反組譯逐指令；203A:03B0、39B5:005C、3CC9:0009、3CC9:0020、3CC9:00B9 各由其差異測試證實）：
 *   idx ← byte ds:A762；tbl ← ds:(4FDA ＋ low16(idx × 6))                    ; 類型表項：power ＝ tbl[0]、A ＝ tbl[2]、B ＝ tbl[4]（word）
 *   T ← 203A:03B0(P1, P2, A, B)                                               ; AX:BX（高:低），16.16 的「距離 ÷ [A × (1 ＋ e)]」
 *   cur ← far ptr (word ds:(477E ＋ 4 × idx) : word ds:(477C ＋ 4 × idx))     ; 曲線：每項 6 bytes ＝ x（dword，16.16，低字在前）＋ y（word）
 *   重複：E ← dword cur[0]；3CC9:00B9(E, T) 之後 jge（有號 32 位元 E >= T）→ 跳出；否則 prev ← cur；cur ← cur ＋ 6（位移 16 位元回繞）
 *   找到（cur ＝ 第一項 E >= T）後：
 *     delta ← E − dword prev[0]；rem ← T − dword prev[0]                       ; 32 位元回繞
 *     dy ← word cur[4] − word prev[4]                                           ; 16 位元
 *     q ← 39B5:005C(rem, delta)                                                 ; (rem << 16) ÷ delta，有號 16.16；除數 0 回 0
 *     p ← 3CC9:0009(q, 符號延伸 dy 成 32 位元)                                   ; 低 32 位元；取高字（整數部分）
 *     v ← p 高字 ＋ word prev[4]（prev 的位址在呼叫之後重讀）                    ; 16 位元
 *     回 低字(3CC9:0020(3CC9:0009(符號延伸 v, 零延伸 tbl[0]) ÷ 100))             ; 無號 32 位元除法
 * 原版的細節（本版照做）：
 *   - 第一項就 E >= T（例如 T ＝ 0 且第一項 x ＝ 0）時，prev 沒有被寫過，[bp+14h]／[bp+16h] 是進入時堆疊上的殘值。docs/re/404-combat-sequence.md 第 20 到 24 行
 *     （「零距離讀未初始化的前項指標」，證據等級 proven，固定樣本的前項為 00D7:0001）已記載。殘值位置：fcmd-base.state 的 SS ＝ 0x63EB、進入點 SP ＝ 0x1F8A，
 *     探針推 4 個參數字 8 bytes、返回位址 4、push bp／si／di 6、sub sp,24h，所以 bp ＝ 0x1F54，[bp+14h] ＝ SS:1F68（位移）、[bp+16h] ＝ SS:1F6A（段）。
 *     被呼叫者的框架都在 bp 以下，不會改到這裡；C 版在進入時先讀下來（C 版的被呼叫者把暫存寫在固定的 1F7x 一帶，不碰 1F68，但先讀最穩）。基底狀態一換，這兩個常數要重算。
 *   - 3CC9:0009 的第二個運算元是 dy 的符號延伸（cwd），3CC9:0009 取低 32 位元，所以負斜率取「有號乘積的高字」，不是往零截斷（docs/re/402-fleet-combat-damage.md 第 55 行）。
 *   - 最後的除法是無號：v 為負時符號延伸成 0xFFFFxxxx，乘 power 後以無號 32 位元除以 100。
 * 語意（強推論）：docs/re/402-fleet-combat-damage.md 第 44 到 55 行：類型表為 DS:4FDA 起、stride 6 的 power／長半軸／短半軸；曲線遠指標表為 DS:477C 起、stride 4；
 *   `q ＝ FixedDist ÷ [a × (1 ＋ e)]`、`fraction ＝ FixedDiv(q − left.x, right.x − left.x)`、`curveY ＝ left.y ＋ signed 高字(fraction × (right.y − left.y))`、
 *   `weaponPower ＝ curveY × power ÷ 100`；docs/internals/38-battle-units.md 第 375 行。與本檔一致。ds:A762 是目前武器的類型編號（用途為強推論）。
 * 副作用：只有 203A:03B0 與 39B5:005C 寫的 ds:8EDC..8EED；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊；前項指標的殘值見上。被呼叫者的堆疊位置都在探針的堆疊帶內，比對時排除。
 * 輸入域：P1、P2 指向艦艇池記錄或其他不在 A0000 到 BFFFF 的位址；曲線與類型表取自記憶體；曲線一定要在遇到 E >= T 的項目後結束（原版不終止的案例被探針丟棄）。
 *   前項殘值不指向堆疊帶與 A0000 到 BFFFF。word 讀取在位移 0xFFFF 時 dosgolem 的高位元組取同一段的位移 0（rd16s）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_203A_03B0(Mem *m, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off, uint16_t p2_seg, uint16_t a, uint16_t b);
extern uint32_t f_39B5_005C(Mem *m, uint16_t a_lo, uint16_t a_hi, uint16_t b_lo, uint16_t b_hi);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0020(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

#define F00EC_SS 0x63EBu  /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F00EC_PREV 0x1F68u /* [bp+14h]：進入點 SP 1F8A − 參數 8 − 返回位址 4 − push bp／si／di 6 − sub sp,24h ＝ bp(1F54)，再加 14h；段在 +2 */
#define F00EC_SF 0x0080u
#define F00EC_OF 0x0800u

static uint16_t w00EC_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static uint32_t w00EC_rd32s(const Mem *m, uint16_t seg, uint16_t off)
{
    return ((uint32_t)w00EC_rd16s(m, seg, (uint16_t)(off + 2u)) << 16) | w00EC_rd16s(m, seg, off);
}

uint16_t f_203A_00EC(Mem *m, uint16_t p1_off, uint16_t p1_seg, uint16_t p2_off, uint16_t p2_seg)
{
    uint16_t prev_off = w00EC_rd16s(m, F00EC_SS, F00EC_PREV);               /* 殘值：[bp+14h] */
    uint16_t prev_seg = w00EC_rd16s(m, F00EC_SS, (uint16_t)(F00EC_PREV + 2u)); /* 殘值：[bp+16h] */
    uint8_t idx = rd8(m, DS(0xA762u));
    uint16_t tbl = (uint16_t)(0x4FDAu + (uint16_t)(idx * 6u));
    uint16_t a = w00EC_rd16s(m, GIN3_DS, (uint16_t)(tbl + 2u));
    uint16_t b = w00EC_rd16s(m, GIN3_DS, (uint16_t)(tbl + 4u));
    uint32_t t = f_203A_03B0(m, p1_off, p1_seg, p2_off, p2_seg, a, b);
    uint16_t cur_off = w00EC_rd16s(m, GIN3_DS, (uint16_t)((uint16_t)(idx << 2) + 0x477Cu));
    uint16_t cur_seg = w00EC_rd16s(m, GIN3_DS, (uint16_t)((uint16_t)(idx << 2) + 0x477Eu));
    uint32_t e, delta, rem, q, prod;
    uint16_t dy, v, fl, power;
    uint64_t dv;

    for (;;) {
        e = w00EC_rd32s(m, cur_seg, cur_off);
        fl = f_3CC9_00B9((uint16_t)(e >> 16), (uint16_t)e, (uint16_t)(t >> 16), (uint16_t)t);
        if (((fl & F00EC_SF) != 0) == ((fl & F00EC_OF) != 0)) /* jge：E >= T */
            break;
        prev_seg = cur_seg;
        prev_off = cur_off;
        cur_off = (uint16_t)(cur_off + 6u);
    }
    delta = e - w00EC_rd32s(m, prev_seg, prev_off);
    rem = t - w00EC_rd32s(m, prev_seg, prev_off);
    dy = (uint16_t)(w00EC_rd16s(m, cur_seg, (uint16_t)(cur_off + 4u)) - w00EC_rd16s(m, prev_seg, (uint16_t)(prev_off + 4u)));
    q = f_39B5_005C(m, (uint16_t)rem, (uint16_t)(rem >> 16), (uint16_t)delta, (uint16_t)(delta >> 16));
    prod = f_3CC9_0009((uint16_t)(q >> 16), (uint16_t)q, (uint16_t)((int16_t)dy < 0 ? 0xFFFFu : 0u), dy);
    v = (uint16_t)((uint16_t)(prod >> 16) + w00EC_rd16s(m, prev_seg, (uint16_t)(prev_off + 4u)));
    power = w00EC_rd16s(m, GIN3_DS, tbl);
    prod = f_3CC9_0009((uint16_t)((int16_t)v < 0 ? 0xFFFFu : 0u), v, 0u, power);
    dv = f_3CC9_0020((uint16_t)(prod >> 16), (uint16_t)prod, 0u, 100u);
    return (uint16_t)(dv >> 32);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_203A_00EC(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("203A_00EC", 4, run)
