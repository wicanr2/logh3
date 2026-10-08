/*
 * 2376:0EE3  由編號算 15 bytes 一筆的表內位移，並把 ES 設成 DS（near，暫存器傳參）
 *
 * 輸入：DL（DH 被 xor dh,dh 清掉）。
 * 輸出：SI = 0x82D3 + 15 * DL；DX = DL（DH = 0）；AX = DS；ES = DS（mov ax,ds；mov es,ax）。
 *       旗標被 shl／sub／add 改過，唯一呼叫者 2376:0E0D 不使用。
 *
 * 比對範圍：AX、DX、SI。ES 不在 diffprobe 的 ret 可選暫存器內（regIndex 沒有 es），
 *   所以「ES = DS」只由反組譯確認，差異測試沒有驗到。這是框架限制，見回報。
 * DS：harness 固定把 DS 設為 DGROUP（GIN3_DS），AX 比對到的就是這個值；
 *   原版呼叫情境的 DS 未另外確認，C 版把 DS 當參數 ds 傳入。
 *
 * 語意（強推論）：82D3 起是半形字型表，一字 15 bytes（15 列 * 8 點）。證據：唯一呼叫者 2376:0E0D 在
 *   DH < 0x80 時，對 DL 與 DH 各呼叫本函式一次（DX 裝兩個單位元組字碼）；DH >= 0x80 時改走 2376:0F2F
 *   的雙位元組路徑，那條路徑每字讀 0x1E = 30 bytes，剛好是 15 bytes 的兩倍。
 *   表的實際內容與 DS 是否為 DGROUP 都沒有逐位元組驗證：未知。
 * 輸入域：DX 的全部 16 位元值；本函式不讀寫記憶體。
 */
#include "gin3_reg.h"

uint64_t f_2376_0EE3(uint16_t ds, uint16_t dx)
{
    uint16_t idx = (uint16_t)(dx & 0xFFu); /* xor dh,dh */
    uint16_t si = (uint16_t)(idx * 15u + 0x82D3u); /* shl si,4；sub si,dx */

    /* 回傳依描述檔 pack:ax,dx,si：由高位到低位 */
    return ((uint64_t)ds << 32) | ((uint64_t)idx << 16) | si;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_2376_0EE3(GIN3_DS, a[0]);
}
GIN3_REGISTER("2376_0EE3", 1, run)
