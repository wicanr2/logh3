/*
 * 35B6:0104  視窗矩形內屬於本視窗槽位的畫面格，byte0 設上 bit7（標髒）（far）
 *
 * 原版：far 呼叫，一個 far 指標參數 obj（位移在前、段在後），無回傳值。
 *   堆疊：序言 push bp／si／di、sub sp,0Ah、mov bp,sp，所以參數在 [bp+14h]（位移）與 [bp+16h]（段）。
 *   obj[+00] byte  種類：等於 1 就直接返回（什麼都不做）。
 *   obj[+04] byte  比對值，obj[+05] byte  起始欄 x，obj[+06] byte  起始列 y，
 *   obj[+07] byte  寬（欄數），obj[+08] byte  高（列數）。
 *   格表是 ds:D81E 起的畫面格表（一格 6 bytes、一列 0xF0 bytes ＝ 40 格）。
 *   起點 ＝ DS:(0xD81E + y × 0xF0 + x × 6)（全部 16 位元回繞）；
 *   for (row = 0; row < obj[+08]; row++)：
 *       c ＝ 列起點；for (col = 0; col < obj[+07]; col++)：
 *           若 (byte0(c) & 0x0F) == obj[+04]：byte0(c) |= 0x80；c += 6。
 *       列起點 += 0xF0。
 *   obj[+04]、[+07]、[+08] 每一圈都從 obj 重讀（`les bx,[bp+14h]` 再取欄位），不是只讀一次。
 *   obj 若蓋到格表（格的 byte0 與 obj 的欄位重疊），寫入 bit7 會改到比對值、寬或高，後面的圈數與比對隨之改變；
 *   本版照原版順序重讀。row、col 是 byte 計數器（無號比較），obj[+07]、[+08] 最大 255。
 * 語意（強推論，欄位表見 docs/internals/47 第 74 到 91 行）：obj 是視窗物件，[+0] ＝ 1 表示已建立但尚未顯示、2 表示已顯示
 *   （第 74 行），[+4] ＝ 槽位編號（第 79 行），[+5]／[+6] ＝ 格 x／y，[+7]／[+8] ＝ 格寬／格高（第 80、81 行）；
 *   畫面格表一格 byte0 的 bit7 ＝ 髒、低四位 ＝ 擁有者槽位（第 91 行）。本函式因此是「把已顯示視窗矩形內屬於該視窗槽位的格
 *   標髒」。起點算式與 docs/re/137 第 171 行 `ds:D81E ＋ 視窗[+6]×0xF0 ＋ 視窗[+5]×6` 相同。
 *   呼叫者 341B:02E5（sub_41005+0x29）。名稱「標髒」「槽位」是文件推論；位移、運算與比較已證實。
 * 不碰 ROM 區、視訊記憶體區：目標永遠在 DS 內（位移 16 位元回繞，線性不超過 0x64B2F）。
 *
 * 輸入域：obj 任意（含蓋到格表）；步數 ≤ 描述檔 step_budget（1,000,000，約 255 × 255 格）。
 * 推論等級：位移與運算已證實（反組譯逐指令＋差異測試）；欄位語意為強推論（依 docs/internals/47）。
 */
#include "gin3_reg.h"

void f_35B6_0104(Mem *m, uint16_t obj_off, uint16_t obj_seg)
{
    uint16_t di, cx, row_off, cell_off;
    uint8_t row, col;

    if (rd8(m, LIN(obj_seg, obj_off)) == 1)
        return;
    di = (uint16_t)(rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 5u))) * 6u); /* mul cx，cx = 6 */
    cx = (uint16_t)(rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 6u))) * 0xF0u); /* mul cx，cx = 0xF0 */
    row_off = (uint16_t)(0xD81Eu + cx);
    row_off = (uint16_t)(row_off + di);
    row = 0;
    while (row < rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 8u)))) {
        cell_off = row_off;
        col = 0;
        while (col < rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 7u)))) {
            if ((rd8(m, DS(cell_off)) & 0x0Fu) == rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 4u))))
                wr8(m, DS(cell_off), (uint8_t)(rd8(m, DS(cell_off)) | 0x80u));
            cell_off = (uint16_t)(cell_off + 6u);
            col++;
        }
        row_off = (uint16_t)(row_off + 0xF0u);
        row++;
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_35B6_0104(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("35B6_0104", 2, run)
