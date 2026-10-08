/*
 * GIN3PS.EXE 297B:142A  艦隊槽號清單的原地過濾（留下 艦隊[+0x01]（陣營）== v 者）
 *
 * 原版：far 呼叫，四個堆疊參數（args[0..3]），回 AX：
 *   args[0..1] ＝ 緩衝區 far 指標（位移、段）。緩衝區是艦隊槽號（byte）的清單，讀寫同一塊記憶體（原地壓縮）
 *   args[2]    ＝ 項數 n（word，只取低位元組：cmp al,[bp+14h] 是 byte 比較）
 *   args[3]    ＝ 比較值 v（word，只取低位元組：cmp al,[bp+16h] 是 byte 比較）
 * 過程：i ＝ 0..n-1（i 是 byte，無號比較 jnb），k ＝ 寫入位置（起點＝緩衝區起點，只在保留時加一）：
 *   idx ＝ 緩衝區[i]；r ＝ ds:C074 + idx × 122（16 位元乘法與加法，idx 最大 255 不會截斷）
 *   艦隊[idx][+0x01] == v（byte 比較，jnz 跳過不相等者）才保留
 *   保留時 緩衝區[k] ← idx（k ≤ i，所以不會蓋到還沒讀的項目）、k ← k+1
 * 回傳：AL ＝ 保留個數 k（byte）。AH 是殘值，呼叫者 297B:131C 只取 AL。
 *   AH 的起點是緩衝區段的高位元組；每一輪「不保留」把 AH 換成 DS 的高位元組（0x54，mov ax,ds），
 *   「保留」時 mov ax,[bp+6] 取回本輪開頭的 AH，所以 AH 不變。差異測試比較整個 AX，C 版照此追蹤 AH。
 * 艦隊記錄 ds:C074 起 32 筆 × 122 bytes（docs/internals/33）。idx ≥ 32 時讀到艦隊表之外，C 版用同一份記憶體模型。
 * 語意：docs/re/342 §(d')（己方過濾）、docs/re/330 line 303；呼叫端傳 v ＝ ds:8F1D（已證實：反組譯逐指令）。
 * 副作用：寫緩衝區前 k 個 byte。輸入域：n、v 的高位元組任意（原版忽略）；緩衝區位移 16 位元回繞。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

uint16_t f_297B_142A(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t n_word, uint16_t value_word)
{
    uint16_t src = buf_off;
    uint16_t dst = buf_off;
    uint8_t count = 0;
    uint8_t i = 0;
    uint8_t ah = (uint8_t)(buf_seg >> 8);

    while (i < (uint8_t)n_word) {
        uint8_t idx = rd8(m, LIN(buf_seg, src));
        uint16_t rec = (uint16_t)(0xC074u + (uint16_t)(idx * 0x7Au));

        if (rd8(m, DS((uint16_t)(rec + 1))) == (uint8_t)value_word) {
            wr8(m, LIN(buf_seg, dst), idx);
            dst++;
            count++;
        } else {
            ah = (uint8_t)(GIN3_DS >> 8);
        }
        src++;
        i++;
    }
    return (uint16_t)(((uint16_t)ah << 8) | count);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_297B_142A(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("297B_142A", 4, run)
