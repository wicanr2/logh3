/*
 * GIN3PS.EXE 28E0:054A  把一名提督標成「沒帶艦隊」：提督[+0x09] 的 bit7 設起來
 *
 * 原版：far 呼叫，一個堆疊字（args[0]），push bp、push si、sub sp,4，參數在 [bp+0Ch]：
 *   args[0] ＝ 提督編號 id（word，只取低位元組：mov bl,[bp+0Ch]）
 * 過程：p ＝ id × 28（mul cx，低 16 位元）；a ＝ ds:[0xB77C + p + 9]（16 位元位移）
 *   test byte [a],80h 為零時 or byte [a],80h；bit7 已設則不寫。
 *   也就是 提督[id][+0x09] |= 0x80。寫入值與原值相同時寫入集合看不出來，所以「只在 bit7 沒設時寫」與
 *   「一律 or」在可觀察的記憶體上沒有差別；本檔照反組譯寫成條件式。
 * 回傳：AX ＝ p（mul 之後沒有再動 AX，是殘值，不是函式的回傳值）。呼叫端不使用，差異測試仍比較整個 AX。
 *   id ≤ 255，p ≤ 7140，不會截斷。
 * 提督記錄 ds:B77C 起、每筆 28 bytes（docs/internals/08）。id ≥ 82 時寫到提督表之外（接著是艦隊表），原版同樣如此。
 * 語意：docs/re/329 第 51、67 行（解散艦隊時對提督與副官各呼叫一次）寫為「+0x09 |= 0x80（沒帶艦隊）」，與本檔一致。
 * 副作用：至多寫 1 個 byte。輸入域：id 的高位元組任意（原版忽略）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。「bit7 ＝ 沒帶艦隊」是 docs 的推論。
 */
#include "gin3_reg.h"

uint16_t f_28E0_054A(Mem *m, uint16_t id_w)
{
    uint8_t id = (uint8_t)id_w;
    uint16_t p = (uint16_t)(id * 0x1Cu);
    uint32_t a = DS((uint16_t)(0xB77Cu + p + 9u));

    if ((rd8(m, a) & 0x80u) == 0u)
        wr8(m, a, (uint8_t)(rd8(m, a) | 0x80u));
    return p;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_054A(m, a[0]); }
GIN3_REGISTER("28E0_054A", 1, run)
