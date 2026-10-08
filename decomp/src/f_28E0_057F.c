/*
 * GIN3PS.EXE 28E0:057F  提督離場：+0x09 設 bit7、+0x00 寫失蹤或 0、+0x02..+0x05 清成 0
 *
 * 原版：far 呼叫，兩個堆疊字（args[0..1]），push bp、push si、sub sp,4，參數從 [bp+0Ch] 起：
 *   args[0] ＝ 提督編號 id（word，只取低位元組：mov bl,[bp+0Ch]）
 *   args[1] ＝ 旗標 f（word，只取低位元組：cmp byte ptr [bp+0Eh],0）
 * 過程：p ＝ id × 28（mul cx，低 16 位元）；r ＝ 0xB77C + p（16 位元）
 *   提督[+0x09] |= 0x80（or byte es:[si+9],80h，無條件）
 *   提督[+0x00] ← (f == 0) ? 0x80 : 0
 *   提督[+0x02] ← 0（word）、提督[+0x04] ← 0（word）
 * 回傳：AX ＝ 剛寫進 +0x00 的值（0x80 或 0）。不是函式的回傳值，呼叫端不使用，差異測試仍比較整個 AX。
 * 提督記錄 ds:B77C 起、每筆 28 bytes（docs/internals/08）。id ≥ 82 時寫到提督表之外（接著是艦隊表），原版同樣如此。
 * 語意：docs/re/329 第 106 行：「+0x09 |= 0x80；+0x00 ← 旗標 0 時 0x80（失蹤）、非 0 時 0；功績清零」，
 *   docs/internals/08 第 35 行：+0x02..+0x05 是功績累計 u32，fate 死亡時由本函式清零，與本檔一致。
 *   docs/re/307 第 50 到 53 行的反組譯節錄與本檔逐行一致。
 * 副作用：寫 +0x09（一個 byte，讀改寫）、+0x00（一個 byte）、+0x02..+0x05（四個 byte）。
 * 輸入域：id、f 的高位元組任意（原版忽略）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。欄位名稱（失蹤、功績）是 docs 的推論。
 */
#include "gin3_reg.h"

uint16_t f_28E0_057F(Mem *m, uint16_t id_w, uint16_t flag_w)
{
    uint8_t id = (uint8_t)id_w;
    uint16_t rec = (uint16_t)(0xB77Cu + (uint16_t)(id * 0x1Cu));
    uint16_t ax;

    wr8(m, DS((uint16_t)(rec + 9u)), (uint8_t)(rd8(m, DS((uint16_t)(rec + 9u))) | 0x80u));
    ax = ((uint8_t)flag_w == 0u) ? 0x80u : 0u;
    wr8(m, DS(rec), (uint8_t)ax);
    wr16(m, DS((uint16_t)(rec + 2u)), 0u);
    wr16(m, DS((uint16_t)(rec + 4u)), 0u);
    return ax;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_057F(m, a[0], a[1]); }
GIN3_REGISTER("28E0_057F", 2, run)
