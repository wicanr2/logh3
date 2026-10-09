/*
 * GIN3PS.EXE 3870:0643  橢圓的離心率 √(1 − B² ÷ A²)（16.16 定點）：A、B 是兩個有號 16 位元整數，先算 T ＝ 1 − B² ÷ A²，再交給 3310:014B 開方
 *
 * 原版（224 bytes，far 呼叫，兩個 word 堆疊參數；序言 push bp／si／di、sub sp,0Ch、mov bp,sp，所以 A 在 [bp+16h]、B 在 [bp+18h]），
 *   回傳 AX ＝ 高字、BX ＝ 低字（描述檔 ret ＝ pack:ax,bx）。區域變數：A32 ＝ [bp+0Ah]:[bp+8]、B32 ＝ [bp+6]:[bp+4]（皆由 cwd 符號延伸）、T ＝ [bp+2]:[bp+0]。
 * 過程（已證實：反組譯逐指令；3CC9:0009、3CC9:0020、3310:014B 各由其差異測試證實）：
 *   A32 ← sext32(A)、B32 ← sext32(B)
 *   B32 的高字 ≠ 0（B 為負）或 B 的低字 >= 0x100（無號 cmp si,100h；jb）→ 大值路徑：
 *     Q ← ((B32 × B32) << 12) ÷ (A32 × A32)           ; 乘法走 3CC9:0009（取低 32 位元），左移 12 位為 32 位元 shl／rcl，除法走 3CC9:0020（無號，除數 0 商 0xFFFFFFFF）
 *     T ← (0x1000 − Q) << 4                            ; 32 位元 sub／sbb，再 shl／rcl 四次
 *   否則（0 <= B < 0x100）→ 小值路徑：
 *     Q ← ((B32 × B32 的低字) << 16) ÷ (A32 × A32)     ; mov ax,bx；xor bx,bx：乘積只留低字，放到高字
 *     T ← 0x10000 − Q                                  ; 32 位元 sub／sbb
 *   回 3310:014B(T)（T 的低字在 [bp+0]、高字在 [bp+2]，push 高字、push 低字）
 *   兩條路徑的 1.0 都是 0x10000：大值路徑的商只有 12 位元小數，最後再 << 4 補回（docs/re/147 第 15 行的說法是避免溢位）。
 *   B² << 12 只留低 32 位元：B >= 0x400 時 B² << 12 已超過 32 位元，高位元被截掉，原版照做，C 版同。
 *   分界測試有兩個：B32 高字 ≠ 0（`or cx,cx; jnz`）與 `cmp si,100h; jb`。B 為負時低字 >= 0x8000，本來就 >= 0x100，所以高字測試不改變結果（冗餘指令，
 *   覆蓋由 B 為負的案例到達）；小值路徑的 B² < 0x10000，「只留低字」是恆等運算。這兩處的突變都是等價突變，突變清單改用有號比較與左移 15。
 * 語意（強推論）：docs/re/94-attack-coefficient-is-a-conic.md 第 43、48、82 行（`3870:0643(W16, W18)` 算離心率 e ＝ √(1 − b² ÷ a²)）、
 *   docs/re/147-fortress-gun-damage.md 第 15 行（b >= 256 時先把中間量縮成 12 位再補回，避免溢位），與本檔一致；
 *   docs/internals/51-borland-long-helpers.md 第 41 行把它與 3870:0614 並列。A ＝ 0 時 A² ＝ 0，3CC9:0020 回商 0xFFFFFFFF，不引發 #DE。
 * 副作用：全部來自 3310:014B → 39B5:005C（ds:8EDC..8EED 共 18 bytes）；本函式自己不寫記憶體。
 *
 * 機器模型：C 版沒有堆疊；被呼叫的 39B5:005C 的堆疊參數放在 fcmd-base.state 進入點 SP ＝ 1F8A 的框架（f_39B5_005C.c 檔頭），原版從本函式呼叫時框架更深，
 *   位置不同，但兩處都在探針的堆疊帶內，比對時排除，C 版自寫自讀，不影響結果。
 * 輸入域：A、B 全部 16 位元值。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3CC9_0009(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint64_t f_3CC9_0020(uint16_t ax, uint16_t bx, uint16_t cx, uint16_t dx);
extern uint32_t f_3310_014B(Mem *m, uint16_t lo, uint16_t hi);

/* 3CC9:0009：AX:BX × CX:DX 的低 32 位元 */
static uint32_t w0643_mul(uint32_t a, uint32_t b)
{
    return f_3CC9_0009((uint16_t)(a >> 16), (uint16_t)a, (uint16_t)(b >> 16), (uint16_t)b);
}

/* 3CC9:0020：AX:BX ÷ CX:DX 的商（無號） */
static uint32_t w0643_div(uint32_t a, uint32_t b)
{
    return (uint32_t)(f_3CC9_0020((uint16_t)(a >> 16), (uint16_t)a, (uint16_t)(b >> 16), (uint16_t)b) >> 32);
}

uint32_t f_3870_0643(Mem *m, uint16_t a_word, uint16_t b_word)
{
    uint32_t a32 = (uint32_t)(int32_t)(int16_t)a_word; /* cwd */
    uint32_t b32 = (uint32_t)(int32_t)(int16_t)b_word; /* cwd */
    uint32_t t;

    if ((uint16_t)(b32 >> 16) != 0 || (uint16_t)b32 >= 0x100u) {
        uint32_t bb = w0643_mul(b32, b32) << 12;
        uint32_t aa = w0643_mul(a32, a32);
        uint32_t q = w0643_div(bb, aa);

        t = (0x1000u - q) << 4;
    } else {
        uint32_t bb = w0643_mul(b32, b32);
        uint32_t aa;
        uint32_t q;

        bb = (uint32_t)(uint16_t)bb << 16;
        aa = w0643_mul(a32, a32);
        q = w0643_div(bb, aa);
        t = 0x10000u - q;
    }
    return f_3310_014B(m, (uint16_t)t, (uint16_t)(t >> 16));
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_3870_0643(m, a[0], a[1]); }
GIN3_REGISTER("3870_0643", 2, run)
