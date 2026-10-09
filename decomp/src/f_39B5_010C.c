/*
 * GIN3PS.EXE 39B5:010C  兩個 32 位元有號數（呼叫者框架內）就地取絕對值，回傳兩者符號的異或位（near，BP 傳參）
 *
 * 呼叫：near，沒有堆疊參數；暫存器 BP ＝ 呼叫者（39B5:000C／005C）的框架指標，兩個運算元在 SS:
 *   第一個 ＝ SS:[BP+6]（低字）、SS:[BP+8]（高字）；第二個 ＝ SS:[BP+0Ah]（低字）、SS:[BP+0Ch]（高字）。
 *   在呼叫者的框架裡就是它的四個堆疊參數 arg_0..arg_6（兩個 long，低字在低位址）。
 * 原版（26 bytes，逐指令）：
 *   mov ax,[bp+8]；xor ax,[bp+0Ch]；and ax,8000h        ; 符號異或 ＝ 兩個高字的 bit15 異或
 *   push ax；push bp
 *   add bp,6；call 39B5:0126                               ; 第一個運算元取絕對值（為負才取二補數）
 *   add bp,4；call 39B5:0126                               ; 第二個運算元取絕對值
 *   pop bp；pop ax；retn                                   ; AX ＝ 符號異或（0 或 0x8000）；BP 還原
 *   回傳 AX ＝ 0x8000 表示兩者異號（乘除結果要變號），否則 0。DX 與旗標是 0126 的殘值，不比對。
 * 呼叫的函式：39B5:0126（已 diff-tested，SS:[BP] 的 32 位元就地取絕對值）。
 * 呼叫者（台帳）：39B5:000C（80 bytes，疑為 32 位元乘法的前置）、39B5:005C。用途判斷（假說）：32 位元有號乘除的前置處理，
 *   符號異或之後由呼叫者決定要不要把結果取負。
 *
 * 機器模型：SS 固定是基底狀態的 0x63EB（diffprobe 不能設定 SS）；BP 與各位移的 16 位元加法回繞；word 存取在位移 0xFFFF 時，
 *   高位元組取同一段的位移 0（dosgolem 的 read16／write16 是兩次 byte 存取，位移各自在段內回繞），用本檔的 rd16s 重現
 *   （0126 內部也是）。SS:BP 的線性位址最高 0x73EAF，不會落入 ROM，也不是視訊記憶體。
 * 輸入域：BP 的 +6..+0xD 八個 byte（含回繞後的位置）不落在探針的堆疊帶（SS:1B8A..1F89，基底狀態 SP ＝ 1F8A 往下 0x400 bytes）。
 *   落在堆疊帶時，探針把該帶的寫入排除在寫入集合之外，而原版 push／call 又會改變讀到的值，C 版無法重現，不涵蓋。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern uint32_t f_39B5_0126(Mem *m, uint16_t ss, uint16_t bp, uint16_t ax, uint16_t dx);

#define F39B5_010C_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */

static uint16_t f39B5_010C_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_39B5_010C(Mem *m, uint16_t bp)
{
    const uint16_t ss = F39B5_010C_SS;
    uint16_t sign = (uint16_t)((f39B5_010C_rd16s(m, ss, (uint16_t)(bp + 8u))
                                ^ f39B5_010C_rd16s(m, ss, (uint16_t)(bp + 0x0Cu))) & 0x8000u);

    /* 兩次 call 之間 AX、DX 的值只被 0126 在「為負」時覆蓋再寫回，不影響記憶體結果；DX 的進入值未知，給 0 */
    (void)f_39B5_0126(m, ss, (uint16_t)(bp + 6u), sign, 0);
    (void)f_39B5_0126(m, ss, (uint16_t)(bp + 0x0Au), sign, 0);
    return sign;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_39B5_010C(m, a[0]); }
GIN3_REGISTER("39B5_010C", 1, run)
