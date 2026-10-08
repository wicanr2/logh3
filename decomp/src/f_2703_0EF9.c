/*
 * GIN3PS.EXE 2703:0EF9  星域「稅率壓成長」基底項（未安定度公式的 f1）
 *
 * 原版：far 呼叫，一個參數＝星域記錄的 far 指標（位移在前、段在後），回 AX。
 * 全程 16 位元有號運算（imul／cwd／idiv），語意如下。名稱與「稅率」「成長率」來自 docs/re/43，
 * 欄位本義未獨立驗證（強推論）；運算本身逐指令對過（已證實）。
 *   owner  ＝ 星域 +0x12
 *   a      ＝ ds:8FD4[owner]（byte），b ＝ ds:8FF8[owner]（byte）
 *   di     ＝ ds:147E + b×4 的 word（只取低字，高字讀了又被覆蓋）
 *   growth ＝ 星域 +0x1E（有號 word）
 *   v0     ＝ growth > 0 ? (星域 +0x27 × 100) ÷ growth : di
 *   v0 >= 400（有號）或 v0 <= 25 → v0 ＝ di        ← docs/re/43 只寫了 <= 25 這一半
 *   p      ＝ (v0 × 100) 的低 16 位元，再有號延伸（imul 後的 cwd 丟掉高字）
 *   q1     ＝ p ÷ di（有號 idiv）
 *   d      ＝ (q1 - 100) ÷ 4（有號，向零取整）+ 100，皆 16 位元回繞
 *   m      ＝ ds:8F18（無號 word）% 10
 *   blend  ＝ ((m + 5) × ds:4B28[a] + (14 - m) × ds:4B28[b])（兩個乘積各留低 16 位元，和也留 16 位元）÷ 19（有號）
 *   回傳   ＝ ((blend × 100) 的低 16 位元再有號延伸) ÷ d（有號 idiv）
 * 由算式推得的兩個範圍事實（已證實，窮舉 byte × word 的商得出）：
 *   - 兩個權重 (m+5) 與 (14-m) 的和恆為 19，ds:4B28 的元素是 byte，所以 blend 最大 255，
 *     blend×100 不會超出 16 位元有號範圍；第二次的「取低 16 位元再延伸」實際上不改變結果。
 *     第一次（v0×100）在 v0 為 328..399 時會溢位。
 *   - growth > 0 路徑上，(稅率 byte × 100) ÷ growth 的商不可能剛好是 399，
 *     所以 v0 的上限判斷在 399 與 400 之間只有 400 這一格有差別。
 * 星域記錄 ds:A79C 起 32 筆 × 126 bytes（docs/internals/22）。
 * 副作用：無（只動自己的堆疊框）。
 * 輸入域：原版在 di == 0、d == 0 或商超出 16 位元有號範圍時觸發 #DE（INT 0）；差異測試會把這種執行丟棄，
 *   C 版這種情況回 0 只為避免 SIGFPE，該路徑不在比較範圍內。
 */
#include "gin3_reg.h"

/* 8086 idiv r16：32 位元被除數 ÷ 16 位元除數，商與餘數都要放進 16 位元。這裡被除數一律是 cwd 延伸的 16 位元值。
 * 會 #DE（除數 0 或商溢位）時回 0。 */
static uint16_t idiv16(int16_t dividend, int16_t divisor)
{
    int32_t q;

    if (divisor == 0)
        return 0;
    q = (int32_t)dividend / divisor;
    if (q > 32767 || q < -32768)
        return 0;
    return (uint16_t)(int16_t)q;
}

uint16_t f_2703_0EF9(Mem *m, uint16_t region_off, uint16_t region_seg)
{
    uint8_t owner = rd8(m, LIN(region_seg, region_off + 0x12));
    uint16_t a = rd8(m, DS((uint16_t)(0x8FD4u + owner)));
    uint16_t b = rd8(m, DS((uint16_t)(0x8FF8u + owner)));
    uint16_t di = rd16(m, DS((uint16_t)(0x147Eu + b * 4u)));
    int16_t growth = (int16_t)rd16(m, LIN(region_seg, region_off + 0x1E));
    int16_t v0;
    int16_t p;
    uint16_t q1;
    uint16_t d;
    uint16_t mod10;
    uint16_t part_a;
    uint16_t part_b;
    uint16_t blend;

    if (growth > 0)
        v0 = (int16_t)idiv16((int16_t)(rd8(m, LIN(region_seg, region_off + 0x27)) * 100u), growth);
    else
        v0 = (int16_t)di;
    if (v0 >= 400 || v0 <= 25)
        v0 = (int16_t)di;

    p = (int16_t)(uint16_t)((int32_t)v0 * 100);
    q1 = idiv16(p, (int16_t)di);
    d = (uint16_t)(q1 - 100u);
    d = (uint16_t)(idiv16((int16_t)d, 4) + 100u);

    mod10 = (uint16_t)(rd16(m, DS(0x8F18)) % 10u);
    part_a = (uint16_t)((uint16_t)(mod10 + 5u) * rd8(m, DS((uint16_t)(0x4B28u + a))));
    part_b = (uint16_t)((uint16_t)(14u - mod10) * rd8(m, DS((uint16_t)(0x4B28u + b))));
    blend = idiv16((int16_t)(uint16_t)(part_a + part_b), 19);

    p = (int16_t)(uint16_t)((int32_t)(int16_t)blend * 100);
    return idiv16(p, (int16_t)d);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2703_0EF9(m, a[0], a[1]); }
GIN3_REGISTER("2703_0EF9", 2, run)
