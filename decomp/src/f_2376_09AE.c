/*
 * 2376:09AE  圖塊矩形的螢幕裁切，算出來源與目的起點（near，無參數，DS = 模組資料段 3D22）
 *
 * 讀：ds:786 = X、ds:788 = Y、ds:78A = W、ds:78C = H（名稱是本檔取的，語意見下；四個都當 16 位元有號數）。
 * 輸出：SI（來源位移）、DI（目的位移）以及就地改寫的 ds:78A（W）、ds:78C（H）。AX、DX 是暫用值，
 *   呼叫者 2376:07FA 緊接著 `call 2376:0A34`，那一支用 SI、DI、ds:78A、ds:78C，不用 AX、DX；
 *   描述檔 ret = pack:si,di，寫入集合比對 ds:78A／78C。
 *
 * 算法（逐指令對過反組譯）：
 *   DI = Y * 80 + (X >> 3, 算術右移) + 0xC80；SI = 0。            所有加法與 shl 都是 16 位元回繞
 *   上緣：若 Y（有號）< 0xA8：
 *       d = 0xA8 - Y；H -= d；t = d << 6；SI += t；DI += t + (t >> 2, 邏輯右移)
 *     否則：若 Y + H（有號比較）>= 0x187：e = Y + H - 0x187 + 1；H -= e。
 *   左右緣：xs = X >> 3（算術右移）：
 *       若 xs（有號）< 1：d = 1 - xs；SI += d；DI += d；W -= d。
 *       否則：若 xs + W（有號比較）>= 0x4E：e = xs + W - 0x4E + 1；W -= e。
 *   上緣若走「Y < 0xA8」那一支，不再做下緣檢查；左右緣兩支互斥。
 *
 * 語意：X 是像素、Y 是畫面列（假說）。DI 是 VGA 畫面（每列 80 bytes）的位移，0xC80 = 40 列 × 80，
 *   即畫面上方保留 40 列；SI 的上緣跳列是每列 0x40 bytes（來源圖塊每列 64 bytes），DI 是每列 80 bytes
 *   （0x40 + 0x40 / 4）。H、W 被改成裁切後的列數與欄數。
 *   由常數推出的可見範圍：欄 1..0x4C（像素 x8..615，因為 xs + W 被壓到 0x4D 以下）；
 *   列 0xA8..0x185（加畫面保留的 40 列，畫面 y208..429，因為 Y + H 被壓到 0x186 以下）。
 *   與 docs/spec/095-fleet-blocks.md 第 39 行（x8..615、畫面 y208..429）及 docs/re/407-fleet-blocks.md
 *   第 31 行（裁切最後 X615／畫面 Y429）相符。
 *
 * 輸入域：四個字的全部 16 位元值。所有存取都在 DS 內固定位移，沒有指標，也不會讀到視訊記憶體區。
 *   邊界值（Y 等於 0xA8 與 0xA7、Y + H 等於 0x187 與 0x186、xs 等於 1 與 0、xs + W 等於 0x4E 與 0x4D、
 *   X 與 Y 為負數）由描述檔 seeds 的 pokes 打到。
 * 推論等級：算法與常數已證實（反組譯逐指令＋差異測試）；欄位語意為假說。
 */
#include "gin3_reg.h"

/* sub ds:[off], v（16 位元） */
static void w09AE_sub16(Mem *m, uint16_t ds, uint16_t off, uint16_t v)
{
    wr16(m, LIN(ds, off), (uint16_t)(rd16(m, LIN(ds, off)) - v));
}

uint32_t f_2376_09AE(Mem *m, uint16_t ds)
{
    uint16_t x = rd16(m, LIN(ds, 0x786u));
    uint16_t y = rd16(m, LIN(ds, 0x788u));
    uint16_t ax, dx;
    uint16_t si = 0, di;

    ax = (uint16_t)(y << 4); /* 四次 add ax,ax */
    dx = ax;
    ax = (uint16_t)(ax << 2); /* 兩次 add ax,ax */
    ax = (uint16_t)(ax + dx); /* Y * 80 */
    di = ax;
    ax = (uint16_t)((int16_t)x >> 3); /* sar ax,3 */
    di = (uint16_t)(di + ax);
    di = (uint16_t)(di + 0xC80u);

    ax = y;
    if ((int16_t)ax < 0xA8) {
        dx = ax;
        ax = (uint16_t)(0xA8u - dx);
        w09AE_sub16(m, ds, 0x78Cu, ax);
        ax = (uint16_t)(ax << 6);
        si = (uint16_t)(si + ax);
        dx = ax;
        dx = (uint16_t)(dx >> 2); /* shr dx,2 */
        ax = (uint16_t)(ax + dx);
        di = (uint16_t)(di + ax);
    } else {
        ax = (uint16_t)(ax + rd16(m, LIN(ds, 0x78Cu)));
        if (!((int16_t)ax < 0x187)) {
            ax = (uint16_t)(ax - 0x187u);
            ax = (uint16_t)(ax + 1u);
            w09AE_sub16(m, ds, 0x78Cu, ax);
        }
    }

    ax = (uint16_t)((int16_t)rd16(m, LIN(ds, 0x786u)) >> 3); /* 重讀 X，sar ax,3 */
    if ((int16_t)ax < 1) {
        dx = ax;
        ax = (uint16_t)(1u - dx);
        si = (uint16_t)(si + ax);
        di = (uint16_t)(di + ax);
        w09AE_sub16(m, ds, 0x78Au, ax);
    } else {
        ax = (uint16_t)(ax + rd16(m, LIN(ds, 0x78Au)));
        if (!((int16_t)ax < 0x4E)) {
            ax = (uint16_t)(ax - 0x4Eu);
            ax = (uint16_t)(ax + 1u);
            w09AE_sub16(m, ds, 0x78Au, ax);
        }
    }
    return ((uint32_t)si << 16) | di;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    /* 描述檔 ds_seg：呼叫者 2376:07FA 載入的模組資料段（IDA 的 seg129） */
    return f_2376_09AE(m, 0x3D22u);
}
GIN3_REGISTER("2376_09AE", 0, run)
