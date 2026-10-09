/*
 * 3BB6:023F  以 DH 為索引查模組資料段的對照表 ds:01B4，把表值放進 DL（near，暫存器傳參）
 *
 * 原版：mov al,dh；mov bx,1B4h；xlat；mov dl,al；retn。
 *   輸入：DH ＝ 索引（0..255，無號）。xlat 讀 byte ds:(BX ＋ AL) ＝ ds:(0x01B4 ＋ DH)，最大 0x02B3，不回繞。
 *   輸出：DL ＝ 表值、AL ＝ 表值（AH 不動）、BX ＝ 0x01B4、DH 不動。其他暫存器不動，不寫記憶體。
 *   描述檔 ret ＝ pack:dx,ax,bx，把 DX／AX／BX 三個暫存器都比對（輸入的 DX 與 AX 是 reg 參數，C 版依輸入算出輸出，
 *   AH 與 DH 原樣穿透）。呼叫者實際只用 DL（3BB6:0212、3BB6:054F 等在呼叫後 `and dl,0DFh`、`or dl,...`）。
 *   DS：呼叫者載入 seg130（dosgolem 下 0x4AFE，載入位置相依），描述檔用 ds_seg 4AFE；C 版把 ds 當參數。
 *
 * 語意（假說）：DH 通常是 `ds:01A6 ＋ 0xB0`（3BB6:054F、3BB6:02BB 的呼叫前置），表值接著與 0xDF／0x20 做 and／or 後交給
 *   3BB6:0248（video-dos 類，寫 I/O 埠）。表的內容與用途（例如某種暫存器位址對照）：未知。
 * 推論等級：xlat 位址、暫存器輸出已證實（逐指令對過）；用途假說。
 * 輸入域：DX 與 AX 的全部 16 位元值（DH 即索引，全部 256 個）；索引表的 256 格由隨機 DH 與突變涵蓋。
 */
#include "gin3_reg.h"

uint64_t f_3BB6_023F(Mem *m, uint16_t ds, uint16_t dx, uint16_t ax)
{
    uint8_t dh = (uint8_t)(dx >> 8);
    uint16_t bx = 0x01B4u;
    uint8_t al = rd8(m, LIN(ds, (uint16_t)(bx + dh))); /* xlat：AL ← ds:[BX ＋ AL]，AL ＝ DH */

    dx = (uint16_t)((dx & 0xFF00u) | al); /* mov dl,al */
    ax = (uint16_t)((ax & 0xFF00u) | al);
    return ((uint64_t)dx << 32) | ((uint64_t)ax << 16) | bx;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3BB6_023F(m, 0x4AFEu, a[0], a[1]); /* 描述檔 ds_seg：seg130；參數順序 dx、ax */
}
GIN3_REGISTER("3BB6_023F", 2, run)
