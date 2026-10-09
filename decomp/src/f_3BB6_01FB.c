/*
 * GIN3PS.EXE 3BB6:01FB  音樂驅動模組（seg130）的重設：清旗標 ds:01AB、重建 8 個聲道記錄、再寫 ds:0BEF ← 0xFF
 *
 * 原版：far 呼叫，無參數、無回傳值（23 bytes）：
 *   push si；push di；push ds；mov ax,seg130；mov ds,ax    ; DS ← seg130（dosgolem 下 0x4AFE，載入位置相依）
 *   call 3BB6:028C                                         ; ds:01AB ← 0
 *   call 3BB6:0003                                         ; 重建 8 個聲道記錄與 ds:01A7..01A9、ds:0BEE
 *   mov byte ds:0BEF,0FFh
 *   pop ds；pop di；pop si；retf
 *   SI／DI／DS 由 push／pop 保存；AX 是 0003 結束時的殘值，描述檔 ret = none，不比對。
 * 兩個被呼叫者是 near 呼叫（同段），各自已 diff-tested，C 版把 DS 當參數（0x4AFE）傳給它們。
 * 呼叫順序：028C 寫 ds:01AB，0003 不讀也不寫 01AB（掃描讀的位移 ≡ 2 mod 4，01AB ≡ 3）；末尾的 0BEF 寫入 0003 也不讀寫（它寫 0BEE、讀 0BF1），
 *   所以三者互相獨立，彼此的順序不影響結果（差異測試無法也不需要分辨）。呼叫次數與所用的段則可以分辨。
 * 呼叫者（台帳）：3B74:0137（hw-dependent，音樂驅動的啟動／切曲流程；用途為假說）。
 *
 * 語意（假說）：seg130 是音樂驅動的資料段，ds:0BEF 的 0xFF 與相鄰的 ds:0BEE（遮罩）、ds:0BF1（聲道要求）同一個區塊；
 *   名稱與用途：未知。寫入的位址與值已證實。
 * 輸入域：同 3BB6:0003（掃描依賴 0xFE 結尾終止，整圈都沒有則原版不終止，被探針以步數上限丟棄）。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；用途為假說。
 */
#include "gin3_reg.h"

extern void f_3BB6_028C(Mem *m, uint16_t ds);
extern void f_3BB6_0003(Mem *m, uint16_t ds);

#define F3BB6_01FB_SEG130 0x4AFEu /* seg130，dosgolem 載入位置下的段值（呼叫者 3BB6:01FB `mov ax,seg seg130`） */

void f_3BB6_01FB(Mem *m)
{
    const uint16_t ds = F3BB6_01FB_SEG130;

    f_3BB6_028C(m, ds);
    f_3BB6_0003(m, ds);
    wr8(m, LIN(ds, 0x0BEFu), 0xFFu);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_3BB6_01FB(m);
    return 0;
}
GIN3_REGISTER("3BB6_01FB", 0, run)
