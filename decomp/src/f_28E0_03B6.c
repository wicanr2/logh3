/*
 * GIN3PS.EXE 28E0:03B6  提督清單的原地過濾（留下 提督[+0x00] != 0 者）
 *
 * 原版：far 呼叫，三個堆疊字（args[0..2]），回 AX（沒有 push si／di，參數從 [bp+0Eh] 起）：
 *   args[0..1] ＝ 清單 far 指標（位移、段）。清單是提督編號（byte）的陣列，讀寫同一塊記憶體（原地壓縮）
 *   args[2]    ＝ 項數 n（word，只取低位元組：cmp al,[bp+12h] 是 byte 比較，無號 jnb）
 * 過程：i ＝ 0..n-1（i 是 byte），寫入位置 k 的起點等於清單起點，只在保留時加一：
 *   id ＝ 清單[i]；p ＝ id × 28（mul cx，低 16 位元）；提督[id][+0x00] ＝ byte ds:[p + 0xB77C]（asm 寫成 [bx-4884h]）
 *   提督[id][+0x00] != 0（asm 是 jz 跳過）才保留；保留時 清單[k] ← id、k ← k+1
 * 回傳：AL ＝ 保留個數 k。AH 是殘值（見下），呼叫者 28E0:0103 只存 AL。
 * AH 殘值（已證實：反組譯逐指令；差異測試比較整個 AX）：
 *   這一支沒有 mov ax,ds。起點是清單段的高位元組；每一輪「不保留」時 AX 留著 mul 的結果 p，AH ＝ p >> 8；
 *   「保留」時 mov ax,[bp+6] 取回本輪開頭的 AH，所以 AH 不變。id 小於 10 時 p < 256，AH ＝ 0。
 * 提督記錄 ds:B77C 起、每筆 28 bytes（docs/internals/08）。id ≥ 82 時讀到提督表之外（接著是艦隊表），原版同樣如此。
 * 語意：docs/re/208 第 20 行、docs/re/43 第 201 行把這一支記為「未讀」；本檔補上條件：留下 +0x00 不為 0 的。
 *   +0x00 在 docs 裡是「職務」（0xFF 無職務、0xFE 排除、0x80 失蹤；docs/re/208、307），這一支的名稱是推論，
 *   條件本身是反組譯證實。
 * 副作用：寫清單前 k 個 byte。輸入域：n 的高位元組任意（原版忽略）；清單位移 16 位元回繞。
 * 推論等級：條件已證實（反組譯逐指令＋差異測試）；欄位語意為強推論。
 */
#include "gin3_reg.h"

uint16_t f_28E0_03B6(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t count_w)
{
    uint8_t count = (uint8_t)count_w;
    uint16_t src = list_off;
    uint16_t dst = list_off;
    uint8_t kept = 0;
    uint8_t ah = (uint8_t)(list_seg >> 8);
    uint8_t i;

    for (i = 0; i < count; i++) {
        uint8_t id = rd8(m, LIN(list_seg, src));
        uint16_t p = (uint16_t)(id * 0x1Cu);

        if (rd8(m, DS((uint16_t)(p + 0xB77Cu))) != 0u) {
            wr8(m, LIN(list_seg, dst), id);
            dst = (uint16_t)(dst + 1u);
            kept = (uint8_t)(kept + 1u);
        } else {
            ah = (uint8_t)(p >> 8);
        }
        src = (uint16_t)(src + 1u);
    }
    return (uint16_t)(((uint16_t)ah << 8) | kept);
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_28E0_03B6(m, a[0], a[1], a[2]); }
GIN3_REGISTER("28E0_03B6", 3, run)
