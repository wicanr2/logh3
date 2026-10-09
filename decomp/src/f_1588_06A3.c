/*
 * GIN3PS.EXE 1588:06A3  戰場地圖載入：把 RLE 壓縮資料解壓縮到目的地，再在地圖上塞橫縱格線
 *
 * 原版：far 呼叫，四個 word 的堆疊參數（序言 push bp、沒有 sub sp，29 bytes），由低位址到高位址：
 *   dst 位移 [bp+6]、dst 段 [bp+8]、src 位移 [bp+0Ah]、src 段 [bp+0Ch]（兩個 far 指標，位移在前）。
 *   push [bp+0Ch]；push [bp+0Ah]；push [bp+8]；push [bp+6]；call 1588:06C0（near 呼叫）；mov sp,bp；call 1588:0731；pop bp；retf
 *   即原樣把四個參數轉給 06C0(dst, src)，然後呼叫 0731()（無參數）。
 *   AX 是 0731 結束時的殘值，描述檔 ret = none。
 * 06C0 把 src 的 RLE 資料解壓縮成 128 × 102 格（13,056 bytes）到 dst；0731 依序呼叫 073C（橫線 0 → 3）與 078C（縱線 0 → 4、3 → 2）
 *   處理 ds:3C1A:3C1C 指的地圖。兩者的目標是否相同由呼叫者決定：呼叫者（台帳 callers ＝ 2：1588:0005、1588:00F6）傳的 dst 是
 *   ds:3C1A 指的地圖緩衝區（docs/re/314 §四；06C0 檔頭），本函式不檢查；差異測試的 dst 也涵蓋與 ds:3C1A 不同的值。
 *
 * 呼叫的函式：1588:06C0、1588:0731（皆已 diff-tested，結構與時序見各自的 C 檔）。
 * 輸入域：同 06C0 與 0731；解壓縮約 158,400 步，單格封包太多的來源會被探針以步數上限丟棄，不在驗證範圍。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；「載入戰場地圖」的用途是強推論（docs/re/314 §四、06C0 檔頭）。
 */
#include "gin3_reg.h"

extern void f_1588_06C0(Mem *m, uint16_t dst_off, uint16_t dst_seg, uint16_t src_off, uint16_t src_seg);
extern void f_1588_0731(Mem *m);

void f_1588_06A3(Mem *m, uint16_t dst_off, uint16_t dst_seg, uint16_t src_off, uint16_t src_seg)
{
    f_1588_06C0(m, dst_off, dst_seg, src_off, src_seg);
    f_1588_0731(m);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1588_06A3(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("1588_06A3", 4, run)
