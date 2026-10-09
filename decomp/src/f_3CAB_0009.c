/*
 * 3CAB:0009  getenv(name)（Datalight C runtime，在 DOS 環境區塊裡找 NAME=value）
 * 呼叫：far，堆疊傳參：name 位移、name 段（以 NUL 結尾的名稱，大小寫有別）。
 * 回傳：找到時 AX ＝ 環境區塊段、BX ＝ '=' 之後的位移（值的第一個字元）；找不到回 AX＝0、BX＝0。
 * 環境區塊的段來自 PSP：PSP 段存在 ds:00D8，環境段 ＝ PSP:002C 的 word；掃描從該段位移 0 起。
 * 做法（逐步照做）：
 *   DI ＝ 0、CX ＝ 0xFFFF（CX 整個函式只設一次，掃描與比對共用，用完就停止前進）。
 *   重複：SI ＝ name；repe cmpsb（逐位元組比對 name 與目前環境項，不等或 CX 用完就停）；
 *     若環境項在不等處是 NUL（即 es:[DI−1] ＝ 0）就回 0:0；
 *     否則若 name 在該處已結束（ds:[SI−1] ＝ 0）且環境項該處是 '=' 就回 ES:DI；
 *     否則 repne scasb 跳到本環境項的結尾 NUL 之後，繼續比下一項。
 * 注意原版的行為：環境項若是 name 的前綴（例如項目 "FO" 對名稱 "FOO"）會在該項的 NUL 處直接回 0:0，不會再看後面的項目，
 * 環境區塊結尾的空項目（雙 NUL）同樣回 0:0。本函式不寫任何記憶體。
 * 輸入域：環境區塊要是 NUL 結束的字串序列且在 64 KB 內有結尾；沒有結尾的區塊原版會無窮迴圈，探針以步數上限丟棄。
 * 推論等級：已證實（差異測試，環境區塊由種子的 pokes 建立；基底狀態的環境區塊是空的）。
 */
#include "gin3_reg.h"

uint32_t f_3CAB_0009(Mem *m, uint16_t noff, uint16_t nseg)
{
    uint16_t es = rd16(m, LIN(rd16(m, DS(0x00D8)), 0x2C));
    uint16_t di = 0, cx = 0xFFFF, si;

    for (;;) {
        si = noff;
        while (cx != 0) {
            uint8_t a = rd8(m, LIN(nseg, si));
            uint8_t e = rd8(m, LIN(es, di));

            si++;
            di++;
            cx--;
            if (a != e)
                break;
        }
        if (rd8(m, LIN(es, (uint16_t)(di - 1))) == 0)
            return 0;
        if (rd8(m, LIN(nseg, (uint16_t)(si - 1))) == 0) {
            if (rd8(m, LIN(es, (uint16_t)(di - 1))) == '=')
                return ((uint32_t)es << 16) | di;
        }
        while (cx != 0) {
            uint8_t e = rd8(m, LIN(es, di));

            di++;
            cx--;
            if (e == 0)
                break;
        }
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3CAB_0009(m, a[0], a[1]);
}
GIN3_REGISTER("3CAB_0009", 2, run)
