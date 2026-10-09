/*
 * GIN3PS.EXE 28E0:000C  六位專任參謀就任：對六個參謀編號各呼叫一次 28E0:05CC（提督[+0x09] ← 0x7F）
 *
 * 原版：far 呼叫，無參數；沒有序言（不 push bp）。依序 `push word ptr ds:[位址]`、call 28E0:05CC、`inc sp` 兩次（清掉一個字的參數）：
 *   順序是 ds:8F4B、ds:8F4C、ds:8F4F、ds:8F50、ds:8F4D、ds:8F4E（注意 4F、50 排在 4D、4E 之前）。
 *   `push word ptr` 推入的是 word（位址的 byte 與下一個 byte），28E0:05CC 只取低位元組當提督編號，所以高位元組（下一個參謀欄位）無關。
 *   回傳：AX 是最後一次 05CC 的回傳（提督編號 × 28，`mul` 之後沒有再動）；呼叫者 042F:0006 之後的使用方式不在本檔查證範圍，
 *   差異測試比較 AX。
 *
 * 語意（已證實：與反組譯逐條對過）：
 *   對 ds:8F4B、8F4C、8F4F、8F50、8F4D、8F4E 六個 byte（參謀編號；ds:8F4B..8F50 的基底值 06 4F 0D 3A 10 3C 是開局六位，見
 *   docs/internals/67 第 103 行）各做一次 提督[編號][+0x09] ← 0x7F（28E0:05CC）。六次寫入互相獨立（提督記錄 ds:B77C 起，
 *   +9 的位址範圍 0xB785 到 0xD369，不會蓋到 ds:8F4B 起的參謀欄位），順序不影響結果。
 *
 * 欄位語意（強推論）：ds:8F4B 到 ds:8F50 是六位專任參謀（軍事、情報等）的提督編號（docs/internals/67 的初始化列；docs/re/295
 *   第 55 到 57 行）；2703:1789 讀 ds:8F4D（情報專任參謀）。「就任」的語意來自 docs/internals/08 第 277 行對 28E0:05CC 的描述。
 *
 * 輸入域：六個參謀編號位置各自的 16 位元值（高位元組無關）；提督記錄與被寫位置由突變與種子涵蓋。
 *   原值與 0x7F 相同時寫入集合看不出來，所以種子先把被寫位置改成別的值。
 * 推論等級：已證實（反組譯逐指令＋差異測試）。
 */
#include "gin3_reg.h"

extern uint16_t f_28E0_05CC(Mem *m, uint16_t id_w);

uint16_t f_28E0_000C(Mem *m)
{
    uint16_t ax;

    f_28E0_05CC(m, rd16(m, DS(0x8F4Bu)));
    f_28E0_05CC(m, rd16(m, DS(0x8F4Cu)));
    f_28E0_05CC(m, rd16(m, DS(0x8F4Fu)));
    f_28E0_05CC(m, rd16(m, DS(0x8F50u)));
    f_28E0_05CC(m, rd16(m, DS(0x8F4Du)));
    ax = f_28E0_05CC(m, rd16(m, DS(0x8F4Eu)));
    return ax;
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_28E0_000C(m);
}
GIN3_REGISTER("28E0_000C", 0, run)
