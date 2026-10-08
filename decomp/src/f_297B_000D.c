/*
 * GIN3PS.EXE 297B:000D  艦隊表初始化（新遊戲）
 *
 * 原版：far 呼叫，沒有參數：
 *   1. memset(ds:C074, 0xFF, 0x0F40)                 ; 艦隊表整張（32 筆 × 122 bytes ＝ 3904 bytes）填 0xFF
 *   2. 對 i ＝ 0..31，p ＝ ds:C074 + 122 × i：
 *        p[+3] ← i                                    ; 這一筆的槽號（docs/internals/33：已證實）
 *        p[+2] ← (i mod 16) + 1                       ; 艦隊番號（docs/internals/33：已證實）
 *   3. ds:8F4A ← 0；ds:8F49 ← 0                       ; 每陣營的艦隊數（docs/spec/029：ds:8F49[陣營]，強推論）
 * 回傳：沒有回傳值。離開時 AX＝0（xor ax,ax），差異測試比較 AX。唯一的呼叫者 0374:023F 呼叫後緊接著 call 2AD2:0004，不使用（反組譯）。
 * 呼叫：3C9C:000A（已 diff-tested）。
 * 副作用：寫艦隊表 3904 個 byte 與 ds:8F49／8F4A。不讀任何記憶體。輸入域：無參數，全部可到達。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);

uint16_t f_297B_000D(Mem *m)
{
    uint16_t record = 0xC074;
    uint8_t i;

    f_3C9C_000A(m, 0xC074, GIN3_DS, 0xFF, 0x0F40);
    for (i = 0; i < 0x20; i++) {
        wr8(m, DS(record + 3), i);
        wr8(m, DS(record + 2), (uint8_t)(i % 16 + 1));
        record = (uint16_t)(record + 0x7A);
    }
    wr8(m, DS(0x8F4A), 0);
    wr8(m, DS(0x8F49), 0);
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_297B_000D(m);
}
GIN3_REGISTER("297B_000D", 0, run)
