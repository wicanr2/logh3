/*
 * GIN3PS.EXE 1588:078C  戰場地圖塞縱向格線（每 15 欄）
 *
 * 原版：far 呼叫，無參數，無回傳值。地圖緩衝區遠指標在 ds:3C1A（位移）／ds:3C1C（段），每列 128 格。
 * 唯一呼叫者 1588:0731（載入地圖之後，與 073C 一起塞格線；1588:0731 先呼叫 073C、後呼叫 078C）。
 *
 * 語意（已證實：docs/re/314 §四「078C：每列欄 0, 15, 30, …（< 128）：圖塊 0 → 4；圖塊 3 → 2」；
 *   本檔與反組譯逐條對過）：
 *   對列 r ＝ 0 … 0x65：
 *     該列第一格位址 ＝ ds:3C1A ＋ (r << 7)（16 位元），段 ds:3C1C（每列重讀這兩個變數）
 *     對欄 c ＝ 0, 15, 30, …（c 是 byte，每次加 15，直到 c >= 0x80；即 0, 15, …, 120 共 9 欄）：
 *       若 byte == 0：寫成 4
 *       否則若 byte == 3：寫成 2
 *   （byte == 3 的格是 073C 先前塞的橫線：橫線與縱線交會的格 3 → 2；1588:0731 的呼叫順序是 073C 先、078C 後。）
 *
 * 輸入域：同 1588:073C。基底地圖已塞過格線，差異測試用種子還原成未塞過的樣子，並放入 0／3／其他值的混合。
 */
#include "gin3_reg.h"

void f_1588_078C(Mem *m)
{
    uint8_t row = 0;

    while (row < 0x66u) {
        uint16_t off = (uint16_t)(rd16(m, DS(0x3C1A)) + (uint16_t)((uint16_t)row << 7));
        uint16_t seg = rd16(m, DS(0x3C1C));
        uint8_t col = 0;

        while (col < 0x80u) {
            if (rd8(m, LIN(seg, off)) == 0)
                wr8(m, LIN(seg, off), 4);
            else if (rd8(m, LIN(seg, off)) == 3)
                wr8(m, LIN(seg, off), 2);
            off = (uint16_t)(off + 0x0Fu);
            col = (uint8_t)(col + 0x0Fu);
        }
        row++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_1588_078C(m);
    return 0;
}
GIN3_REGISTER("1588_078C", 0, run)
