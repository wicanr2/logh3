/*
 * GIN3PS.EXE 321F:0392  登陸／圍攻場景的兩點取景：取艦隊 A 的艦艇與行星 B 的兩個座標點，算朝向（取負）、兩點中點、
 *   與中點到兩端點的距離（整數部分 × 3 ÷ 2）排成的四個像素值，寫進 ds:D7EA..D7FD
 *
 * 原版（385 bytes，far 呼叫，兩個遠指標堆疊參數 A（位移 [bp+16h]、段 [bp+18h]）、B（位移 [bp+1Ah]、段 [bp+1Ch]）；序言 push bp／si／di、sub sp,0Ch、
 *   mov bp,sp）。沒有回傳值（描述檔 ret ＝ none）。區域變數：[bp+0]／[bp+2] ＝ 中點 x／y、[bp+4]／[bp+6] ＝ 距離（低字／高字）、[bp+8]／[bp+0Ah] ＝ 艦艇指標。
 * 過程（已證實：反組譯逐指令；2BE1:011D、3310:0006、3CC9:0009、3CC9:0070、3CC9:00B9 各由其差異測試證實）：
 *   ship ← (word ds:CFB6):(word ds:CFB4 ＋ low16(A[+0x10]（word）× 0x4E))
 *   byte A[+1] == 0：(ds:D7EE, ds:D7F0) ← (ship[+0x34], ship[+0x36])；(ds:D7F2, ds:D7F4) ← (B[+0x46], B[+0x48])
 *   否則           ：(ds:D7EE, ds:D7F0) ← (B[+0x46], B[+0x48])；(ds:D7F2, ds:D7F4) ← (ship[+0x34], ship[+0x36])
 *     （每個 word 都是先讀、再寫；來源與 ds:D7EE.. 重疊時，後讀的看得到先寫的）
 *   ang ← 2BE1:011D(a ＝ D7F2 − D7EE, b ＝ D7F4 − D7F0)；byte ds:D7EA ← (0 − ang) 的低位元組
 *   中點：[bp+0] ← (D7EE ＋ D7F2) 的低 16 位元 >> 1（邏輯右移）；[bp+2] ← (D7F0 ＋ D7F4) 的低 16 位元 >> 1
 *   d ← 3310:0006(D7EE, D7F0, D7F2, D7F4)（16.16 距離）
 *   3CC9:00B9(d, 200.0) 之後 jle（有號 32 位元 d <= 200.0）：
 *     q1 ← 3310:0006(mid_x, mid_y, D7EE, D7F0) 的整數部分（AX）；ds:D7F6 ← 0x140 − low16((q1 × 3) ÷ 2)（3CC9:0009 乘、3CC9:0070 有號除）；ds:D7F8 ← 0x12A
 *     q2 ← 3310:0006(mid_x, mid_y, D7F2, D7F4) 的整數部分；ds:D7FA ← low16((q2 × 3) ÷ 2) ＋ 0x140；ds:D7FC ← 0x12A
 *   否則：ds:D7F6 ← 0xAA；ds:D7F8 ← 0x12A；ds:D7FA ← 0x1D6；ds:D7FC ← 0x12A
 *   （原版在第一次呼叫 3310:0006 前多推一個 0x140，呼叫後只清 8 bytes，留在堆疊上由後面的 `pop ax` 取出；C 版直接用常數。）
 * 語意（強推論）：艦隊 A 的 +0x10 是艦艇池索引、+1 是陣營（0 ＝ 帝國或同盟之一），池艦艇 +0x34／+0x36 是座標；B 是行星（+0x46／+0x48 座標，
 *   docs/spec/019-planet-assault.md 第 50 行）。0x140 ＝ 320、0x12A ＝ 298 是畫面常數（強推論，docs/internals/data/landing-frame.json 與 fleet-scene.json 有相同常數）；
 *   ds:D7EA..D7FD 的各格語意未查（假說：登陸場景的取景框，兩個點在畫面上對稱擺放）。
 * 副作用：ds:D7EA、D7EE..D7F4、D7F6..D7FC；3310:0006 內部的 ds:8EDC..8EED 暫存由被呼叫者的 C 版寫入。
 * 邊界：d 恰為 200.0 時（只有座標軸方向，dx 或 dy 為 200、另一個為 0），兩條路徑輸出相同（中點到兩端的整數部分都是 100，乘 3 除 2 得 150，0x140 − 150 = 0xAA、150 + 0x140 = 0x1D6），所以 jle 與 jl 在這一點不可區分（等價突變）。
 *
 * 輸入域：基底 fcmd-base.state（被呼叫的 39B5 C 版寫死進入點 SP ＝ 1F8A 的堆疊位址）；A 取艦隊表、B 取行星表（seeds 另給特殊位置）。word 讀取在位移 0xFFFF 時
 *   dosgolem 的高位元組取同一段的位移 0（rd16s）。輸出全在 DGROUP，不在 ROM 區。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3310_0006(Mem *m, uint16_t x1, uint16_t y1, uint16_t x2, uint16_t y2);
extern uint16_t f_3CC9_00B9(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0070(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

static uint16_t f321F_0392_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f321F_0392_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

#define F0392_ZF 0x0040u
#define F0392_SF 0x0080u
#define F0392_OF 0x0800u

/* (ip × 3) ÷ 2 的低 16 位元：3CC9:0009 乘，3CC9:0070 有號除，只取商的低字 */
static uint16_t f321F_0392_scale(uint16_t ip)
{
    uint32_t prod = f_3CC9_0009(0, ip, 0, 3);
    uint64_t q = f_3CC9_0070((uint16_t)(prod >> 16), (uint16_t)prod, 0, 2);

    return (uint16_t)(q >> 32);
}

void f_321F_0392(Mem *m, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg)
{
    uint16_t ship = (uint16_t)(f321F_0392_rd16s(m, GIN3_DS, 0xCFB4u)
                               + (uint16_t)(f321F_0392_rd16s(m, a_seg, (uint16_t)(a_off + 0x10u)) * 0x4Eu));
    uint16_t ship_seg = f321F_0392_rd16s(m, GIN3_DS, 0xCFB6u);
    uint16_t mid_x, mid_y, sh, d_hi, ang;
    uint32_t d;
    uint16_t fl;

    if (rd8(m, LIN(a_seg, (uint16_t)(a_off + 1u))) == 0) {
        f321F_0392_wr16s(m, GIN3_DS, 0xD7EEu, f321F_0392_rd16s(m, ship_seg, (uint16_t)(ship + 0x34u)));
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F0u, f321F_0392_rd16s(m, ship_seg, (uint16_t)(ship + 0x36u)));
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F2u, f321F_0392_rd16s(m, b_seg, (uint16_t)(b_off + 0x46u)));
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F4u, f321F_0392_rd16s(m, b_seg, (uint16_t)(b_off + 0x48u)));
    } else {
        f321F_0392_wr16s(m, GIN3_DS, 0xD7EEu, f321F_0392_rd16s(m, b_seg, (uint16_t)(b_off + 0x46u)));
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F0u, f321F_0392_rd16s(m, b_seg, (uint16_t)(b_off + 0x48u)));
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F2u, f321F_0392_rd16s(m, ship_seg, (uint16_t)(ship + 0x34u)));
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F4u, f321F_0392_rd16s(m, ship_seg, (uint16_t)(ship + 0x36u)));
    }
    ang = f_2BE1_011D(m,
                      (uint16_t)(f321F_0392_rd16s(m, GIN3_DS, 0xD7F2u) - f321F_0392_rd16s(m, GIN3_DS, 0xD7EEu)),
                      (uint16_t)(f321F_0392_rd16s(m, GIN3_DS, 0xD7F4u) - f321F_0392_rd16s(m, GIN3_DS, 0xD7F0u)));
    wr8(m, DS(0xD7EAu), (uint8_t)(0u - ang));
    mid_x = (uint16_t)((uint16_t)(f321F_0392_rd16s(m, GIN3_DS, 0xD7EEu) + f321F_0392_rd16s(m, GIN3_DS, 0xD7F2u)) >> 1);
    mid_y = (uint16_t)((uint16_t)(f321F_0392_rd16s(m, GIN3_DS, 0xD7F0u) + f321F_0392_rd16s(m, GIN3_DS, 0xD7F4u)) >> 1);
    d = f_3310_0006(m,
                    f321F_0392_rd16s(m, GIN3_DS, 0xD7EEu), f321F_0392_rd16s(m, GIN3_DS, 0xD7F0u),
                    f321F_0392_rd16s(m, GIN3_DS, 0xD7F2u), f321F_0392_rd16s(m, GIN3_DS, 0xD7F4u));
    fl = f_3CC9_00B9((uint16_t)(d >> 16), (uint16_t)d, 0x00C8u, 0x0000u);
    if ((fl & F0392_ZF) || (((fl & F0392_SF) != 0) != ((fl & F0392_OF) != 0))) { /* jle */
        d = f_3310_0006(m, mid_x, mid_y, f321F_0392_rd16s(m, GIN3_DS, 0xD7EEu), f321F_0392_rd16s(m, GIN3_DS, 0xD7F0u));
        d_hi = (uint16_t)(d >> 16);
        sh = f321F_0392_scale(d_hi);
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F6u, (uint16_t)(0x0140u - sh));
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F8u, 0x012Au);
        d = f_3310_0006(m, mid_x, mid_y, f321F_0392_rd16s(m, GIN3_DS, 0xD7F2u), f321F_0392_rd16s(m, GIN3_DS, 0xD7F4u));
        d_hi = (uint16_t)(d >> 16);
        sh = f321F_0392_scale(d_hi);
        f321F_0392_wr16s(m, GIN3_DS, 0xD7FAu, (uint16_t)(sh + 0x0140u));
        f321F_0392_wr16s(m, GIN3_DS, 0xD7FCu, 0x012Au);
    } else {
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F6u, 0x00AAu);
        f321F_0392_wr16s(m, GIN3_DS, 0xD7F8u, 0x012Au);
        f321F_0392_wr16s(m, GIN3_DS, 0xD7FAu, 0x01D6u);
        f321F_0392_wr16s(m, GIN3_DS, 0xD7FCu, 0x012Au);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_321F_0392(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("321F_0392", 4, run)
