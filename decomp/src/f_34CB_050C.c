/*
 * GIN3PS.EXE 34CB:050C  視窗內容（彩色，一格 4 bytes）搬進畫面格表：有外框時先畫外框（34CB:0BA0），再把內容緩衝區複製進已鎖住（bit5 ＝ 1）的格子
 *
 * 原版：far 呼叫，參數是視窗結構的 far 指標（位移 [bp+1Ah]、段 [bp+1Ch]；序言 push bp／si／di、sub sp,10h）。
 *   唯一呼叫者 341B:01FA。回傳值沒人用，描述檔 ret = none。與 34CB:01A4 互為鏡像：34CB:01A4 只處理沒鎖住的格子，並呼叫另一支外框函式，其餘逐指令相同。
 *
 * 讀視窗結構（docs/internals/47-window-frame-tiles.md「結構位移」；欄位名稱是強推論）：
 *   +3 外框樣式起始碼（0 ＝ 不畫框）  +4 槽位編號（擁有者）  +5 格 x  +6 格 y  +B 內容格寬  +C 內容格高（皆 byte）
 *   +E 內容緩衝區位移（word）  +10 內容緩衝區段（word）；內容緩衝區一格 4 bytes：字碼 word（內容 +0／+1）、顏色 byte（+2）、顏色高位 byte（+3）。
 * 畫面格表：ds:D81E 起，40 欄 × 20 列、一格 6 bytes、一列 0xF0 bytes（docs/internals/31、47）。
 *
 * 動作（已證實：反組譯逐條對過）：
 *   1. (+3) ≠ 0（有外框）：格起點 cur ＝ D81E ＋ ((+6)＋1) × 0xF0 ＋ ((+5)＋1) × 6（16 位元回繞；加 1 是 16 位元，最大 256），
 *      並呼叫 34CB:0BA0(win) 畫外框（cur 在呼叫前算好）；否則 cur ＝ D81E ＋ (+6) × 0xF0 ＋ (+5) × 6，不畫外框。
 *      段是呼叫時的 DS（GIN3_DS）。
 *   2. 外框之後才讀內容指標 (+10):(+E)、cw ＝ (+B)、ch ＝ (+C)（快取一次）。
 *   3. 外層 i ＝ 0 起，i < ch（無號 byte）：列起點 ＝ cur；cur ＋＝ 0xF0；內層 j ＝ 0 起，j < cw：
 *        若格[+0] 的 bit5 != 0：
 *            格[+0] ＝ (+4) | 0x90          （(+4) 每格重讀；0x80 ＝ 髒、0x10 ＝ 走字模、低位元組 ＝ 擁有者，強推論）
 *            格 word[+2] ＝ 內容 word[+0]    （內容每格重讀）
 *            格[+4] ＝ 內容[+2]；格[+5] ＝ 內容[+3]
 *        格 ＋＝ 6；內容 ＋＝ 4（兩者位移 16 位元回繞）
 *   內容指標不在每列重設，一路連續前進（內容緩衝區的列寬等於 cw 才對得上；本函式不檢查）。格[+1]、格[+3] 不碰。
 *   與 34CB:07EE 的差別：07EE 只更新擁有者符合且內容有變的格，並以差量方式標髒；本函式無條件覆寫（已鎖住（bit5 ＝ 1）的格子）。
 *   原版逐步讀寫記憶體：視窗結構、內容緩衝區可與畫面格表別名，後面的讀取看得到前面的寫入（例如 (+4) 每格重讀），本版照原順序。
 *
 * 輸入域：視窗結構任意 byte 值，位移全部 16 位元回繞，word 存取在位移 0xFFFF 時高位元組取同一段的位移 0。
 *   外框函式 0BA0 的域外情形（寬或高為 0 時走訪整個資料段而可能不返回）同樣不涵蓋；視窗指標或內容緩衝區段指到 A0000 到 BFFFF
 *   的案例被探針丟棄；內容緩衝區蓋到呼叫堆疊（SS:1B8A..1F89）的情形不涵蓋。
 * 呼叫的函式：34CB:0BA0（已 diff-tested）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位名稱為強推論。
 */
#include "gin3_reg.h"

extern void f_34CB_0BA0(Mem *m, uint16_t w_off, uint16_t w_seg);

static uint8_t f34CB050C_wb(const Mem *m, uint16_t seg, uint16_t off, uint16_t n)
{
    return rd8(m, LIN(seg, (uint16_t)(off + n)));
}

/* 段內 word 讀寫：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go） */
static uint16_t f34CB050C_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f34CB050C_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_34CB_050C(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    const uint16_t ds = GIN3_DS;
    uint16_t cur;
    uint16_t cseg, coff;
    uint8_t cw, ch;
    uint8_t i = 0;

    if (f34CB050C_wb(m, w_seg, w_off, 3) != 0) {
        cur = (uint16_t)(0xD81Eu + (uint16_t)((f34CB050C_wb(m, w_seg, w_off, 6) + 1u) * 0xF0u)
                         + (uint16_t)((f34CB050C_wb(m, w_seg, w_off, 5) + 1u) * 6u));
        f_34CB_0BA0(m, w_off, w_seg);
    } else {
        cur = (uint16_t)(0xD81Eu + (uint16_t)(f34CB050C_wb(m, w_seg, w_off, 6) * 0xF0u)
                         + (uint16_t)(f34CB050C_wb(m, w_seg, w_off, 5) * 6u));
    }
    cseg = f34CB050C_rd16s(m, w_seg, (uint16_t)(w_off + 0x10u));
    coff = f34CB050C_rd16s(m, w_seg, (uint16_t)(w_off + 0x0Eu));
    cw = f34CB050C_wb(m, w_seg, w_off, 0x0B);
    ch = f34CB050C_wb(m, w_seg, w_off, 0x0C);

    while (i < ch) {
        uint16_t cell = cur;
        uint8_t j = 0;

        cur = (uint16_t)(cur + 0xF0u);
        while (j < cw) {
            if ((rd8(m, LIN(ds, cell)) & 0x20u) != 0) {
                wr8(m, LIN(ds, cell), (uint8_t)(f34CB050C_wb(m, w_seg, w_off, 4) | 0x90u));
                f34CB050C_wr16s(m, ds, (uint16_t)(cell + 2u), f34CB050C_rd16s(m, cseg, coff));
                wr8(m, LIN(ds, (uint16_t)(cell + 4u)), rd8(m, LIN(cseg, (uint16_t)(coff + 2u))));
                wr8(m, LIN(ds, (uint16_t)(cell + 5u)), rd8(m, LIN(cseg, (uint16_t)(coff + 3u))));
            }
            cell = (uint16_t)(cell + 6u);
            coff = (uint16_t)(coff + 4u);
            j++;
        }
        i++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_34CB_050C(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("34CB_050C", 2, run)
