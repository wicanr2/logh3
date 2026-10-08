/*
 * GIN3PS.EXE 3A92:0062  定點乘加（減）：( [di]*bx - [si]*cx + 0x200 ) 取第 8..23 位元後算術右移 2
 *
 * 呼叫：near，暫存器傳參，沒有堆疊參數、沒有 C 序言（手寫組語，不是 Datalight C 編出來的）。
 *   輸入 SI、DI ＝ ds 內的位移，各指向一個 16 位元有號數；BX、CX ＝ 16 位元有號數。
 *   輸出 AX。讀 ds:[SI]、ds:[DI] 各一個 word；無寫入。
 *   組語另外弄髒 BX、CX（＝ 第一個乘積的低字／高字）、DX 與旗標，呼叫者 3A92:0001／0042 事先 push bx／cx 再還原，
 *   差異測試只比對 AX。
 *
 * 運算（已證實，差異測試；全部 16 位元暫存器運算的逐步對照）：
 *   P1 = (int16)ds:[SI] * (int16)CX          ; mov ax,[si]; imul cx（單運算元有號乘法，32 位元積）
 *   P2 = (int16)ds:[DI] * (int16)BX          ; mov ax,[di]; imul bx
 *   S  = P2 - P1 + 0x200                     ; sub/sbb 與 add/adc，32 位元回繞
 *   mid = (S >> 8) & 0xFFFF                  ; mov al,ah; mov ah,dl：取 S 的第 8..23 位元
 *   AX = (int16)mid >> 2                     ; sar ax,1 兩次
 *   結果等於 S 的第 10..23 位元，以第 23 位元做符號擴展。|S| < 2^23 時就是 (S >> 10)（算術右移）；
 *   |S| >= 2^23 時 S 的高位元被截掉（與 (int32)S >> 10 不同）。docs/re/409 把範圍限在 0..2047，沒有涵蓋這段，
 *   這裡的差異測試涵蓋了。
 *
 * 語意：
 *   強推論（docs/re/155、docs/internals/41，以及 3A92:0001 的反組譯）：SI＝ds:573A+2a、DI＝ds:573A+2(a+0x40) 時，
 *   [SI]＝sin(a)×1024、[DI]＝cos(a)×1024（256 刻一圈的正弦表），BX＝x、CX＝y，
 *   本函式算 x' ＝ (cos·x − sin·y + 512) >> 10；3A92:0081 算 y'。這個命名（旋轉）未被差異測試證明，
 *   差異測試證明的只有上面的算術。
 *
 * 輸入域：
 *   SI、DI 涵蓋整個 16 位元：約一半取自 choices（正弦表的 256 個位移 0x573A..0x5938，加 0、1、0xFFFE、0xFFFF），其餘為均勻隨機。
 *   位移 0xFFFF 的 word 讀取依 8086／dosgolem 的語意在段內回繞（高位元組取 ds:0000）；真機 286 以上會產生例外，
 *   原版不會這樣呼叫，這裡照執行器的行為建模。
 *   BX、CX 涵蓋整個 16 位元，含負數與 -32768。
 */
#include "gin3_reg.h"

/* ds:off 的 word，兩個位元組的位移各自在 64 KB 段內回繞（與 dosgolem／8086 一致）。 */
static uint16_t rd16w(const Mem *m, uint16_t off)
{
    uint16_t lo = rd8(m, DS(off));
    uint16_t hi = rd8(m, DS((uint16_t)(off + 1)));

    return (uint16_t)(lo | (hi << 8));
}

uint16_t f_3A92_0062(Mem *m, uint16_t si, uint16_t di, uint16_t bx, uint16_t cx)
{
    int32_t p_si = (int32_t)(int16_t)rd16w(m, si) * (int32_t)(int16_t)cx; /* mov ax,[si]; imul cx */
    int32_t p_di = (int32_t)(int16_t)rd16w(m, di) * (int32_t)(int16_t)bx; /* mov ax,[di]; imul bx */
    uint32_t sum = (uint32_t)p_di - (uint32_t)p_si;                       /* sub ax,bx; sbb dx,cx */
    uint16_t mid;

    sum = sum + 0x200u;                                                   /* add ax,200h; adc dx,0 */
    mid = (uint16_t)(sum >> 8);                                           /* mov al,ah; mov ah,dl */
    return (uint16_t)((mid >> 2) | ((mid & 0x8000u) ? 0xC000u : 0u));     /* sar ax,1; sar ax,1 */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    return f_3A92_0062(m, a[0], a[1], a[2], a[3]);
}
GIN3_REGISTER("3A92_0062", 4, run)
