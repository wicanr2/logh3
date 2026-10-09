/*
 * 2376:00E4  畫面格表的矩形計數：把格表中「低 4 位元等於 AH」的格的第 3 個 byte 加一（near，暫存器傳參）
 *
 * 輸入（暫存器，DS = DGROUP）：與 2376:00B2 相同——AH = 鍵值 k（AL 不讀）；BL = 起點列、BH = 起點欄；
 *   DL = 列數、DH = 每列格數。唯一呼叫者 2376:0306；2376:0164 內呼叫 0306 的兩處都在 es:[si+1Eh] == 0 的條件下（asm 讀得）。
 * 原版（逐指令，結構同 00B2，只有每格的動作不同）：
 *   DI ← 2376:0150(BX)（= 0xD81E + 0xF0 * BL + 6 * BH，16 位元回繞）；
 *   外圈 DL 次、內圈 DH 次（都是 do-while + `loop`，0 會跑 65536 圈）：
 *     SI ← DI（每圈外圈開頭）；DI ← DI + 0xF0；每格：([SI] & 0x0F) == AH → inc byte [SI+3]（byte 回繞，0xFF → 0）；SI += 6
 *   AX、CX、旗標是暫用值，呼叫者之後只 `retn`；描述檔 ret = none，只比對記憶體寫入。
 *
 * 語意（假說）：[SI+3] 是格的「疊加計數」（名稱沒有文件記載）；2376:0164 依 es:[si+1Eh] == 0 的條件呼叫 0306，
 *   與 00B2 成對（00B2 標可見＋髒，00E4 加計數），細節未知。docs 沒有逐位元組記載，本檔以反組譯與差異測試為準。
 * 輸入域：同 2376:00B2。SI 恆為偶數，[SI+3] 的位移最多 0xFFFF 不會跨界。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為假說。
 */
#include "gin3_reg.h"

extern uint64_t f_2376_0150(uint16_t ax, uint16_t cx, uint16_t dx);

void f_2376_00E4(Mem *m, uint16_t ax, uint16_t bx, uint16_t dx)
{
    uint8_t key = (uint8_t)(ax >> 8); /* AH */
    uint16_t di = (uint16_t)f_2376_0150(ax, bx, dx); /* mov ch,bh；mov cl,bl：CX = BX */
    uint16_t rows = (uint16_t)(dx & 0xFFu);          /* mov cl,dl；xor ch,ch */

    do {
        uint16_t si = di;
        uint16_t cols = (uint16_t)(dx >> 8);         /* mov cl,dh；xor ch,ch */

        di = (uint16_t)(di + 0xF0u);
        do {
            if ((uint8_t)(rd8(m, DS(si)) & 0x0Fu) == key)
                wr8(m, DS((uint16_t)(si + 3u)), (uint8_t)(rd8(m, DS((uint16_t)(si + 3u))) + 1u));
            si = (uint16_t)(si + 6u);
        } while ((uint16_t)(--cols) != 0);
    } while ((uint16_t)(--rows) != 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2376_00E4(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("2376_00E4", 3, run)
