/*
 * GIN3PS.EXE 0B20:0006  敵艦隊頁的重建：列出（所在星域 ＝ v 的）敵方現役艦隊，重算它們的索敵等級並把等級非 0 者收進 ds:95B6 起的清單
 *
 * 原版（53 bytes，far 呼叫，一個堆疊字 v；push bp、sub sp,22h，所以 v 在 [bp+28h]）：
 *   區域變數（堆疊）：[bp+0] ＝ n（byte）、[bp+2 .. bp+0x21] ＝ 艦隊槽號清單（32 bytes）。
 *   ds:95B5 ← 0                                       ; byte
 *   mov al,2 ; push ax                                ; 旗標 word：AL ＝ 2（AH 殘值，131C 只取低位元組）
 *   push v ; push ss ; push bp+2
 *   n ← 297B:131C(SS:bp+2, v, 旗標 2)                 ; 旗標 2 ＝ 只留艦隊[+1] != ds:8F1D 的敵方；已 diff-tested，AL ＝ 個數
 *   [bp+0] ← AL                                       ; 堆疊上的區域變數（本函式之後不讀）
 *   push ax（AL ＝ n、AH 殘值）; push ss ; push bp+2
 *   r ← 0B20:003B(SS:bp+2, n)                         ; 已 diff-tested，回 AX ＝ 清單筆數（AH ＝ 0），003B 只取 n 的低位元組
 *   ds:95B4 ← AL
 *   回 AX 是 003B 的 AX（描述檔 ret ＝ ax）。
 * 參數：args[0] ＝ v（131C 只取低位元組：0xFF 不過濾，其餘只留所在星域 ＝ v 的艦隊）。
 * 副作用：ds:95B5 ← 0；ds:95B4 ← 003B 的筆數；003B 重算各敵方艦隊的 +0x70（索敵等級）並寫 ds:95B6 起的清單；
 *   堆疊帶內的清單與 [bp+0]（比對時排除）。
 * 呼叫者（台帳）：04F8:05ED、073F:02EE（hw-dependent）。docs/re/319-fleet-list-tabs.md 第 21、74、162 行
 *   「敵艦隊 `0B20:0006`」、「`0B20:0006(星域)`：`297B:131C(&清單, 星域, 2)`（旗標 2 ＝ 敵軍）→ `0B20:003B` 逐支…」，與本檔一致；
 *   docs/re/325-enemy-tab-rows-and-window.md 第 26 行「`0B20:0006(0)` 重建清單時就從艦隊表與艦艇池讀到新值」。
 *   ds:95B4／95B5 的用途（筆數與捲動起點）為假說。
 * 堆疊：清單緩衝區位置 ＝ 進入點 SP − 0x28（參數 2、far 回位址 4、bp 2、區域變數 0x22 後 bp，再 + 2）；進入點 SP 以下
 *   0x400 bytes 是堆疊帶，比對時排除；本版把清單寫在同一個線性位置，131C 寫、003B 讀回。SS 0x63EB、進入點 SP 0x1F68 由
 *   向量檔 stack_hi ＝ 0x65E18 反推。
 * 輸入域：v 全部 16 位元值；艦隊表、艦艇池取 S3 基底；ds:8F1D 任意。
 * 推論等級：已證實（反組譯逐指令＋差異測試）；欄位語意為假說。
 */
#include "gin3_reg.h"

extern uint8_t f_297B_131C(Mem *m, uint16_t buf_off, uint16_t buf_seg, uint16_t v_word, uint16_t flags_word);
extern uint16_t f_0B20_003B(Mem *m, uint16_t ids_off, uint16_t ids_seg, uint16_t n_word);

#define W0006_SS 0x63EBu
#define W0006_SP 0x1F68u

uint16_t f_0B20_0006(Mem *m, uint16_t v)
{
    uint16_t bp = (uint16_t)(W0006_SP - 0x2Au);
    uint16_t buf = (uint16_t)(bp + 2u);
    uint8_t n;
    uint16_t r;

    wr8(m, DS(0x95B5u), 0);
    n = f_297B_131C(m, buf, W0006_SS, v, 2u);
    wr8(m, LIN(W0006_SS, bp), n);
    r = f_0B20_003B(m, buf, W0006_SS, n);
    wr8(m, DS(0x95B4u), (uint8_t)r);
    return r;
}

static uint64_t run(Mem *m, const uint16_t *a) { return f_0B20_0006(m, a[0]); }
GIN3_REGISTER("0B20_0006", 1, run)
