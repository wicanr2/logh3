/*
 * 2376:051F  由編號算 160 bytes 一筆的表內位移（near，暫存器傳參）
 *
 * 輸入：AL（AH 被 xor ah,ah 丟棄，AX 高位元組不影響結果）。
 * 輸出：SI = 0x79E + 160 * AL（16 位元回繞，AL 最大 255 時不會回繞）。
 *       AX 被當暫存使用，回來時 = 64 * AL（兩個呼叫者 2376:03F6、0486 之後都馬上覆寫 AX，
 *       不依賴它），不納入比對；旗標同樣不納入。
 *
 * 語意（強推論）：79E 是一張以 160 bytes 為一筆的圖形表（在 2376 模組自己的資料段）。
 *   證據是呼叫端用法：03F6 用 rep movsw 搬 0x50 個字（160 bytes）；0486 在呼叫後把 DS 設成模組資料段，
 *   以 lodsw 讀 4 個平面各 20 列（4 * 20 * 2 = 160 bytes）寫入 A000。
 *   AL 是呼叫者從物件 +2 讀出的編號，編號的意義與表的內容：未知。
 * 輸入域：AX 的全部 16 位元值；本函式不讀寫記憶體。
 */
#include "gin3_reg.h"

uint16_t f_2376_051F(uint16_t ax)
{
    uint16_t idx = (uint16_t)(ax & 0xFFu); /* xor ah,ah */

    return (uint16_t)(idx * 160u + 0x79Eu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)m;
    return f_2376_051F(a[0]);
}
GIN3_REGISTER("2376_051F", 1, run)
