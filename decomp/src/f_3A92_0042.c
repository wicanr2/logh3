/*
 * GIN3PS.EXE 3A92:0042  二維旋轉（暫存器版）：以 byte 角度 a 旋轉 (x, y)，回 (x', y')
 *
 * 原版：far 呼叫，沒有堆疊參數，沒有 C 序言。暫存器傳參：AL ＝ 角度 a（xor ah,ah，AH 被忽略）、BX ＝ x、CX ＝ y（皆當有號 16 位元）。
 * 輸出暫存器：BX ＝ x'、CX ＝ y'、SI ＝ 0x573A + 2 × a、DI ＝ 0x573A + 2 × ((a + 0x40) mod 256)；AX ＝ y'（mov cx,ax 之前的值）；
 *   DX 是 3A92:0081 最後留下的 32 位元和的高字，不比對（描述檔 ret = pack:bx,cx,si,di，最多 4 個暫存器）。
 *
 * 過程（已證實：反組譯逐指令；3A92:0062、0081 由其各自的差異測試證實）：
 *   SI ← 0x573A + a + a；AL ← AL + 0x40（只改 AL，8 位元回繞，AH 保持 0）；DI ← 0x573A + a2 + a2（a2 ＝ 新的 AL）
 *   push bx; push cx; call 3A92:0062（x' ＝ (cos·x − sin·y + 512) >> 10；SI 指 sin 表項、DI 指 cos 表項）；pop cx; pop bx（還原 BX、CX）
 *   push ax; call 3A92:0081（y' ＝ (sin·x + cos·y + 512) >> 10）；mov cx,ax；pop bx（BX ← x'）
 *   所以 0081 看到的 BX、CX 是原始的 x、y，不是 0062 弄髒的值。
 * 表 ds:573A 起 256 筆有號 word，值約 sin × 1024（docs/internals/41，強推論）；"旋轉"這個命名未被差異測試證明，
 *   差異測試證明的是算術本身。
 * 副作用：無寫入。輸入域：AX、BX、CX 全部 16 位元值（AH 忽略）；表內容取自記憶體（基底狀態與突變）。
 *   角度最大 255，所以最多讀到 ds:0x5938 與 0x593A（含）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_3A92_0062(Mem *m, uint16_t si, uint16_t di, uint16_t bx, uint16_t cx);
extern uint16_t f_3A92_0081(Mem *m, uint16_t si, uint16_t di, uint16_t bx, uint16_t cx);

/* 回 pack:bx,cx,si,di（BX 在最高 16 位元） */
uint64_t f_3A92_0042(Mem *m, uint16_t ax_in, uint16_t bx, uint16_t cx)
{
    uint8_t a = (uint8_t)ax_in;
    uint16_t si = (uint16_t)(0x573Au + a + a);
    uint8_t a2 = (uint8_t)(a + 0x40u);
    uint16_t di = (uint16_t)(0x573Au + a2 + a2);
    uint16_t x1 = f_3A92_0062(m, si, di, bx, cx);
    uint16_t y1 = f_3A92_0081(m, si, di, bx, cx);

    return ((uint64_t)x1 << 48) | ((uint64_t)y1 << 32) | ((uint64_t)si << 16) | di;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3A92_0042(m, a[0], a[1], a[2]); }
GIN3_REGISTER("3A92_0042", 3, run)
