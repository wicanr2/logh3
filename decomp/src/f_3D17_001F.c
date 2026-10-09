/*
 * 3D17:001F  seg:off 換成 21 位元線性位址的片段（3D17:0003 的前半）
 * 呼叫：far，暫存器傳參。輸入 AX＝段、BX＝位移。CX 原值保留（push／pop）。
 * 輸出：BX ＝ 線性位址的低 16 位元；AX ＝ 線性位址 bit16 以上（0..16，AH 恆為 0）。
 *   線性位址 ＝ 段 × 16 ＋ 位移，不遮到 20 位元（段 0xFFFF、位移 0xFFFF 時 AX＝0x10）。
 * 做法：rol ax,4 之後 AL 低半是段的高 4 位元，其餘是 (段 << 4) 的低 16 位元；BX 加上後，進位加進 AL。
 * 呼叫者只有 3D17:0003（正規化 far 指標）。旗標沒有被呼叫者使用，不在比對範圍。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

uint64_t f_3D17_001F(uint16_t ax, uint16_t bx, uint16_t cx)
{
    uint16_t rot = (uint16_t)((ax << 4) | (ax >> 12)); /* rol ax,4 */
    uint16_t lo = (uint16_t)(rot & 0xFu);              /* and ax,0Fh；AH 因此為 0 */
    uint32_t sum = (uint32_t)bx + (uint16_t)(rot ^ lo);

    bx = (uint16_t)sum;
    ax = (uint16_t)((lo + (sum >> 16)) & 0xFFu); /* adc al,ah（ah ＝ 0） */
    return ((uint64_t)ax << 32) | ((uint64_t)bx << 16) | cx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_3D17_001F(a[0], a[1], a[2]);
}
GIN3_REGISTER("3D17_001F", 3, run)
