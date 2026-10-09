/*
 * 34CB:00AB  視窗內容的顯示（顏色 0 那條路，docs/re/138）：把視窗內容緩衝區的字碼低位元組搬進沒鎖住的畫面格
 *
 * 原版：far 呼叫，參數是視窗結構的 far 指標（位移 [bp+1Eh]、段 [bp+20h]；序言 push bp／si／di、sub sp,14h）。
 *   唯一呼叫者 341B:0145 內的 sub_40E65（+0x5F），依視窗 +1（顏色）選路：docs/re/138 記「0xFF → 34CB:02C4、
 *   0 → 34CB:00AB、其餘 → 34CB:01A4」，本函式是 0 那條。回傳值沒人用，描述檔 ret = none。
 *
 * 讀視窗結構（docs/internals/47-window-frame-tiles.md「結構位移」；名稱是本檔取的）：
 *   +4  槽位編號  +5 格 x  +6 格 y  +7 格寬  +8 格高（皆 byte）
 *   +9  內容緩衝區的起始位移補正（byte）  +A、+B  兩個 byte，相乘是內容緩衝區的起始位移補正
 *   +B  內容格寬（也是內容每列前進的 byte 數）
 *   +E  內容緩衝區位移（word）  +10 內容緩衝區段（word）
 *
 * 動作（已證實：反組譯逐條對過）：
 *   格起點   = D81E + (+6) × 0xF0 + (+5) × 6（16 位元回繞），段 = 呼叫時的 DS（GIN3_DS）。
 *   內容起點 = (+10):( (+E) + (+9) + (+A) × (+B) )（位移 16 位元回繞）。
 *   外層 i = 0 起，i < h；h 是進函式時讀一次的 (+8)（快取在區域變數）：
 *       列起點 = 目前格位移；格位移 += 0xF0；內容列起點 = 目前內容位移；內容位移 += (+B)（重讀）。
 *       內層 j = 0 起，j < (+7)（每一輪都重讀）：
 *           若 (格[+0] & 0x20) == 0：
 *               格[+0] = (+4) | 0x80（髒 ＋ 擁有者）；格[+1] = 1；格[+2] = 內容緩衝區的 1 個 byte；格[+3] = 0
 *           格位移 += 6；內容位移 += 1
 *   寫入順序：格[+0]、格[+1]、（讀內容 byte）、格[+2]、格[+3]。內容緩衝區可能與畫面格表重疊，順序照原版。
 *   進函式時還存了一份 (+7) 到區域變數，之後沒有人讀（迴圈用的是重讀的 (+7)）。
 *
 * 語意（強推論）：+1 = 0 的視窗（顏色 0）把內容緩衝區當一格 1 byte 的純字元表，只更新字碼低位元組，
 *   高位元組（格[+3]）清 0、旗標[+1] 設 1（半形字，docs/internals/31 記 +1 bit0 = 半形字），
 *   屬性[+4] 與[+5] 不動。0x20 是鎖住位元（docs/re/138），鎖住的格子不覆寫。
 * 輸入域：視窗結構任意 byte 值，位移全部 16 位元回繞。寬 255、高 255 時超過探針步數上限的案例被丟棄；
 *   視窗指標或內容緩衝區段指到 A0000 到 BFFFF（視訊記憶體）的案例被探針丟棄。
 */
#include "gin3_reg.h"

static uint8_t f34CB00AB_wb(const Mem *m, uint16_t seg, uint16_t off, uint16_t n)
{
    return rd8(m, LIN(seg, (uint16_t)(off + n)));
}

/* 段內 word 讀取：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go） */
static uint16_t f34CB00AB_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_34CB_00AB(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    const uint16_t ds = GIN3_DS;
    uint16_t cur = (uint16_t)(0xD81Eu + (uint16_t)(f34CB00AB_wb(m, w_seg, w_off, 6) * 0xF0u)
                              + (uint16_t)(f34CB00AB_wb(m, w_seg, w_off, 5) * 6u));
    const uint8_t h = f34CB00AB_wb(m, w_seg, w_off, 8);
    const uint16_t cseg = f34CB00AB_rd16s(m, w_seg, (uint16_t)(w_off + 0x10u));
    uint16_t coff = (uint16_t)(f34CB00AB_rd16s(m, w_seg, (uint16_t)(w_off + 0x0Eu))
                               + f34CB00AB_wb(m, w_seg, w_off, 9)
                               + (uint16_t)(f34CB00AB_wb(m, w_seg, w_off, 0x0A) * f34CB00AB_wb(m, w_seg, w_off, 0x0B)));
    uint8_t i = 0;

    while (i < h) {
        uint16_t cell = cur;
        uint16_t cptr = coff;
        uint8_t j = 0;

        cur = (uint16_t)(cur + 0xF0u);
        coff = (uint16_t)(coff + f34CB00AB_wb(m, w_seg, w_off, 0x0B));
        while (j < f34CB00AB_wb(m, w_seg, w_off, 7)) {
            if ((rd8(m, LIN(ds, cell)) & 0x20u) == 0) {
                wr8(m, LIN(ds, cell), (uint8_t)(f34CB00AB_wb(m, w_seg, w_off, 4) | 0x80u));
                wr8(m, LIN(ds, (uint16_t)(cell + 1u)), 1);
                wr8(m, LIN(ds, (uint16_t)(cell + 2u)), rd8(m, LIN(cseg, cptr)));
                wr8(m, LIN(ds, (uint16_t)(cell + 3u)), 0);
            }
            cell = (uint16_t)(cell + 6u);
            cptr = (uint16_t)(cptr + 1u);
            j++;
        }
        i++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_34CB_00AB(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("34CB_00AB", 2, run)
