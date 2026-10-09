/*
 * GIN3PS.EXE 1261:032B  艦隊的完整度重算：對艦隊每個分隊呼叫 1261:03FB，取平均寫進艦隊 +0x3C，並更新艦隊旗標 bit0／2／3
 *
 * 原版：far 呼叫，一個 far 指標參數＝艦隊記錄（序言 push bp／si／di、sub sp,0Ch，所以在 [bp+16h]／[bp+18h]），回 AX（見下）。
 * 逐條對應反組譯：
 *   di ← 0（word，加總）、i ← 0（byte）、cnt ← 0（byte，有艘數的分隊數）、組 ← 艦隊 + 0x10（16 位元位移，段＝艦隊段）
 *   迴圈：
 *     i >= byte 艦隊[+0x0F]（無號，每圈重讀）→ 結束
 *     idx  ← word 組[0]；艦艇 ← (word ds:CFB4 ＋ low16(idx × 0x4E)) : word ds:CFB6          ; 池基底每圈重讀
 *     word 艦艇[+8] ≠ 0（現有艘數非 0）：
 *         r ← 1261:03FB(艦隊 far 指標, 艦艇 far 指標)             ; 該分隊的士氣百分比（見 f_1261_03FB.c）
 *         di += r 的低位元組（`xor ah,ah; add di,ax`）；cnt++
 *     i++；組 += 4
 *   cnt == 0：艦隊[+0] |= 0x08；結束（沒有任何有艘數的分隊，旗標 bit3 ＝ 全滅？名稱為假說）
 *   否則：
 *     艦隊[+0] &= 0xF7
 *     avg ← di ÷ cnt（16 位元無號除；cnt 是 byte ≥ 1，商不會溢位）
 *     艦隊[+0x3C] ← avg 的低位元組
 *     old ← 艦隊[+0]
 *     avg 的低位元組 < 0x32（無號）→ 艦隊[+0] |= 1，否則 艦隊[+0] &= 0xFE        ; 完整度低於 50 的旗標
 *     (old ^ 艦隊[+0]) 的 bit0 有設（這次改變了 bit0）→ 艦隊[+0] |= 4
 *
 * 回傳 AX（描述檔 ret = ax；呼叫者 1261:0067 不用，297B:0E8B 的呼叫點也沒有取用，AX 是 codegen 殘值，但差異測試照比）：
 *   cnt == 0：AX ＝ (AH << 8) | i。AL ＝ 離開迴圈時的 i（`mov al,[bp+0]`）；AH 是進入點 `mov ax,[bp+18h]` 設的艦隊段高位元組，
 *     之後每個分隊被 `mul`（idx × 0x4E）改成乘積的高位元組（呼叫了 03FB 的分隊隨後 `xor ah,ah` 清成 0，但 cnt == 0 時沒有任何分隊呼叫過 03FB）。
 *   cnt != 0：AX ＝ old ^ 艦隊[+0]（兩個 byte 零延伸後的互斥或，高位元組 0，在設定 bit2 之前取）。
 *   平均 avg ＝ di ÷ cnt 不會超過 255（每個加數都是 byte、cnt 是它們的個數），所以「商只取低位元組」與「商飽和」等價，沒有可驗的差別。
 *
 * 呼叫者：1261:0067、297B:0E8B（docs/spec/102「受損艦隊按 1261:032B 的編制數重算完整度，排除獨立斥候；低於 50 設 bit0」、
 *   docs/internals/53 第 64 行「士氣（1261:032B，艦隊 +0x3C）」、docs/internals/33 第 306 行）。
 *
 * 推論等級
 *   已證實（差異測試）：欄位位移與寬度、迴圈條件（無號、每圈重讀組數）、現有艘數非 0 的篩選、03FB 的參數、只取低位元組相加、
 *     平均的除法、0x32 門檻與比較方向、三個旗標位元的更新順序、回傳 AX 的組成。
 *   強推論（docs/spec/102、internals/53）：艦隊 +0x3C 是完整度（士氣百分比）、bit0 ＝ 低於 50、bit2 ＝ bit0 這次有變、
 *     艦艇 +8 是現有艘數。bit3 的語意（「沒有任何有艘數的分隊」）為假說。
 * 輸入域：艦隊指標任意（位移 16 位元回繞，word 在位移 0xFFFF 時高位元組取位移 0）；池基底 ds:CFB4／CFB6 任意；組數 0..255。
 *   03FB 的寫入（艦艇 +6）落在 ROM 區（池段 >= 0xF000）時，原版會丟棄寫入，而 03FB 的 C 版沒有 ROM 判斷；
 *   本批測試沒有產生這種案例（記錄的案例只把 ds:CFB6 的低位元組由 0x8B 突變成 0x0B，高位元組沒動），該情形不在輸入域內。
 */
#include "gin3_reg.h"

extern uint16_t f_1261_03FB(Mem *m, uint16_t fleet_off, uint16_t fleet_seg, uint16_t ship_off, uint16_t ship_seg);

/* word 讀取：位移各自 16 位元回繞（seg:FFFF 的高位元組在同一段的位移 0），與 dosgolem 一致 */
static uint16_t f032B_rd16s(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint16_t)(rd8(m, LIN(seg, off)) | (rd8(m, LIN(seg, (uint16_t)(off + 1u))) << 8));
}

uint16_t f_1261_032B(Mem *m, uint16_t fleet_off, uint16_t fleet_seg)
{
    uint16_t di = 0;
    uint8_t i = 0;
    uint8_t cnt = 0;
    uint8_t ah = (uint8_t)(fleet_seg >> 8);
    uint16_t slot = (uint16_t)(fleet_off + 0x10u);

    while (i < rd8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x0Fu)))) {
        uint16_t prod = (uint16_t)(f032B_rd16s(m, fleet_seg, slot) * 0x4Eu);
        uint16_t pool_seg = f032B_rd16s(m, GIN3_DS, 0xCFB6u);
        uint16_t ship = (uint16_t)(f032B_rd16s(m, GIN3_DS, 0xCFB4u) + prod);

        ah = (uint8_t)(prod >> 8);
        if (f032B_rd16s(m, pool_seg, (uint16_t)(ship + 8u)) != 0) {
            uint16_t r = f_1261_03FB(m, fleet_off, fleet_seg, ship, pool_seg);

            di = (uint16_t)(di + (r & 0xFFu));
            cnt = (uint8_t)(cnt + 1u);
            /* 原版此處 `xor ah,ah`，之後 AH 只在 cnt == 0 的出口用得到，而 cnt == 0 表示一次都沒呼叫 03FB，所以不必記 */
        }
        i = (uint8_t)(i + 1u);
        slot = (uint16_t)(slot + 4u);
    }

    if (cnt == 0) {
        wr8(m, LIN(fleet_seg, fleet_off), (uint8_t)(rd8(m, LIN(fleet_seg, fleet_off)) | 0x08u));
        return (uint16_t)(((uint16_t)ah << 8) | i);
    } else {
        uint8_t avg, old, now;

        wr8(m, LIN(fleet_seg, fleet_off), (uint8_t)(rd8(m, LIN(fleet_seg, fleet_off)) & 0xF7u));
        avg = (uint8_t)(di / cnt);
        wr8(m, LIN(fleet_seg, (uint16_t)(fleet_off + 0x3Cu)), avg);
        old = rd8(m, LIN(fleet_seg, fleet_off));
        if (avg < 0x32u)
            wr8(m, LIN(fleet_seg, fleet_off), (uint8_t)(rd8(m, LIN(fleet_seg, fleet_off)) | 0x01u));
        else
            wr8(m, LIN(fleet_seg, fleet_off), (uint8_t)(rd8(m, LIN(fleet_seg, fleet_off)) & 0xFEu));
        now = rd8(m, LIN(fleet_seg, fleet_off));
        if (((old ^ now) & 1u) != 0)
            wr8(m, LIN(fleet_seg, fleet_off), (uint8_t)(rd8(m, LIN(fleet_seg, fleet_off)) | 0x04u));
        return (uint16_t)(old ^ now);
    }
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1261_032B(m, a[0], a[1]); }
GIN3_REGISTER("1261_032B", 2, run)
