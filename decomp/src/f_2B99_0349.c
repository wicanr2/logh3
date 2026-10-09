/*
 * GIN3PS.EXE 2B99:0349  中點矩陣查「下一跳」：沿 mid[起點][目標] 一路往中繼點收斂，回第一個直接相鄰的點
 *
 * 原版：far 呼叫，四個堆疊字（push bp、push si、sub sp,2，沒有 push di，所以參數從 [bp+0Ah] 起）：
 *   args[0] ＝ dest（word，只取低位元組：mov al,[bp+0Ah]）
 *   args[1] ＝ start（word，只取低位元組：mov bl,[bp+0Ch]）
 *   args[2..3] ＝ 矩陣 far 指標（位移、段）
 * 過程：
 *   v ← 矩陣[start × 32 + dest]（byte；索引的位移用 16 位元運算，超過 0xFFFF 回繞）
 *   v == 0xFF → 回 dest 的低位元組（AX 高位元組為 0）
 *   否則 遞迴呼叫自己 (v, start, 矩陣)，回傳遞迴的回傳值（AX 原樣帶回）
 *     遞迴時 push word ptr [bp+0]，高位元組是區域變數的另一個未初始化 byte，但被呼叫者只讀低位元組，所以沒有影響。
 * 本檔把尾遞迴寫成迴圈，行為相同（原版每層佔 18 bytes 堆疊，堆疊寫入不在比較範圍）；鏈超過 10 萬層（原版必為無窮遞迴）時回 0xFFFF 放棄。
 *
 * 輸入域：矩陣內容任意，但只涵蓋「原版在步數上限內終止」的情形。矩陣含環（v 序列重複而不碰到 0xFF）時原版無窮遞迴
 *   （堆疊會一路寫穿記憶體），差異測試因超過步數上限而丟棄這類案例；本版此時也不會終止，不在比較範圍。
 *   遞迴超過約 56 層時，原版的堆疊寫入會超出探針排除的堆疊帶，本版不涵蓋（真實矩陣的深度遠小於此，見 docs/internals/26，最遠 12 步）。
 * 證據：asm 逐指令；docs/re/28 第 55 到 70 行把查表寫成「表[起點 × 32 + 終點]」，docs/re/183 第 17 到 24 行把遞迴寫成
 *   「否則：回 2B99:0349(v, b, m)」，與本檔一致。矩陣是中點矩陣（見證矩陣）是 docs/internals/26 的結論。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_2B99_0349(Mem *m, uint16_t dest_word, uint16_t start_word, uint16_t mat_off, uint16_t mat_seg)
{
    uint8_t start = (uint8_t)start_word;
    uint8_t dest = (uint8_t)dest_word;
    unsigned depth;

    for (depth = 0; depth < 100000u; depth++) {
        uint16_t si = (uint16_t)(mat_off + (uint16_t)((uint16_t)start << 5) + dest);
        uint8_t v = rd8(m, LIN(mat_seg, si));

        if (v == 0xFFu)
            return dest;
        dest = v;
    }
    /* 原版此時無窮遞迴（輸入域外）；本版在 10 萬層後放棄，只為避免突變版無窮迴圈 */
    return 0xFFFFu;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_2B99_0349(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("2B99_0349", 4, run)
