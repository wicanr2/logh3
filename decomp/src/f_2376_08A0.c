/*
 * 2376:08A0  依 X 的低 3 位元改寫位移程式碼的指令（自我修改程式碼，near，無參數，DS = 模組資料段 3D22）
 *
 * 與 2376:05FC 是平行實作（見 f_2376_05FC.c），只有三處不同：快取變數是 ds:784（05FC 用 ds:782）、
 * 改寫位址是 cs:091A／cs:092E 起（05FC 是 cs:0691／cs:06A5 起）、翻轉的分界寫成 `cmp ax,4; jbe`
 * （k <= 4 不翻轉，等價於 05FC 的 `cmp ax,5; jb`）。
 *
 * 讀：ds:786 = X（只取低 3 位元，k = X & 7）、ds:784 = 上一次改寫用的 k（快取）。
 * 若 k 等於 ds:784：什麼都不寫，直接返回。
 * 否則：ds:784 ← k，並依 k 改寫程式碼段 2376 的兩段指令位元組：
 *     k <= 4（無號）：bx = 0xC089（89 C0，mov ax,ax）、dx = 0xC8D1（D1 C8，ror ax,1）
 *     k >= 5        ：k ← 8 - k（結果 3、2、1）、bx = 0xC486（86 C4，xchg ah,al）、dx = 0xC0D1（D1 C0，rol ax,1）
 *   cs:091A 與 cs:092E 各寫一個 word bx；然後 si = 0x91C、di = 0x930，cx = 4 - k：先各寫 cx 個 word 的
 *   0xC089，si、di 每次加 2；再各寫 k 個 word 的 dx。cx = 0 時第一段跳過；k = 0 時第二段跳過。
 *   結果：兩段各 5 個 word（cs:091A..0923 與 cs:092E..0937），共最多 10 個 word、20 bytes。
 *   暫存器在返回時是暫用值；唯一呼叫者 2376:087A 在呼叫之後 `pop si` 並重算 cl、dh、dl。
 *   描述檔 ret = none，只比對記憶體寫入。
 *
 * 語意與文件：docs/re/407-fleet-blocks.md 第 30 行寫「依 X mod 8 改 091A／092E 等六個 bytes」。
 *   實際寫入範圍是上述兩段共 20 bytes（兩個單 word 加上四個迴圈 word，各兩份），不是六個 bytes；
 *   「091A／092E」兩個起點與位址相符。基底狀態裡 cs:091A 起是 `8b c0 d1 c0 d1 c0 d1 c0 d1 c0`
 *   （ds:784 = 0xFFFF，還沒改寫過，是編譯出來的原樣：mov ax,ax 的另一種編碼 8B C0 與四個 rol ax,1）。
 *
 * 輸入域：ds:786 與 ds:784 的全部 16 位元值。
 * 推論等級：位元組值與位移已證實（反組譯逐指令＋差異測試，種子把改寫區預先填 0xFF 以便看出每個 word）；
 *   各機器碼的助記符來自位元組解碼，用途為強推論。
 */
#include "gin3_reg.h"

void f_2376_08A0(Mem *m, uint16_t ds, uint16_t cs)
{
    uint16_t k = (uint16_t)(rd16(m, LIN(ds, 0x786u)) & 7u);
    uint16_t bx, dx, si, di, cx;

    if (k == rd16(m, LIN(ds, 0x784u)))
        return;
    wr16(m, LIN(ds, 0x784u), k);
    bx = 0xC089u;
    dx = 0xC8D1u;
    if (k > 4u) {
        k = (uint16_t)(8u - k);
        bx = 0xC486u;
        dx = 0xC0D1u;
    }
    wr16(m, LIN(cs, 0x091Au), bx);
    wr16(m, LIN(cs, 0x092Eu), bx);
    si = 0x091Cu;
    di = 0x0930u;
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
    /* ds：描述檔 ds_seg（呼叫者 2376:07FA 載入的模組資料段）；cs：dosgolem 載入位置下的 2376 程式段 */
    f_2376_08A0(m, 0x3D22u, 0x2376u);
    return 0;
}
GIN3_REGISTER("2376_08A0", 0, run)
