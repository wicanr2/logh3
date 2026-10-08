/*
 * GIN3PS.EXE 2B5E:000C  設定行星表遠指標（ds:5336／ds:5338）
 *
 * 原版：far 呼叫，沒有參數，無條件寫兩個 word：
 *   ds:5336 ← 2
 *   ds:5338 ← seg131 的段值（重定位後；在 dosgolem 載入位置、映像在 0349 下為 0x5313）
 * ds:5336／5338 是整張行星表的遠指標，執行期行星表在 5313:0002 起，81 筆 × 82 bytes
 * （docs/internals/29、docs/re/264、docs/re/44：已證實）。
 * 回傳：沒有回傳值。離開時 AX＝段值、BX＝2 是寫入前載入暫存器的殘值。
 *   唯一的呼叫者 2AF5:0007 呼叫後直接 mov ax,40h 改寫 AX，不使用（反組譯）。差異測試仍比較 AX:BX
 *   （pack:ax,bx），忠實表達這個殘值。
 * 呼叫者 2AF5:0007 緊接著把 ds:5336／5338 當緩衝區位址傳給 332E:037D（該函式本批未讀，語意未知）。
 * 位置相依：段值 0x5313 是重定位後的結果，與 GIN3_DS（0x54B3）一樣只在 dosgolem 的載入位置成立；
 *   換載入位置要連同 GIN3_DS 一起換。
 * 副作用：只寫上述兩個 word。不讀任何記憶體。輸入域：無參數，全部可到達。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

#define GIN3_SEG131 0x5313u

uint32_t f_2B5E_000C(Mem *m)
{
    wr16(m, DS(0x5336), 2);
    wr16(m, DS(0x5338), GIN3_SEG131);
    return ((uint32_t)GIN3_SEG131 << 16) | 2u; /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2B5E_000C(m);
}
GIN3_REGISTER("2B5E_000C", 0, run)
