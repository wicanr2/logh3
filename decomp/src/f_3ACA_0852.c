/*
 * GIN3PS.EXE 3ACA:0852  從引數清單讀出下一個整數引數並前進（printf 的取引數，類似 va_arg）
 *
 * 呼叫：far，三個堆疊字：
 *   [bp+0Eh, 10h]  ap     far 指標，指向「引數清單游標」這個 far 指標變數（slot）：
 *                         slot+0 word＝目前引數的位移 p_off，slot+2 word＝目前引數的段 p_seg
 *   [bp+12h]       flags  整個 word 參與測試（test di,10h／test di,200h），其他位元不看
 * 行為（反組譯逐指令）：
 *   flags & 0x10（long）：value ＝ 32 位元 [p_seg:p_off]（低字在低位址，hi ＝ [p+2]、lo ＝ [p]）；
 *                          slot 的位移字 += 4（16 位元回繞，不進位到段），flags 的 0x200 在這條路徑上不看。
 *   否則（int）：         w ＝ 16 位元 [p_seg:p_off]；cwd 符號延伸成 hi:lo ＝ (w 的 bit15 ? 0xFFFF : 0):w；
 *                          flags & 0x200（unsigned）時 hi ← 0（lo 不變）；slot 的位移字 += 2。
 *                          原版這條路徑上有 `and byte ptr [bp+0],0FFh`，對低位元組是空操作（旗標以外沒有效果）。
 * 回傳：AX ＝ hi、BX ＝ lo（AX:BX，不是 DX:AX）。呼叫者 3ACA:0079（printf 本體）的 3 處。
 * 暫存器：回傳 AX、BX；CX、DX、SI、DI、ES 為殘值，不屬語意輸出。
 *
 * 機器模型（段內回繞）：每個 word 存取的兩個 byte 的位移各自 16 位元回繞（位移 0xFFFF 的高位元組取同段位移 0），
 *   `les si,es:[bx]` 的段字在 bx+2、`mov cx,es:[si+2]` 的位移運算也是 16 位元，皆以 rd16s／wr16s 重現。
 * 輸入域：slot 位於 RAM 且不在視訊記憶體（A0000 到 BFFFF）與 ROM（F0000 以上）：唯一的寫入就是 slot 的位移字，
 *   所以不需要 ROM 寫入模型；p 指向的位置只讀，由記憶體內容決定（隨機突變會造出任意 p，讀到視訊記憶體區的案例由探針丟棄）。
 *   slot 不在 SS 的堆疊區內。flags 與 slot 位址皆為全部 16 位元值（描述檔取代表值）。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）；「printf 取引數」是用途判斷（強推論：呼叫者是 printf 本體，
 *   flags 0x10／0x200 吻合 long／unsigned 修飾）。
 */
#include "gin3_reg.h"

/* 段內存取：兩個位元組的位移各自 16 位元回繞（dosgolem 的 read16／write16） */
static uint16_t f_3ACA_0852_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1))) << 8));
}

static void f_3ACA_0852_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1)), (uint8_t)(v >> 8));
}

/* 回傳 (AX<<16)|BX ＝ (hi<<16)|lo */
uint32_t f_3ACA_0852(Mem *m, uint16_t ap_off, uint16_t ap_seg, uint16_t flags)
{
    uint16_t p_off, p_seg, lo, hi;

    p_off = f_3ACA_0852_rd16s(m, ap_seg, ap_off);
    p_seg = f_3ACA_0852_rd16s(m, ap_seg, (uint16_t)(ap_off + 2));
    if (flags & 0x10) {
        hi = f_3ACA_0852_rd16s(m, p_seg, (uint16_t)(p_off + 2));
        lo = f_3ACA_0852_rd16s(m, p_seg, p_off);
        f_3ACA_0852_wr16s(m, ap_seg, ap_off, (uint16_t)(f_3ACA_0852_rd16s(m, ap_seg, ap_off) + 4));
    } else {
        lo = f_3ACA_0852_rd16s(m, p_seg, p_off);
        hi = (lo & 0x8000u) ? 0xFFFFu : 0u; /* cwd */
        if (flags & 0x200) {
            /* and byte [bp+0],0FFh：對 lo 的低位元組是空操作，lo 整個不變 */
            hi = 0;
        }
        f_3ACA_0852_wr16s(m, ap_seg, ap_off, (uint16_t)(f_3ACA_0852_rd16s(m, ap_seg, ap_off) + 2));
    }
    return ((uint32_t)hi << 16) | lo;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3ACA_0852(m, a[0], a[1], a[2]); }
GIN3_REGISTER("3ACA_0852", 3, run)
