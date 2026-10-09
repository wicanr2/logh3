/*
 * GIN3PS.EXE 3ACA:08B6  把整數寫成 "+ddd"／"-ddd"（正負號加 3 位十進位數字）到輸出游標，並前進游標 4 個位元組
 *
 * 呼叫：far，兩個堆疊字：
 *   [bp+0Eh, 10h]  cur_p  far 指標，指向「輸出游標」這個 far 指標變數（slot）：slot+0 word＝游標位移、slot+2 word＝游標段
 *   [bp+12h]       v      16 位元整數（整個 word）
 * 行為（反組譯逐指令）：
 *   1. 把 slot 的段、位移抄進區域變數 cseg、coff。
 *   2. v 是負數（有號：or ax,ax; jns）就取二補數（neg；v＝0x8000 取負後還是 0x8000）並用 '-'，否則 '+'。寫 1 個位元組到 cseg:coff，coff 加 1（16 位元回繞）。
 *   3. 有號除法（idiv，向零取整）：d1 ＝ (v ÷ 100 的低位元組) + '0'；r ＝ v mod 100；d2 ＝ (r ÷ 10 的低位元組) + '0'；d3 ＝ (r mod 10 的低位元組) + '0'。
 *      依序各寫 1 個位元組到 cseg:coff，coff 每次加 1（16 位元回繞，不進位到段）。
 *      v＝0x8000（取負後仍是有號 −32768）時商與餘數是負的：商 −327（低位元組 0xB9）→ d1＝0xE9；餘數 −68 → 商 −6（0xFA）→ d2＝0x2A、
 *      餘數 −8（0xF8）→ d3＝0x28。原版照算不報錯。
 *      v 的絕對值大於 999 時 d1 超過 '9'（例如 5000 → 商 50 → 'b'），原版照寫。
 *   4. 最後寫回 slot：先 slot+2 ← cseg（原值），再 slot+0 ← coff（新位移，等於原位移加 4）。
 * 沒有回傳值（AX 為殘值，DX 為最後一次除法的餘數 + '0'，不屬於語意輸出）。呼叫者 3ACA:0079（printf 本體），1 處。
 * 用途判斷（強推論）：printf 浮點 %e 格式的指數 "e+ddd"（正負號加 3 位數）；呼叫者上下文本檔未查，因此仍算強推論。
 *
 * 機器模型：
 *   - 每個 word 讀寫的兩個 byte 位移各自 16 位元回繞（位移 0xFFFF 的高位元組取同段位移 0）。
 *   - 線性位址 0xF0000 以上（dosgolem 的 BIOS ROM 區）的寫入被安靜丟棄；游標由記憶體內容決定，可能指到那裡，
 *     所以每個 byte 寫入都經過本檔的 f_3ACA_08B6_wr8m（ROM 區無效）。這是機器的性質，不是原版函式的邏輯。
 *   - 線性位址遮到 20 位元（LIN 的行為）；FFFF:0010 以上回繞到 0。
 * 輸入域：slot 與游標寫入範圍不重疊函式自己的堆疊框架（SS:1B8A..1F89 附近）與函式本體的程式碼；
 *   寫入視訊記憶體（A0000 到 BFFFF）的案例由探針丟棄。其餘 slot 位址、游標（含別名：游標指到 slot 自己）與 v 皆為全部 16 位元值。
 *   別名（游標落在 slot 內）時原版先寫 4 個字元、最後才寫 slot，本版照順序做。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

static uint16_t f_3ACA_08B6_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1))) << 8));
}

/* 單一 byte 寫入；線性 0xF0000 以上（ROM 區）無效 */
static void f_3ACA_08B6_wr8m(Mem *m, uint32_t lin, uint8_t v)
{
    lin &= 0xFFFFFu;
    if (lin < 0xF0000u)
        m->b[lin] = v;
}

static void f_3ACA_08B6_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f_3ACA_08B6_wr8m(m, LIN(seg, off), (uint8_t)v);
    f_3ACA_08B6_wr8m(m, LIN(seg, (uint16_t)(off + 1)), (uint8_t)(v >> 8));
}

void f_3ACA_08B6(Mem *m, uint16_t cp_off, uint16_t cp_seg, uint16_t value)
{
    uint16_t cseg, coff, nv;
    int q100, r100, q10, r10;
    uint8_t c[4];
    int i;

    coff = f_3ACA_08B6_rd16s(m, cp_seg, cp_off);
    cseg = f_3ACA_08B6_rd16s(m, cp_seg, (uint16_t)(cp_off + 2));
    nv = value;
    if ((int16_t)value < 0) {
        nv = (uint16_t)(0u - value); /* neg ax：0x8000 取負後仍是 0x8000 */
        c[0] = '-';
    } else {
        c[0] = '+';
    }
    q100 = (int)(int16_t)nv / 100; /* idiv：有號、向零取整 */
    r100 = (int)(int16_t)nv % 100;
    q10 = r100 / 10;
    r10 = r100 % 10;
    c[1] = (uint8_t)(q100 + 0x30);
    c[2] = (uint8_t)(q10 + 0x30);
    c[3] = (uint8_t)(r10 + 0x30);
    for (i = 0; i < 4; i++) {
        f_3ACA_08B6_wr8m(m, LIN(cseg, (uint16_t)(coff + i)), c[i]);
    }
    coff = (uint16_t)(coff + 4);
    f_3ACA_08B6_wr16s(m, cp_seg, (uint16_t)(cp_off + 2), cseg);
    f_3ACA_08B6_wr16s(m, cp_seg, cp_off, coff);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3ACA_08B6(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("3ACA_08B6", 3, run)
