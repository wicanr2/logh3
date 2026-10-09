/*
 * GIN3PS.EXE 1930:000B  在場分隊清單裡艦種 5 的艦艇，各呼叫一次 1930:0062（補飛彈與艦載機存量）
 *
 * 原版：far 呼叫，無參數，回傳 AX 是殘值（見下）。序言 push bp／si／di、sub sp,8，區域變數：
 *   di ← 0（清單序號，word）、[bp+0] ← 0x98BA（清單位移，word）、[bp+2] ← DS（清單段）
 *   迴圈：
 *     di >= word ds:9896（無號，jnb）→ 結束，AX ＝ di（`mov ax,di` 之後 cmp，所以離開時 AX ＝ di）
 *     idx  ← word ds:[清單位移]                           ; 在場分隊清單，一項 2 bytes（艦艇池索引）
 *     艦艇 ← (word ds:CFB4 ＋ low16(idx × 0x4E)) : word ds:CFB6    ; 每圈重讀池基底
 *     byte 艦艇[+4]（艦種）== 5 → 1930:0062(艦艇 far 指標)
 *     di++；清單位移 += 2
 * 寫入只發生在 0062；本函式自己不寫記憶體。
 *
 * 回傳：離開迴圈時 AX ＝ di（比較用的 `mov ax,di`），描述檔 ret = ax。呼叫者 12AA:0245 不用回傳值，
 *   AX ＝ di 是 codegen 的殘值，不屬於語意，但差異測試照比。
 * 呼叫者：12AA:0245（回合開頭依序 2526:000E、1930:000B、1AAB:000E、2264:000A，docs/re/339 §1.4、docs/re/145 第 225 行）。
 *
 * 推論等級
 *   已證實（差異測試）：清單基底 0x98BA 與筆數 ds:9896（word 無號比較）、池基底的來源與每圈重讀、艦種欄位位移 +4 與常數 5、
 *     呼叫 0062 的參數（艦艇 far 指標）、`AX ＝ di`。
 *   強推論（docs/internals/38-battle-units.md、docs/internals/67 的 1930 列）：ds:98BA／9896 是在場分隊清單與筆數；
 *     艦種 5 ＝ 補給艦（0062 的語意是假說，見 f_1930_0062.c）。
 * 輸入域：清單項目、池基底與筆數任意；基底狀態的筆數是 0，清單由種子與突變構造。筆數大到超過探針步數上限
 *   （200,000）的執行由探針丟棄。
 */
#include "gin3_reg.h"

extern void f_1930_0062(Mem *m, uint16_t ship_off, uint16_t ship_seg);

/* DS 段內的 word 讀取：位移 16 位元回繞，seg:FFFF 的高位元組在位移 0（與 dosgolem 一致） */
static uint16_t f1930_000B_ds16(const Mem *m, uint16_t off)
{
    return (uint16_t)(rd8(m, DS(off)) | (rd8(m, DS((uint16_t)(off + 1u))) << 8));
}

uint16_t f_1930_000B(Mem *m)
{
    uint16_t di = 0;
    uint16_t list = 0x98BAu;

    while (di < f1930_000B_ds16(m, 0x9896u)) {
        uint16_t idx = f1930_000B_ds16(m, list);
        uint16_t ship_off = (uint16_t)(f1930_000B_ds16(m, 0xCFB4u) + (uint16_t)(idx * 0x4Eu));
        uint16_t ship_seg = f1930_000B_ds16(m, 0xCFB6u);

        if (rd8(m, LIN(ship_seg, (uint16_t)(ship_off + 4u))) == 5)
            f_1930_0062(m, ship_off, ship_seg);
        di = (uint16_t)(di + 1u);
        list = (uint16_t)(list + 2u);
    }
    return di;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_1930_000B(m);
}
GIN3_REGISTER("1930_000B", 0, run)
