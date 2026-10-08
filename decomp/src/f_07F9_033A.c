/*
 * GIN3PS.EXE 07F9:033A  提督的副官成不成立（far，一個 far 指標參數，回 AX = 0 或 1，不寫記憶體）
 *
 * 參數：admiral = 提督記錄的 far 指標（位移、段）。提督表在 ds:B77C，一筆 28 bytes、82 筆
 *   （docs/internals/08、30）。副官記錄在 DS 內（ds:B77C + 副官編號 × 28），不隨 admiral 的段走。
 * 欄位（相對提督記錄）：
 *   adm[+0x07] = 副官編號（0xFF = 沒有）、adm[+0x06]、adm[+0x0A]
 *   adj = ds:B77C + adm[+0x07] × 28：adj[+0x00]、adj[+0x06]、adj[+0x09]、adj[+0x0A]
 * 回 1 的條件（五個全成立，其餘回 0）：
 *   adm[+0x07] != 0xFF
 *   adj[+0x09] 的 bit7 有設
 *   adj[+0x00] == 0xFF
 *   adj[+0x0A] == adm[+0x0A]
 *   adj[+0x06] <  adm[+0x06]（無號 byte 比較，jb）
 * 位移相加依 16 位元回繞（LIN 把位移遮成 16 位元）。
 *
 * 推論等級
 *   已證實（差異測試）：上列欄位位移、常數與比較方向。
 *   強推論（docs/re/182）：adj[+0x09] bit7 = 沒有帶艦隊、+0x0A = 所在星域、+0x00 = 職務（0xFF 為無）；
 *     adj[+0x06] 的語意（階級／地位）docs/re/182 標假設，這裡不採用，只當 byte 比較。
 * 輸入域：admiral 取提督表 82 筆（段為 DS，另有一個偏移 DS+0x10 的段驗證段參數確實被使用）以及
 *   位移接近 0xFFFF 的邊緣值（驗證欄位位移的 16 位元回繞）；副官編號 0..255 全域（記憶體內容由
 *   突變決定，編號 >= 82 時讀到提督表之後的記憶體，原版同樣如此）。
 */
#include "gin3_reg.h"

uint16_t f_07F9_033A(Mem *m, uint16_t adm_off, uint16_t adm_seg)
{
    uint8_t adj_id = rd8(m, LIN(adm_seg, adm_off + 7u));
    uint16_t adj = (uint16_t)(0xB77Cu + adj_id * 28u);

    if (adj_id == 0xFF) return 0;
    if ((rd8(m, DS(adj + 9u)) & 0x80u) == 0) return 0;
    if (rd8(m, DS(adj + 0u)) != 0xFF) return 0;
    if (rd8(m, DS(adj + 0x0Au)) != rd8(m, LIN(adm_seg, adm_off + 0x0Au))) return 0;
    return rd8(m, DS(adj + 6u)) < rd8(m, LIN(adm_seg, adm_off + 6u)) ? 1 : 0;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_07F9_033A(m, a[0], a[1]); }
GIN3_REGISTER("07F9_033A", 2, run)
