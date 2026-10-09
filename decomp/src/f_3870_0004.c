/*
 * GIN3PS.EXE 3870:0004  把描述槽表 ds:F254 起的 128 bytes 全填 0xFF（8 槽 × 16 bytes 的「全部空閒」）
 * 呼叫：far，無參數。內容：memset(ds:F254, 0xFF, 0x80)，呼叫 3C9C:000A（已 diff-tested，C 版直接呼叫）。
 * 回傳：不使用。原版返回時暫存器是 memset 的出口值：AX ＝ 段（DS）、BX ＝ 位移（0xF254），描述檔以 pack:ax,bx 一併比對。
 * 副作用：ds:F254 到 ds:F2D3 共 128 bytes ← 0xFF。
 * 欄位語意：docs/re/135 第 80 行：ds:F254 + 代號 × 16 是 3870:001A 存放的 16 bytes 描述，槽的 [+7] 為 0xFF 表示空閒
 *   （3870:033E 以此找空槽）。所以本函式是「清空全部槽」。呼叫者 0374:0191（強推論）。
 * 基底狀態（fcmd-base.state）該區已是 0xFF，寫入集合要靠種子把前態 poke 成別的值才看得出來；
 *   鄰近位元組（ds:F253、ds:F2D4）在基底狀態不是 0xFF，用來抓多寫一格。
 * 輸入域：無輸入。推論等級：位移、長度、值已證實（asm 與差異測試）；槽表的用途為強推論。
 */
#include "gin3_reg.h"

extern uint32_t f_3C9C_000A(Mem *m, uint16_t off, uint16_t seg, uint16_t val, uint16_t count);

uint32_t f_3870_0004(Mem *m)
{
    return f_3C9C_000A(m, 0xF254, GIN3_DS, 0x00FF, 0x0080); /* AX<<16 | BX */
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    return f_3870_0004(m);
}
GIN3_REGISTER("3870_0004", 0, run)
