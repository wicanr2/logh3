/*
 * 2376:109C  把 15 個 byte 隔格搬進字形緩衝區，再清四個 word（near，暫存器傳參，DS = DGROUP）
 *
 * 輸入：DS:SI = 來源（15 bytes，SI 是呼叫者 2376:0E0D 用 2376:0EE3 算出的位址）、
 *       DI = 目的的位移（ES = DS，原版開頭 `mov ax,ds; mov es,ax`；呼叫者給 0x82A1 或 0x82A0）。
 * 原版：`cld`，然後 15 次 { movsb（[di] ← [si]，si++、di++）; inc di }，所以目的每次前進 2 bytes、
 *   來源前進 1 byte，每次 movsb 之後才 inc di。逐個 byte 先讀後寫，來源與目的重疊時
 *   後面讀到的是前面剛寫入的值，本版照同一順序逐 byte 讀寫，不用 memcpy。
 *   接著把 ds:829C、ds:829E、ds:82C0、ds:82C2 四個 word 寫成 0（字形緩衝區的第 0、1、18、19 列）。
 * 輸出：記憶體寫入（15 個隔格的 byte 與四個 word）。暫存器：CX = 0，SI 加 15、DI 加 30，AX 是暫用值，
 *   ES 被改成 DS 沒有還原。呼叫者 2376:0E0D 兩次呼叫之後，SI 由下一次 `call 2376:0EE3` 重算，
 *   DI 由 `mov di, 82A0h` 重設或 `pop di` 還原，都不使用這裡留下的值。描述檔 ret = none，只比對記憶體寫入。
 *
 * 語意（強推論）：ds:829C 是 20 列的字形緩衝區，一列兩個 byte（docs/internals/25-text-rendering.md 第 14 行：
 *   16 × 15 的字形讀進 20 列緩衝區，第 0、1、18、19 列留白）。這個函式把 15 列的單 byte 資料
 *   放到緩衝區的第 2 到 16 列的某一個 byte（DI 的奇偶決定放在列的第 0 個還是第 1 個 byte，呼叫者
 *   先用 0x82A1 放一次、再用 0x82A0 放一次，兩次拼成 16 位元寬），並清掉上下各兩列。
 *   0x82A0 = 0x829C + 4，剛好是第 2 列的起點，與此推論一致。字形格式的細節：docs/internals/25 未逐 byte 記載。
 *
 * 輸入域：SI、DI 的全部 16 位元值。位移超過 0xFFFF 時回繞（byte 存取，段內回繞與線性位址連續在 DGROUP
 *   範圍內沒有差別，因為 DGROUP 段尾 0x64B2F 遠小於 1 MB）。來源或目的落在堆疊上的案例，
 *   堆疊不在 DGROUP 的 64 KB 內，不會發生。
 * 推論等級：搬移與清除的位移、次數、順序已證實（反組譯逐指令＋差異測試）；用途為強推論。
 */
#include "gin3_reg.h"

void f_2376_109C(Mem *m, uint16_t si, uint16_t di)
{
    uint16_t cx;

    for (cx = 15u; cx != 0; cx--) {
        wr8(m, DS(di), rd8(m, DS(si))); /* movsb */
        si = (uint16_t)(si + 1u);
        di = (uint16_t)(di + 2u); /* movsb 的 +1 加上 inc di */
    }
    wr16(m, DS(0x829Cu), 0);
    wr16(m, DS(0x829Eu), 0);
    wr16(m, DS(0x82C0u), 0);
    wr16(m, DS(0x82C2u), 0);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    f_2376_109C(m, a[0], a[1]);
    return 0;
}
GIN3_REGISTER("2376_109C", 2, run)
