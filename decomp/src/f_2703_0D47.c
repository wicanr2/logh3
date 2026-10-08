/*
 * GIN3PS.EXE 2703:0D47  行星成長率（千分率）
 *
 * 原版：far 呼叫，一個參數＝行星記錄的 far 指標（位移在前、段在後），回 AX。
 * 語意（已證實：反組譯逐指令；階梯表內容與 docs/re/33、docs/internals/data/growth-table.json 一致）：
 *   表在 ds:4A68，一列 6 bytes ＝ 門檻（32 位元）＋ 成長率（16 位元）。指標 row 從 0x4A68 起，
 *   每圈讀「下一列的門檻」ds:[row+6]（低字）、ds:[row+8]（高字）與行星經濟力（+0x1A 低字、+0x1C 高字）
 *   做 32 位元無號比較，下一列門檻 >= 經濟力 就停；否則 row += 6（16 位元回繞）。
 *   停下來的成長率 ＝ ds:[row+4]（word）。
 *   行星 +0x16 < 3000（0x0BB8，無號）：成長率 ＝ ((成長率 × 3) 取低 16 位元) >> 1。
 * 欄位標籤：asm 讀的是 +0x16。docs/internals/29 定義 +0x16 為「防衛力上限」（+0x14 才是現在值），
 *   docs/re/33、144 把它寫成「防衛力」。這裡以位移為準，不下語意結論（標籤差異已列入回報）。
 * 行星記錄：82 bytes（docs/internals/29），執行期整張表由 ds:5336(位移)／ds:5338(段) 指向。
 * 副作用：無，只讀。
 * 輸入域：門檻表被改壞而沒有任何一列滿足條件時原版會一路走 16 位元回繞的指標不停（步數超出上限、
 *   差異測試會丟棄）；C 版在走滿一整圈（32768 列）後回 0 以免卡死，那條路徑不在比較範圍內。
 */
#include "gin3_reg.h"

uint16_t f_2703_0D47(Mem *m, uint16_t planet_off, uint16_t planet_seg)
{
    uint16_t row = 0x4A68;
    uint16_t rate = 0;
    uint16_t steps;

    for (steps = 0; steps < 0x8000u; steps++) {
        uint32_t next_lo = rd16(m, DS((uint16_t)(row + 6)));
        uint32_t next_hi = rd16(m, DS((uint16_t)(row + 8)));
        uint32_t econ_lo = rd16(m, LIN(planet_seg, planet_off + 0x1A));
        uint32_t econ_hi = rd16(m, LIN(planet_seg, planet_off + 0x1C));
        uint32_t next = (next_hi << 16) | next_lo;
        uint32_t econ = (econ_hi << 16) | econ_lo;

        if (next >= econ) {
            rate = rd16(m, DS((uint16_t)(row + 4)));
            if (rd16(m, LIN(planet_seg, planet_off + 0x16)) < 0x0BB8u)
                rate = (uint16_t)((uint16_t)(rate * 3u) >> 1);
            return rate;
        }
        row = (uint16_t)(row + 6);
    }
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2703_0D47(m, a[0], a[1]); }
GIN3_REGISTER("2703_0D47", 2, run)
