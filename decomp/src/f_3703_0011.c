/*
 * GIN3PS.EXE 3703:0011  把 seg129 內的兩個 word 設成初值：[DDBE] ← 0、[DDC0] ← 0x60BE
 * 呼叫：far，無參數。序言 push es，結尾 pop es，所以 ES 還原。回傳時 AX ＝ 0x60BE（最後一次 mov ax,60BE）。
 * 副作用（asm 逐條）：
 *   es ← 0x3D22（seg129，重定位後的值，見 3703:000D）
 *   es:[DDBE] ← 0（word）
 *   es:[DDC0] ← 0x60BE（word）
 * 線性位址：0x3D220 + 0xDDBE ＝ 0x4AFDE 與 0x4AFE0，在映像的資料區（高於程式碼區 0x3D200），沒有落在 ROM 或視訊記憶體。
 * 基底狀態（fcmd-base.state）這兩格已經是 0 與 0x60BE（初始化過），所以寫入集合要靠種子先把前態 poke 成別的值才看得出來。
 * 兩個欄位的語意未知（只有 5 個呼叫者：341B:0004、0145、01CF、02E5、35D1:026D）。推論等級：位移與值已證實，語意未知。
 * 輸入域：無輸入。
 */
#include "gin3_reg.h"

uint16_t f_3703_0011(Mem *m)
{
    wr16(m, LIN(0x3D22, 0xDDBE), 0x0000);
    wr16(m, LIN(0x3D22, 0xDDC0), 0x60BE);
    return 0x60BE;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_3703_0011(m);
}
GIN3_REGISTER("3703_0011", 0, run)
