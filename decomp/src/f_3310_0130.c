/*
 * GIN3PS.EXE 3310:0130  整數開平方：把引數 x 當 16.16 定點數的整數部分（高字），低字補 0，交給 3310:014B 開平方
 *
 * 原版（27 bytes，far 呼叫，一個堆疊字；序言 push bp／si、mov bp,sp，沒有 sub sp，所以 x 在 [bp+8]）：
 *   si ← x；ax ← x；bx ← 0（`xor bx,bx`）
 *   push ax ；push bx ；call far 3310:014B       ; 依推入順序，先推的 ax 在高位址：dword n 的高字 ＝ x、低字 ＝ 0
 *   mov sp,bp ；pop si ；pop bp ；retf
 *   回傳 AX:BX（高字在 AX、低字在 BX）原樣取自 014B（描述檔 ret ＝ pack:ax,bx）。
 * 過程（已證實：反組譯逐指令；3310:014B 由其差異測試證實）：
 *   回 f_3310_014B(低字 0, 高字 x)，也就是 16.16 定點開平方 sqrt(x) × 65536 的牛頓法結果。
 * 語意（強推論）：docs/re/299-region-unrest-writers-and-recompute.md 第 245 到 248 行「3310:0130(x) 把 x 當高位字組成 32 位元、
 *   交給 3310:014B」，與本檔一致；docs/re/43-unrest-full-disassembly.md 第 125 行把 3310:0130 稱為 helper(n × p × q ÷ 1000)；
 *   docs/spec/006-unrest-at-load.md 第 77 行、docs/spec/042-region-unrest-recompute.md 第 112 行說唯一呼叫者 2703:10FC 推入的是清成 0 的那一格。
 * 副作用：全部來自 014B（即 39B5:005C 寫 ds:8EDC..8EED 共 18 bytes）；本函式自己不寫記憶體。
 *
 * 輸入域：x 全部 16 位元值。x ＝ 0 時 014B 回 0:0；x >= 0x8000 時 n 當有號數是負的（014B 的迴圈一立即離開）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3310_014B(Mem *m, uint16_t lo, uint16_t hi);

uint32_t f_3310_0130(Mem *m, uint16_t x)
{
    return f_3310_014B(m, 0, x);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3310_0130(m, a[0]); }
GIN3_REGISTER("3310_0130", 1, run)
