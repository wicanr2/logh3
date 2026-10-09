/*
 * 36AD:02AB  畫面座標換成 VGA 畫面位移 DI（near，無參數，DS = DGROUP；純讀取）
 *
 * 原版（逐指令，全部 16 位元回繞）：
 *   ax = ds:8AE0 << 4（四次 add ax,ax）; dx = ax; ax <<= 2（兩次 add ax,ax）; dx += ax    ; dx = ds:8AE0 * 80
 *   ax = ds:8ADE; ax = sar(ax, 3)（三次 sar ax,1，算術右移）; ax += dx; di = ax; di += 0xC80
 * 輸出：DI = ds:8AE0 * 80 + (有號)(ds:8ADE >> 3) + 0xC80；AX = DI - 0xC80；DX = ds:8AE0 * 80。
 *   AX、DX 是暫用值，唯一呼叫者 36AD:0279 緊接著 call 36AD:0343，後者一進入就改寫 DX 與 AX，只用 DI；
 *   描述檔 ret = pack:ax,dx,di，連暫用值一併比對。
 * 讀 ds:8ADE（X 座標，有號 word）、ds:8AE0（Y 座標，有號 word）：名稱為假說，由 36AD:0279 在呼叫前把
 *   ds:8AD8 − ds:8AF4 存進 8ADE、ds:8ADA − ds:8AF6 存進 8AE0 推得（座標減去裁切原點）。
 * 語意（強推論）：每列 80 bytes（640 像素 / 8），0xC80 = 40 列 × 80，與 2376:09AE 的 DI 算式
 *   （Y × 80 + (X >> 3, 算術) + 0xC80）相同；X、Y 為像素與列的對應見 docs/re/407 的畫面保留列。
 *
 * 輸入域：兩個 word 的全部 16 位元值；沒有指標，不讀視訊記憶體區，不寫記憶體。
 * 推論等級：算式已證實（反組譯逐指令＋差異測試）；欄位名稱為假說。
 */
#include "gin3_reg.h"

uint64_t f_36AD_02AB(Mem *m)
{
    uint16_t ax, dx, di;

    ax = rd16(m, DS(0x8AE0u));
    ax = (uint16_t)(ax << 4); /* 四次 add ax,ax */
    dx = ax;
    ax = (uint16_t)(ax << 2); /* 兩次 add ax,ax */
    dx = (uint16_t)(dx + ax); /* Y * 80 */
    ax = (uint16_t)((int16_t)rd16(m, DS(0x8ADEu)) >> 3); /* 三次 sar ax,1 */
    ax = (uint16_t)(ax + dx);
    di = ax;
    di = (uint16_t)(di + 0xC80u);
    return ((uint64_t)ax << 32) | ((uint64_t)dx << 16) | di;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_36AD_02AB(m);
}
GIN3_REGISTER("36AD_02AB", 0, run)
