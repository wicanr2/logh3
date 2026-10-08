/*
 * GIN3PS.EXE 2AF5:00B6  還原「目前星域的行星切片」（遠指標與顆數）
 *
 * 原版：far 呼叫，沒有參數：
 *   ds:CFF8（位移）← ds:51B0；ds:CFFA（段）← ds:51B2         ; 切片遠指標，由備份還原
 *   ds:CFFC（byte）← ds:CFFD                                  ; 切片的行星顆數，由備份還原
 * 備份由 2AF5:0036 設定：舊值存到 ds:51B0／51B2 與 ds:CFFD（docs/internals/29、docs/re/330：已證實）。
 * 回傳：沒有回傳值。離開時 AX、BX 是殘值：BX＝ds:51B0 的值，
 *   AX＝(ds:51B2 的高位元組 << 8) | 剛複製的顆數 byte（mov al,ds:CFFD 只換掉 AL）。
 *   13 個呼叫者中至少 6 個緊接著改寫 AX（mov ax,… 或 mov al,…）或呼叫其他函式，其餘是否使用殘值未逐一追查（未知）；
 *   差異測試比較 AX:BX 以忠實表達殘值。
 * 副作用：寫 ds:CFF8（word）、ds:CFFA（word）、ds:CFFC（byte）。
 * 輸入域：無參數，記憶體內容全部可到達。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint32_t f_2AF5_00B6(Mem *m)
{
    uint16_t saved_seg = rd16(m, DS(0x51B2));
    uint16_t saved_off = rd16(m, DS(0x51B0));
    uint8_t saved_count;

    wr16(m, DS(0xCFFA), saved_seg);
    wr16(m, DS(0xCFF8), saved_off);
    saved_count = rd8(m, DS(0xCFFD));
    wr8(m, DS(0xCFFC), saved_count);
    return ((uint32_t)(uint16_t)((saved_seg & 0xFF00u) | saved_count) << 16) | saved_off; /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2AF5_00B6(m);
}
GIN3_REGISTER("2AF5_00B6", 0, run)
