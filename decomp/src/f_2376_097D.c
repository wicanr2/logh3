/*
 * 2376:097D  矩形可見性檢查，回傳進位旗標（near，無參數，讀 DS 內四個字）
 *
 * 讀：DS:786 = X、DS:788 = Y、DS:78A = W、DS:78C = H（名稱是本檔取的，語意見下）。
 *   四個都當 16 位元有號數；X 先做算術右移 3（sar ax,3）。
 * 判斷（全部是有號比較，任何一條不成立就 CF = 1）：
 *   (X >> 3) <= 0x4E
 *   (X >> 3) + W >= 1          （加法 16 位元回繞）
 *   Y <= 0x187
 *   Y + H >= 0xA8              （加法 16 位元回繞）
 * 輸出：CF = 0 表示四條都成立（clc），CF = 1 表示有一條不成立（stc）。
 *   唯一呼叫者 2376:07FA 接 jb（CF），不使用其他旗標與 AX。
 *   為了與描述檔 ret=pack:flags 逐位元比對，本版回傳 OF、SF、ZF、CF 四個旗標：
 *   OF／SF／ZF 是最後一次執行的 cmp 留下的（sar 與 add 的旗標會被後面的 cmp 蓋掉），
 *   CF 由 clc／stc 決定。AX 在原版被當暫存，不納入比對。
 *
 * DS：原版呼叫者 07FA 在 call 之前 mov ds 成模組自己的資料段（IDA 的 seg129，dosgolem 的段值是 0x3D22，
 *   不是 DGROUP），那四個字就是 07FA 剛從它的參數存進去的 word_4A516／4A518／4A51A／4A51C。
 *   描述檔用 ds_seg 3D22 呼叫；C 版把 DS 當參數 ds。
 *
 * 語意（假說）：07FA 的參數依序是 X、Y、W、H；這是對 (X/8, Y) 為左上角的矩形做螢幕可見性
 *   檢查。上界 0x4E、0x187 與下界 1、0xA8 的來源沒有查，
 *   是不是螢幕邊界：未知。
 * 輸入域：四個字的全部 16 位元值。
 *
 * 邊界案例由描述檔的 seeds 手工給定（X>>3 等於 0x4E 與 0x4F、X>>3 加 W 等於 0 與 1、Y 等於 0x187 與 0x188、
 *   Y 加 H 等於 0xA7 與 0xA8、X 為負數），因為隨機突變很難剛好打到這些相等邊界。
 */
#include "gin3_reg.h"

#define F_CF 0x0001u
#define F_ZF 0x0040u
#define F_SF 0x0080u
#define F_OF 0x0800u

/* cmp a,b 對 OF／SF／ZF／CF 的影響（16 位元） */
static uint16_t cmp_flags(uint16_t a, uint16_t b)
{
    uint16_t r = (uint16_t)(a - b);
    uint16_t f = 0;

    if (a < b)
        f |= F_CF;
    if (r == 0)
        f |= F_ZF;
    if (r & 0x8000u)
        f |= F_SF;
    if (((a ^ b) & (a ^ r)) & 0x8000u)
        f |= F_OF;
    return f;
}

uint16_t f_2376_097D(Mem *m, uint16_t ds)
{
    int16_t v = (int16_t)rd16(m, LIN(ds, 0x786));
    uint16_t f;

    v = (int16_t)(v >> 3); /* sar ax,3 */
    f = cmp_flags((uint16_t)v, 0x4E);
    if (v > 0x4E)
        return (uint16_t)(f | F_CF);

    v = (int16_t)(uint16_t)((uint16_t)v + rd16(m, LIN(ds, 0x78A)));
    f = cmp_flags((uint16_t)v, 1);
    if (v < 1)
        return (uint16_t)(f | F_CF);

    v = (int16_t)rd16(m, LIN(ds, 0x788));
    f = cmp_flags((uint16_t)v, 0x187);
    if (v > 0x187)
        return (uint16_t)(f | F_CF);

    v = (int16_t)(uint16_t)((uint16_t)v + rd16(m, LIN(ds, 0x78C)));
    f = cmp_flags((uint16_t)v, 0xA8);
    if (v < 0xA8)
        return (uint16_t)(f | F_CF);

    return (uint16_t)(f & ~F_CF);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_2376_097D(m, 0x3D22u); /* 描述檔 ds_seg：呼叫者 2376:07FA 載入的模組資料段 */
}
GIN3_REGISTER("2376_097D", 0, run)
