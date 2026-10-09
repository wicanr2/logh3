/*
 * GIN3PS.EXE 37A9:01F9  游標在指定視窗內時，找出清單中第一個涵蓋游標所在格的項目（回項目的遠指標，沒有則回 0:0）
 *
 * 原版：far 呼叫，六個 word 的堆疊參數（序言 push bp／si／di、sub sp,2、mov bp,sp，所以第一個參數在 [bp+0Ch]），由低位址到高位址：
 *   win 位移 [bp+0Ch]、win 段 [bp+0Eh]（視窗結構的遠指標）、list 位移 [bp+10h]、list 段 [bp+12h]（項目清單的遠指標）、
 *   x [bp+14h]、y [bp+16h]（畫面座標）。回傳 AX:BX ＝ 項目的段:位移，找不到為 0:0。逐指令：
 *   1. AX:BX ← 341B:039D(x, y)。若 AX ≠ win 段 或 BX ≠ win 位移 → 回 0:0       ; 命中的視窗不是呼叫者指定的那一個（段、位移分開比，不比線性位址）
 *   2. 341B:044A(win, x, y, out_x ＝ SS:BP+0, out_y ＝ SS:BP+1)                  ; 兩個輸出是區域變數 [bp+0]、[bp+1]（byte），即游標在 win 內的格座標
 *   3. 迴圈（p ＝ list；每圈後只有位移加 0x0C，16 位元回繞，段不變）：
 *        p[0] ＝ 0xFF → 回 0:0；p[1] ＝ 0xFF → 回 0:0                            ; 結尾標記（x 或 y 任一為 0xFF）
 *        p[4] ＝ 0 → 下一筆                                                      ; 停用的項目
 *        cell_x < p[0]（無號 byte）→ 下一筆；p[0] ＋ p[2] <= cell_x（16 位元無號）→ 下一筆
 *        cell_y < p[1]（無號 byte）→ 下一筆；p[1] ＋ p[3] <= cell_y（16 位元無號）→ 下一筆
 *        否則回 AX ＝ list 段、BX ＝ p 的位移（項目的遠指標）
 *   每圈重讀 cell_x、cell_y（區域變數）；項目每筆 12 bytes：[0] 欄、[1] 列、[2] 寬、[3] 高、[4] 啟用旗標，其餘 [5..11] 本函式不讀。
 *   欄／列／寬／高與啟用旗標的名稱是強推論（項目是矩形熱區：半開區間 [p0, p0＋p2) × [p1, p1＋p3)；042F:0142 的檔頭另記
 *   「熱區清單 ds:03D8」），其他欄位：未知。
 * 呼叫者（台帳）：27 個（042F:0391、04C8:000A、0654:07F7、06E8:017C／01E8、073F:0AFD、07F9:085D、090A:0098、0A02:08E7、0B20:0336、
 *   0C0A:015D／0351、0C7F:0148／03CB、0D05:0141／0388、0D92:047C、0DEA:0F80／16AB、0F7A:0445、0FF4:087A、113A:0899、1373:014D／026E、
 *   2AF5:0226、3826:023B、3A9C:009D）。
 *
 * 機器模型：C 版沒有堆疊。兩個輸出區域變數放在 SS:1F76／1F77，這是原版的實際位置（基底狀態 SP ＝ 1F8A，減返回位址 4、六個參數 12、
 *   push bp／si／di 6、sub sp,2 → bp ＝ 1F76），SS 固定是基底狀態的 0x63EB。這個位址在探針的堆疊帶（SS:1B8A..1F89）內，
 *   寫入集合比對時雙方都排除。原版每圈改寫堆疊上的 list 位移參數，C 版用區域變數。位移加法（p＋1…＋4、＋0x0C）都是 16 位元回繞，
 *   以段內位移讀取（seg:FFFF 之後是 seg:0000）。
 * 輸入域：全部 16 位元值；串列要有結尾標記（沒有則原版繞圈，被探針以步數上限丟棄）。視窗串列要是良構。清單（list）與視窗結構不涵蓋
 *   指向堆疊帶（SS:1B8A..1F89）的案例，那會讓原版改掉自己的區域變數。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位名稱與用途為強推論／未知。
 */
#include "gin3_reg.h"

extern uint32_t f_341B_039D(Mem *m, uint16_t x, uint16_t y);
extern void f_341B_044A(Mem *m, uint16_t woff, uint16_t wseg, uint16_t x, uint16_t y,
                        uint16_t xo, uint16_t xs, uint16_t yo, uint16_t ys);

#define F37A9_01F9_SS 0x63EBu /* fcmd-base.state 的 SS（diffprobe 無法設定） */
#define F37A9_01F9_BP 0x1F76u /* 原版 mov bp,sp 之後的 BP（區域變數 [bp+0]、[bp+1]） */

/* 段內位移讀取：位移加法 16 位元回繞 */
static uint8_t f37A9_01F9_b(const Mem *m, uint16_t seg, uint16_t off, uint16_t k)
{
    return rd8(m, LIN(seg, (uint16_t)(off + k)));
}

uint32_t f_37A9_01F9(Mem *m, uint16_t win_off, uint16_t win_seg, uint16_t list_off, uint16_t list_seg,
                     uint16_t x, uint16_t y)
{
    uint32_t hit = f_341B_039D(m, x, y); /* AX:BX ＝ 段:位移 */
    uint16_t p = list_off;
    unsigned n;

    if ((uint16_t)(hit >> 16) != win_seg || (uint16_t)hit != win_off)
        return 0;
    f_341B_044A(m, win_off, win_seg, x, y,
                F37A9_01F9_BP, F37A9_01F9_SS, (uint16_t)(F37A9_01F9_BP + 1u), F37A9_01F9_SS);
    for (n = 0; n < 0x4001u; n++) { /* p 以 0x0C 前進，位移 16 位元回繞，整圈（16384 筆）後回到原點；此界線只讓突變版不當機 */
        uint8_t px = f37A9_01F9_b(m, list_seg, p, 0);

        if (px == 0xFFu)
            return 0;
        if (f37A9_01F9_b(m, list_seg, p, 1) == 0xFFu)
            return 0;
        if (f37A9_01F9_b(m, list_seg, p, 4) != 0
            && !(rd8(m, LIN(F37A9_01F9_SS, F37A9_01F9_BP)) < px)
            && !((uint16_t)(px + f37A9_01F9_b(m, list_seg, p, 2)) <= rd8(m, LIN(F37A9_01F9_SS, F37A9_01F9_BP)))
            && !(rd8(m, LIN(F37A9_01F9_SS, (uint16_t)(F37A9_01F9_BP + 1u))) < f37A9_01F9_b(m, list_seg, p, 1))
            && !((uint16_t)(f37A9_01F9_b(m, list_seg, p, 1) + f37A9_01F9_b(m, list_seg, p, 3))
                 <= rd8(m, LIN(F37A9_01F9_SS, (uint16_t)(F37A9_01F9_BP + 1u)))))
            return ((uint32_t)list_seg << 16) | p; /* AX:BX */
        p = (uint16_t)(p + 0x0Cu);
    }
    return 0; /* 繞整圈都沒有結尾標記也沒有符合的項目：原版不終止（探針以步數上限丟棄），域外 */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_37A9_01F9(m, a[0], a[1], a[2], a[3], a[4], a[5]);
}
GIN3_REGISTER("37A9_01F9", 6, run)
