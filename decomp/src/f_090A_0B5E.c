/*
 * GIN3PS.EXE 090A:0B5E  全域稅率重算：帝國、同盟、無主三段（無參數、無回傳值）
 *
 * 原版：far 呼叫，無參數，沒有區域變數：
 *   ds:8F1D ← 0；090A:0B93(level ＝ byte ds:8FD4)       ; 帝國
 *   ds:8F1D ← 1；090A:0B93(level ＝ byte ds:8FD5)       ; 同盟
 *   ds:8F1D ← 2；090A:0DC0(level ＝ 2)                  ; 無主
 *   retf
 * level 參數是 `mov al, ds:8FD4h; xor ah, ah; push ax`，所以是 byte 零延伸成 word（不是讀 word ds:8FD4）。
 * 0B93 依 byte ds:8F1D 挑出擁有者 ＝ 該值的星域（每圈重讀，見 f_090A_0B93.c）；0DC0 固定處理擁有者 ＝ 2，
 * 不讀 ds:8F1D。ds:8F1D 結尾留在 2。
 *
 * 回傳：retf 時 AX 是 0DC0 的殘值（描述檔 ret = none）。
 * 呼叫者：0654:0007（每月結算）與 2703:0000（docs/spec/038、docs/internals/27）。
 *
 * 推論等級
 *   已證實（差異測試）：寫 ds:8F1D 的值與順序、level 的來源與寬度、無主段的常數 2。
 *   強推論（docs/spec/038、docs/internals/29）：ds:8FD4／8FD5 是帝國／同盟的難度等級（level），ds:8F1D 是「目前陣營」。
 * 輸入域：ds:8FD4／8FD5 任意 byte；其餘同 0B93／0DC0（星域與提督表內容任意）。
 */
#include "gin3_reg.h"

extern void f_090A_0B93(Mem *m, uint16_t level);
extern void f_090A_0DC0(Mem *m, uint16_t level);

void f_090A_0B5E(Mem *m)
{
    wr8(m, DS(0x8F1D), 0);
    f_090A_0B93(m, rd8(m, DS(0x8FD4)));
    wr8(m, DS(0x8F1D), 1);
    f_090A_0B93(m, rd8(m, DS(0x8FD5)));
    wr8(m, DS(0x8F1D), 2);
    f_090A_0DC0(m, 2);
}

static uint64_t run(Mem *m, const uint16_t *a)
{
    (void)a;
    f_090A_0B5E(m);
    return 0;
}
GIN3_REGISTER("090A_0B5E", 0, run)
