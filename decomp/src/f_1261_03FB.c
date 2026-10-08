/*
 * GIN3PS.EXE 1261:03FB  艦隊的一個分隊（艦艇池一筆）的士氣百分比，並寫回該筆的 +0x06
 *
 * 原版：far 呼叫，兩個 far 指標參數：艦隊記錄（ds:C074，122 bytes 一筆）fleet、艦艇池記錄（78 bytes 一筆）ship。回 AX。
 * 呼叫者 1261:032B（對艦隊每個分隊呼叫一次再取平均寫回艦隊 +0x3C，docs/re/330 §6、docs/re/153）。
 * 呼叫 3CC9:0009（32 位元乘）、3CC9:0020（32 位元無號除），其數學結果用 gin3_lmul／gin3_uldiv 表達（兩者皆已 diff-tested）。
 *
 * 語意（已證實：docs/re/330 §6「1261:03FB(f, s)」與 docs/re/344 §五「1261:03FB(艦隊far, 艦艇far)」，與反組譯逐條對過）：
 *   adm  ← ds:(0xB77C ＋ fleet[+4] × 28)          提督表一筆 28 bytes；fleet[+4] 是 byte；位移 16 位元
 *   full ← ship[+0x0A]（word，滿編）；cur ← ship[+0x08]（word，現有艘數）
 *   t    ← uint32( (uint16)(full − cur) ) × sext32( (uint16)(200 − adm[+0x0B]) ) 的低 32 位元   ； 200 − 能力 以有號延伸
 *   q    ← (t ÷ 100) 的低 16 位元                      ； 無號 32 位元除
 *   q    ← min(q, full)                                 ； 無號 16 位元比較（cmp bx,ax / jbe）
 *   v    ← uint32( (uint16)(full − q) ) × 100 ÷ full    ； 無號 32 位元除；full ＝ 0 時商為 0xFFFFFFFF
 *   ship[+0x06] ← v 的低位元組（byte）
 *   回 AX ＝ v 的低 16 位元（full ＝ 0 時 0xFFFF，寫入 0xFF）
 *   提督表 adm[+0x0B] 是「耐打」能力（docs/re/153）。
 *
 * 注意：docs/re/153 與 docs/re/330 寫「提督 ＝ ds:B77C ＋ 艦隊[+0x04] × 28」，反組譯一致；
 *   欄位名「提督」是文件的叫法，表的語意以文件為準（強推論），本版只重現算式。
 *
 * 輸入域：fleet 取艦隊表 32 筆，ship 取艦艇池 352 筆；fleet[+4]、adm[+0x0B]、ship[+8]／[+0x0A] 取 0 … 0xFFFF 全域
 *   （含 full ＝ 0、cur ＞ full、能力 ＞ 200 的邊界）。
 */
#include "gin3_reg.h"

uint16_t f_1261_03FB(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t ship_off, uint16_t ship_seg)
{
    uint16_t adm = (uint16_t)(0xB77Cu + (uint16_t)((uint16_t)rd8(m, LIN(fleet_seg, fleet_off + 4u)) * 0x1Cu));
    uint16_t full = rd16(m, LIN(ship_seg, ship_off + 0x0Au));
    uint16_t lost = (uint16_t)(full - rd16(m, LIN(ship_seg, ship_off + 8u)));
    uint16_t factor = (uint16_t)(0xC8u - rd8(m, DS(adm + 0x0Bu)));
    uint32_t t = gin3_lmul(lost, (uint32_t)(int32_t)(int16_t)factor);
    uint16_t q = (uint16_t)gin3_uldiv(t, 100u);
    uint16_t full2;
    uint32_t v;

    if (q > full)
        q = full;
    full2 = rd16(m, LIN(ship_seg, ship_off + 0x0Au));
    v = gin3_uldiv(gin3_lmul((uint16_t)(full2 - q), 100u), full2);
    wr8(m, LIN(ship_seg, ship_off + 6u), (uint8_t)v);
    return (uint16_t)v;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1261_03FB(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("1261_03FB", 4, run)
