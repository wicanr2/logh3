/*
 * GIN3PS.EXE 1F6F:024C  依亂數與艦隊的一個 byte 欄位，從候選清單挑「第 1／2／3 筆」之一（回 0、1 或 2）
 *
 * 原版：far 呼叫，五個字的堆疊參數（由低位址到高位址）：
 *   ship  ＝ 艦艇池記錄的 far 指標（位移在前、段在後）
 *   list  ＝ 候選清單的 far 指標（原版呼叫者推入 ds:A49A；本函式沒有讀它）
 *   count ＝ 候選筆數（完整 word，無號比較）
 * 回 AX ＝ 0、1 或 2。呼叫者 1F6F:005F 推入的五個字與此一致（反組譯 sub_2C2BF）。
 *
 * 逐條對應反組譯（IDA sub_2C4AC）：
 *   fleet ← ds:C074 + low16(艦艇[+0x01](byte) × 0x7A)       ; 艦艇 +0x01 是所屬艦隊槽號
 *   r     ← 3CD5:000C()                                       ; rand()，0..0x7FFF，會推進 ds:94DC 的種子
 *   q     ← low16( (r × fleet[+0x45](byte)) ÷ 100 )           ; 32 位元乘、無號除，商只取低 16 位元
 *   若 q < 0x2000 且 count >= 3：回 2
 *   否則若 q < 0x4000 且 count >= 2：回 1
 *   否則回 0                                                   ; 比較全是無號
 * rand() 的 cwd 沒有影響（r <= 0x7FFF）。
 *
 * 推論等級
 *   已證實（差異測試的位元組、回傳值與寫入集合）：公式、兩個門檻 0x2000／0x4000、count 的門檻 3／2、
 *     商截成 16 位元（32767 × 255 ÷ 100 ＝ 83555 超過 16 位元）、rand() 恰好被呼叫一次。
 *   強推論（docs/re/412-fleet-auto-command.md 第 33 行「024C | 值=低word(RND×FLEET+45/100)，8192／16384門檻選第3／2／1筆」）：
 *     fleet +0x45 是一個百分比類的欄位；回傳值用來挑候選清單的第幾筆。本檔與該行逐項相符。
 *   未知：fleet +0x45 的欄位語意（docs 未命名）。
 *
 * 呼叫的函式：3CD5:000C（rand）已 diff-tested；這裡直接呼叫 f_3CD5_000C。
 *
 * 輸入域：艦艇指標取自艦艇池（描述檔 table），艦隊槽號 +0x01 與 fleet +0x45 由突變導向讀取集改動；
 *   槽號超過 31 時讀艦隊表之後的記憶體，本版同樣照做。count 取 0..65535。
 */
#include "gin3_reg.h"

extern uint16_t f_3CD5_000C(Mem *m);

uint16_t f_1F6F_024C(Mem *m, uint16_t ship_off, uint16_t ship_seg, uint16_t list_off, uint16_t list_seg, uint16_t count)
{
    uint16_t fleet = (uint16_t)(0xC074 + (uint16_t)(rd8(m, LIN(ship_seg, ship_off + 0x01)) * 0x7Au));
    uint16_t r = f_3CD5_000C(m);
    uint16_t q = (uint16_t)gin3_uldiv(gin3_lmul(r, rd8(m, DS(fleet + 0x45))), 100u);

    (void)list_off;
    (void)list_seg;
    if (q < 0x2000u && count >= 3)
        return 2;
    if (q < 0x4000u && count >= 2)
        return 1;
    return 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1F6F_024C(m, a[0], a[1], a[2], a[3], a[4]); }
GIN3_REGISTER("1F6F_024C", 5, run)
