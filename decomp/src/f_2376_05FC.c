/*
 * 2376:05FC  依 X 的低 3 位元改寫位移程式碼的指令（自我修改程式碼，near，無參數，DS = 模組資料段 3D22）
 *
 * 讀：ds:786 = X（只取低 3 位元，k = X & 7）、ds:782 = 上一次改寫用的 k（快取）。
 * 若 k 等於 ds:782：什麼都不寫，直接返回。
 * 否則：ds:782 ← k，並依 k 把「程式碼段 2376 的兩段指令位元組」改寫成不同的位移指令：
 *     k < 5 （無號）：bx = 0xC089（機器碼 89 C0，mov ax,ax）、dx = 0xC8D1（D1 C8，ror ax,1）
 *     k >= 5       ：k ← 8 - k（結果 3、2、1）、bx = 0xC486（86 C4，xchg ah,al）、dx = 0xC0D1（D1 C0，rol ax,1）
 *   cs:0691 與 cs:06A5 各寫一個 word bx；
 *   然後 si = 0x693、di = 0x6A7，cx = 4 - k：先在 cs:[si]、cs:[di] 各寫 cx 個 word 的 0xC089（mov ax,ax），
 *   si、di 每次加 2；再各寫 k 個 word 的 dx。cx = 0 時第一段整個跳過；k = 0 時第二段整個跳過。
 *   結果：兩段各 5 個 word（cs:0691..069A 與 cs:06A5..06AE），第一個 word 是 bx，其後 (4 - k) 個 mov ax,ax，
 *   最後 k 個 dx。k 的範圍是 0..4（k >= 5 時換成 8 - k = 3..1），所以 cx 與 k 都在 0..4。
 *   AX、BX、CX、DX、SI、DI、旗標在返回時是暫用值（呼叫者 2376:05D6 之後 `pop si` 並重算 cl、dh、dl），
 *   描述檔 ret = none，只比對記憶體寫入。
 *
 * 語意：被改寫的 cs:0691..069A 與 cs:06A5..06AE 落在 2376:05D6 呼叫的 loc_30928（偏移 0x658 起，
 *   到 2376:06CF 之前）內，基底狀態裡那裡的位元組是 `ad 89 c0 89 c0 89 c0 89 c0 89 c0 8a f8`：
 *   lodsw 之後接 5 個 mov ax,ax，也就是 k = 0 時的樣子。這支是依 X 的低 3 位元調整位移次數的自修改程式
 *   （強推論）。docs/re/407-fleet-blocks.md 第 30 行把 2376:08A0 記成「依 X mod 8 改 091A／092E 等六個 bytes，
 *   屬對齊用自修改指令」；05FC 是同一型的另一份（見 f_2376_08A0.c）。原版把結果寫進程式碼段 CS（寫入集合裡
 *   出現的是線性位址 0x23760 + 位移 的位元組）。被改寫的指令在本批次不執行，只比對寫入。
 * 呼叫者：2376:05D6（只在 X 的低 4 位元不為 0 時呼叫）。
 *
 * 程式碼段 CS 在 dosgolem 的載入位置是 0x2376（映像基底 0x0349 之下的重定位後段值）；本檔以參數 cs 傳入，
 * 不寫死在函式內。程式碼段與 DS（3D22）的線性範圍不重疊（0x23760+0xFFFF < 0x3D220），沒有別名。
 *
 * 輸入域：ds:786 與 ds:782 的全部 16 位元值（k 只取 ds:786 的低 3 位元；ds:782 若不在 0..7，永遠不等於 k）。
 * 推論等級：位元組值與位移已證實（反組譯逐指令＋差異測試，種子把改寫區預先填 0xFF 以便看出每個 word）；
 *   各機器碼的助記符（mov ax,ax、ror／rol ax,1、xchg ah,al）來自位元組解碼，用途（X 對齊）為強推論。
 */
#include "gin3_reg.h"

void f_2376_05FC(Mem *m, uint16_t ds, uint16_t cs)
{
    uint16_t k = (uint16_t)(rd16(m, LIN(ds, 0x786u)) & 7u);
    uint16_t bx, dx, si, di, cx;

    if (k == rd16(m, LIN(ds, 0x782u)))
        return;
    wr16(m, LIN(ds, 0x782u), k);
    bx = 0xC089u;
    dx = 0xC8D1u;
    if (k >= 5u) {
        k = (uint16_t)(8u - k);
        bx = 0xC486u;
        dx = 0xC0D1u;
    }
    wr16(m, LIN(cs, 0x0691u), bx);
    wr16(m, LIN(cs, 0x06A5u), bx);
    si = 0x0693u;
    di = 0x06A7u;
    cx = (uint16_t)(4u - k);
    if (cx != 0) {
        bx = 0xC089u;
        do {
            wr16(m, LIN(cs, si), bx);
            wr16(m, LIN(cs, di), bx);
            si = (uint16_t)(si + 2u);
            di = (uint16_t)(di + 2u);
        } while (--cx != 0);
    }
    cx = k;
    if (cx != 0) {
        do {
            wr16(m, LIN(cs, si), dx);
            wr16(m, LIN(cs, di), dx);
            si = (uint16_t)(si + 2u);
            di = (uint16_t)(di + 2u);
        } while (--cx != 0);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    /* ds：描述檔 ds_seg（呼叫者 2376:0538 載入的模組資料段）；cs：dosgolem 載入位置下的 2376 程式段 */
    f_2376_05FC(m, 0x3D22u, 0x2376u);
    return 0;
}
GIN3_REGISTER("2376_05FC", 0, run)
