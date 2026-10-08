/*
 * GIN3PS.EXE 28E0:05CC  專任參謀就任：提督[+0x09] ← 0x7F
 *
 * 原版：far 呼叫，一個堆疊字（args[0]），push bp、push si、sub sp,4，參數在 [bp+0Ch]：
 *   args[0] ＝ 提督編號 id（word，只取低位元組：mov bl,[bp+0Ch]）
 * 過程：p ＝ id × 28（mul cx，低 16 位元）；ds:[0xB77C + p + 9] ← 0x7F（mov byte，不讀原值，整個 byte 覆寫）
 * 回傳：AX ＝ p（mul 之後沒有再動 AX，是殘值，不是函式的回傳值）。呼叫端不使用，差異測試仍比較整個 AX。
 *   id ≤ 255，p ≤ 7140，不會截斷。
 * 提督記錄 ds:B77C 起、每筆 28 bytes（docs/internals/08）。id ≥ 82 時寫到提督表之外（接著是艦隊表），原版同樣如此。
 * 語意：docs/internals/08 第 277 行：「專任參謀就任時 28E0:05CC 把它寫成 0x7F（清掉 bit7，低七位填 127」，
 *   docs/re/268 第 83、136 行同；與本檔一致。
 * 副作用：寫 1 個 byte。原值與 0x7F 相同時寫入集合看不出來，所以差異測試的種子要先把原值改成別的值。
 * 輸入域：id 的高位元組任意（原版忽略）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。「+0x09 ＝ 艦隊槽／閒置旗標」是 docs 的推論。
 */
#include "gin3_reg.h"

uint16_t f_28E0_05CC(Mem *m, uint16_t id_w)
{
    uint8_t id = (uint8_t)id_w;
    uint16_t p = (uint16_t)(id * 0x1Cu);

    wr8(m, DS((uint16_t)(0xB77Cu + p + 9u)), 0x7Fu);
    return p;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_05CC(m, a[0]); }
GIN3_REGISTER("28E0_05CC", 1, run)
