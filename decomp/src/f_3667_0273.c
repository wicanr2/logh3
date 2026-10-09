/*
 * 3667:0273  選單項目表的名稱字串，逐字寫進視窗內容緩衝區（far）
 *
 * 原版：far 呼叫，一個 far 指標參數 obj（位移在前、段在後），無回傳值。
 *   堆疊：序言 push bp／si／di、sub sp,12h、mov bp,sp，所以參數在 [bp+1Ch]（位移）與 [bp+1Eh]（段）。
 *   obj[+0B] byte  內容格寬 w（欄數），obj[+0C] byte  內容格高 h（列數）
 *   obj[+0E] word、obj[+10] word  內容緩衝區 far 指標 p0（位移、段）
 *   obj[+1D] byte  顏色 A，obj[+1F] byte  顏色 B
 *   obj[+20] word、obj[+22] word  選單項目表 far 指標 q（位移、段）；項目 6 bytes：[+0] 旗標、[+2] word／[+4] word
 *     名稱字串 far 指標 r（位移、段）
 *   for (row = 0; row < obj[+0C]; row++)：
 *       c ＝ p0（目前列的緩衝區起點）；r ＝ q[+2]／q[+4]；
 *       for (col = 0; col < obj[+0B]; col++)：
 *           b0 ＝ r[0]；若 b0 != 0：
 *               c 處的 word ← (b0 << 8) + r[1]（小端：c[0] ＝ r[1]、c[1] ＝ b0）；
 *               旗標 ＝ q[0]（寫了 word 之後才讀）；c[+2] ← (旗標 != 0) ? obj[+1D] : obj[+1F]（同樣在 word 寫了之後才讀）；
 *               r 的位移 += 2。
 *           （b0 == 0：不寫 c、r 不前進，只有 c 前進；這一列剩下的欄位全部略過，因為 r 不動。）
 *           c 的位移 += 4。
 *       p0 的位移 += obj[+0B] × 4（16 位元）；q 的位移 += 6。
 *   obj[+0B]、[+0C]、[+1D]、[+1F]、q[0]、r[0]、r[1] 全部在用到的當下才讀，不快取；
 *   緩衝區若與 obj、項目表或字串重疊，寫入會改變後面的圈數、顏色、旗標與字元，本版照原版順序重讀。
 *   word 存取在位移 0xFFFF 時高位元組取同一段的位移 0（rd16s／wr16s）；byte 存取的位移 16 位元回繞。
 * 語意：docs/re/117 第 26 到 30 行「一筆選單項目：六個位元組：可選旗標（byte，0 ＝ 反白畫成不可選）＋補一個 0 ＋
 *   指向名稱的遠指標」，docs/re/137 第 120 到 130 行（選單物件 0x24 bytes，比訊息框多 [+1B]、[+1D] 色等欄位）。
 *   名稱字串是雙位元組字碼序列（b0、b1 ＝ 一個字碼），一格 4 bytes（字碼 word、顏色 byte、旗標 byte，
 *   docs/re/137）；旗標非 0 取顏色 A（選單呼叫參數 0x97），旗標 0 取顏色 B（停用色 0x98，docs/internals/58 第 143 行）。
 *   「b0 == 0 時 r 不前進」會讓字串遇到前導 0 的字碼就整列停筆，這是 asm 行為；字碼 b0 實際是否可為 0
 *   （半形字元存法）未查證。名稱「顏色 A／B」的對應是強推論，位移、運算與順序已證實。
 * 寫入落在線性 0xF0000 以上（BIOS ROM 區）時，dosgolem 安靜丟棄；本版逐 byte 用 put8 表達。
 *   寫入或讀取落在 0xA0000 到 0xBFFFF 的案例由探針丟棄。
 *
 * 輸入域：obj 任意（含緩衝區蓋到 obj、項目表與字串）；步數 ≤ 描述檔 step_budget（1,000,000，約 28,000 格）。
 * 推論等級：位移與運算已證實（反組譯逐指令＋差異測試）；欄位語意為強推論。
 */
#include "gin3_reg.h"

/* 段內回繞的 word 讀取：seg:FFFF 的高位元組在 seg:0000（dosgolem internal/cpu/cpu.go）。 */
static uint16_t f3667_0273_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

/* 線性 0xF0000 以上是 ROM：寫入無效（dosgolem internal/machine/bios_rom.go）。 */
static void f3667_0273_put8(Mem *m, uint32_t lin, uint8_t v)
{
    if ((lin & 0xFFFFFu) < 0xF0000u)
        wr8(m, lin, v);
}

static void f3667_0273_wr16s(Mem *m, uint16_t seg, uint16_t off, uint16_t v)
{
    f3667_0273_put8(m, LIN(seg, off), (uint8_t)v);
    f3667_0273_put8(m, LIN(seg, (uint16_t)(off + 1u)), (uint8_t)(v >> 8));
}

void f_3667_0273(Mem *m, uint16_t obj_off, uint16_t obj_seg)
{
    uint16_t p0_off, p0_seg, q_off, q_seg, c_off, c_seg, r_off, r_seg, ax;
    uint8_t row, col, b0, al;

    p0_seg = f3667_0273_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x10u));
    p0_off = f3667_0273_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x0Eu));
    q_seg = f3667_0273_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x22u));
    q_off = f3667_0273_rd16s(m, obj_seg, (uint16_t)(obj_off + 0x20u));
    row = 0;
    while (row < rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Cu)))) {
        c_seg = p0_seg;
        c_off = p0_off;
        r_seg = f3667_0273_rd16s(m, q_seg, (uint16_t)(q_off + 4u));
        r_off = f3667_0273_rd16s(m, q_seg, (uint16_t)(q_off + 2u));
        col = 0;
        while (col < rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Bu)))) {
            b0 = rd8(m, LIN(r_seg, r_off));
            if (b0 != 0) {
                ax = (uint16_t)((uint16_t)(b0 << 8) + rd8(m, LIN(r_seg, (uint16_t)(r_off + 1u))));
                f3667_0273_wr16s(m, c_seg, c_off, ax);
                if (rd8(m, LIN(q_seg, q_off)) != 0)
                    al = rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x1Du)));
                else
                    al = rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x1Fu)));
                f3667_0273_put8(m, LIN(c_seg, (uint16_t)(c_off + 2u)), al);
                r_off = (uint16_t)(r_off + 2u);
            }
            c_off = (uint16_t)(c_off + 4u);
            col++;
        }
        p0_off = (uint16_t)(p0_off + (uint16_t)(rd8(m, LIN(obj_seg, (uint16_t)(obj_off + 0x0Bu))) << 2));
        row++;
        q_off = (uint16_t)(q_off + 6u);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_3667_0273(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("3667_0273", 2, run)
