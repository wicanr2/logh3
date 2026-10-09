/*
 * GIN3PS.EXE 2BE1:011D  由位移 (a, b) 求 byte 角度（256 刻一圈）：atan(|b| ÷ |a|) 查表後依象限換算
 *
 * 原版：far 呼叫，兩個堆疊字 a、b（有號 16 位元；push bp／si／di、sub sp,0Ah，所以 a 在 [bp+14h]、b 在 [bp+16h]），
 *   回 AL（AH 是殘值，描述檔 ret = pack:al）。
 * 過程（已證實：反組譯逐指令；2BE1:01E1、3CC9:0020 由其各自的差異測試證實）：
 *   A ← |a|、B ← |b|（`neg`，16 位元：0x8000 取負後仍是 0x8000）
 *   A == 0                → 比值 ← 0xFFFF
 *   否則 Q ← (sext32(B) << 10) ÷ sext32(A)    ; 無號 32 位元除法 3CC9:0020；cwd 符號延伸：B 或 A 為 0x8000 時延伸成 0xFFFF8000
 *     Q 的高字 == 0 → 比值 ← Q 的低字；否則（高字 >= 1；asm 對高字 == 1 也走夾限分支）→ 比值 ← 0xFFFF
 *   k ← 2BE1:01E1(比值)（反正切表查找，0..0x40）
 *   依 a、b 的原始正負號（有號比較 0，jns／js）換算：
 *     a >= 0、b >= 0 → k            a < 0、b >= 0 → 0x80 − k（8 位元）
 *     a < 0、b < 0   → k + 0x80     a >= 0、b < 0 → −k（8 位元）
 *   回 AL
 * 語意（強推論，docs/re/129、docs/re/330 第 313 行附近；名稱不是證據）：0x40 是 90 度，所以是以 x 軸為 0 度、
 *   y 向下（b < 0 時角度為負 = 0x100 − k）的方位角。
 * 副作用：無寫入。輸入域：a、b 全部 16 位元值；表 ds:593A 起 0x41 個 word 取自記憶體。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_2BE1_01E1(Mem *m, uint16_t ratio);

uint8_t f_2BE1_011D(Mem *m, uint16_t a_word, uint16_t b_word)
{
    int16_t a = (int16_t)a_word;
    int16_t b = (int16_t)b_word;
    uint16_t abs_a = (a < 0) ? (uint16_t)(0u - a_word) : a_word;
    uint16_t abs_b = (b < 0) ? (uint16_t)(0u - b_word) : b_word;
    uint16_t ratio;
    uint8_t k;

    if (abs_a == 0) {
        ratio = 0xFFFFu;
    } else {
        uint32_t num = (uint32_t)(int32_t)(int16_t)abs_b << 10; /* cwd；shl／rcl 十次，32 位元 */
        uint32_t den = (uint32_t)(int32_t)(int16_t)abs_a;       /* cwd */
        uint32_t q = gin3_uldiv(num, den);                      /* 3CC9:0020 */

        ratio = ((q >> 16) == 0) ? (uint16_t)q : 0xFFFFu;
    }
    k = (uint8_t)f_2BE1_01E1(m, ratio);
    if (a >= 0) {
        if (b >= 0)
            return k;
        return (uint8_t)(0u - k);
    }
    if (b >= 0)
        return (uint8_t)(0x80u - k);
    return (uint8_t)(k + 0x80u);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2BE1_011D(m, a[0], a[1]); }
GIN3_REGISTER("2BE1_011D", 2, run)
