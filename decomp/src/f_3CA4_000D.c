/*
 * 3CA4:000D  strncpy(dst, src, n)（Datalight C runtime）
 * 呼叫：far，堆疊傳參：dst 位移、dst 段、src 位移、src 段、n（word）。回傳 AX ＝ dst 段、BX ＝ dst 位移。
 * 做法：k ＝ 在 src 掃 NUL 所得的位元組數（含 NUL；找不到時 65535）。
 *   c ＝ (n >= k) ? k : n（無號比較），複製 c 個位元組（由低往高逐位元組），
 *   再從 dst + c 起補 n − c 個 0。所以 src 夠短時連 NUL 一併複製並補零到 n；src 太長時只複製 n 個位元組，不補 NUL。
 * 位移在段內 16 位元回繞。語意等價於標準 C 的 strncpy（回傳值為 seg:off 對）。
 * 找不到 NUL 的 64 KB 邊界（掃描上限 65535 個位元組）由描述檔 pokes 的 repeat 種子涵蓋。推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

uint32_t f_3CA4_000D(Mem *m, uint16_t doff, uint16_t dseg, uint16_t soff, uint16_t sseg, uint16_t n)
{
    uint16_t k = 0, cx = 0xFFFF, c, pad, i;

    while (cx != 0) {
        uint8_t b = rd8(m, LIN(sseg, (uint16_t)(soff + k)));

        k++;
        cx--;
        if (b == 0)
            break;
    }
    c = (n >= k) ? k : n;
    pad = (uint16_t)(n - c);
    for (i = 0; i < c; i++)
        wr8(m, LIN(dseg, (uint16_t)(doff + i)), rd8(m, LIN(sseg, (uint16_t)(soff + i))));
    for (i = 0; i < pad; i++)
        wr8(m, LIN(dseg, (uint16_t)(doff + c + i)), 0);
    return ((uint32_t)dseg << 16) | doff; /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3CA4_000D(m, a[0], a[1], a[2], a[3], a[4]);
}
GIN3_REGISTER("3CA4_000D", 5, run)
