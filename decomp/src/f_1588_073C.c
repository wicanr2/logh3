/*
 * GIN3PS.EXE 1588:073C  戰場地圖塞橫向格線（每 12 列的整列）
 *
 * 原版：far 呼叫，無參數，無回傳值。地圖緩衝區遠指標在 ds:3C1A（位移）／ds:3C1C（段），每列 128 格。
 * 唯一呼叫者 1588:0731（載入地圖之後，與 078C 一起塞格線；1588:0731 先呼叫 073C、後呼叫 078C）。
 *
 * 語意（已證實：docs/re/314 §四「073C：列 0, 12, 24, …（< 102）整列：圖塊 0 → 3」；本檔與反組譯逐條對過）：
 *   對列 r ＝ 0, 12, 24, …（r 是 byte，每次加 12，直到 r >= 0x66）：
 *     取該列第一格的位址 ＝ ds:3C1A ＋ (r << 7)（16 位元），段 ds:3C1C（每列都重讀這兩個變數，指標不跨列累加）
 *     該列的 128 格每格：若 byte == 0，寫成 3。
 *   即 r ＝ 0, 12, …, 96 共 9 列。
 *
 * 輸入域：地圖指標取基底狀態（戰場地圖已載入的局面）。基底地圖已經塞過格線，所以差異測試用種子把格線格子
 *   還原成未塞過的樣子（0），再由突變導向讀取集改動內容與指標。
 */
#include "gin3_reg.h"

void f_1588_073C(Mem *m)
{
    uint8_t row = 0;

    while (row < 0x66u) {
        uint16_t off = (uint16_t)(rd16(m, DS(0x3C1A)) + (uint16_t)((uint16_t)row << 7));
        uint16_t seg = rd16(m, DS(0x3C1C));
        uint8_t col = 0;

        while (col < 0x80u) {
            if (rd8(m, LIN(seg, off)) == 0)
                wr8(m, LIN(seg, off), 3);
            off = (uint16_t)(off + 1u);
            col++;
        }
        row = (uint8_t)(row + 0x0Cu);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_1588_073C(m);
    return 0;
}
GIN3_REGISTER("1588_073C", 0, run)
