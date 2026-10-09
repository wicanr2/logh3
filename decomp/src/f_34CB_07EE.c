/*
 * 34CB:07EE  視窗內容（彩色，一格 4 bytes）的更新顯示：只把「內容與畫面格不同」的本視窗格子更新
 *
 * 原版：far 呼叫，參數是視窗結構的 far 指標（位移 [bp+1Ch]、段 [bp+1Eh]；序言 push bp／si／di、sub sp,12h）。
 *   唯一呼叫者 341B:02E5 內的 sub_41005（+0x45）。回傳值沒人用，描述檔 ret = none。
 *   docs/internals/31-text-plane.md 開頭記「視窗元件各自有一份小的字元緩衝區，34CB:07EE 把它疊進這塊全螢幕平面」。
 *
 * 讀視窗結構（docs/internals/47-window-frame-tiles.md「結構位移」；名稱是本檔取的）：
 *   +3 外框樣式起始碼（0 = 不畫框）  +4 槽位編號  +5 格 x  +6 格 y（byte）
 *   +B 內容格寬  +C 內容格高（byte；docs/internals/47 記為「內容格寬／格高」）
 *   +E 內容緩衝區位移（word）  +10 內容緩衝區段（word）；內容緩衝區一格 4 bytes：
 *       字碼 word（內容 +0／+1）、顏色 byte（+2）、顏色高位 byte（+3）（docs/internals/47、docs/re/138）
 *
 * 動作（已證實：反組譯逐條對過）：
 *   格起點 = D81E + (行 y') × 0xF0 + (欄 x') × 6（16 位元回繞），段 = 呼叫時的 DS（GIN3_DS）；
 *       (+3) != 0（有外框）時 x' = (+5) + 1、y' = (+6) + 1（內容在外框內縮一格；加 1 是 16 位元，x' 最大 256），
 *       否則 x' = (+5)、y' = (+6)。
 *   內容指標 = (+10):(+E)。cw = (+B)、ch = (+C) 在進函式時讀一次（快取）。
 *   外層 i = 0 起，i < ch：格位移 = 列起點；
 *       內層 j = 0 起，j < cw：
 *           若 (格[+0] & 0x0F) == (+4)（每格重讀 +4）：
 *               若 格 word[+2] != 內容 word[+0]，或 格[+4] != 內容[+2]，或 格[+5] != 內容[+3]：
 *                   格[+0] |= 0x80；格 word[+2] = 內容 word[+0]（重讀）；格[+4] = 內容[+2]；格[+5] = 內容[+3]
 *           格位移 += 6；內容指標 += 4
 *       列起點 += 0xF0
 *   內容指標不在每列重設，一路連續前進（內容緩衝區的列寬等於 cw 才對得上；本函式不檢查）。
 *   與 34CB:06D8 的差別：本函式是彩色版（比較與更新字碼 word、兩個顏色 byte），不碰格[+1]、格[+3] 的獨立欄位。
 *
 * 語意（強推論）：外框內縮一格的位置換算與 docs/re/138 的「內容格由 34CB:01A4 搬進來」相符；
 *   本函式是內容改動後的差量重畫版（只在不同時才標髒）。呼叫者 341B:02E5 的完整用途未在本檔驗證：未知。
 * 輸入域：視窗結構任意 byte 值，位移全部 16 位元回繞，word 存取在位移 0xFFFF 時高位元組取同一段的位移 0。
 *   寬 255、高 255 時超過探針步數上限的案例被丟棄；視窗指標或內容緩衝區段指到 A0000 到 BFFFF（視訊記憶體）的案例被探針丟棄。
 */
#include "gin3_reg.h"

static uint8_t f34CB07EE_wb(const Mem *m, uint16_t seg, uint16_t off, uint16_t n)
{
    return rd8(m, LIN(seg, (uint16_t)(off + n)));
}

/* 段內 word 讀寫：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go） */
static uint16_t f34CB07EE_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static void f34CB07EE_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_34CB_07EE(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    const uint16_t ds = GIN3_DS;
    uint16_t cur;
    uint16_t cseg;
    uint16_t coff;
    uint8_t cw, ch;
    uint8_t i = 0;

    if (f34CB07EE_wb(m, w_seg, w_off, 3) != 0)
        cur = (uint16_t)(0xD81Eu + (uint16_t)((f34CB07EE_wb(m, w_seg, w_off, 5) + 1u) * 6u)
                         + (uint16_t)((f34CB07EE_wb(m, w_seg, w_off, 6) + 1u) * 0xF0u));
    else
        cur = (uint16_t)(0xD81Eu + (uint16_t)(f34CB07EE_wb(m, w_seg, w_off, 5) * 6u)
                         + (uint16_t)(f34CB07EE_wb(m, w_seg, w_off, 6) * 0xF0u));
    cseg = f34CB07EE_rd16s(m, w_seg, (uint16_t)(w_off + 0x10u));
    coff = f34CB07EE_rd16s(m, w_seg, (uint16_t)(w_off + 0x0Eu));
    cw = f34CB07EE_wb(m, w_seg, w_off, 0x0B);
    ch = f34CB07EE_wb(m, w_seg, w_off, 0x0C);

    while (i < ch) {
        uint16_t cell = cur;
        uint8_t j = 0;

        while (j < cw) {
            if ((rd8(m, LIN(ds, cell)) & 0x0Fu) == f34CB07EE_wb(m, w_seg, w_off, 4)) {
                const uint16_t sip = (uint16_t)(cell + 2u);

                if (f34CB07EE_rd16s(m, ds, sip) != f34CB07EE_rd16s(m, cseg, coff)
                    || rd8(m, LIN(ds, (uint16_t)(cell + 4u))) != rd8(m, LIN(cseg, (uint16_t)(coff + 2u)))
                    || rd8(m, LIN(ds, (uint16_t)(cell + 5u))) != rd8(m, LIN(cseg, (uint16_t)(coff + 3u)))) {
                    wr8(m, LIN(ds, cell), (uint8_t)(rd8(m, LIN(ds, cell)) | 0x80u));
                    f34CB07EE_wr16s(m, ds, sip, f34CB07EE_rd16s(m, cseg, coff));
                    wr8(m, LIN(ds, (uint16_t)(cell + 4u)), rd8(m, LIN(cseg, (uint16_t)(coff + 2u))));
                    wr8(m, LIN(ds, (uint16_t)(cell + 5u)), rd8(m, LIN(cseg, (uint16_t)(coff + 3u))));
                }
            }
            cell = (uint16_t)(cell + 6u);
            coff = (uint16_t)(coff + 4u);
            j++;
        }
        cur = (uint16_t)(cur + 0xF0u);
        i++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_34CB_07EE(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("34CB_07EE", 2, run)
