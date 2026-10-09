/*
 * GIN3PS.EXE 2376:02DE  畫面格表矩形操作的包裝：由 ES:SI 指的物件取出鍵值與矩形，呼叫 2376:00B2（near，暫存器傳參）
 *
 * 輸入（暫存器）：ES:SI ＝ 顯示物件；AL ＝ 進入值（不使用，AH 被覆蓋）；DS ＝ DGROUP。
 * 原版（20 bytes，逐指令）：
 *   mov ah, es:[si+2]             ; 鍵值 k
 *   mov bx, es:[si+0x24]            ; word：BL ＝ [si+0x24]、BH ＝ [si+0x25]
 *   mov dx, es:[si+0x26]            ; word：DL ＝ [si+0x26]、DH ＝ [si+0x27]
 *   xchg bh,bl                    ; BH ＝ [si+0x24]（欄）、BL ＝ [si+0x25]（列）
 *   xchg dh,dl                    ; DH ＝ [si+0x26]（每列格數）、DL ＝ [si+0x27]（列數）
 *   call 2376:00B2                  ; 見該檔（已 diff-tested）
 *   retn
 *   返回時 AX、BX、DX、旗標是 00B2 的殘值；呼叫者之後只用記憶體，描述檔 ret ＝ none，只比對記憶體寫入。
 * 欄位名稱（欄、列、每列格數、列數）取自 00B2 檔頭（強推論：2376:0150 的 CL ＝ 列、CH ＝ 欄）；物件其他欄位語意：未知。
 * 呼叫者（台帳）：2376:0164、2376:0364（皆 game-logic）。00B2 與 00E4 成對（見各自檔頭）。
 *
 * 機器模型：兩個 word 讀取在位移 0xFFFF 時，高位元組取同段位移 0（dosgolem 的 read16 是兩次 byte 存取，位移各自在段內回繞），
 *   本檔用 f02DE_rd16s 重現；其餘位移加法都是 16 位元回繞。物件的讀取全在呼叫 00B2 之前，物件與畫面格表別名也不影響。
 * 輸入域：ES、SI 的全部 16 位元值（指到 A0000 到 BFFFF 的案例被探針丟棄）；列數或每列格數為 0 時原版繞 65536 圈，
 *   超過步數上限被探針丟棄，不在比較範圍。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位名稱為強推論。
 */
#include "gin3_reg.h"

extern void f_2376_00B2(Mem *m, uint16_t ax, uint16_t bx, uint16_t dx);

/* 段內 word 讀取：seg:FFFF 的高位元組在 seg:0000 */
static uint16_t f02DE_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

static uint16_t f02DE_xchg(uint16_t w) { return (uint16_t)((w >> 8) | (w << 8)); }

void f_2376_02DE(Mem *m, uint16_t ax, uint16_t es, uint16_t si)
{
    uint16_t bx, dx;

    ax = (uint16_t)((ax & 0x00FFu) | ((uint16_t)rd8(m, LIN(es, (uint16_t)(si + 2u))) << 8));
    bx = f02DE_xchg(f02DE_rd16s(m, es, (uint16_t)(si + 0x24u)));
    dx = f02DE_xchg(f02DE_rd16s(m, es, (uint16_t)(si + 0x26u)));
    f_2376_00B2(m, ax, bx, dx);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2376_02DE(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("2376_02DE", 3, run)
