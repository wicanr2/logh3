/*
 * 3BB6:07D7  清除模組資料段的 ds:03EC 起 0x800 bytes，再清 ds:0BEC、設旗標 ds:01AB ← 1（near，無參數，無回傳值）
 *
 * 原版：mov di,3ECh；mov cx,800h；mov ax,ds；mov es,ax；cld；xor al,al；rep stosb；
 *   mov ds:0BECh,al；mov byte ptr ds:1ABh,1；retn。
 *   寫入：ds:03EC..0BEB（0x800 bytes）← 0；ds:0BEC ← 0（al，緊接在填充區之後一個 byte，所以連續清 0x801 bytes）；
 *   ds:01AB ← 1。不讀記憶體。ES 被設成 DS（不比對，呼叫者之後不依賴：未知，ES 不在 ret 內）。
 *   DS：呼叫者 3BB6:0703 在 call 之前 `mov ds,seg130`（dosgolem 下 0x4AFE，載入位置相依）；
 *   描述檔用 ds_seg 4AFE，C 版把 ds 當參數。
 *
 * 語意（假說）：3BB6:0703 先呼叫本函式再呼叫 3BB6:0003（重建 8 個聲道記錄）；ds:03EC 起的 0x800 bytes 是
 *   什麼緩衝區：未知。ds:01AB 與 3BB6:028C（寫 0）是同一個旗標。
 * 推論等級：寫入範圍與值已證實（逐指令對過）；用途假說。
 * 輸入域：無輸入（不讀記憶體）。基底狀態的 03EC..0BEC 本來就全是 0、ds:01AB 是 0，單看基底看不出少寫或寫偏，
 *   描述檔的 seeds 先把 03E0 起的範圍填 0xAA、邊界兩側各一個 byte 設成非 0，再個別 poke 邊界位址。
 *   `rep stosb` 的 DF 由前面的 cld 清掉；DI 從 0x3EC 遞增 0x800 次不會超過 0xFFFF，不涉及回繞。
 */
#include "gin3_reg.h"

void f_3BB6_07D7(Mem *m, uint16_t ds)
{
    uint16_t di = 0x03ECu;
    uint16_t cx = 0x0800u;

    while (cx != 0) {
        wr8(m, LIN(ds, di), 0);
        di = (uint16_t)(di + 1u);
        cx = (uint16_t)(cx - 1u);
    }
    wr8(m, LIN(ds, 0x0BECu), 0); /* mov ds:0BECh,al，al ＝ 0 */
    wr8(m, LIN(ds, 0x01ABu), 1);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_3BB6_07D7(m, 0x4AFEu); /* 描述檔 ds_seg：呼叫者 3BB6:0703 載入的 seg130 */
    return 0;
}
GIN3_REGISTER("3BB6_07D7", 0, run)
