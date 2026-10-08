/*
 * GIN3PS.EXE 1930:0062  一艘艦艇依艦種表補回飛彈（+0x0E）與艦載機（+0x12）存量，每項有「每次上限」
 *
 * 原版：far 呼叫，一個參數＝艦艇池記錄的 far 指標（位移在前、段在後），無回傳值（描述檔 ret = none）。
 * 艦種表在 DS 段 ds:4E3A，一列 26 bytes，列號 ＝ 陣營 × 208 ＋ 艦種 × 26（docs/internals/40-ship-kind-table.md）。
 *
 * 逐條對應反組譯（IDA sub_25ED2）：
 *   列   ← 0x4E3A + low16(艦艇[+0x00](byte) × 0xD0) + low16(艦艇[+0x04](byte) × 0x1A)    ; 16 位元回繞
 *   第一段（飛彈，32 位元欄位 +0x0E..+0x11）：
 *     n     ← 艦艇[+0x08](word)
 *     cap   ← 列[+0x10](byte) × n                  ; 32 位元乘（3CC9:0009）
 *     per   ← 列[+0x0F](byte) × n                  ; 同一個 n（原版把 n 存在區域變數，不重讀）
 *     need  ← cap − 艦艇[+0x0E..+0x11]               ; 32 位元減，會借位回繞
 *     give  ← (need <= per) ? need : per              ; 32 位元無號比較（先比高字，高字相同才比低字）
 *     艦艇[+0x0E..+0x11] += give                       ; 32 位元加（add 低字、adc 高字）
 *   第二段（艦載機，+0x12..+0x15）：
 *     n 重讀艦艇[+0x08]；cap ← 列[+0x14] × n；per ← 列[+0x13] × n；need ← cap − 艦艇[+0x12..+0x15]；
 *     give ← (need <= per) ? need : per；艦艇[+0x12..+0x15] += give
 *   存量超過 cap 時 need 是借位回繞後的大數，give 於是取 per（同 docs/re/349 §四的「無號比較」那一族）。
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：上述運算、欄位位移、兩段互相獨立、n 在第二段重讀。
 *   強推論（docs/internals/64-ship-pool.md：+0x0E 飛彈存量、+0x12 艦載機存量；docs/internals/40-ship-kind-table.md：
 *     列 +0x10 飛彈每艘幾發、+0x14 艦載機每艘幾架；docs/re/349 §五 的同族搬運函式）：
 *     cap 是滿庫存，per 是「每次最多補多少」（列 +0x0F、+0x13 × 艘數）。
 *   強推論（反組譯 sub_25E7B）：呼叫者 1930:000B 走在場分隊清單（ds:98BA／9896），只對艦種 == 5 的艦艇呼叫本函式；
 *     回合開頭由 12AA:0245 依序呼叫 2526:000E、1930:000B、1AAB:000E、2264:000A（docs/re/339 §1.4、docs/re/145 第 225 行）。
 *     所以本函式是「補給艦（艦種 5）回合開頭自補」的假說，docs 沒有描述 1930:000B／0062，這個語意標 hypothesis。
 *   未知：列 +0x0F、+0x13 的欄位名稱與語意。
 *
 * 呼叫的函式：3CC9:0009（32 位元乘）已 diff-tested，這裡用 gin3_lmul 表達同一個數學結果。
 *
 * 輸入域：艦艇指標任意（描述檔從池取），陣營／艦種 byte 任意（列號 16 位元回繞，讀 DS 段內任意位置）；
 *   艘數 n 取 0..65535；存量任意。函式本身不檢查艦種，艦種 == 5 是呼叫者的條件。
 */
#include "gin3_reg.h"

static uint32_t rd32w(const Mem *m, uint16_t seg, uint16_t off)
{
    return (uint32_t)rd16(m, LIN(seg, off)) | ((uint32_t)rd16(m, LIN(seg, off + 2)) << 16);
}

static void add32w(Mem *m, uint16_t seg, uint16_t off, uint32_t v)
{
    uint32_t lo = (uint32_t)rd16(m, LIN(seg, off)) + (uint16_t)v;

    wr16(m, LIN(seg, off), (uint16_t)lo);
    wr16(m, LIN(seg, off + 2), (uint16_t)(rd16(m, LIN(seg, off + 2)) + (uint16_t)(v >> 16) + (lo >> 16)));
}

void f_1930_0062(Mem *m, uint16_t ship_off, uint16_t ship_seg)
{
    uint16_t row = (uint16_t)(0x4E3A + (uint16_t)(rd8(m, LIN(ship_seg, ship_off)) * 0xD0u) +
                              (uint16_t)(rd8(m, LIN(ship_seg, ship_off + 0x04)) * 0x1Au));
    uint16_t n = rd16(m, LIN(ship_seg, ship_off + 0x08));
    uint32_t cap = gin3_lmul(rd8(m, DS(row + 0x10)), n);
    uint32_t per = gin3_lmul(rd8(m, DS(row + 0x0F)), n);
    uint32_t need = cap - rd32w(m, ship_seg, ship_off + 0x0E);
    uint32_t give = need <= per ? need : per;

    add32w(m, ship_seg, ship_off + 0x0E, give);

    n = rd16(m, LIN(ship_seg, ship_off + 0x08));
    cap = gin3_lmul(rd8(m, DS(row + 0x14)), n);
    per = gin3_lmul(rd8(m, DS(row + 0x13)), n);
    need = cap - rd32w(m, ship_seg, ship_off + 0x12);
    give = need <= per ? need : per;
    add32w(m, ship_seg, ship_off + 0x12, give);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_1930_0062(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("1930_0062", 2, run)
