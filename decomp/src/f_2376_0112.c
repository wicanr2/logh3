/*
 * 2376:0112  畫面格表的矩形查詢：矩形內有沒有「可見、低 4 位元等於 AH、第 2 個 byte 的 bit0 已設」的格（near，暫存器傳參）
 *
 * 輸入（暫存器，DS = DGROUP）：與 2376:00B2 相同——AH = 鍵值 k（AL 不讀）；BL = 起點列、BH = 起點欄；
 *   DL = 列數、DH = 每列格數。唯一呼叫者 2376:0164 的後半（`call 2376:0112` 後 `mov es:[si+1],al`，只存 AL）。
 * 原版（逐指令，結構同 00B2）：
 *   DI ← 2376:0150(BX)；外圈 DL 次、內圈 DH 次（do-while + `loop`）；每格依序檢查
 *     [SI] 的 bit7 已設（test al,80h；jz 跳過）、([SI] & 0x0F) == AH、[SI+1] 的 bit0 已設（test byte [si+1],1）；
 *   三者都成立：AX ← 1 並立即返回；整個矩形都沒有：AX ← 0。
 *   返回 AX：成立路徑 AX = 1（`mov ax,1`），不成立路徑 `xor ax,ax`，所以 AH 恆為 0，比對整個 AX。
 * 輸出：只讀記憶體，沒有寫入。
 *
 * 語意（強推論）：與 00B2（把 [SI] |= 0x80、[SI+1] |= 1）成對——這是「矩形內是否有已標髒、可見、同鍵值的格」的查詢。
 *   2376:0164 用它的結果更新 es:[si+1]（物件的另一個旗標）。名稱為假說。
 * 輸入域：同 2376:00B2。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint64_t f_2376_0150(uint16_t ax, uint16_t cx, uint16_t dx);

uint16_t f_2376_0112(Mem *m, uint16_t ax, uint16_t bx, uint16_t dx)
{
    uint8_t key = (uint8_t)(ax >> 8); /* AH */
    uint16_t di = (uint16_t)f_2376_0150(ax, bx, dx); /* mov ch,bh；mov cl,bl：CX = BX */
    uint16_t rows = (uint16_t)(dx & 0xFFu);          /* mov cl,dl；xor ch,ch */

    do {
        uint16_t si = di;
        uint16_t cols = (uint16_t)(dx >> 8);         /* mov cl,dh；xor ch,ch */

        di = (uint16_t)(di + 0xF0u);
        do {
            uint8_t b = rd8(m, DS(si));

            if ((b & 0x80u) != 0 && (uint8_t)(b & 0x0Fu) == key &&
                (rd8(m, DS((uint16_t)(si + 1u))) & 0x01u) != 0)
                return 1;
            si = (uint16_t)(si + 6u);
        } while ((uint16_t)(--cols) != 0);
    } while ((uint16_t)(--rows) != 0);
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2376_0112(m, a[0], a[1], a[2]); }
GIN3_REGISTER("2376_0112", 3, run)
