/*
 * GIN3PS.EXE 39B5:0126  呼叫者框架內一個 32 位元有號數的就地取絕對值（為負才取二補數）
 *
 * 呼叫：near，沒有堆疊參數。暫存器傳參：BP＝呼叫者設好的框架指標，資料在 SS:[BP+0]（低字）與 SS:[BP+2]（高字）。
 *   test word [bp+2],8000h  高字的 bit 15 沒設 → 什麼都不做就返回（AX、DX 不變）。
 *   有設 → AX:DX ← 0 − (SS:[BP+2]:SS:[BP+0])（32 位元二補數，借位從低字帶到高字），
 *          再把 AX 寫回 SS:[BP+0]、DX 寫回 SS:[BP+2]。
 * 0x80000000 的負數取負後還是 0x80000000（寫入集合看不出來，暫存器仍輸出）。
 * 呼叫者：只有 39B5:010C（台帳 callers＝2：它內部兩個 call）。010C 在呼叫前先 push AX（兩運算元高字的符號異或）與 BP，
 * 並用 add bp,6／add bp,4 把 BP 依序移到兩個運算元位置，返回前 pop ax 還原 AX；
 * 所以回傳時的 AX、DX 不屬於語意輸出，差異測試比 AX 與 DX 只為了驗證運算本身。
 * 用途判斷（假說）：32 位元有號乘除的前置處理——兩個運算元取絕對值，符號異或另存在 010C 的堆疊上；
 * 39B5:010C 之上的 39B5:000C、005C 是否為乘法與除法本體，本檔未查。
 *
 * 機器模型：SS 固定是基底狀態的 0x63EB（diffprobe 不能設 SS），BP 的 word 存取在位移 0xFFFF 時，
 * 第二個位元組取同一段的位移 0（dosgolem 的 read16/write16 兩次 byte 存取，位移各自在段內回繞），
 * 用本檔的 rd16s／wr16s 重現。SS:BP 的線性位址最高 0x73EAF，不會落入 ROM（0xF0000 以上），也不是視訊記憶體。
 *
 * 輸入域：BP 的 +0..+3 四個 byte 不落在探針的堆疊區（SS:1B8A..1F89，基底狀態 SP＝1F8A 往下 0x400 bytes）。
 * 落在堆疊區時，探針把該區的寫入排除在寫入集合之外，而原版 call 推入的返回位址又會改變讀到的值，C 版無法重現，
 * 不涵蓋。其餘 BP 值、AX、DX 皆為全部 16 位元值。
 * 推論等級：行為已證實（反組譯逐指令＋差異測試）；用途是強推論。
 */
#include "gin3_reg.h"

#define F39B5_0126_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */

static uint16_t f_39B5_0126_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1))) << 8));
}

static void f_39B5_0126_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    wr8(m, LIN(seg, off), (uint8_t)v);
    wr8(m, LIN(seg, (uint16_t)(off + 1)), (uint8_t)(v >> 8));
}

/* 回傳 (AX<<16)|DX */
uint32_t f_39B5_0126(Mem *m, uint16_t ss, uint16_t bp, uint16_t ax, uint16_t dx)
{
    uint16_t lo, hi, borrow;

    if ((f_39B5_0126_rd16s(m, ss, (uint16_t)(bp + 2)) & 0x8000u) == 0)
        return ((uint32_t)ax << 16) | dx;
    lo = f_39B5_0126_rd16s(m, ss, bp);
    hi = f_39B5_0126_rd16s(m, ss, (uint16_t)(bp + 2));
    borrow = (lo != 0) ? 1 : 0;   /* sub ax,[bp+0]：0 − lo 在 lo != 0 時借位 */
    ax = (uint16_t)(0 - lo);
    dx = (uint16_t)(0 - hi - borrow); /* sbb dx,[bp+2] */
    f_39B5_0126_wr16s(m, ss, bp, ax);
    f_39B5_0126_wr16s(m, ss, (uint16_t)(bp + 2), dx);
    return ((uint32_t)ax << 16) | dx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_39B5_0126(m, F39B5_0126_SS, a[2], a[0], a[1]);
}
GIN3_REGISTER("39B5_0126", 3, run)
