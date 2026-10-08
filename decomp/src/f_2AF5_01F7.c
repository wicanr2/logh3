/*
 * GIN3PS.EXE 2AF5:01F7  行星的查詢半徑（防衛力夾在 1500..6000 再除以 25）
 *
 * 原版：far 呼叫，一個參數＝行星記錄的 far 指標（位移、段；args[0]＝位移、args[1]＝段），回 AX。
 *   v ＝ 行星[+0x14]（word，防衛力）
 *   v ＜ 1500（無號比較 jb）→ 1500；v ＞ 6000（無號比較 ja）→ 6000
 *   回傳 v ÷ 25（無號 16 位元除法，商一定在 60..240，不會 #DE）
 * 語意與 docs/re/134（「防衛力夾在 1500–6000 再 ÷ 25」）、docs/re/290、docs/re/146 的 rules.PlanetQueryRadius 一致；
 * 行星表在 ds:5336/5338 指向的 5313:0002 起，81 筆 × 82 bytes（docs/internals/29）。
 * 反組譯直接可見但 docs 沒寫的細節：兩次比較都是無號，所以 v ≥ 0x8000 一律夾到 6000（回 240），不會當負數夾到 1500。
 * 離開時 DX＝餘數（殘值，不比較）。副作用：無。
 * 輸入域：任何 far 指標，讀一個 word；v 全部 16 位元值都可到達。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2AF5_01F7(Mem *m, uint16_t planet_off, uint16_t planet_seg)
{
    uint16_t defense = rd16(m, LIN(planet_seg, planet_off + 0x14));

    if (defense < 1500u)
        defense = 1500u;
    if (defense > 6000u)
        defense = 6000u;
    return (uint16_t)(defense / 25u);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2AF5_01F7(m, a[0], a[1]); }
GIN3_REGISTER("2AF5_01F7", 2, run)
