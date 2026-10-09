/*
 * 2376:00B2  畫面格表的矩形標記：把格表中「低 4 位元等於 AH」的格設成可見並標髒（near，暫存器傳參）
 *
 * 輸入（暫存器，DS = DGROUP）：
 *   AH = 鍵值 k（AL 不讀）；BL = 起點列、BH = 起點欄（呼叫者 2376:02DE 把 es:[si+24h] 的 word 經 xchg bh,bl 後給 BX）；
 *   DL = 列數、DH = 每列格數（同樣來自 es:[si+26h]，xchg dh,dl 後給 DX）。
 * 原版：
 *   1. mov ch,bh；mov cl,bl；call 2376:0150，得 DI = 0xD81E + 0xF0 * BL + 6 * BH（16 位元回繞）。
 *      格表 ds:D81E 起，一格 6 bytes，一列 0xF0 bytes（40 格）。
 *   2. 外圈 DL 次（do-while，`loop`：DL = 0 時 CX 先是 0，減 1 變 0xFFFF，會跑 65536 圈），每圈：
 *        SI ← DI；DI ← DI + 0xF0；內圈 DH 次（同樣 do-while），每格：
 *          若 ([SI] & 0x0F) == AH：[SI] |= 0x80；[SI+1] |= 0x01
 *          SI ← SI + 6
 *   3. 返回。AX、CX 與旗標是暫用值，呼叫者（2376:02DE 與 2376:02F2）之後只 `retn`，不使用它們；
 *      描述檔 ret = none，只比對記憶體寫入。SI、DI 由 push／pop 保存。
 * 所有位址都是 DS 內 16 位元位移（超過 0xFFFF 回繞）。SI 恆為偶數（0xD81E 與 0xF0、6 的倍數相加），
 *   所以 [SI+1] 不會跨越位移 0xFFFF。
 *
 * 語意（強推論）：docs/re/334-annihilation-crash-is-executor.md 第 33 到 34 行：「字元平面 ds:D81E（40×20×6）從 BX
 *   對應的格起，DL 列 × DH 格：([si] & 0Fh) == AH → [si] |= 80h、[si+1] |= 01h」，與本版一致（已證實：
 *   反組譯逐指令）。bit7 與 bit0 的名稱（可見、髒）為假說；AH 是物件 +2 的數值（02DE 的 `mov ah,es:[si+2]`）。
 *
 * 輸入域：AX、BX 的全部 16 位元值；DL >= 1 且 DL × DH 在每案 20 萬步內跑得完（DL = 0 或 DH = 0 時原版繞 65536 圈，
 *   超出步數上限被探針丟棄，本版仍照 do-while 與 16 位元計數器寫，行為一致但不在比較範圍）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint64_t f_2376_0150(uint16_t ax, uint16_t cx, uint16_t dx);

void f_2376_00B2(Mem *m, uint16_t ax, uint16_t bx, uint16_t dx)
{
    uint8_t key = (uint8_t)(ax >> 8); /* AH */
    uint16_t di = (uint16_t)f_2376_0150(ax, bx, dx); /* mov ch,bh；mov cl,bl：CX = BX */
    uint16_t rows = (uint16_t)(dx & 0xFFu);          /* mov cl,dl；xor ch,ch */

    do {
        uint16_t si = di;
        uint16_t cols = (uint16_t)(dx >> 8);         /* mov cl,dh；xor ch,ch */

        di = (uint16_t)(di + 0xF0u);
        do {
            if ((uint8_t)(rd8(m, DS(si)) & 0x0Fu) == key) {
                wr8(m, DS(si), (uint8_t)(rd8(m, DS(si)) | 0x80u));
                wr8(m, DS((uint16_t)(si + 1u)), (uint8_t)(rd8(m, DS((uint16_t)(si + 1u))) | 0x01u));
            }
            si = (uint16_t)(si + 6u);
        } while ((uint16_t)(--cols) != 0);
    } while ((uint16_t)(--rows) != 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2376_00B2(m, a[0], a[1], a[2]);
    return 0;
}
GIN3_REGISTER("2376_00B2", 3, run)
