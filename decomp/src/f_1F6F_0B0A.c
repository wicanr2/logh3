/*
 * GIN3PS.EXE 1F6F:0B0A  在場分隊清單裡，位置落在兩艘艦艇座標所圍矩形內的分隊：收進一張 word 清單，回傳收了幾筆
 *
 * 原版：far 呼叫，六個字的堆疊參數（由低位址到高位址）：
 *   a   ＝ 艦艇 A 的 far 指標（位移在前、段在後）
 *   b   ＝ 艦艇 B 的 far 指標
 *   out ＝ 輸出 word 清單的 far 指標
 * 回 AX ＝ 收了幾筆（16 位元計數）。
 * 清單是 ds:98BA 起的 word 陣列（每筆一個艦艇池索引），筆數是 word ds:9896。
 *
 * 逐條對應反組譯（IDA sub_2CD6A）：
 *   矩形 ← 以 A、B 的位置（艦艇 +0x34 為 X、+0x36 為 Y，都是無號 word）為對角的軸對齊矩形：
 *     xmin ＝ min(A.x, B.x)，xmax ＝ max(A.x, B.x)，ymin ＝ min(A.y, B.y)，ymax ＝ max(A.y, B.y)（無號比較）
 *   found ← 0；i ← 0；清單指標 ← ds:98BA
 *   while i < ds:9896（無號 word 比較，每圈重讀）：
 *     idx   ← 清單[i]
 *     艦艇  ← (ds:CFB6 : ds:CFB4 + low16(idx × 0x4E))     ; 池基底每圈重讀
 *     略過 若 艦艇 == A（段與位移都相等）或 艦艇 == B
 *     略過 若 艦艇[+0x1E] 的 bit7 已設
 *     略過 若 艦艇.x < xmin 或 艦艇.y < ymin 或 艦艇.x > xmax 或 艦艇.y > ymax（無號）
 *     否則 out[0](word) ← idx；out 指標 += 2；found++
 *     i++，清單指標 += 2（16 位元回繞）
 *   A、B 的座標在函式開頭各讀一次；A 與 B 本身若在清單裡，因為被「等於 A 或 B」排除，不會收進輸出。
 *
 * 推論等級
 *   已證實（差異測試的位元組、回傳值與寫入集合）：矩形的建法、迴圈結構、略過條件與順序、欄位位移。
 *   強推論（docs/internals/38-battle-units.md：+0x34／+0x36 現在位置 X／Y、+0x1E bit7 ＝ 不進攻擊名單／不被索敵看到）：
 *     本函式找「兩艘艦艇之間」的候選分隊。名稱不是證據，只有位移與比較是證據。
 *   未知：呼叫者 1F6F:091C 的用途（本函式只有一個呼叫者）。
 *
 * 輸入域：A、B 取自艦艇池（描述檔 table）；清單筆數 ds:9896 任意（步數超過 200,000 的由探針丟棄）；
 *   清單內容與池索引任意。輸出指標的寫入範圍由描述檔的暫存緩衝區與其後的記憶體構成。
 *   基底狀態 fcmd-s3-base 的 ds:9896 是 0，清單內容靠突變與描述檔 seeds 的 pokes 造出。
 */
#include "gin3_reg.h"

uint16_t f_1F6F_0B0A(Mem *m, uint16_t a_off, uint16_t a_seg, uint16_t b_off, uint16_t b_seg, uint16_t out_off, uint16_t out_seg)
{
    uint16_t ax_ = rd16(m, LIN(a_seg, a_off + 0x34));
    uint16_t ay_ = rd16(m, LIN(a_seg, a_off + 0x36));
    uint16_t bx_ = rd16(m, LIN(b_seg, b_off + 0x34));
    uint16_t by_ = rd16(m, LIN(b_seg, b_off + 0x36));
    uint16_t xmin = ax_ > bx_ ? bx_ : ax_;
    uint16_t xmax = ax_ > bx_ ? ax_ : bx_;
    uint16_t ymin = ay_ > by_ ? by_ : ay_;
    uint16_t ymax = ay_ > by_ ? ay_ : by_;
    uint16_t found = 0;
    uint16_t i = 0;
    uint16_t list = 0x98BA;

    while (i < rd16(m, DS(0x9896))) {
        uint16_t idx = rd16(m, DS(list));
        uint16_t seg = rd16(m, DS(0xCFB6));
        uint16_t ship = (uint16_t)(rd16(m, DS(0xCFB4)) + (uint16_t)(idx * 0x4Eu));

        if (!(seg == a_seg && ship == a_off) && !(seg == b_seg && ship == b_off) &&
            !(rd8(m, LIN(seg, ship + 0x1E)) & 0x80) &&
            !(rd16(m, LIN(seg, ship + 0x34)) < xmin) && !(rd16(m, LIN(seg, ship + 0x36)) < ymin) &&
            !(rd16(m, LIN(seg, ship + 0x34)) > xmax) && !(rd16(m, LIN(seg, ship + 0x36)) > ymax)) {
            wr16(m, LIN(out_seg, out_off), idx);
            out_off = (uint16_t)(out_off + 2);
            found++;
        }
        i++;
        list = (uint16_t)(list + 2);
    }
    return found;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_1F6F_0B0A(m, a[0], a[1], a[2], a[3], a[4], a[5]); }
GIN3_REGISTER("1F6F_0B0A", 6, run)
