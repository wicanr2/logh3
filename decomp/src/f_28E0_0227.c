/*
 * GIN3PS.EXE 28E0:0227  提督清單的原地過濾（留下 提督[+0x0A] == 參數 者）
 *
 * 原版：far 呼叫，四個堆疊字（args[0..3]），回 AX：
 *   args[0..1] ＝ 清單 far 指標（位移、段）。清單是提督編號（byte）的陣列，讀寫同一塊記憶體（原地壓縮）
 *   args[2]    ＝ 項數 n（word，只取低位元組：cmp al,[bp+14h] 是 byte 比較，無號 jnb）
 *   args[3]    ＝ 比較值 v（word，只取低位元組：cmp al,[bp+16h] 是 byte 比較）
 * 過程：i ＝ 0..n-1（i 是 byte），寫入位置 k 的起點等於清單起點，只在保留時加一：
 *   id ＝ 清單[i]；r ＝ ds:B77C + id × 28（16 位元乘法與加法；id 最大 255，不會截斷）
 *   提督[id][+0x0A] == v（byte 比較）才保留；保留時 清單[k] ← id、k ← k+1（k ≤ i，不會蓋到還沒讀的項目）
 * 回傳：AL ＝ 保留個數 k。AH 是殘值（見下），呼叫者 28E0:0103 只存 AL。
 * AH 殘值（已證實：反組譯逐指令；差異測試比較整個 AX）：
 *   起點是清單段的高位元組（mov ax,[bp+12h]）；每一輪「不保留」時 AH 變成 DS 的高位元組（mov ax,ds，0x54），
 *   「保留」時 mov ax,[bp+6] 取回本輪開頭的 AH，所以 AH 不變。
 * 提督記錄 ds:B77C 起、每筆 28 bytes（docs/internals/08）；+0x0A ＝ 所在星域（docs/re/208 第 38 行「順帶證實」）。
 *   id ≥ 82 時讀到提督表之外（接著是艦隊表），C 版用同一份記憶體模型，原版同樣如此。
 * 語意來源：docs/re/208（28E0:0103 的星域參數 != 0xFF 時先過這一道）；本檔以反組譯與差異測試為準。
 * 副作用：寫清單前 k 個 byte。輸入域：n、v 的高位元組任意（原版忽略）；清單位移 16 位元回繞。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。「+0x0A ＝ 所在星域」的名稱是推論（docs/re/208）。
 */
#include "gin3_reg.h"

uint16_t f_28E0_0227(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w, uint16_t want_w)
{
    uint8_t count = (uint8_t)count_w;
    uint8_t want = (uint8_t)want_w;
    uint16_t src = list_off;
    uint16_t dst = list_off;
    uint8_t kept = 0;
    uint8_t ah = (uint8_t)(list_seg >> 8);
    uint8_t i;

    for (i = 0; i < count; i++) {
        uint8_t id = rd8(m, LIN(list_seg, src));
        uint16_t rec = (uint16_t)(0xB77Cu + (uint16_t)(id * 0x1Cu));

        if (rd8(m, DS((uint16_t)(rec + 0x0Au))) == want) {
            wr8(m, LIN(list_seg, dst), id);
            dst = (uint16_t)(dst + 1u);
            kept = (uint8_t)(kept + 1u);
        } else {
            ah = (uint8_t)(GIN3_DS >> 8);
        }
        src = (uint16_t)(src + 1u);
    }
    return (uint16_t)(((uint16_t)ah << 8) | kept);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_0227(m, a[0], a[1], a[2], a[3]); }
GIN3_REGISTER("28E0_0227", 4, run)
