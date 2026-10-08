/*
 * GIN3PS.EXE 2703:1ADB  叛亂對一顆行星的效果：變無主、經濟力剩三分之二、防衛力扣掉上限的三分之一
 *
 * 原版：far 呼叫，一個參數＝行星記錄的 far 指標（位移在前、段在後）；無有意義的回傳。
 * 語意（已證實：反組譯逐指令；docs/re/292、internals/29 一致，補充了下列兩處細節）：
 *   行星[+0x11] ← 2
 *   經濟力 E ＝ dword(行星[+0x1A])；E ← E − 有號除法(E ÷ 3)        （3CC9:0070，商向零取整，有號）
 *   d ＝ word(行星[+0x14])；M ＝ word(行星[+0x16])
 *   (d × 3 的低 16 位元) < M（無號）→ 行星[+0x14] ← 0
 *   否則                           → 行星[+0x14] ← d − M ÷ 3        （無號 16 位元除法與回繞）
 * 細節（與 docs/re/292 的文字差異）：
 *   - 比較用的是 d × 3 的**低 16 位元**（mul 後 DX 被下一行覆蓋），d > 21845 時會截斷，
 *     不是「d × 3 < M」的 32 位元比較。
 *   - d × 3 恰等於 M 時走 else 分支，結果 d − d ＝ 0，與 then 分支相同；只有 16 位元截斷後相等
 *     （例：d ＝ 0x5556、M ＝ 2）才看得出 < 與 <= 的差別。
 *   - 經濟力用有號除法：E 的最高位元為 1 時商為負，E − 商 變大。
 * 欄位名稱：+0x11 擁有者、+0x14 防衛力、+0x16 防衛力上限、+0x1A 經濟力，見 docs/internals/29（防衛力與上限的
 *   標籤以位移為準，docs/internals/29 與 docs/re/33 的標籤有出入，這裡不下語意結論）。
 * 副作用：寫行星 +0x11、+0x1A..+0x1D、+0x14..+0x15。
 * 輸入域：全部可到達（除數是常數 3）。
 * 推論等級：已證實（差異測試）。
 */
#include "gin3_reg.h"
#include "gin3_rt.h"

void f_2703_1ADB(Mem *m, uint16_t planet_off, uint16_t planet_seg)
{
    uint16_t econ_off = (uint16_t)(planet_off + 0x1A);
    uint16_t def_off = (uint16_t)(planet_off + 0x14);
    uint32_t econ, q, r;
    uint16_t def, cap;

    wr8(m, LIN(planet_seg, (uint16_t)(planet_off + 0x11)), 2);
    econ = ((uint32_t)rd16(m, LIN(planet_seg, (uint16_t)(econ_off + 2))) << 16) | rd16(m, LIN(planet_seg, econ_off));
    gin3_ldivmod(econ, 3, &q, &r);
    econ = econ - q;
    wr16(m, LIN(planet_seg, econ_off), (uint16_t)econ);
    wr16(m, LIN(planet_seg, (uint16_t)(econ_off + 2)), (uint16_t)(econ >> 16));

    def = rd16(m, LIN(planet_seg, def_off));
    cap = rd16(m, LIN(planet_seg, (uint16_t)(planet_off + 0x16)));
    if ((uint16_t)(def * 3u) < cap)
        wr16(m, LIN(planet_seg, def_off), 0);
    else
        wr16(m, LIN(planet_seg, def_off), (uint16_t)(def - (uint16_t)(cap / 3u)));
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2703_1ADB(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2703_1ADB", 2, run)
