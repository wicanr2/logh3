/*
 * GIN3PS.EXE 3870:0614  橢圓的焦距 c ＝ ⌊a × e⌋：e ＝ 3870:0643(a, b)（16.16 定點離心率），乘上 sext32(a)，回 32 位元乘積的高字
 *
 * 原版（47 bytes，far 呼叫，兩個 word 堆疊參數；序言 push bp／si／di、mov bp,sp，沒有 sub sp，所以 arg_0 ＝ a 在 [bp+0Ah]、arg_2 ＝ b 在 [bp+0Ch]）：
 *   回傳 AX（描述檔 ret ＝ ax）。
 * 過程（已證實：反組譯逐指令；3870:0643、3CC9:0009 由其各自的差異測試證實）：
 *   push b、push a、call 3870:0643（回 AX:BX ＝ e 的高字:低字）、mov sp,bp
 *   DI ← AX（e 的高字）；AX:BX ← DI:BX（BX 是 0643 回的低字，中間沒有被動過）＝ e
 *   CX:DX ← sext32(a)（cwd；DX:AX ＝ cwd 後 CX ＝ 符號字，DX ＝ a）
 *   AX:BX ← 3CC9:0009(e, sext32(a))（取低 32 位元）
 *   回傳 AX ＝ 乘積的高字（`mov bx,ax; xor ax,ax; mov ax,bx`：BX 重新載入後 AX ＝ 3CC9 回傳的 AX，即高字）
 * 語意（強推論）：docs/re/147-fortress-gun-damage.md 第 16 行「3870:0614(a, b) ＝ 焦距 c ＝ ⌊a × e⌋（取 16.16 乘積的高位字）」、
 *   docs/re/145-fortress-gun-aim-and-fire.md 第 86 行（c ＝ 3870:0614(長半軸, 短半軸)），與本檔一致。
 * 副作用：全部來自 3870:0643 → 3310:014B → 39B5:005C（ds:8EDC..8EED 共 18 bytes）；本函式自己不寫記憶體。
 * 輸入域：a、b 全部 16 位元值。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3870_0643(Mem *m, uint16_t a_word, uint16_t b_word);
extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);

uint16_t f_3870_0614(Mem *m, uint16_t a, uint16_t b)
{
    uint32_t e = f_3870_0643(m, a, b);
    uint32_t sa = (uint32_t)(int32_t)(int16_t)a; /* cwd */
    uint32_t prod = f_3CC9_0009((uint16_t)(e >> 16), (uint16_t)e, (uint16_t)(sa >> 16), (uint16_t)sa);

    return (uint16_t)(prod >> 16);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3870_0614(m, a[0], a[1]); }
GIN3_REGISTER("3870_0614", 2, run)
