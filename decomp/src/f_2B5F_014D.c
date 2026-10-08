/*
 * GIN3PS.EXE 2B5F:014D  次月份造艦計劃（七個艦種的金額、艘數與實付總額）
 *
 * 原版：far 呼叫，沒有參數，回 AX:BX（32 位元，AX＝高字、BX＝低字）。F ＝ ds:8F1D（目前座位，byte，零延伸）：
 *   預算 B ＝ dword(ds:8F5A + 4 × F)（船艦建造費），先複製到 ds:CFFE（低字）／ds:D000（高字）；ds:D03A ← 100
 *   提督 ＝ ds:B77C + ds:8F4B[F] × 28（軍事專任參謀）；營運 ops ＝ 提督[+0x0E]（byte）
 *   對 j ＝ 0..6（七個艦種）：
 *     p ＝ ds:8FD6[F × 7 + j]（byte，這一列的撥款百分比）；ds:D03A ← ds:D03A - p（byte，回繞）
 *     金額 A[j] ＝ (B × p) ÷ 100                      ; 32 位元乘法回繞、無號 32 位元除法 → ds:D002 + 4j
 *     單價 u ＝ ( 列[j]+2 的 word × 100 的低 16 位元 ) ÷ ops   ; 16 位元無號除法（DX 先清零）；列 ＝ ds:4E3A + F × 0xD0 + j × 0x1A
 *     艘數 C[j] ＝ A[j] ÷ u                          ; 無號 32 位元除法（u ＝ 0 時 3CC9:0020 回 0xFFFFFFFF）→ ds:D01E + 4j
 *     實付總額 T += C[j] × u                         ; 32 位元回繞
 *     ds:D01E + 4j ← C[j] × 10                       ; 再讀回乘 10 寫回（docs/re/265：「艘數 ＝ 金額 ÷ 單價 × 10」）
 *   回傳 T
 * 語意見 docs/re/265、docs/spec/037、docs/internals/32／08／40（已證實：反組譯逐指令）。docs 沒有記的細節（反組譯可見）：
 *   (1) 基價 × 100 只留低 16 位元（mul 的 DX 被 xor dx,dx 丟掉），基價 ≥ 656 時與數學上的 ×100 不同；
 *   (2) ops ＝ 0 時 div 觸發 #DE（除以零），本版與差異測試都不涵蓋（輸入域排除，原版這時會進除法錯誤的中斷）；
 *   (3) ds:D03A 是 byte，七列撥款百分比加總超過 100 時回繞。
 * 呼叫：3CC9:0009（乘）、3CC9:0020（除）；以 gin3_lmul／gin3_uldiv 表達（兩者已 diff-tested）。
 * 副作用：寫 ds:CFFE／D000（預算複本）、ds:D002..D01D（金額）、ds:D01E..D039（艘數 × 10）、ds:D03A。
 * 輸入域：任意記憶體內容，但提督[+0x0E] ≠ 0（為 0 時原版 #DE，探針丟棄該案）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

static uint32_t rd32w(const Mem *m, uint16_t off)
{
    return (uint32_t)rd16(m, DS(off)) | ((uint32_t)rd16(m, DS((uint16_t)(off + 2))) << 16);
}

static void wr32w(Mem *m, uint16_t off, uint32_t v)
{
    wr16(m, DS((uint16_t)(off + 2)), (uint16_t)(v >> 16));
    wr16(m, DS(off), (uint16_t)v);
}

uint32_t f_2B5F_014D(Mem *m)
{
    uint16_t seat = rd8(m, DS(0x8F1D));
    uint16_t budget_off = (uint16_t)(seat * 4u);
    uint16_t budget_hi = rd16(m, DS((uint16_t)(budget_off + 0x8F5Cu)));
    uint16_t budget_lo = rd16(m, DS((uint16_t)(budget_off + 0x8F5Au)));
    uint16_t share = (uint16_t)(0x8FD6u + (uint16_t)(seat * 7u));
    uint16_t row = (uint16_t)(0x4E3Au + (uint16_t)(seat * 0xD0u));
    uint16_t amount_at = 0xD002;
    uint16_t ships_at = 0xD01E;
    uint16_t admiral = (uint16_t)(0xB77Cu + (uint16_t)(rd8(m, DS((uint16_t)(0x8F4Bu + seat))) * 0x1Cu));
    uint32_t total = 0;
    uint8_t j;

    wr16(m, DS(0xD000), budget_hi);
    wr16(m, DS(0xCFFE), budget_lo);
    wr8(m, DS(0xD03A), 0x64);
    for (j = 0; j < 7; j++) {
        uint8_t percent = rd8(m, DS(share));
        uint32_t budget;
        uint16_t unit;
        uint16_t scaled;
        uint8_t ops;
        uint32_t count;
        uint32_t amount;

        wr8(m, DS(0xD03A), (uint8_t)(rd8(m, DS(0xD03A)) - percent));
        budget = ((uint32_t)rd16(m, DS(0xD000)) << 16) | rd16(m, DS(0xCFFE));
        percent = rd8(m, DS(share));
        share = (uint16_t)(share + 1);
        wr32w(m, amount_at, gin3_uldiv(gin3_lmul(budget, percent), 100));
        ops = rd8(m, DS((uint16_t)(admiral + 0x0E)));
        scaled = (uint16_t)(rd16(m, DS((uint16_t)(row + 2))) * 100u);
        unit = ops ? (uint16_t)(scaled / ops) : 0; /* ops ＝ 0：原版 #DE，輸入域外；回 0 只是不讓 C 版當機 */
        amount = rd32w(m, amount_at);
        amount_at = (uint16_t)(amount_at + 4);
        count = gin3_uldiv(amount, unit);
        wr32w(m, ships_at, count);
        total += gin3_lmul(count, unit);
        count = rd32w(m, ships_at);
        wr32w(m, ships_at, gin3_lmul(count, 10));
        ships_at = (uint16_t)(ships_at + 4);
        row = (uint16_t)(row + 0x1A);
    }
    return total;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2B5F_014D(m);
}
GIN3_REGISTER("2B5F_014D", 0, run)
