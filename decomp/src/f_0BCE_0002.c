/*
 * GIN3PS.EXE 0BCE:0002  提督清單頁的重建：列出「本方、[+0]==0xFF、[+9] 的 bit7 已設」的提督編號到 ds:95DA 起的緩衝區，筆數存 ds:95D8
 *
 * 原版（32 bytes，far 呼叫，一個堆疊字 a）：
 *   push bp ; mov bp,sp
 *   xor ax,ax ; ds:95D9 ← AL（0）
 *   push ax                         ; limit word ＝ 0（AX 剛清成 0）
 *   mov al,25h ; push ax            ; flags word ＝ 0x0025（AH ＝ 0）
 *   push a                          ; a word（[bp+6]）
 *   push ds ; push 95DAh            ; 輸出緩衝區 far 指標（段 ＝ DGROUP、位移 0x95DA）
 *   call 28E0:0103                  ; 已 diff-tested；回 AL ＝ 留下的個數
 *   ds:95D8 ← AL ; pop bp ; retf
 *   回傳 AX 是 28E0:0103 的 AX（AL ＝ 個數，AH 是殘值）；描述檔 ret ＝ pack:al。
 * 旗標 0x25 ＝ bit0 | bit2 | bit5，對應 28E0:0103 的篩選器：bit0（028E）留 提督[+1] == ds:8F1D；bit2（035C）留 提督[+0] == 0xFF；
 *   bit5（047E）留 提督[+9] 的 bit7 已設。limit ＝ 0 在這個旗標下不使用（bit6 未設）。a（只取低位元組）!= 0xFF 時先過 0227（提督[+0x0A] == a）。
 * 副作用：ds:95D9 ← 0；ds:95DA 起 n 個 byte（清單，n ≤ 82，被篩選器就地壓縮）；ds:95D8 ← n。
 * 呼叫者（台帳）：04F8:05ED、073F:02EE、073F:0522（皆 hw-dependent）。docs/re/319-fleet-list-tabs.md 第 21、58、75、158 行：
 *   「提督 `0BCE:0002`，帶 `ds:9588` ＝ 清單看的」，即提督頁重建；a 的語意（星域）依 28E0:0103 的檔頭與 docs/re/208。
 *   提督記錄欄位的名稱不是證據，位移與比較才是；ds:95D8／95D9／95DA 的用途為假說（強推論：清單頁的筆數與捲動起點）。
 * 輸入域：a 全部 16 位元值；提督表 ds:B77C 起 82 筆 × 28 bytes 取自記憶體；ds:8F1D 任意。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為假說。
 */
#include "gin3_reg.h"

extern uint8_t f_28E0_0103(Mem *m, uint16_t list_off, uint16_t list_seg, uint16_t a_w, uint16_t flags_w, uint16_t limit_w);

uint8_t f_0BCE_0002(Mem *m, uint16_t a)
{
    uint8_t n;

    wr8(m, DS(0x95D9u), 0);
    n = f_28E0_0103(m, 0x95DAu, GIN3_DS, a, 0x25u, 0u);
    wr8(m, DS(0x95D8u), n);
    return n;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0BCE_0002(m, a[0]); }
GIN3_REGISTER("0BCE_0002", 1, run)
