/*
 * 34CB:06D8  視窗內容的更新顯示：只把「內容與畫面格不同」的本視窗格子更新（字碼低位元組）
 *
 * 原版：far 呼叫，參數是視窗結構的 far 指標（位移 [bp+22h]、段 [bp+24h]；序言 push bp／si／di、sub sp,18h）。
 *   唯一呼叫者 341B:02E5 內的 sub_41005（+0x54）。回傳值沒人用，描述檔 ret = none。
 *
 * 讀視窗結構（docs/internals/47-window-frame-tiles.md「結構位移」；名稱是本檔取的）：
 *   +4 槽位編號  +5 格 x  +6 格 y  +7 格寬  +8 格高（皆 byte）
 *   +9 內容起點的補正（byte，強推論：內容 x 位移）  +A、+B：+A × +B 加進內容起點（強推論：內容 y 位移 × 內容格寬）
 *   +B 同時是內容每列前進的 byte 數（內容格寬）
 *   +E 內容緩衝區位移（word）  +10 內容緩衝區段（word）
 *
 * 動作（已證實：反組譯逐條對過）：
 *   格起點   = D81E + (+6) × 0xF0 + (+5) × 6（16 位元回繞），段 = 呼叫時的 DS（GIN3_DS）。
 *   內容起點 = (+10):( (+E) + (+9) + (+A) × (+B) )（位移 16 位元回繞）。
 *   外層 i = 0 起，i < (+8)（每一輪重讀 +8）：
 *       格位移 = 列起點；內容位移 = 內容列起點；
 *       內層 j = 0 起，j < (+7)（每一輪重讀 +7）：
 *           若 (格[+0] & 0x0F) == (+4)（擁有者符合；cl = (+4) 整個 byte，每格重讀）：
 *               若 格[+2] != 內容 byte：
 *                   格[+0] = (+4) | 0x80；格[+1] = 1；格[+2] = 內容 byte（重讀）；格[+3] = 0
 *           格位移 += 6；內容位移 += 1
 *       列起點 += 0xF0；內容列起點 += (+B)（重讀）
 *   與 34CB:00AB 的差別：不看鎖住位元（0x20），改看擁有者；而且只在字碼低位元組真的不同時才寫（不同才標髒）。
 *   寫入順序：格[+0]、格[+1]、（重讀內容 byte）、格[+2]、格[+3]。內容緩衝區與畫面格表重疊時順序有影響，照原版。
 *
 * 語意（強推論）：顏色 0 視窗內容更新的差量版：只重畫變動的字，不動鎖住與別人的格子。
 *   呼叫者 341B:02E5 的用途（視窗內容改動後的重畫）未在本檔驗證：未知。
 * 輸入域：視窗結構任意 byte 值，位移全部 16 位元回繞。寬 255、高 255 時超過探針步數上限的案例被丟棄；
 *   視窗指標或內容緩衝區段指到 A0000 到 BFFFF（視訊記憶體）的案例被探針丟棄。
 */
#include "gin3_reg.h"

static uint8_t f34CB06D8_wb(const Mem *m, uint16_t seg, uint16_t off, uint16_t n)
{
    return rd8(m, LIN(seg, (uint16_t)(off + n)));
}

/* 段內 word 讀取：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go） */
static uint16_t f34CB06D8_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_34CB_06D8(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    const uint16_t ds = GIN3_DS;
    uint16_t cur = (uint16_t)(0xD81Eu + (uint16_t)(f34CB06D8_wb(m, w_seg, w_off, 6) * 0xF0u)
                              + (uint16_t)(f34CB06D8_wb(m, w_seg, w_off, 5) * 6u));
    const uint16_t cseg = f34CB06D8_rd16s(m, w_seg, (uint16_t)(w_off + 0x10u));
    uint16_t crow = (uint16_t)(f34CB06D8_rd16s(m, w_seg, (uint16_t)(w_off + 0x0Eu))
                               + f34CB06D8_wb(m, w_seg, w_off, 9)
                               + (uint16_t)(f34CB06D8_wb(m, w_seg, w_off, 0x0A) * f34CB06D8_wb(m, w_seg, w_off, 0x0B)));
    uint8_t i = 0;

    while (i < f34CB06D8_wb(m, w_seg, w_off, 8)) {
        uint16_t cell = cur;
        uint16_t cptr = crow;
        uint8_t j = 0;

        while (j < f34CB06D8_wb(m, w_seg, w_off, 7)) {
            const uint8_t owner = f34CB06D8_wb(m, w_seg, w_off, 4);

            if ((rd8(m, LIN(ds, cell)) & 0x0Fu) == owner
                && rd8(m, LIN(ds, (uint16_t)(cell + 2u))) != rd8(m, LIN(cseg, cptr))) {
                wr8(m, LIN(ds, cell), (uint8_t)(owner | 0x80u));
                wr8(m, LIN(ds, (uint16_t)(cell + 1u)), 1);
                wr8(m, LIN(ds, (uint16_t)(cell + 2u)), rd8(m, LIN(cseg, cptr)));
                wr8(m, LIN(ds, (uint16_t)(cell + 3u)), 0);
            }
            cell = (uint16_t)(cell + 6u);
            cptr = (uint16_t)(cptr + 1u);
            j++;
        }
        cur = (uint16_t)(cur + 0xF0u);
        crow = (uint16_t)(crow + f34CB06D8_wb(m, w_seg, w_off, 0x0B));
        i++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_34CB_06D8(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("34CB_06D8", 2, run)
