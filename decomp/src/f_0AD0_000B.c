/*
 * GIN3PS.EXE 0AD0:000B  艦隊清單頁的重建：列出「本方」現役艦隊的槽號到 ds:9594 起的緩衝區，筆數存 ds:9592
 *
 * 原版（31 bytes，far 呼叫，一個堆疊字 v）：
 *   push bp ; mov bp,sp
 *   ds:9593 ← 0                         ; byte
 *   mov al,1 ; push ax                  ; flags word：AL ＝ 1，AH 是進入值的殘值（297B:131C 只取低位元組）
 *   push v                              ; v word（[bp+6]）
 *   push ds ; push 9594h                ; 輸出緩衝區 far 指標（段 ＝ DGROUP、位移 0x9594）
 *   call 297B:131C                      ; 已 diff-tested；回 AL ＝ 個數
 *   ds:9592 ← AL ; pop bp ; retf
 *   回傳 AX 是 131C 的 AX（AL ＝ 個數，AH 是殘值）；描述檔 ret ＝ pack:al，只比 AL。
 * 參數：args[0] ＝ v（word；131C 只取低位元組：0xFF 不過濾，其餘只留所在星域 ＝ v 的艦隊）。
 *   flags ＝ 1：131C 只留艦隊[+1]（陣營）＝ ds:8F1D 的「本方」艦隊。
 * 副作用：ds:9593 ← 0；ds:9594 起 n 個 byte（131C 寫的清單，n ≤ 32）；ds:9592 ← n。
 * 呼叫者（台帳）：04F8:05ED、073F:02EE、073F:0522（皆 hw-dependent）。docs/re/319-fleet-list-tabs.md 第 72、158 行：
 *   `0AD0:000B(ds:9588)` 是艦隊頁重建，ds:9588 ＝ 清單看的星域（參數 v 的來源）。ds:9593 的清零與 ds:9594 起的清單、
 *   ds:9592 的個數的用途為假說（強推論：清單頁的筆數與捲動起點）。
 * 輸入域：v 全部 16 位元值；艦隊表取自記憶體；ds:8F1D 任意。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為假說。
 */
#include "gin3_reg.h"

extern uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word);

uint8_t f_0AD0_000B(Mem *m, uint16_t v)
{
    uint8_t n;

    wr8(m, DS(0x9593u), 0);
    n = f_297B_131C(m, 0x9594u, GIN3_DS, v, 1u);
    wr8(m, DS(0x9592u), n);
    return n;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0AD0_000B(m, a[0]); }
GIN3_REGISTER("0AD0_000B", 1, run)
