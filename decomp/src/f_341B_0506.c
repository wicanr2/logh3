/*
 * GIN3PS.EXE 341B:0506  沿視窗串列把每個「種類 2」的視窗節點鎖住它所在的畫面格（呼叫 34CB:000D）
 *
 * 原版：far 呼叫，一個參數＝視窗節點的 far 指標（位移在 [bp+6]、段在 [bp+8]；只有 push bp，沒有區域變數）。無回傳值。
 *   唯一呼叫者 341B:0145（台帳 callers = 1）。
 * 過程（已證實：反組譯逐指令；34CB:000D 由其自己的差異測試證實）：
 *   do {
 *     若 節點[+0] == 2（byte）：34CB:000D(節點)
 *     節點 ← 節點 +0x12（位移 word）／+0x14（段 word）     ; 呼叫 000D 之後才讀，所以 000D 的寫入會影響這次讀到的指標
 *   } while (節點 不是 0:0)                                   ; 兩個 word 同時為 0 才結束
 *   do-while：進來時指標是 0:0 也會先讀 0000:0000 比對一次（不先檢查）。
 *   節點的 +0x12／+0x14 與 341B:0353 走的 +0x16／+0x18 是同一個雙向串列的兩個方向：基底狀態頭是 54B3:F144（編號 4），
 *   +0x16／+0x18 往尾（94F6）走，+0x12／+0x14 往頭走（94F6 → 9518 → 953A → 955C → F144 → 0:0）。方向的命名是強推論。
 * 副作用：只有 34CB:000D 的寫入（畫面格表 ds:D81E 起，格 [+0] |= 0x20）。
 * 輸入域：節點取基底狀態的視窗串列與其突變；串列有環或太長（超過探針 200,000 步）的案例被丟棄。
 *   word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（rd16s）。讀到 A0000 到 BFFFF 的案例由探針丟棄。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern void f_34CB_000D(Mem *m, uint16_t w_off, uint16_t w_seg);

static uint16_t w0506_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

void f_341B_0506(Mem *m, uint16_t w_off, uint16_t w_seg)
{
    do {
        uint16_t nseg, noff;

        if (rd8(m, LIN(w_seg, w_off)) == 2)
            f_34CB_000D(m, w_off, w_seg);
        nseg = w0506_rd16s(m, w_seg, (uint16_t)(w_off + 0x14u));
        noff = w0506_rd16s(m, w_seg, (uint16_t)(w_off + 0x12u));
        w_seg = nseg;
        w_off = noff;
    } while ((w_seg | w_off) != 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_341B_0506(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("341B_0506", 2, run)
