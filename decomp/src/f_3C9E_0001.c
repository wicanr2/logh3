/*
 * 3C9E:0001  strlen(s)（Datalight C runtime）
 * 呼叫：far，堆疊傳參：s 位移、s 段。回傳 AX ＝ 長度（不含結尾 NUL）。
 * 做法：AL＝0，CX＝0xFFFF，cld，repne scasb；回傳 ~CX − 1。
 * 掃描最多 65535 個位元組（CX 用完為止），位移在段內 16 位元回繞（DI 回繞）。
 * 找不到 NUL 時掃完 65535 個位元組，回傳 0xFFFE（= 65535 − 1）；NUL 恰好在第 65535 個位元組（索引 65534）時也是 0xFFFE。
 * 這兩個 64 KB 邊界由描述檔的 pokes（repeat）建立 65535／65536 個非零位元組的種子，已用差異測試證實。
 * 所有行為（含位移回繞與 64 KB 上限）：已證實（差異測試）。語意等價於標準 C 的 strlen。
 */
#include "gin3_reg.h"

uint16_t f_3C9E_0001(Mem *m, uint16_t off, uint16_t seg)
{
    uint16_t k = 0, cx = 0xFFFF;

    while (cx != 0) {
        uint8_t b = rd8(m, LIN(seg, (uint16_t)(off + k)));

        k++;
        cx--;
        if (b == 0)
            break;
    }
    return (uint16_t)(k - 1); /* ~cx - 1，cx ＝ 0xFFFF − k */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3C9E_0001(m, a[0], a[1]);
}
GIN3_REGISTER("3C9E_0001", 2, run)
