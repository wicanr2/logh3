/*
 * 3C9F:0008  strcat(dst, src)（Datalight C runtime）
 * 呼叫：far，堆疊傳參：dst 位移、dst 段、src 位移、src 段。回傳 AX ＝ dst 段、BX ＝ dst 位移。
 * 做法（逐步與原版相同）：
 *   1. 在 dst 找 NUL：AL＝0，CX＝0xFFFF，repne scasb；DI 退一格（dec di）指向該 NUL，存起來。
 *   2. 在 src 找 NUL：同法，CX 取 ~CX ＝ 掃過的位元組數 k（含 NUL；找不到時為 65535）。
 *   3. 從 src 複製 k 個位元組到 dst 的 NUL 處（rep movsb，由低往高逐位元組，來源與目的重疊時照序讀寫）。
 * 長度都是先量好再複製，所以 src 與 dst 重疊時，量到的長度是複製前的內容。位移在段內 16 位元回繞。
 * 語意等價於標準 C 的 strcat（回傳值為 seg:off 對）。找不到 NUL 的 64 KB 邊界（掃描上限 65535 個位元組）由描述檔 pokes 的 repeat 種子涵蓋。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

/* repne scasb（AL＝0，CX＝0xFFFF）：回傳掃過的位元組數，含找到的 NUL；最多 65535 */
static uint16_t f_3C9F_0008_scan_nul(const Mem *m, uint16_t seg, uint16_t off)
{
    uint16_t k = 0, cx = 0xFFFF;

    while (cx != 0) {
        uint8_t b = rd8(m, LIN(seg, (uint16_t)(off + k)));

        k++;
        cx--;
        if (b == 0)
            break;
    }
    return k;
}

uint32_t f_3C9F_0008(Mem *m, uint16_t doff, uint16_t dseg, uint16_t soff, uint16_t sseg)
{
    uint16_t k1 = f_3C9F_0008_scan_nul(m, dseg, doff);
    uint16_t di = (uint16_t)(doff + k1 - 1);
    uint16_t k2 = f_3C9F_0008_scan_nul(m, sseg, soff);
    uint16_t i;

    for (i = 0; i < k2; i++)
        wr8(m, LIN(dseg, (uint16_t)(di + i)), rd8(m, LIN(sseg, (uint16_t)(soff + i))));
    return ((uint32_t)dseg << 16) | doff; /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3C9F_0008(m, a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3C9F_0008", 4, run)
