/*
 * GIN3PS.EXE 28E0:0103  列出可用的提督編號，再依旗標套用最多八道篩選器（原地壓縮清單），回傳留下的個數
 *
 * 原版：far 呼叫，五個堆疊字：清單輸出緩衝區的 far 指標（位移、段）、星域參數 a、旗標 flags、上限 limit。
 *   序言 push bp、sub sp,0Ah、mov bp,sp，所以參數在 [bp+10h]（位移）、[bp+12h]（段）、[bp+14h]（a）、[bp+16h]（flags）、[bp+18h]（limit）。
 *   區域變數：[bp+0] ＝ 個數 n（byte）、[bp+1] ＝ 迴圈索引 i（byte）、[bp+2..5] ＝ 清單寫入指標、[bp+6..9] ＝ 提督記錄指標。
 *   回傳 AL ＝ n；AH 是最後一次被呼叫者留下的殘值（差異測試用 ret = pack:al 只比 AL，呼叫者只存 AL）。8 個呼叫者。
 *
 * 語意（已證實：與反組譯逐條對過）：
 *   1. 建清單：i 從 0 到 0x51（byte 無號，`cmp al,52h；jnb`），提督記錄 ds:B77C 起每筆 28 bytes；記錄[+0] != 0xFE 的 i 依序寫進
 *      清單（byte），n ← 留下的個數。
 *   2. a（只取低位元組）!= 0xFF：n ← 28E0:0227(清單, n, a)（保留 提督[+0x0A] == a 者）。
 *   3. flags 的低位元組（`test byte [bp+16h]`）依序檢查（高位元組無關）：
 *        bit0：n ← 28E0:028E(清單, n, ds:8F1D)（保留 提督[+0x01] == v）
 *        bit1：n ← 28E0:02F5(清單, n, ds:8F1D)（保留 提督[+0x01] != v）
 *        bit2：n ← 28E0:035C(清單, n)（保留 提督[+0x00] == 0xFF）
 *        bit3：n ← 28E0:03B6(清單, n)（保留 提督[+0x00] != 0）
 *        bit4：n ← 28E0:0410(清單, n)（保留 (+0x09 & 0x80) == 0 且 +0x09 != 0x7F）
 *        bit5：n ← 28E0:047E(清單, n)（保留 +0x09 的 bit7 有設者）
 *        bit6：n ← 28E0:04E3(清單, n, limit)（保留 +0x06 < limit 者，無號 byte）
 *      bit7 不使用。每道篩選器的結果（AL）存回 n。
 *   4. 回 n。
 *
 * 呼叫被呼叫者的方式（逐一對過 asm）：每道篩選器都是 far 呼叫，參數從右往左推：先推（需要的話）比較值，再推 `word ptr [bp+0]`
 *   （n 與 i 兩個 byte 並成一個 word：低位元組 n、高位元組 i = 0x52，被呼叫者只取低位元組當項數），再推清單段與位移。
 *   0227 的比較值是 `push ax`（AL = a，AH 殘值，被呼叫者只取低位元組）；028E、02F5 的比較值是 `push word ptr ds:8F1D`；
 *   04E3 的上限是 `push word ptr [bp+18h]`。
 *
 * 欄位語意（強推論）：docs/re/208（28E0:0103 的星域參數 != 0xFF 時先過 0227 那一道）；0227 以後各道篩選器的語意見各自的檔頭。
 *   提督記錄[+0] == 0xFE 是空位（本函式只認這個值）。
 *
 * 輸入域：清單緩衝區放暫存區（至少 82 bytes，被呼叫者就地壓縮）；a、flags、limit 的全部 16 位元值（a 與 flags 只取低位元組）；
 *   提督記錄 82 筆與 ds:8F1D 由突變與種子涵蓋。
 * 推論等級：已證實（反組譯逐指令＋差異測試，八道篩選器各有差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_28E0_0227(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t want_w);
extern uint16_t f_28E0_028E(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t want_w);
extern uint16_t f_28E0_02F5(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t want_w);
extern uint16_t f_28E0_035C(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w);
extern uint16_t f_28E0_03B6(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w);
extern uint16_t f_28E0_0410(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w);
extern uint16_t f_28E0_047E(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w);
extern uint16_t f_28E0_04E3(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t limit_w);

uint8_t f_28E0_0103(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t a_w, uint16_t flags_w, uint16_t limit_w)
{
    uint8_t n = 0;
    uint8_t i = 0;
    uint16_t out = list_off;
    uint16_t rec = 0xB77Cu;
    uint8_t flags = (uint8_t)flags_w;

    while (i < 0x52u) {
        if (rd8(m, DS(rec)) != 0xFEu) {
            wr8(m, LIN(list_seg, out), i);
            out = (uint16_t)(out + 1u);
            n = (uint8_t)(n + 1u);
        }
        i = (uint8_t)(i + 1u);
        rec = (uint16_t)(rec + 0x1Cu);
    }
    if ((uint8_t)a_w != 0xFFu)
        n = (uint8_t)f_28E0_0227(m, list_off, list_seg, (uint16_t)(n | ((uint16_t)i << 8)), (uint16_t)(uint8_t)a_w);
    if ((flags & 1u) != 0)
        n = (uint8_t)f_28E0_028E(m, list_off, list_seg, (uint16_t)(n | ((uint16_t)i << 8)), rd16(m, DS(0x8F1Du)));
    if ((flags & 2u) != 0)
        n = (uint8_t)f_28E0_02F5(m, list_off, list_seg, (uint16_t)(n | ((uint16_t)i << 8)), rd16(m, DS(0x8F1Du)));
    if ((flags & 4u) != 0)
        n = (uint8_t)f_28E0_035C(m, list_off, list_seg, (uint16_t)(n | ((uint16_t)i << 8)));
    if ((flags & 8u) != 0)
        n = (uint8_t)f_28E0_03B6(m, list_off, list_seg, (uint16_t)(n | ((uint16_t)i << 8)));
    if ((flags & 0x10u) != 0)
        n = (uint8_t)f_28E0_0410(m, list_off, list_seg, (uint16_t)(n | ((uint16_t)i << 8)));
    if ((flags & 0x20u) != 0)
        n = (uint8_t)f_28E0_047E(m, list_off, list_seg, (uint16_t)(n | ((uint16_t)i << 8)));
    if ((flags & 0x40u) != 0)
        n = (uint8_t)f_28E0_04E3(m, list_off, list_seg, (uint16_t)(n | ((uint16_t)i << 8)), limit_w);
    return n;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_0103(m, a[0], a[1], a[2], a[3], a[4]); }
GIN3_REGISTER("28E0_0103", 5, run)
