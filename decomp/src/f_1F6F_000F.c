/*
 * GIN3PS.EXE 1F6F:000F  在場分隊清單的每一筆：預估剩餘艘數 +0x3E 設成現有艘數 +0x08
 *
 * 原版：far 呼叫，無參數，無回傳值（描述檔 ret = none）。
 * 清單是 ds:98BA 起的 word 陣列（每筆一個艦艇池索引），筆數是 word ds:9896。
 *
 * 逐條對應反組譯（IDA sub_2C26F）：
 *   i ← 0；清單指標 ← ds:98BA
 *   while i < ds:9896（無號 word 比較，每圈重讀）：
 *     idx   ← 清單[i]
 *     艦艇  ← (ds:CFB6 : ds:CFB4 + low16(idx × 0x4E))     ; 池基底每圈重讀；mul 只取 AX，位移 16 位元回繞
 *     艦艇[+0x3E](word) ← 艦艇[+0x08](word)
 *     i++，清單指標 += 2（16 位元回繞）
 *
 * 推論等級
 *   已證實（差異測試的位元組與寫入集合）：迴圈結構、欄位位移、無條件複製。
 *   強推論（docs/spec/101-fleet-auto-command.md 的 "1F6F:004D" 一列、docs/re/412-fleet-auto-command.md）：
 *     艦艇 +0x08 是現有艘數，+0x3E 是命令前的預估剩餘艘數；本函式是「每次艦隊命令入口把 SHIP+08 拷到 +3E」
 *     的清單版。名稱不是證據，只有位移與複製關係是證據。
 *   未知：呼叫者 1ED2:0001 的時機（本函式只有一個呼叫者）。
 *
 * 輸入域：清單筆數 ds:9896 取 0..65535 任意值（步數超過 200,000 的由探針丟棄）；清單內容與池索引任意。
 *   池索引超出 352 筆時讀寫池之後的記憶體，本版同樣照做（平面視訊記憶體的讀取除外，見 docs/re/445 限制）。
 *   基底狀態 fcmd-s3-base 的 ds:9896 是 0，清單內容靠突變與描述檔 seeds 的 pokes 造出。
 */
#include "gin3_reg.h"

void f_1F6F_000F(Mem *m)
{
    uint16_t i = 0;
    uint16_t list = 0x98BA;

    while (i < rd16(m, DS(0x9896))) {
        uint16_t idx = rd16(m, DS(list));
        uint16_t seg = rd16(m, DS(0xCFB6));
        uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));

        wr16(m, LIN(seg, ship + 0x3E), rd16(m, LIN(seg, ship + 0x08)));
        i++;
        list = (uint16_t)(list + 2);
    }
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_1F6F_000F(m);
    return 0;
}
GIN3_REGISTER("1F6F_000F", 0, run)
