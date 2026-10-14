/*
 * GIN3PS.EXE 31DB:01B6  戰術層挑目標點：以艦隊位置與參考點為對角線框一個（補成正方形的）方框，從集結點表 ds:D70C 收框內（嚴格）的候選，
 *   隨機抽一個，要艦隊到它、它到參考點兩條直線都走得通（1588:0891）才採用；抽到「第 n 號」（不存在的候選）或候選抽光就用參考點
 *
 * 原版（573 bytes，far 呼叫，八個 word 堆疊參數；序言 push bp／si／di、sub sp,54h、mov bp,sp，所以第一個參數在 [bp+5Eh]）：
 *   F ＝ 艦隊記錄的 far 指標（位移 [bp+5Eh]、段 [bp+60h]；位置 F[+0x5C]、F[+0x5E]）、a ＝ 參考點 x（[bp+62h]）、b ＝ 參考點 y（[bp+64h]）、
 *   out1 ＝ 目標 x 的 far 指標（[bp+66h]、[bp+68h]）、out2 ＝ 目標 y 的 far 指標（[bp+6Ah]、[bp+6Ch]）。沒有回傳值（AX 是殘值；描述檔 ret ＝ none）。
 *   區域變數：[bp+0] ＝ 掃描索引 i（byte）、[bp+1] ＝ 候選數（byte）、[bp+2] ＝ 抽到的 idx、[bp+4] ＝ 完成旗標、[bp+6]／[bp+8] ＝ 寬 w／高 h（無號差）、
 *   [bp+0Ah]／[bp+0Eh] ＝ X 範圍 lo／hi、[bp+0Ch]／[bp+10h] ＝ Y 範圍 lo／hi、[bp+12h] ＝ 表項指標、[bp+16h] 起 32 bytes ＝ 候選清單（byte 表索引）、
 *   [bp+36h]／[bp+38h] ＝ 清單寫入指標。
 * 過程（已證實：反組譯逐指令；1588:0891、3CA9:0002、3CD5:000C 由其各自的差異測試證實）：
 *   (fx, fy) ← F[+0x5C], F[+0x5E]
 *   X：fx >= a（無號 jb）→ w ← fx − a、lo ← a、hi ← fx；否則 w ← a − fx、lo ← fx、hi ← a        （同樣處理 Y：h、[bp+0Ch]、[bp+10h]）
 *   (int16) w >= (int16) h（`jl` 不成立）：
 *       mid ← (fy ＋ b) 的 16 位元和邏輯右移 1；Y.hi ← mid ＋ trunc(w ÷ 2)（cwd；sub ax,dx；sar ax,1，有號向零取整）；Y.lo ← Y.hi − w     ; Y 撐成邊長 w
 *   否則：
 *       mid ← (fx ＋ a) 的 16 位元和邏輯右移 1；X.hi ← mid ＋ trunc(h ÷ 2)；X.lo ← X.hi − h                                          ; X 撐成邊長 h
 *   n ← 0；對 i ＝ 0 .. （byte ds:D78C）−1（無號 byte jnb；表項 ds:D70C ＋ 4i 為 (x, y) 兩個 word）：
 *       不是「x ＝ fx 且 y ＝ fy」，且 x > X.lo 且 x < X.hi 且 y > Y.lo 且 y < Y.hi（全部有號 jle／jge）→ 清單[n++] ← i
 *   重複：
 *       r ← 3CD5:000C()；idx ← r % (n ＋ 1)（無號 div，n 是 byte，除數 1..256 不會 #DE）
 *       idx ＝ n → *out1 ← a、*out2 ← b，結束
 *       p ← 表項 ds:D70C ＋ 4 × 清單[idx]
 *       1588:0891(fx, fy, p.x, p.y) ≠ 0 且 1588:0891(p.x, p.y, a, b) ≠ 0 → *out1 ← p.x、*out2 ← p.y，結束
 *       否則 n ← n − 1；k ← n − idx；k ≠ 0 → 3CA9:0002(清單 ＋ idx, 清單 ＋ idx ＋ 1, k)（刪掉清單[idx]，後面的往前搬）
 * 說明：「不是 x ＝ fx 且 y ＝ fy」這個檢查在所有輸入下都是多餘的：W ≥ H 時 X 範圍就是 (min(fx,a), max(fx,a)) 的嚴格內部，fx 必是端點，被 x > lo／x < hi 排除；
 *   W < H 時 Y 範圍同理排除 fy。本版仍照原版寫（C 版沒有第二個 fy 比較的突變可抓，見 31DB_01B6.notes.md）。
 * 語意（強推論）：docs/re/100-tactical-rally-points.md 第 53 到 90 行、docs/internals/43-battle-terrain-and-rally.md 第 74 到 83 行（以艦隊與參考點為對角線框方框、短邊撐成長邊、
 *   候選嚴格在界內且不等於艦隊位置、rand() % (n+1)、抽到最後一號就設成參考點、兩條直線都通才採用、不通就刪掉再抽），與本檔一致。
 *   「不等於艦隊位置」這句 docs 寫成獨立條件，實際上被範圍的嚴格不等式涵蓋。
 * 副作用：*out1、*out2 各一個 word；ds:94DC 的亂數種子；0891 的 ds:8EDC..8EED 暫存；堆疊上的清單（堆疊帶，比對排除）。
 *   word 存取在位移 0xFFFF 時，高位元組取同一段的位移 0（dosgolem 行為）；線性 0xF0000 以上（ROM）的寫入被安靜丟棄，本檔按 byte 模型化。
 *
 * 機器模型：C 版沒有堆疊。候選清單放在原版的實際位置（基底狀態 fcmd-standby.state：SS 固定 0x63EB，進入點 SP ＝ 1E76，探針推 8 個參數字 16 bytes、返回位址 4、
 *   push bp／si／di 6、sub sp,54h，所以 bp ＝ 1E08，清單從 SS:1E1E 起，共 32 bytes），交給 f_3CA9_0002 搬移。清單在探針堆疊帶（SS:1A76..1E75）內，比對時雙方排除。
 *   1588:0891 的 C 版自己存還原 SS:1F78..1F8F（見 f_1588_0891.c），這裡不用再包。
 * 輸入域：F 取艦隊表（ds:C074，32 筆 × 122 bytes）內指標；表項筆數 ds:D78C 列入 freeze 且 seeds 只給 0 到 32（原版清單只有 32 bytes，筆數超過 32 時會蓋到
 *   [bp+36h] 的寫入指標，C 版不涵蓋）；表項、參考點、亂數種子取記憶體與突變。0891 的迴圈在目標超出戰場時不終止，超過探針步數上限的案例由探針丟棄。
 *   輸出指標指向暫存區、ROM 或 seeds 給的位址，不指向堆疊帶與 A0000 到 BFFFF。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint16_t f_1588_0891(Mem *m, uint16_t x, uint16_t y, uint16_t x1, uint16_t y1);
extern uint16_t f_3CD5_000C(Mem *m);
extern uint32_t f_3CA9_0002(Mem *m, uint16_t doff, uint16_t dseg, uint16_t soff, uint16_t sseg, uint16_t count);

#define F01B6_SS 0x63EBu   /* fcmd-standby.state 的 SS（diffprobe 無法設定） */
#define F01B6_LIST 0x1E1Eu /* 進入點 SP 1E76 − 16 − 4 − 6 − 54h ＝ bp(1E08)，清單在 bp+16h */

static uint16_t w01B6_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 0xF0000 以上是 ROM：寫入無效 */
static void w01B6_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void w01B6_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    w01B6_put8(m, LIN(seg, off), (uint8_t)v);
    w01B6_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

/* cwd；sub ax,dx；sar ax,1：有號向零取整的除以 2 */
static uint16_t w01B6_half(uint16_t v)
{
    int16_t s = (int16_t)v;

    return (uint16_t)((s + (s < 0 ? 1 : 0)) >> 1);
}

void f_31DB_01B6(Mem *m, uint16_t f_off, uint16_t f_seg, uint16_t a, uint16_t b, uint16_t o1_off, uint16_t o1_seg, uint16_t o2_off, uint16_t o2_seg)
{
    uint16_t fx = w01B6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu));
    uint16_t fy, w, h, xlo, xhi, ylo, yhi, mid, tx, ty;
    uint8_t cnt = 0, i, tbl_n;
    unsigned guard;

    if (fx >= a) {
        w = (uint16_t)(fx - a);
        xlo = a;
        xhi = fx;
    } else {
        w = (uint16_t)(a - fx);
        xlo = fx;
        xhi = a;
    }
    fy = w01B6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu));
    if (fy >= b) {
        h = (uint16_t)(fy - b);
        ylo = b;
        yhi = fy;
    } else {
        h = (uint16_t)(b - fy);
        ylo = fy;
        yhi = b;
    }
    if ((int16_t)w >= (int16_t)h) {
        mid = (uint16_t)((uint16_t)(w01B6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)) + b) >> 1);
        yhi = (uint16_t)(w01B6_half(w) + mid);
        ylo = (uint16_t)(yhi - w);
    } else {
        mid = (uint16_t)((uint16_t)(w01B6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)) + a) >> 1);
        xhi = (uint16_t)(w01B6_half(h) + mid);
        xlo = (uint16_t)(xhi - h);
    }

    tbl_n = rd8(m, DS(0xD78Cu));
    for (i = 0; i < tbl_n; i = (uint8_t)(i + 1u)) {
        uint16_t e = (uint16_t)(0xD70Cu + 4u * i);

        tx = w01B6_rd16s(m, GIN3_DS, e);
        ty = w01B6_rd16s(m, GIN3_DS, (uint16_t)(e + 2u));
        if (w01B6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)) == tx && w01B6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)) == ty)
            continue;
        if ((int16_t)tx <= (int16_t)xlo || (int16_t)tx >= (int16_t)xhi)
            continue;
        if ((int16_t)ty <= (int16_t)ylo || (int16_t)ty >= (int16_t)yhi)
            continue;
        wr8(m, LIN(F01B6_SS, (uint16_t)(F01B6_LIST + cnt)), i);
        cnt = (uint8_t)(cnt + 1u);
    }

    for (guard = 0; guard < 4000u; guard++) {
        uint16_t r = f_3CD5_000C(m);
        uint16_t idx = (uint16_t)(r % ((uint16_t)cnt + 1u));
        uint16_t e, k;

        if (idx == cnt) {
            w01B6_wr16s(m, o1_seg, o1_off, a);
            w01B6_wr16s(m, o2_seg, o2_off, b);
            return;
        }
        e = (uint16_t)(0xD70Cu + 4u * rd8(m, LIN(F01B6_SS, (uint16_t)(F01B6_LIST + idx))));
        tx = w01B6_rd16s(m, GIN3_DS, e);
        ty = w01B6_rd16s(m, GIN3_DS, (uint16_t)(e + 2u));
        if (f_1588_0891(m, w01B6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Cu)), w01B6_rd16s(m, f_seg, (uint16_t)(f_off + 0x5Eu)), tx, ty) != 0) {
            if (f_1588_0891(m, tx, ty, a, b) != 0) {
                w01B6_wr16s(m, o1_seg, o1_off, tx);
                w01B6_wr16s(m, o2_seg, o2_off, ty);
                return;
            }
        }
        cnt = (uint8_t)(cnt - 1u);
        k = (uint16_t)((uint16_t)cnt - idx);
        if (k != 0)
            (void)f_3CA9_0002(m, (uint16_t)(F01B6_LIST + idx), F01B6_SS, (uint16_t)(F01B6_LIST + idx + 1u), F01B6_SS, k);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_31DB_01B6(m, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7]);
    return 0;
}
GIN3_REGISTER("31DB_01B6", 8, run)
