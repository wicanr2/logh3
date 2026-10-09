/*
 * 3CA2:0009  strcpy(dst, src)（Datalight C runtime）
 * 呼叫：far，堆疊傳參：dst 位移、dst 段、src 位移、src 段。回傳 AX ＝ dst 段、BX ＝ dst 位移。
 * 做法：先在 src 找 NUL（AL＝0，CX＝0xFFFF，repne scasb），k ＝ ~CX ＝ 掃過的位元組數（含 NUL；找不到時 65535），
 * 再從 src 複製 k 個位元組到 dst（rep movsb，由低往高逐位元組）。長度先量好才複製，重疊時照序讀寫。
 * 位移在段內 16 位元回繞。語意等價於標準 C 的 strcpy（回傳值為 seg:off 對）。
 * 找不到 NUL 的 64 KB 邊界（掃描上限 65535 個位元組）由描述檔 pokes 的 repeat 種子涵蓋。推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

uint32_t f_3CA2_0009(Mem *m, uint16_t doff, uint16_t dseg, uint16_t soff, uint16_t sseg)
{
    uint16_t k = 0, cx = 0xFFFF, i;

    while (cx != 0) {
        uint8_t b = rd8(m, LIN(sseg, (uint16_t)(soff + k)));

        k++;
        cx--;
        if (b == 0)
            break;
    }
    for (i = 0; i < k; i++)
        wr8(m, LIN(dseg, (uint16_t)(doff + i)), rd8(m, LIN(sseg, (uint16_t)(soff + i))));
    return ((uint32_t)dseg << 16) | doff; /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3CA2_0009(m, a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3CA2_0009", 4, run)
