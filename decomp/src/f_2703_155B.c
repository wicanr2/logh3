/*
 * GIN3PS.EXE 2703:155B  把一筆艦艇的艘數還回星域的待命存量（TakeRegionShips 的反向）
 *
 * 原版：far 呼叫，兩個 far 指標參數（各佔兩個字，位移在前、段在後）：
 *   第一個＝星域記錄，第二個＝艦艇記錄。無有意義的回傳（AX 是 艦艇[+0x00] × 28 的殘值，呼叫端 2AD2:018C 不用）。
 * 語意（已證實：反組譯逐指令；docs/re/329、spec/031 一致）：
 *   欄位 ＝ 星域 + 0x30 + 艦艇[+0x00] × 28 + 艦艇[+0x04] × 4      （位移 16 位元回繞）
 *   dword(欄位) ＋= 艦艇[+0x08]（word，零擴充成 32 位元；以 add／adc 進位到高字）
 * 欄位名稱：艦艇[+0x00]＝陣營、[+0x04]＝艦種、[+0x08]＝現有艘數，來自 docs/re/215、329；
 *   星域 +0x30 起是 2 陣營 × 7 艦種 × 4 bytes 的待命存量表，來自 docs/internals/22。
 *   艦艇[+0x00]、[+0x04] 當 byte 零擴充使用，不檢查範圍（遊戲裡艦種 7 不呼叫這支，docs/re/329）。
 * 副作用：只寫星域記錄內那一個 dword。
 * 輸入域：兩個指標都指到記錄表內；dword 的兩個字不跨越 64KB 段尾（本框架的記憶體模型不做段內回繞）。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"

void f_2703_155B(Mem *m, uint16_t region_off, uint16_t region_seg, uint16_t ship_off, uint16_t ship_seg)
{
    uint16_t count = rd16(m, LIN(ship_seg, (uint16_t)(ship_off + 8)));
    uint16_t kind4 = (uint16_t)(rd8(m, LIN(ship_seg, (uint16_t)(ship_off + 4))) << 2);
    uint16_t side28 = (uint16_t)(rd8(m, LIN(ship_seg, ship_off)) * 0x1C);
    uint16_t field = (uint16_t)(region_off + 0x30);
    uint16_t lo, hi;
    uint32_t sum;

    field = (uint16_t)(field + side28);
    field = (uint16_t)(field + kind4);
    lo = rd16(m, LIN(region_seg, field));
    hi = rd16(m, LIN(region_seg, (uint16_t)(field + 2)));
    sum = (uint32_t)lo + count;
    wr16(m, LIN(region_seg, field), (uint16_t)sum);
    wr16(m, LIN(region_seg, (uint16_t)(field + 2)), (uint16_t)(hi + (sum >> 16)));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2703_155B(m, a[0], a[1], a[2], a[3]);
    return 0;
}
GIN3_REGISTER("2703_155B", 4, run)
